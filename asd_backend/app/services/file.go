package services

import (
	"asd/app/dto"
	"asd/app/models"
	"asd/app/vo"
	"asd/utils/common"
	"asd/utils/tencent"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

var FileService = new(fileService)

type fileService struct{}

// GetList 获取用户文件列表
func (s *fileService) GetList(req dto.FilePageReq, userId int) ([]vo.FileVo, int64, error) {
	o := orm.NewOrm()

	query := o.QueryTable(new(models.File)).Filter("user_id", userId)

	// 按文件类型筛选
	if req.FileType != "" {
		query = query.Filter("asset_type", req.FileType)
	}

	// 按可见性筛选
	if req.Visibility != "" {
		query = query.Filter("visibility", req.Visibility)
	}

	// 按状态筛选
	if req.Status != "" {
		query = query.Filter("status", req.Status)
	}

	if req.UserID != 0 {
		query = query.Filter("user_id", req.UserID)
	}

	count, _ := query.Count()

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	query = query.OrderBy("-created_at").Limit(req.Limit, (req.Page-1)*req.Limit)

	var list []models.File
	_, err := query.All(&list)

	var result []vo.FileVo
	for _, v := range list {
		downloadUrl, _ := tencent.GetDownloadURL(v.FilePath)

		item := vo.FileVo{
			FileID:      v.FileID,
			FileName:    v.FileName,
			FilePath:    v.FilePath,
			FileSize:    v.FileSize,
			FileType:    v.FileType,
			ContentType: v.ContentType,
			Description: v.Description,
			DownloadUrl: downloadUrl,
			Visibility:  v.Visibility,
			Status:      v.Status,
			CreatedAt:   v.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		result = append(result, item)
	}

	return result, count, err
}

// GetDetail 获取文件详情
func (s *fileService) GetDetail(fileId string, userId int) (*vo.FileVo, error) {
	record := &models.File{FileID: fileId}
	fmt.Println("fileId", fileId)
	if err := record.Get(); err != nil {
		return nil, err
	}

	// 验证文件所有权
	if record.UserID != userId {
		return nil, errors.New("无权限访问该文件")
	}

	// 获取下载URL
	downloadUrl, err := tencent.GetDownloadURL(record.FilePath)
	if err != nil {
		// 获取下载链接失败，记录错误并继续
		logs.Error("获取下载链接失败 (ID: %s): %v\n", record.FileID, err)
	}

	result := &vo.FileVo{
		FileID:      record.FileID,
		FileName:    record.FileName,
		FilePath:    record.FilePath,
		FileSize:    record.FileSize,
		FileType:    record.FileType,
		ContentType: record.ContentType,
		Description: record.Description,
		Metadata:    record.Metadata,
		Visibility:  record.Visibility,
		DownloadUrl: downloadUrl,
		CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
		Content:     record.Content,
	}

	return result, nil
}

// Add 添加文件
func (s *fileService) Add(req dto.FileAddReq, userId int) (string, error) {
	assetID := common.GetUUID()
	if req.FileID != "" {
		assetID = req.FileID
	}
	record := &models.File{
		FileID:      assetID,
		UserID:      userId,
		FileName:    req.FileName,
		FilePath:    req.FilePath,
		FileSize:    req.FileSize,
		FileType:    req.FileType,
		ContentType: req.ContentType,
		Content:     req.Content,
		Description: req.Description,
		Metadata:    req.Metadata,
		Visibility:  req.Visibility,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := record.Insert()
	if err != nil {
		return "", fmt.Errorf("添加文件失败: %v", err)
	}

	return assetID, nil
}

// Delete 删除文件
func (s *fileService) Delete(assetId string, userId int) error {
	record := &models.File{FileID: assetId}
	if err := record.Get(); err != nil {
		return err
	}

	// 验证文件所有权
	if record.UserID != userId {
		return errors.New("无权限删除该文件")
	}

	_, err := record.Delete()
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	return nil
}

// ChangeStatus 更改文件状态
func (s *fileService) ChangeStatus(assetId string, status string, userId int) error {
	record := &models.File{FileID: assetId}
	if err := record.Get(); err != nil {
		return err
	}

	// 验证文件所有权
	if record.UserID != userId {
		return errors.New("无权限修改该文件")
	}

	record.Status = status
	record.UpdatedAt = time.Now()

	_, err := record.Update()
	if err != nil {
		return fmt.Errorf("更新文件状态失败: %v", err)
	}

	return nil
}

// GetDownloadURLsByFileIds 根据文件ID列表获取下载地址列表
func (s *fileService) GetDownloadURLsByFileIds(fileIds string, userId int) ([]string, error) {
	if fileIds == "" {
		return []string{}, nil
	}

	// 解析文件ID字符串，假设格式为逗号分隔的ID列表
	ids := strings.Split(fileIds, ",")
	if len(ids) == 0 {
		return []string{}, nil
	}

	// 一次性查询所有文件
	o := orm.NewOrm()
	var assets []models.File

	// 构建IN查询条件
	qs := o.QueryTable(new(models.File)).Filter("file_id__in", ids)

	// 执行查询
	_, err := qs.All(&assets)
	if err != nil {
		return nil, fmt.Errorf("批量获取文件失败: %v", err)
	}

	// 创建文件ID到文件对象的映射，方便后续查找
	assetMap := make(map[string]*models.File)
	for i := range assets {
		assetMap[assets[i].FileID] = &assets[i]
	}

	// 构建下载地址列表
	var urls []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}

		asset, exists := assetMap[id]
		if !exists {
			// 文件不存在，跳过
			continue
		}

		// 验证文件所有权或公开性
		if asset.UserID != userId && asset.Visibility != "public" {
			// 无权限访问，跳过
			continue
		}

		// 获取下载URL
		url, err := tencent.GetDownloadURL(asset.FilePath)
		if err != nil {
			// 获取下载链接失败，记录错误并继续
			fmt.Printf("获取下载链接失败 (ID: %s): %v\n", id, err)
			continue
		}

		// 只添加URL到结果列表
		urls = append(urls, url)
	}

	return urls, nil
}

// GetFilesByIds 根据文件ID列表批量获取文件详情
func (s *fileService) GetFilesByIds(fileIds string, userId int) ([]vo.FileVo, error) {
	if fileIds == "" {
		return []vo.FileVo{}, nil
	}

	// 解析文件ID字符串，假设格式为逗号分隔的ID列表
	ids := strings.Split(fileIds, ",")
	if len(ids) == 0 {
		return []vo.FileVo{}, nil
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
		return []vo.FileVo{}, nil
	}

	// 一次性查询所有文件
	o := orm.NewOrm()
	var files []models.File

	// 构建IN查询条件
	qs := o.QueryTable(new(models.File)).Filter("file_id__in", validIds)

	// 执行查询
	_, err := qs.All(&files)
	if err != nil {
		return nil, fmt.Errorf("批量获取文件失败: %v", err)
	}

	// 构建结果列表
	var result []vo.FileVo
	for _, file := range files {
		// 验证文件所有权或公开性
		if file.UserID != userId && file.Visibility != "public" {
			// 无权限访问，跳过
			continue
		}

		// 获取下载URL
		downloadUrl, err := tencent.GetDownloadURL(file.FilePath)
		if err != nil {
			// 获取下载链接失败，记录错误并继续
			fmt.Printf("获取下载链接失败 (ID: %s): %v\n", file.FileID, err)
			continue
		}

		// 构建文件视图对象
		item := vo.FileVo{
			FileID:      file.FileID,
			FileName:    file.FileName,
			FilePath:    file.FilePath,
			FileSize:    file.FileSize,
			FileType:    file.FileType,
			ContentType: file.ContentType,
			Description: file.Description,
			Visibility:  file.Visibility,
			Status:      file.Status,
			DownloadUrl: downloadUrl,
			CreatedAt:   file.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		result = append(result, item)
	}

	return result, nil
}
