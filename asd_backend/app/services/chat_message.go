package services

import (
	"asd/app/dto"
	"asd/app/models"
	"asd/app/vo"
	"asd/utils/common"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

var ChatMessage = new(ChatMessageService)

type ChatMessageService struct{}

func (s *ChatMessageService) GetList(req dto.ChatMessagePageReq, userId int) ([]vo.ChatMessageVo, int64, error) {
	o := orm.NewOrm()

	query := o.QueryTable(new(models.ChatMessage)).Filter("mark", 1).Filter("chat_id", req.ChatID)

	count, _ := query.Count()

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	query = query.OrderBy("id").Limit(req.Limit, (req.Page-1)*req.Limit)

	var list []models.ChatMessage
	_, err := query.All(&list)

	var result []vo.ChatMessageVo
	for _, v := range list {
		if v.UserId != userId && v.UserId != 0 {
			return nil, 0, fmt.Errorf("对话不属于当前用户")
		}

		// 处理资源文件下载地址和详情
		var files []vo.FileVo
		if v.FileIds != "" {
			// 获取文件详细信息
			files, err = FileService.GetFilesByIds(v.FileIds, userId)
			if err != nil {
				fmt.Printf("获取文件详情失败: %v\n", err)
				// 继续处理，不中断流程
			}
		}

		var reports []vo.ReportVo
		if v.ReportIds != "" {
			// 获取文件详细信息
			reports, err = Report.GetReportsByIds(v.ReportIds, userId)
			if err != nil {
				fmt.Printf("获取文件详情失败: %v\n", err)
				// 继续处理，不中断流程
			}
		}

		item := vo.ChatMessageVo{
			MessageID:  v.MessageID,
			ChatID:     v.ChatID,
			UserId:     v.UserId,
			Prompt:     v.Prompt,
			Completion: v.Completion,
			Files:      files, // 添加文件详细信息
			Reports:    reports,
			CreatedAt:  v.CreatedAt.Format("2006-01-02 15:04:05"),
			Reasoning:  v.Reasoning,
		}
		result = append(result, item)
	}

	return result, count, err
}

// 获取最近的消息记录
func (s *ChatMessageService) GetRecentList(req dto.ChatMessagePageReq, userId int) ([]vo.ChatMessageVo, int64, error) {
	o := orm.NewOrm()

	query := o.QueryTable(new(models.ChatMessage)).Filter("mark", 1).Filter("chat_id", req.ChatID).Exclude("completion", "")

	count, _ := query.Count()

	query = query.OrderBy("-id").Limit(req.Limit, (req.Page-1)*req.Limit)

	var list []models.ChatMessage
	_, err := query.All(&list)

	// 对结果按 id 升序排序，恢复消息的时间顺序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})

	var result []vo.ChatMessageVo
	for _, v := range list {
		if v.UserId != userId && v.UserId != 0 {
			return nil, 0, fmt.Errorf("对话不属于当前用户")
		}

		item := vo.ChatMessageVo{
			MessageID:  v.MessageID,
			ChatID:     v.ChatID,
			UserId:     v.UserId,
			Prompt:     v.Prompt,
			Completion: v.Completion,
			CreatedAt:  v.CreatedAt.Format("2006-01-02 15:04:05"),
			RawPrompt:  v.RawPrompt,
			Reasoning:  v.Reasoning,
		}
		result = append(result, item)
	}

	return result, count, err
}

func (s *ChatMessageService) Add(
	chatId string, prompt, completion string, fileIDs []string, reportIds []string,
	userId int) (string, error) {
	messageID := common.GetUUID()
	record := &models.ChatMessage{
		MessageID:  messageID,
		ChatID:     chatId,
		UserId:     userId,
		Prompt:     prompt,
		Completion: completion,
		// 将文件ID数组拼接成逗号分隔的字符串
		FileIds:   strings.Join(fileIDs, ","),
		ReportIds: strings.Join(reportIds, ","),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Mark:      1,
	}

	_, err := record.Insert()

	return messageID, err
}

func (s *ChatMessageService) UpdateRawPrompt(messageID string, rawPrompt string, userId int) (int64, error) {
	record := &models.ChatMessage{MessageID: messageID}
	if err := record.Get(); err != nil {
		return 0, err
	}

	if record.UserId != userId {
		return 0, errors.New("无权限操作")
	}

	return record.UpdateRawPrompt(rawPrompt)
}

func (s *ChatMessageService) UpdateCompletion(messageID string, completion string, reasoning string, userId int) (int64, error) {
	record := &models.ChatMessage{MessageID: messageID}
	if err := record.Get(); err != nil {
		return 0, err
	}

	if record.UserId != userId {
		return 0, errors.New("无权限操作")
	}

	return record.UpdateCompletion(completion, reasoning)
}

func (s *ChatMessageService) Delete(id int64, userId int) (int64, error) {
	record := &models.ChatMessage{Id: id}
	if err := record.Get(); err != nil {
		return 0, err
	}

	if record.UserId != userId {
		return 0, errors.New("无权限操作")
	}

	return record.Delete()
}

// 根据对话ID删除所有消息
func (s *ChatMessageService) DeleteByChatID(chatID string, userId int) (int64, error) {
	o := orm.NewOrm()

	// 批量更新 mark 字段
	num, err := o.QueryTable(new(models.ChatMessage)).
		Filter("chat_id", chatID).
		Filter("mark", 1).
		Update(orm.Params{
			"mark":       0,
			"updated_at": time.Now(),
		})

	if err != nil {
		return 0, fmt.Errorf("删除消息失败: %v", err)
	}

	return num, nil
}

// GetRecentMessages 获取最近的消息记录
func (s *ChatMessageService) GetRecentMessages(chatID string, limit int, userId int) ([]map[string]string, error) {
	// 创建查询请求
	messageReq := dto.ChatMessagePageReq{
		ChatID: chatID,
		Page:   1,
		Limit:  limit,
	}

	// 获取消息列表
	messages, _, err := s.GetRecentList(messageReq, userId)
	if err != nil {
		return nil, err
	}

	// 构建历史消息内容
	var historyMessages []map[string]string
	for _, msg := range messages {
		historyMessages = append(historyMessages, map[string]string{
			"role":    "user",
			"content": msg.RawPrompt,
		})
		historyMessages = append(historyMessages, map[string]string{
			"role":    "assistant",
			"content": msg.Completion,
		})
	}

	return historyMessages, nil
}

func (s *ChatMessageService) GetLatestMessage(chatID string, userId int) (*models.ChatMessage, error) {
	o := orm.NewOrm()
	// 查询最新的消息记录
	var message models.ChatMessage
	err := o.QueryTable(new(models.ChatMessage)).
		Filter("chat_id", chatID).
		Filter("message_id__isnull", false).
		Filter("mark", 1).
		OrderBy("-id").
		Limit(1).
		One(&message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// 获取消息详情
func (s *ChatMessageService) GetDetail(messageID string, userId int) (*vo.ChatMessageVo, error) {
	record := &models.ChatMessage{MessageID: messageID}
	if err := record.Get(); err != nil {
		return nil, err
	}
	if record.UserId != userId {
		return nil, errors.New("无权限操作")
	}
	// 处理资源文件下载地址和详情
	var files []vo.FileVo
	if record.FileIds != "" {
		// 获取文件详细信息
		files, _ = FileService.GetFilesByIds(record.FileIds, userId)
	}
	var reports []vo.ReportVo
	if record.ReportIds != "" {
		// 获取文件详细信息
		reports, _ = Report.GetReportsByIds(record.ReportIds, userId)
	}
	item := vo.ChatMessageVo{
		MessageID:  record.MessageID,
		ChatID:     record.ChatID,
		UserId:     record.UserId,
		Prompt:     record.Prompt,
		Completion: record.Completion,
		Files:      files, // 添加文件详细信息
		Reports:    reports,
		CreatedAt:  record.CreatedAt.Format("2006-01-02 15:04:05"),
		Reasoning:  record.Reasoning,
	}
	return &item, nil
}
