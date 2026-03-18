package models

import (
	"asd/utils"
	"database/sql"
	"errors"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type User struct {
	Id            int            `orm:"column(id);auto" description:"主键ID"`
	WechatOpenId  sql.NullString `orm:"column(wechat_open_id);size(36);null;unique" description:"微信登录OpenID"`
	WechatUnionId sql.NullString `orm:"column(wechat_union_id);size(36);null" description:"微信登录UnionID"`
	Realname      string         `orm:"column(realname);size(150);null" description:"真实姓名"`
	Nickname      string         `orm:"column(nickname);size(150);null" description:"昵称"`
	Gender        int            `orm:"column(gender);null" description:"性别:1男 2女 3保密"`
	Avatar        string         `orm:"column(avatar);type(text);null" description:"头像"`
	Mobile        sql.NullString `orm:"column(mobile);size(11);unique;null" description:"手机号码"`
	Email         string         `orm:"column(email);size(30);null" description:"邮箱地址"`
	Birthday      time.Time      `orm:"column(birthday);type(date);null" description:"出生日期"`
	ProvinceCode  string         `orm:"column(province_code);size(50);null" description:"省份编号"`
	CityCode      string         `orm:"column(city_code);size(50);null" description:"市区编号"`
	DistrictCode  string         `orm:"column(district_code);size(50);null" description:"区县编号"`
	Address       string         `orm:"column(address);size(255);null" description:"详细地址"`
	CityName      string         `orm:"column(city_name);size(150);null" description:"所属城市"`
	Username      string         `orm:"column(username);size(50);null" description:"登录用户名"`
	Password      string         `orm:"column(password);size(150);null" description:"登录密码"`
	Jwt           string         `orm:"column(jwt);size(255);null" description:"JWT"`
	Intro         string         `orm:"column(intro);size(500);null" description:"个人简介"`
	Status        int            `orm:"column(status);null" description:"状态：1正常 2禁用"`
	Note          string         `orm:"column(note);size(500);null" description:"备注"`
	Sort          int            `orm:"column(sort);null" description:"排序号"`
	LoginNum      int            `orm:"column(login_num);null" description:"登录次数"`
	LoginIp       string         `orm:"column(login_ip);size(20);null" description:"最近登录IP"`
	LoginTime     time.Time      `orm:"column(login_time);type(datetime);null" description:"最近登录时间"`
	CreateUser    int            `orm:"column(create_user);null" description:"添加人"`
	CreateTime    time.Time      `orm:"column(create_time);type(datetime);null" description:"创建时间"`
	UpdateUser    int            `orm:"column(update_user);null" description:"更新人"`
	UpdateTime    time.Time      `orm:"column(update_time);type(datetime);null" description:"更新时间"`
	Mark          int            `orm:"column(mark)" description:"有效标识(1正常 0删除)"`
}

func (t *User) TableName() string {
	return "users"
}

func init() {
	orm.RegisterModel(new(User))
}

// 根据条件查询单条数据
func (t *User) Get() error {
	o := orm.NewOrm()
	query := o.QueryTable(new(User))

	if t.Id != 0 {
		query = query.Filter("id", t.Id)
	} else if t.WechatOpenId.String != "" {
		query = query.Filter("wechat_open_id", t.WechatOpenId.String)
	} else if t.Mobile.String != "" {
		query = query.Filter("mobile", t.Mobile.String)
	} else {
		return errors.New("没有提供查询条件")
	}
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
func (t *User) Insert() (int64, error) {
	id, err := orm.NewOrm().Insert(t)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// INSERT_YOUR_REWRITE_HERE

// 更新数据
func (t *User) Update() (int64, error) {
	o := orm.NewOrm()
	rows, err := o.Update(t)
	if err := utils.HandleDBUpdateError(rows, err); err != nil {
		return 0, err
	}
	return rows, nil
}

// 删除记录
func (t *User) Delete() (int64, error) {
	o := orm.NewOrm()
	rows, err := o.Delete(t)
	if err := utils.HandleDBDeleteError(rows, err); err != nil {
		return 0, err
	}
	return rows, nil
}
