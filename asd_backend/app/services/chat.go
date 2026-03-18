package services

import (
	"asd/app/dto"
	"asd/app/models"
	"asd/app/vo"
	"asd/utils/common"
	"errors"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

var Chat = new(chatService)

type chatService struct{}

func (s *chatService) GetList(req dto.ChatPageReq, userId int) ([]vo.ChatVo, int64, error) {
	o := orm.NewOrm()
	query := o.QueryTable(new(models.Chat)).Filter("mark", 1).Filter("user_id", userId)

	if req.Title != "" {
		query = query.Filter("title__icontains", req.Title)
	}

	count, _ := query.Count()
	query = query.OrderBy("-id").Limit(req.Limit, req.Offset)

	var list []models.Chat
	_, err := query.All(&list)

	// 数据处理
	var result []vo.ChatVo
	for _, v := range list {
		// 创建ChatVo对象并设置基本属性
		item := vo.ChatVo{
			ChatID:    v.ChatID,
			Title:     v.Title,
			UserId:    v.UserId,
			CreatedAt: v.CreatedAt.Local().Format("2006-01-02 15:04:05"),
		}

		result = append(result, item)
	}

	return result, count, err
}

func (s *chatService) Detail(chatID string, userId int) (*vo.ChatVo, []vo.ChatMessageVo, error) {

	chat := &models.Chat{ChatID: chatID}

	if err := chat.Get(); err != nil {
		return nil, nil, err
	}

	if chat.UserId != userId {
		return nil, nil, errors.New("无权限操作")
	}

	if chat.Mark != 1 {
		return nil, nil, errors.New("对话已删除")
	}

	detail := vo.ChatVo{
		ChatID:    chat.ChatID,
		Title:     chat.Title,
		UserId:    chat.UserId,
		CreatedAt: chat.CreatedAt.Local().Format("2006-01-02 15:04:05"),
	}

	query := dto.ChatMessagePageReq{
		ChatID: chatID,
		Limit:  1000,
		Page:   1,
	}
	messages, _, err := ChatMessage.GetList(query, userId)

	// 收集所有消息ID
	messageIDs := make([]string, len(messages))
	for i, msg := range messages {
		messageIDs[i] = msg.MessageID
	}

	return &detail, messages, err
}

func (s *chatService) Add(req dto.ChatAddReq, userId int) (string, error) {
	chatID := common.GetUUID()
	chat := &models.Chat{
		ChatID:    chatID,
		Title:     req.Title,
		UserId:    userId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Mark:      1,
	}

	_, err := chat.Insert()

	return chatID, err
}

func (s *chatService) Update(req dto.ChatUpdateReq, userId int) (int64, error) {
	// 将字符串ID转换为int64类型

	chat := &models.Chat{ChatID: req.ChatID}
	if err := chat.Get(); err != nil {
		return 0, err
	}

	if chat.UserId != userId {
		return 0, errors.New("无权限操作")
	}

	chat.Title = req.Title
	chat.UpdatedAt = time.Now()
	return chat.Update()
}

// 更新对话上下文
func (s *chatService) UpdateContext(chatId string, context string, userId int) (int64, error) {
	chat := &models.Chat{ChatID: chatId}
	if err := chat.Get(); err != nil {
		return 0, err
	}

	if chat.UserId != userId {
		return 0, errors.New("无权限操作")
	}

	chat.UpdatedAt = time.Now()
	return chat.Update()
}

func (s *chatService) Delete(chatID string, userId int) (int64, error) {
	chat := &models.Chat{ChatID: chatID}
	if err := chat.Get(); err != nil {
		return 0, err
	}
	fmt.Println(chatID)

	if chat.UserId != userId {
		return 0, errors.New("无权限操作")
	}

	return chat.Delete()
}

// 开始对话，系统回复，直接写入表中
func (s *chatService) StartChat(req dto.ChatMessageStartReq, userId int) (string, error) {
	chatId := common.GetUUID()
	chat := &models.Chat{
		ChatID:    chatId,
		Title:     req.Title, // @todo 需要有一个生成规则
		UserId:    userId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Mark:      1,
	}

	_, err := chat.Insert()

	if err != nil {
		return "", err
	}

	return chatId, err
}

// UpdateUsageStats 更新消息使用统计
func (s *chatService) UpdateUsageStats(chatID string, usageStats *dto.UsageStats, userId int) error {
	if usageStats == nil || usageStats.TotalTokens == 0 {
		logs.Info("没有使用统计数据需要更新")
		return nil
	}

	// 获取最新的消息记录
	o := orm.NewOrm()
	var message models.ChatMessage
	err := o.QueryTable(new(models.ChatMessage)).
		Filter("chat_id", chatID).
		Filter("user_id", userId).
		OrderBy("-id").
		One(&message)

	if err != nil {
		logs.Error("获取最新消息失败: %v", err)
		return err
	}

	// 创建使用统计记录
	usageRecord := &models.MessageUsageStats{
		MessageID:        message.MessageID,
		ChatID:           chatID,
		UserID:           fmt.Sprintf("%d", userId), // 转换为字符串
		Model:            usageStats.Model,
		PromptTokens:     usageStats.PromptTokens,
		CompletionTokens: usageStats.CompletionTokens,
		TotalTokens:      usageStats.TotalTokens,
		Cost:             usageStats.Cost,
		Currency:         usageStats.Currency,
	}

	// 先尝试查询是否已存在记录
	existingRecord := &models.MessageUsageStats{MessageID: message.MessageID}
	err = existingRecord.GetByMessageID()

	if err == nil {
		// 记录已存在，更新
		existingRecord.PromptTokens = usageStats.PromptTokens
		existingRecord.CompletionTokens = usageStats.CompletionTokens
		existingRecord.TotalTokens = usageStats.TotalTokens
		existingRecord.Model = usageStats.Model
		existingRecord.Cost = usageStats.Cost
		existingRecord.Currency = usageStats.Currency

		_, err = existingRecord.Update()
		if err != nil {
			logs.Error("更新使用统计数据失败: %v", err)
			return err
		}
		logs.Info("已更新使用统计数据，消息ID: %s, 总tokens: %d", message.MessageID, usageStats.TotalTokens)
	} else {
		// 记录不存在，插入新记录
		_, err = usageRecord.Insert()
		if err != nil {
			logs.Error("保存使用统计数据失败: %v", err)
			return err
		}
		logs.Info("已保存使用统计数据，消息ID: %s, 总tokens: %d", message.MessageID, usageStats.TotalTokens)
	}

	return nil
}
