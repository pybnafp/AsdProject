package controllers

import (
	"asd/app/dto"
	"asd/app/services"
	"asd/utils/common"
	"asd/utils/gfile"
	"os/exec"

	"github.com/beego/beego/v2/core/logs"
)

var FileApi = new(FileApiController)

type FileApiController struct {
	BaseController
}

// 添加文件
func (c *FileApiController) Upload() {
	userId := c.GetUserId()
	fileID := common.GetUUID()
	info, err := gfile.FileUpload(&c.Controller, "file", fileID, userId)

	if err != nil {
		c.ErrorJson(500, err.Error())
		return
	}

	req := dto.FileAddReq{
		FileID:      fileID,
		FileName:    info.FileName,
		FilePath:    info.FilePath,
		FileSize:    info.FileSize,
		ContentType: info.ContentType,
		FileType:    info.FileType,
		Visibility:  "private",
		Metadata:    "",
		Description: "",
	}

	cmd := exec.Command("markitdown", info.TmpFilePath)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		logs.Error("用markitdown获取上传文件内容失败", cmd.String(), err)
	} else {
		req.Content = string(bytes)
	}

	fileId, err := services.FileService.Add(req, userId)
	if err != nil {
		logs.Error(err)
		c.ErrorJson(500, err.Error())
		return
	}

	// 获取文件详情
	file, err := services.FileService.GetDetail(fileId, userId)
	if err != nil {
		logs.Error(err)
		c.ErrorJson(500, err.Error())
		return
	}

	c.JSON(common.JsonResult{
		Code: 0,
		Msg:  "文件上传成功",
		Data: file,
	})

}

// 查询文件
func (c *FileApiController) Get() {
	var req dto.FileQueryReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, err.Error())
		return
	}

	userId := c.GetUserId()
	file, err := services.FileService.GetDetail(req.FileID, userId)
	if err != nil {
		c.ErrorJson(404, err.Error())
		return
	}

	c.JSON(common.JsonResult{
		Code: 0,
		Data: file,
	})
}

// 删除文件
func (c *FileApiController) Delete() {
	var req dto.FileQueryReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, err.Error())
		return
	}

	userId := c.GetUserId()
	err := services.FileService.Delete(req.FileID, userId)
	if err != nil {
		c.ErrorJson(500, err.Error())
		return
	}

	c.JSON(common.JsonResult{
		Code: 0,
		Msg:  "文件删除成功",
	})
}
