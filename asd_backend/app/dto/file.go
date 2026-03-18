package dto

import "github.com/gookit/validate"

// FilePageReq 文件列表请求参数
type FilePageReq struct {
	UserID     int    `json:"user_id"`
	Page       int    `json:"page" form:"page"`             // 页码
	Limit      int    `json:"limit" form:"limit"`           // 每页数量
	FileType   string `json:"file_type" form:"file_type"`   // 文件类型
	Visibility string `json:"visibility" form:"visibility"` // 可见性
	Status     string `json:"status" form:"status"`         // 状态
}

// 添加文件请求
type FileAddReq struct {
	FileID      string `json:"file_id" validate:"required"`
	UserID      int    `json:"user_id" validate:"required"`
	FileName    string `json:"file_name" validate:"required"`
	FilePath    string `json:"file_path" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required"`
	FileType    string `json:"file_type" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
	Content     string `json:"content"`
	Description string `json:"description"`
	Metadata    string `json:"metadata"`
	Visibility  string `json:"visibility"`
	Status      string `json:"status"`
}

func (v FileAddReq) Messages() map[string]string {
	return validate.MS{
		"UserID.required": "用户ID不能为空",
	}
}

// 查询文件请求
type FileQueryReq struct {
	FileID string `json:"file_id"`
}
