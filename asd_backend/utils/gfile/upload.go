package gfile

import (
	"asd/conf"
	"asd/utils"
	"asd/utils/common"
	"asd/utils/tencent"
	"fmt"
	"os"
	"path/filepath"
	"time"

	beego "github.com/beego/beego/v2/adapter"
)

// FileInfo 文件信息结构体
type FileInfo struct {
	FilePath    string `json:"file_path"`     // COS 文件路径
	FileName    string `json:"file_name"`     // 原始文件名
	FileType    string `json:"file_type"`     // 文件类型
	FileSize    int64  `json:"file_size"`     // 文件大小(字节)
	ContentType string `json:"content_type"`  // MIME 类型
	Url         string `json:"url"`           // MIME 类型
	TmpFilePath string `json:"tmp_file_path"` // 临时文件的路径
}

func FileUpload(c *beego.Controller, uploadfile string, fileID string, userId int) (*FileInfo, error) {
	file, header, err := c.GetFile(uploadfile)
	if err != nil {
		return nil, fmt.Errorf("文件获取失败: %w", err)
	}
	defer file.Close()

	// 获取文件名和扩展名
	fileName := header.Filename
	ext := GetFullExt(fileName)

	// 创建临时文件保存上传内容
	filename := fmt.Sprintf("%s%s", fileID, ext)
	tempFilePath := conf.CONFIG.ApiConfig.FileCachePath + filename

	// 确保临时目录存在
	if err := os.MkdirAll(filepath.Dir(tempFilePath), 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	err = c.SaveToFile(uploadfile, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("文件保存失败: %w", err)
	}

	// 上传文件到 COS
	cosFileName := utils.GetUserIDKey(userId) + "/" + filename
	filePath, err := tencent.UploadFileToCOS(cosFileName, tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("文件上传到 COS 失败: %w", err)
	}

	url, err := tencent.GetDownloadURL(filePath)
	if err != nil {
		return nil, err
	}

	// 返回文件信息
	fileInfo := &FileInfo{
		FilePath:    filePath,
		FileType:    ext,
		FileName:    fileName,
		FileSize:    header.Size,
		ContentType: header.Header.Get("Content-Type"),
		Url:         url,
		TmpFilePath: tempFilePath,
	}

	return fileInfo, nil
}

func Upload(c *beego.Controller, uploadfile string, path string, fileType string) common.JsonResult {
	file, header, err := c.GetFile(uploadfile) // 获取上传的文件
	if err != nil {
		// 返回结果
		return common.JsonResult{
			Code: -1,
			Msg:  "文件获取失败",
		}
	}
	defer file.Close()

	// 获取文件名和扩展名
	fileName := header.Filename
	ext := GetFullExt(fileName) // 使用自定义函数获取复合扩展名

	if ext != fileType {
		// 返回结果
		return common.JsonResult{
			Code: -1,
			Msg:  "文件格式错误: 请上传 " + fileType + " 文件",
		}
	}

	// 创建临时文件保存上传内容
	tempFilePath := "/tmp/" + fmt.Sprintf("%d%s", time.Now().Unix(), ext)
	err = c.SaveToFile(uploadfile, tempFilePath)
	if err != nil {
		// 返回结果
		return common.JsonResult{
			Code: -1,
			Msg:  "文件保存失败",
		}
	}

	// 上传文件到 COS
	cosFileName := fmt.Sprintf(path+"/%d%s", time.Now().Unix(), ext) // 生成 COS 中的文件名
	filePath, err := tencent.UploadFileToCOS(cosFileName, tempFilePath)
	if err != nil {
		// 返回结果
		return common.JsonResult{
			Code: -1,
			Msg:  "文件上传到 COS 失败",
		}
	}

	// 删除临时文件
	os.Remove(tempFilePath)

	url, err := tencent.GetDownloadURL(filePath)
	if err != nil {
		// 返回结果
		return common.JsonResult{
			Code: -1,
			Msg:  "生成下载地址失败",
		}
	}

	// 返回文件的存储路径	// 返回结果
	return common.JsonResult{
		Code: 0,
		Data: map[string]string{"filepath": filePath, "url": url},
		Msg:  "文件上传成功",
	}
}
