package models

import (
	"asd/utils"
	"errors"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Chat struct {
	Id        int64     `orm:"column(id);auto" description:"主键ID"`
	ChatID    string    `orm:"column(chat_id);size(128)" description:"对外的 chatid"`
	Title     string    `orm:"column(title);size(255)" description:"标题"`
	UserId    int       `orm:"column(user_id);default(0)" description:"用户ID"`
	CreatedAt time.Time `orm:"column(created_at);type(datetime)" description:"创建时间"`
	UpdatedAt time.Time `orm:"column(updated_at);type(datetime)" description:"更新时间"`
	Mark      int8      `orm:"column(mark);default(1)" description:"有效标识(1正常 0删除)"`
}

func (t *Chat) TableName() string {
	return "chats"
}

func init() {
	orm.RegisterModel(new(Chat))
}

// 根据条件查询单条数据
func (t *Chat) Get() error {
	o := orm.NewOrm()
	query := o.QueryTable(new(Chat))

	// 判断使用哪个参数进行查询
	if t.Id != 0 {
		query = query.Filter("id", t.Id)
	} else if t.ChatID != "" {
		query = query.Filter("chat_id", t.ChatID)
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
func (t *Chat) Insert() (int64, error) {
	id, err := orm.NewOrm().Insert(t)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// 更新数据
func (t *Chat) Update() (int64, error) {
	o := orm.NewOrm()
	rows, err := o.Update(t)
	if err := utils.HandleDBUpdateError(rows, err); err != nil {
		return 0, err
	}
	return rows, nil
}

// 删除记录
func (t *Chat) Delete() (int64, error) {
	o := orm.NewOrm()

	// 假设字段 mark 是用来标记删除状态的，0 表示已删除
	t.Mark = 0
	t.UpdatedAt = time.Now()

	// 只更新 mark 字段
	rows, err := o.Update(t, "Mark", "UpdatedAt")
	fmt.Println("删除行数:", rows)
	if err != nil {
		return 0, errors.New("删除失败:" + err.Error())
	}
	if rows == 0 {
		return 0, errors.New("没有操作任何记录")
	}

	return rows, nil
}
