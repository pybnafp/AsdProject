package dto

// 智能体对话请求
type AgentChatReq struct {
	ChatID           string              `json:"chat_id"`                    // 对话ID
	Prompt           string              `json:"prompt" validate:"required"` // 用户输入
	Messages         []map[string]string `json:"messages"`                   // 添加历史消息字段
	EnableGuideLines bool                `json:"enable_guidelines"`
	EnableResearches bool                `json:"enable_researches"`
}
