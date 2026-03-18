package dto

// UsageStats 表示API调用的使用统计
type UsageStats struct {
	Model            string  `json:"model"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	ReasoningTokens  int     `json:"reasoning_tokens"`
	Cost             float64 `json:"cost"`
	Currency         string  `json:"currency"`
}
