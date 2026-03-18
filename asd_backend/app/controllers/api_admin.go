package controllers

import (
	"asd/app/dto"
	"asd/app/models"
	"asd/app/services"
	"asd/utils/common"
	"asd/utils/gfile"
	"asd/utils/gstr"
	"fmt"
	"os/exec"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

var AdminApi = new(AdminApiController)

type AdminApiController struct {
	BaseController
}

func (c *AdminApiController) AddReport() {
	var req dto.ReportAddReq
	if err := c.ParseForm(&req); err != nil {
		fmt.Printf("%+v\n", err)
		c.ErrorJson(400, "请求参数错误"+err.Error())
		return
	}
	reportDate, err := time.Parse("2006-01-02", req.ReportDate)
	if err != nil {
		c.ErrorJson(400, "report_date格式错误"+err.Error())
		return
	}

	input := services.AddReportInput{
		UserID:     req.UserID,
		Name:       req.Name,
		ReportId:   req.ReportID,
		Type:       req.Type,
		ReportDate: reportDate,
		Status:     models.ReportStatusReviewed,
	}

	if info, err := gfile.FileUpload(&c.Controller, "original_file", common.GetUUID(), req.UserID); err != nil {
		c.ErrorJson(500, err.Error())
		return
	} else {
		input.OriginalFileName = info.FileName
		input.OriginalFilePath = info.FilePath
		input.OriginalFileSize = info.FileSize
	}

	if info, err := gfile.FileUpload(&c.Controller, "report_file", common.GetUUID(), req.UserID); err != nil {
		c.ErrorJson(500, err.Error())
		return
	} else {
		input.ReportFileName = info.FileName
		input.ReportFilePath = info.FilePath
		input.ReportFileSize = info.FileSize

		cmd := exec.Command("markitdown", info.TmpFilePath)
		bytes, err := cmd.CombinedOutput()
		if err != nil {
			logs.Error("用markitdown获取报告内容失败", cmd.String(), err)
		} else {
			input.Content = string(bytes)
		}
	}

	reportVo, err := services.Report.AddReport(input)
	if err != nil {
		c.ErrorJson(500, err.Error())
		return
	}

	c.JSON(common.JsonResult{
		Code:  0,
		Data:  reportVo,
		Count: 1,
	})
}

func (c *AdminApiController) FixPubMedLinks() {
	var messages []models.ChatMessage
	_, err := orm.NewOrm().QueryTable("chat_messages").All(&messages)
	if err != nil {
		c.ErrorJson(500, err.Error())
		return
	}

	for _, msg := range messages {
		if msg.Completion == "" {
			continue
		}

		// Fix completion linkes
		msg.Completion = gstr.FixPubMedLinks(msg.Completion)

		if _, err := orm.NewOrm().Update(&msg, "Completion"); err != nil {
			logs.Error("Failed to update chat message", err)
			continue
		}
	}

	c.JSON(common.JsonResult{
		Code: 0,
		Msg:  "Success",
	})
}

func (c *AdminApiController) GetRagWorkersStatus() {
	statuses := services.RagService.GetRagWorkersStatus()
	c.JSON(common.JsonResult{
		Code: 0,
		Data: statuses,
	})
}
