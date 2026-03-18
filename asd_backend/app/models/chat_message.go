package models

import (
	"asd/utils"
	"errors"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type ChatMessage struct {
	Id         int64     `orm:"column(id);auto" description:"主键ID"`
	MessageID  string    `orm:"column(message_id);size(128)" description:"对外的消息 ID"`
	ChatID     string    `orm:"column(chat_id);size(128)" description:"对话 ID"`
	UserId     int       `orm:"column(user_id);default(0)" description:"用户 ID，系统为 0"`
	Prompt     string    `orm:"column(prompt);type(text)" description:"用户输入"`
	RawPrompt  string    `orm:"column(raw_prompt);type(text)" description:"提交给LLM的Prompt"`
	Completion string    `orm:"column(completion);type(text)" description:"回复内容"`
	Reasoning  string    `orm:"column(reasoning);type(text)" description:"思考内容"`
	FileIds    string    `orm:"column(file_ids);type(text)" description:"对应的文件，多个逗号隔开"`
	ReportIds  string    `orm:"column(report_ids);type(text)" description:"对应的报告，多个逗号隔开"`
	Mark       int8      `orm:"column(mark);default(1)" description:"有效标识(1正常 0删除)"`
	CreatedAt  time.Time `orm:"column(created_at);type(datetime)" description:"创建时间"`
	UpdatedAt  time.Time `orm:"column(updated_at);type(datetime)" description:"更新时间"`
}

func (cr *ChatMessage) TableName() string {
	return "chat_messages"
}

func init() {
	orm.RegisterModel(new(ChatMessage))
}

// 根据条件查询单条数据
func (t *ChatMessage) Get() error {
	o := orm.NewOrm()
	query := o.QueryTable(new(ChatMessage))

	// 判断使用哪个参数进行查询
	if t.Id != 0 {
		query = query.Filter("id", t.Id)
	} else if t.MessageID != "" {
		query = query.Filter("message_id", t.MessageID)
	} else {
		return errors.New("没有提供查询条件")
	}

	// 查询单条记录
	err := query.One(t)
	if err == orm.ErrMultiRows {
		// 多条的时候报错
		return errors.New("查询到了多条记录")
	}
	if err == orm.ErrNoRows {
		// 没有找到记录
		return errors.New("未查询到记录")
	}

	return nil
}

// 插入数据
func (t *ChatMessage) Insert() (int64, error) {
	id, err := orm.NewOrm().Insert(t)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// 更新数据
func (t *ChatMessage) Update() (int64, error) {
	o := orm.NewOrm()
	rows, err := o.Update(t)
	if err := utils.HandleDBUpdateError(rows, err); err != nil {
		return 0, err
	}
	return rows, nil
}

func (t *ChatMessage) UpdateRawPrompt(rawPrompt string) (int64, error) {
	o := orm.NewOrm()

	// 更新回复内容
	t.RawPrompt = rawPrompt
	t.UpdatedAt = time.Now()

	// 只更新 RawPrompt 和 UpdatedAt 字段
	rows, err := o.Update(t, "RawPrompt", "UpdatedAt")
	if err != nil {
		return 0, errors.New("更新内容失败:" + err.Error())
	}
	if rows == 0 {
		return 0, errors.New("没有更新任何记录")
	}

	return rows, nil
}

// UpdateCompletion 更新回复内容
func (t *ChatMessage) UpdateCompletion(completion string, reasoning string) (int64, error) {
	o := orm.NewOrm()

	// 更新回复内容
	t.Completion = completion
	t.Reasoning = reasoning
	t.UpdatedAt = time.Now()

	// 只更新 Completion, Reasoning 和 UpdatedAt 字段
	rows, err := o.Update(t, "Completion", "Reasoning", "UpdatedAt")
	if err != nil {
		return 0, errors.New("更新内容失败:" + err.Error())
	}
	if rows == 0 {
		return 0, errors.New("没有更新任何记录")
	}

	return rows, nil
}

// 删除记录
func (t *ChatMessage) Delete() (int64, error) {
	o := orm.NewOrm()

	// 假设字段 mark 是用来标记删除状态的，0 表示已删除
	t.Mark = 0

	// 只更新 mark 字段
	rows, err := o.Update(t, "Mark")
	if err != nil {
		return 0, errors.New("删除失败:" + err.Error())
	}
	if rows == 0 {
		return 0, errors.New("没有操作任何记录")
	}

	return rows, nil
}
