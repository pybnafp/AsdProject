package dto

import "github.com/gookit/validate"

// 添加对话记录请求
type ChatMessageAddReq struct {
	ChatID string `json:"chat_id" validate:"required"`
	Params string `json:"params" validate:"required"`
}

func (v ChatMessageAddReq) Messages() map[string]string {
	return validate.MS{
		"ChatID.required": "对话ID不能为空",
		"Params.required": "请求参数不能为空",
	}
}

// 开始对话请求
type ChatMessageStartReq struct {
	InitContent string `json:"init_content" validate:"required"`
	Title       string `json:"title" validate:"required"`
}

func (v ChatMessageStartReq) Messages() map[string]string {
	return validate.MS{
		"Context.required":     "对话上下文不能为空",
		"InitContent.required": "初始内容不能为空",
		"Title.required":       "标题不能为空",
	}
}

// 对话记录分页请求
type ChatMessagePageReq struct {
	ChatID string `json:"chat_id" validate:"required"`
	Page   int    `json:"page" `
	Limit  int    `json:"limit" `
}

// 删除记录请求
type ChatMessageDeleteReq struct {
	RecordId string `json:"record_id" validate:"required"`
}
