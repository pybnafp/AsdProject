package services

import (
	"asd/app/dto"
	"asd/app/models"
	"asd/app/vo"
	"asd/utils/tencent"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

var Report = new(reportService)

type reportService struct{}

func (s *reportService) GetList(req dto.ReportPageReq, userId int) ([]vo.ReportVo, int64, error) {
	o := orm.NewOrm()
	query := o.QueryTable(new(models.Report)).Filter("user_id", userId)

	if req.Type > 0 {
		query = query.Filter("type", req.Type)
	}
	if req.Year > 0 {
		query = query.Filter("report_date__gte", fmt.Sprintf("%d-01-01", req.Year))
		query = query.Filter("report_date__lt", fmt.Sprintf("%d-01-01", req.Year+1))
	}

	count, _ := query.Count()
	query = query.OrderBy("-id").Limit(req.Limit, req.Offset)

	var list []models.Report
	_, err := query.All(&list)

	// 数据处理
	var result []vo.ReportVo
	for _, v := range list {
		// 创建ChatVo对象并设置基本属性
		item := vo.ReportVo{
			ReportID:         v.ReportId,
			Name:             v.Name,
			Type:             v.Type,
			Status:           v.Status,
			OriginalFileName: v.OriginalFileName,
			OriginalFileSize: v.OriginalFileSize,
			ReportFileName:   v.ReportFileName,
			ReportFileSize:   v.ReportFileSize,
			ReportDate:       v.ReportDate.Local().Format("2006-01-02"),
			CreatedAt:        v.CreatedAt.Local().Format("2006-01-02 15:04:05"),
		}

		if url, err := tencent.GetDownloadURL(v.OriginalFilePath); err != nil {
			return nil, 0, err
		} else {
			item.OriginalFileUrl = url
		}

		if url, err := tencent.GetDownloadURL(v.ReportFilePath); err != nil {
			return nil, 0, err
		} else {
			item.ReportFileUrl = url
		}

		result = append(result, item)
	}

	return result, count, err
}

type AddReportInput struct {
	ReportId         string
	UserID           int
	Name             string
	Type             int
	ReportDate       time.Time
	OriginalFileName string
	OriginalFilePath string
	OriginalFileSize int64
	ReportFileName   string
	ReportFilePath   string
	ReportFileSize   int64
	Content          string
	Status           int
}

func (s *reportService) AddReport(req AddReportInput) (*vo.ReportVo, error) {
	report := models.Report{
		ReportId:         req.ReportId,
		UserID:           req.UserID,
		Type:             req.Type,
		ReportDate:       req.ReportDate,
		OriginalFileName: req.OriginalFileName,
		OriginalFilePath: req.OriginalFilePath,
		OriginalFileSize: req.OriginalFileSize,
		ReportFileName:   req.ReportFileName,
		ReportFilePath:   req.ReportFilePath,
		ReportFileSize:   req.ReportFileSize,
		Content:          req.Content,
		Status:           req.Status,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	_, err := report.Insert()
	if err != nil {
		return nil, err
	}

	reportVo := vo.ReportVo{
		ReportID:         report.ReportId,
		UserID:           report.UserID,
		Type:             report.Type,
		Status:           report.Status,
		OriginalFileName: report.OriginalFileName,
		OriginalFileSize: report.OriginalFileSize,
		ReportFileName:   report.ReportFileName,
		ReportFileSize:   report.ReportFileSize,
		ReportDate:       report.ReportDate.Format("2006-01-02"),
		CreatedAt:        report.CreatedAt.Format("2006-01-02 15:04:05"),
		Content:          report.Content,
	}

	if url, err := tencent.GetDownloadURL(report.OriginalFilePath); err != nil {
		return nil, err
	} else {
		reportVo.OriginalFileUrl = url
	}

	if url, err := tencent.GetDownloadURL(report.ReportFilePath); err != nil {
		return nil, err
	} else {
		reportVo.ReportFileUrl = url
	}

	return &reportVo, nil
}

// GetReportsByIds 根据报告ID列表批量获取报告详情
func (s *reportService) GetReportsByIds(reportIds string, userId int) ([]vo.ReportVo, error) {
	if reportIds == "" {
		return []vo.ReportVo{}, nil
	}

	// 解析报告ID字符串，假设格式为逗号分隔的ID列表
	ids := strings.Split(reportIds, ",")
	if len(ids) == 0 {
		return []vo.ReportVo{}, nil
	}

	// 过滤空ID
	var validIds []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			validIds = append(validIds, id)
		}
	}

	if len(validIds) == 0 {
		return []vo.ReportVo{}, nil
	}

	// 一次性查询所有报告
	o := orm.NewOrm()
	var reports []models.Report

	// 构建IN查询条件
	qs := o.QueryTable(new(models.Report)).Filter("report_id__in", validIds)

	// 执行查询
	_, err := qs.All(&reports)
	if err != nil {
		return nil, fmt.Errorf("批量获取报告失败: %v", err)
	}

	// 构建结果列表
	var result []vo.ReportVo
	for _, report := range reports {
		// TODO: !验证报告所有权或公开性

		// 构建报告视图对象
		item := vo.ReportVo{
			ReportID:         report.ReportId,
			UserID:           report.UserID,
			Type:             report.Type,
			Status:           report.Status,
			OriginalFileName: report.OriginalFileName,
			OriginalFileSize: report.OriginalFileSize,
			ReportFileName:   report.ReportFileName,
			ReportFileSize:   report.ReportFileSize,
			ReportDate:       report.ReportDate.Format("2006-01-02"),
			CreatedAt:        report.CreatedAt.Format("2006-01-02 15:04:05"),
			Content:          report.Content,
		}

		if url, err := tencent.GetDownloadURL(report.OriginalFilePath); err != nil {
			return nil, err
		} else {
			item.OriginalFileUrl = url
		}

		if url, err := tencent.GetDownloadURL(report.ReportFilePath); err != nil {
			return nil, err
		} else {
			item.ReportFileUrl = url
		}
		result = append(result, item)
	}

	return result, nil
}

func (s *reportService) GetDetail(reportId string, userId int) (*vo.ReportVo, error) {
	record := &models.Report{ReportId: reportId}
	if err := record.Get(); err != nil {
		return nil, err
	}
	// TODO: 验证权限
	item := &vo.ReportVo{
		ReportID:         record.ReportId,
		UserID:           record.UserID,
		Type:             record.Type,
		Status:           record.Status,
		OriginalFileName: record.OriginalFileName,
		OriginalFileSize: record.OriginalFileSize,
		ReportFileName:   record.ReportFileName,
		ReportFileSize:   record.ReportFileSize,
		ReportDate:       record.ReportDate.Format("2006-01-02"),
		CreatedAt:        record.CreatedAt.Format("2006-01-02 15:04:05"),
		Content:          record.Content,
	}
	if url, err := tencent.GetDownloadURL(record.OriginalFilePath); err != nil {
		return nil, err
	} else {
		item.OriginalFileUrl = url
	}
	if url, err := tencent.GetDownloadURL(record.ReportFilePath); err != nil {
		return nil, err
	} else {
		item.ReportFileUrl = url
	}
	return item, nil
}
