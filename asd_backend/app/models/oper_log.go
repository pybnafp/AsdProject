package models

import (
	"asd/utils"
	"errors"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// OperLog 操作日志记录
type OperLog struct {
	Id           int       `orm:"column(id);auto" description:"主键ID"`
	UserId       int       `orm:"column(user_id);default(0)" description:"用户ID"`
	Model        string    `orm:"column(model);size(150)" description:"操作模块"`
	OperType     int       `orm:"column(oper_type);default(0)" description:"操作类型：0其它 1新增 2修改 3删除 4查询 5设置状态 6导入 7导出 8设置权限 9设置密码"`
	OperMethod   string    `orm:"column(oper_method);size(30);null" description:"操作方法"`
	Username     string    `orm:"column(username);size(255);null" description:"操作账号"`
	OperName     string    `orm:"column(oper_name);size(50);null" description:"操作用户"`
	OperUrl      string    `orm:"column(oper_url);size(255);null" description:"请求URL"`
	OperIp       string    `orm:"column(oper_ip);size(50)" description:"主机地址"`
	OperLocation string    `orm:"column(oper_location);size(255)" description:"操作地点"`
	RequestParam string    `orm:"column(request_param);size(2000)" description:"请求参数"`
	Result       string    `orm:"column(result);size(2000)" description:"返回参数"`
	Status       int8      `orm:"column(status);default(0)" description:"日志状态：0正常日志 1错误日志"`
	UserAgent    string    `orm:"column(user_agent);type(text);null" description:"代理信息"`
	Note         string    `orm:"column(note);size(2000);null" description:"备注"`
	CreateUser   int       `orm:"column(create_user)" description:"添加人"`
	CreateTime   time.Time `orm:"column(create_time);type(datetime);auto_now_add" description:"操作时间"`
	UpdateUser   int       `orm:"column(update_user);default(0)" description:"更新人"`
	UpdateTime   time.Time `orm:"column(update_time);type(datetime);auto_now" description:"更新时间"`
	Mark         int8      `orm:"column(mark);default(1)" description:"有效标识"`
}

// TableName 获取对应数据库表名
func (t *OperLog) TableName() string {
	return "sys_oper_logs"
}

// 初始化
func init() {
	orm.RegisterModel(new(OperLog))
}

// 获取单条记录
func (t *OperLog) Get() error {
	err := orm.NewOrm().QueryTable(new(OperLog)).Filter("id", t.Id).One(t)
	if err == orm.ErrMultiRows {
		return errors.New("查询到了多条记录")
	}
	if err == orm.ErrNoRows {
		return errors.New("未查询到记录")
	}
	return nil
}

// 插入数据
func (t *OperLog) Insert() (int64, error) {
	id, err := orm.NewOrm().Insert(t)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// 更新数据
func (t *OperLog) Update() (int64, error) {
	o := orm.NewOrm()
	rows, err := o.Update(t)
	if err := utils.HandleDBUpdateError(rows, err); err != nil {
		return 0, err
	}
	return rows, nil
}

// 删除数据
func (t *OperLog) Delete() (int64, error) {
	o := orm.NewOrm()
	rows, err := o.Delete(t)
	if err := utils.HandleDBDeleteError(rows, err); err != nil {
		return 0, err
	}
	return rows, nil
}
