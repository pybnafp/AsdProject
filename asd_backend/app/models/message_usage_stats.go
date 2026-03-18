package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// MessageUsageStats 消息消耗统计表
type MessageUsageStats struct {
	Id               int64     `orm:"column(id);auto" description:"ID"`
	MessageID        string    `orm:"column(message_id);size(36);unique" description:"消息ID"`
	ChatID           string    `orm:"column(chat_id);size(36)" description:"聊天ID"`
	UserID           string    `orm:"column(user_id);size(36)" description:"用户ID"`
	Model            string    `orm:"column(model);size(100)" description:"使用的模型"`
	PromptTokens     int       `orm:"column(prompt_tokens);default(0)" description:"输入消息tokens数"`
	CompletionTokens int       `orm:"column(completion_tokens);default(0)" description:"输出消息tokens数"`
	TotalTokens      int       `orm:"column(total_tokens);default(0)" description:"总tokens数"`
	Cost             float64   `orm:"column(cost);digits(10);decimals(6);default(0.000000)" description:"费用"`
	Currency         string    `orm:"column(currency);size(10);default(USD)" description:"货币单位"`
	CreatedAt        time.Time `orm:"column(created_at);type(timestamp);auto_now_add" description:"创建时间"`
}

func (t *MessageUsageStats) TableName() string {
	return "message_usage_stats"
}

func init() {
	orm.RegisterModel(new(MessageUsageStats))
}

// 插入消息使用统计数据
func (t *MessageUsageStats) Insert() (int64, error) {
	id, err := orm.NewOrm().Insert(t)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// 根据消息ID查询统计数据
func (t *MessageUsageStats) GetByMessageID() error {
	o := orm.NewOrm()
	return o.QueryTable(new(MessageUsageStats)).Filter("message_id", t.MessageID).One(t)
}

// 更新统计数据
func (t *MessageUsageStats) Update() (int64, error) {
	o := orm.NewOrm()
	rows, err := o.Update(t)
	if err != nil {
		return 0, err
	}
	return rows, nil
}
