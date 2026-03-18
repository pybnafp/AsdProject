package models

import (
	"errors"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

const (
	ReportTypeComprehensive       = 1 // 综合报告
	ReportTypeQuestionnaireScales = 2 // 问卷量表
	ReportTypeFacialData          = 3 // 面容数据
	ReportTypeEyeTrackingData     = 4 // 眼动数据
	ReportTypeBrainImagingData    = 5 // 脑影像数据
	ReportTypeBahavioralVideo     = 6 // 行为视频
	ReportTypeGeneticData         = 7 // 基因数据
)

const (
	ReportStatusUnderReview = 1 // 待审核
	ReportStatusReviewed    = 2 // 已审核
)

type Report struct {
	Id               int64     `orm:"column(id);auto" description:"ID"`
	ReportId         string    `orm:"column(report_id);size(36);unique" description:"报告ID"`
	UserID           int       `orm:"column(user_id);" description:"用户ID"`
	Name             string    `orm:"column(name);" description:"报告人员姓名"`
	OriginalFileName string    `orm:"column(original_file_name)" description:"原数据文件名"`
	OriginalFilePath string    `orm:"column(original_file_path)" description:"原数据文件路径"`
	OriginalFileSize int64     `orm:"column(original_file_size)" description:"原数据文件大小"`
	ReportFileName   string    `orm:"column(report_file_name)" description:"文字报告文件名"`
	ReportFilePath   string    `orm:"column(report_file_path)" description:"文字报告文件路径"`
	ReportFileSize   int64     `orm:"column(report_file_size)" description:"文字报告文件大小"`
	Content          string    `orm:"column(content)" description:"报告具体内容"`
	Type             int       `orm:"column(type)" description:"报告类型"`
	Status           int       `orm:"column(status)" description:"状态"`
	ReportDate       time.Time `orm:"column(report_date);type(date)" description:"报告日期"`
	CreatedAt        time.Time `orm:"column(created_at);type(datetime);default(current_timestamp)" description:"创建时间"`
	UpdatedAt        time.Time `orm:"column(updated_at);type(datetime);default(current_timestamp)" description:"更新时间"`
}

func (r *Report) TableName() string {
	return "reports"
}

func init() {
	orm.RegisterModel(new(Report))
}

// Insert 插入数据
func (u *Report) Insert() (int64, error) {
	id, err := orm.NewOrm().Insert(u)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Get 根据条件查询单条数据
func (u *Report) Get() error {
	o := orm.NewOrm()
	query := o.QueryTable(new(Report))

	if u.Id != 0 {
		query = query.Filter("id", u.Id)
	} else if u.ReportId != "" {
		query = query.Filter("report_id", u.ReportId)
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
