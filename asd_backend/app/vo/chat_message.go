package vo

type ChatMessageVo struct {
	MessageID  string     `json:"message_id" description:"主键ID"`
	ChatID     string     `json:"chat_id" description:"对话ID"`
	UserId     int        `json:"user_id" description:"用户ID"`
	Prompt     string     `json:"prompt" description:"提交内容"`
	Completion string     `json:"completion" description:"回复内容"`
	Files      []FileVo   `json:"files" description:"文件列表"`
	Reports    []ReportVo `json:"reports" description:"报告列表"`
	CreatedAt  string     `json:"created_at" description:"创建时间"`
	RawPrompt  string     `json:"-" description:"包含文件、报告的提交内容"`
	Reasoning  string     `json:"reasoning" description:"深度思考"`
}
