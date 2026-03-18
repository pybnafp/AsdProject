package dto

import "github.com/gookit/validate"

// 对话分页请求
type ChatPageReq struct {
	Title  string `json:"title"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// 对话分页请求
type ChatDetailReq struct {
	ChatID string `json:"chat_id" validate:"required"`
}

func (v ChatDetailReq) Messages() map[string]string {
	return validate.MS{
		"ChatID.required": "对话ID不能为空",
	}
}

// 添加对话请求
type ChatAddReq struct {
	Title string `json:"title" validate:"required"`
}

func (v ChatAddReq) Messages() map[string]string {
	return validate.MS{
		"Title.required": "对话标题不能为空",
	}
}

// 更新对话请求
type ChatUpdateReq struct {
	ChatID string `json:"chat_id" validate:"required"`
	Title  string `json:"title" validate:"required"`
}

func (v ChatUpdateReq) Messages() map[string]string {
	return validate.MS{
		"ChatID.required": "对话ID不能为空",
		"Title.required":  "对话标题不能为空",
	}
}

// 更新对话请求
type ChatDeleteReq struct {
	ChatID string `json:"chat_id" validate:"required"`
}

func (v ChatDeleteReq) Messages() map[string]string {
	return validate.MS{
		"ChatID.required": "对话ID不能为空",
	}
}

type StreamChatReq struct {
	Prompt           string   `json:"prompt" validate:"required"`
	ChatID           string   `json:"chat_id"`
	FileIDs          []string `json:"file_ids"`
	ReportIDs        []string `json:"report_ids"`
	EnableGuidelines bool     `json:"enable_guidelines"`
	EnableResearches bool     `json:"enable_researches"`
}

func (v StreamChatReq) Messages() map[string]string {
	return validate.MS{
		"Message.required": "消息不能为空",
	}
}

// StopStreamReq 停止流式对话请求
type StopStreamReq struct {
	ChatID    string `json:"chat_id" validate:"required"`
	MessageID string `json:"message_id" validate:"required"`
}

// StreamReadReq 流式对话读取请求
type StreamReadReq struct {
	ChatID    string `json:"chat_id" validate:"required"`    // 对话ID
	MessageID string `json:"message_id" validate:"required"` // 消息ID
}
