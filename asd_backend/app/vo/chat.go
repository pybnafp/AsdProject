package vo

type ChatVo struct {
	ChatID    string `json:"chat_id" description:"主键ID"`
	Title     string `json:"title" description:"标题"`
	UserId    int    `json:"user_id" description:"用户ID"`
	CreatedAt string `json:"created_at" description:"创建时间"`
}
