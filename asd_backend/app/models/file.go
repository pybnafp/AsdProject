package models

import (
	"errors"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type File struct {
	Id          int64     `orm:"column(id);auto" description:"ID"`
	FileID      string    `orm:"column(file_id);size(36);unique" description:"文件ID"`
	UserID      int       `orm:"column(user_id);" description:"用户ID"`
	FileName    string    `orm:"column(file_name);size(255)" description:"文件名"`
	FilePath    string    `orm:"column(file_path);size(512)" description:"存储路径"`
	FileSize    int64     `orm:"column(file_size)" description:"文件大小"`
	FileType    string    `orm:"column(file_type);size(100)" description:"文件类型"`
	ContentType string    `orm:"column(content_type);size(100)" description:"MIME类型"`
	Description string    `orm:"column(description);type(text)" description:"描述"`
	Metadata    string    `orm:"column(metadata);type(text)" description:"元数据"`
	Content     string    `orm:"column(content);type(text);default()" description:"文件解析得到的Markdown格式内容"`
	Visibility  string    `orm:"column(visibility);default(private)" description:"可见性"`
	Status      string    `orm:"column(status);default(active)" description:"状态"`
	CreatedAt   time.Time `orm:"column(created_at);type(datetime);default(current_timestamp)" description:"创建时间"`
	UpdatedAt   time.Time `orm:"column(updated_at);type(datetime);default(current_timestamp)" description:"更新时间"`
}

func (u *File) TableName() string {
	return "files"
}

func init() {
	orm.RegisterModel(new(File))
}

// Insert 插入数据
func (u *File) Insert() (int64, error) {
	id, err := orm.NewOrm().Insert(u)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Get 根据条件查询单条数据
func (u *File) Get() error {
	o := orm.NewOrm()
	query := o.QueryTable(new(File))

	if u.Id != 0 {
		query = query.Filter("id", u.Id)
	} else if u.FileID != "" {
		query = query.Filter("file_id", u.FileID)
	} else {
		return errors.New("没有提供查询条件")
	}

	err := query.One(u)
	if err == orm.ErrMultiRows {
		return errors.New("查询到了多条记录")
	}
	if err == orm.ErrNoRows {
		return errors.New("未查询到记录")
	}

	return nil
}

// Update 更新数据
func (u *File) Update() (int64, error) {
	o := orm.NewOrm()
	rows, err := o.Update(u)
	if err != nil {
		return 0, err
	}
	return rows, nil
}

// Delete 删除记录
func (u *File) Delete() (int64, error) {
	o := orm.NewOrm()
	rows, err := o.Delete(u)
	if err != nil {
		return 0, err
	}
	return rows, nil
}
