package controllers

import (
	"asd/app/dto"
	"asd/app/models"
	"asd/app/services"
	"asd/utils/common"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/gookit/validate"
)

var ChatApi = new(ChatApiController)

type ChatApiController struct {
	BaseController
}

// 对话列表
func (c *ChatApiController) List() {
	var req dto.ChatPageReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, "请求参数错误"+err.Error())
		return
	}

	if req.Offset < 0 {
		req.Offset = 0
	}

	if req.Limit == 0 {
		req.Limit = 30
	}

	if req.Limit > 1000 {
		req.Limit = 1000
	}

	// 参数校验
	v := validate.Struct(req)
	if !v.Validate() {
		c.ErrorJson(400, v.Errors.One())
		return
	}

	userId := c.GetUserId() // 从中间件获取用户ID
	list, count, err := services.Chat.GetList(req, userId)
	if err != nil {
		c.ErrorJson(400, err.Error())
		return
	}

	c.JSON(common.JsonResult{
		Code:  0,
		Data:  list,
		Count: count,
	})
}

// 更新对话
func (c *ChatApiController) Update() {
	var req dto.ChatUpdateReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, err.Error())
		return
	}

	fmt.Println(req)
	if req.Title == "" {
		c.ErrorJson(400, "Title 不能为空")
		return
	}

	userId := c.GetUserId()
	_, err := services.Chat.Update(req, userId)
	if err != nil {
		c.ErrorJson(400, err.Error())
		return
	}

	c.JSON(common.JsonResult{
		Code: 0,
		Msg:  "更新成功",
	})
}

// 删除对话
func (c *ChatApiController) Delete() {
	var req dto.ChatDeleteReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, err.Error())
		return
	}

	// 验证对话权限
	chat := &models.Chat{ChatID: req.ChatID}
	if err := chat.Get(); err != nil {
		c.ErrorJson(400, err.Error())
		return
	}

	userId := c.GetUserId()
	if chat.UserId != userId {
		c.ErrorJson(400, "无权限操作")
		return
	}

	_, err := services.Chat.Delete(req.ChatID, userId)
	if err != nil {
		c.ErrorJson(500, err.Error())
		return
	}

	_, err = services.ChatMessage.DeleteByChatID(req.ChatID, userId)

	if err != nil {
		c.ErrorJson(500, err.Error())
		return
	}

	c.JSON(common.JsonRes{
		Code: 0,
		Msg:  "删除成功",
	})
}

// 对话记录列表 根据对话 ID 查询
func (c *ChatApiController) MessageList() {
	var req dto.ChatMessagePageReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, "请求参数错误")
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}

	if req.Page > 100 {
		req.Page = 100
	}

	if req.Limit == 0 {
		req.Limit = 30
	}

	if req.Limit > 1000 {
		req.Limit = 1000
	}

	// 参数校验
	v := validate.Struct(req)
	if !v.Validate() {
		c.ErrorJson(400, v.Errors.One())
		return
	}

	userId := c.GetUserId()
	list, count, err := services.ChatMessage.GetList(req, userId)
	if err != nil {
		c.ErrorJson(400, err.Error())
		return
	}

	c.JSON(common.JsonResult{
		Code:  0,
		Data:  list,
		Count: count,
	})
}

// 根据对话 ID 查询
func (c *ChatApiController) Detail() {
	var req dto.ChatDetailReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, "请求参数错误")
		return
	}

	// 参数校验
	v := validate.Struct(req)
	if !v.Validate() {
		c.ErrorJson(400, v.Errors.One())
		return
	}

	userId := c.GetUserId()
	chat, messages, err := services.Chat.Detail(req.ChatID, userId)
	if err != nil {
		c.ErrorJson(400, err.Error())
		return
	}

	c.JSON(common.JsonRes{
		Code: 0,
		Data: map[string]interface{}{
			"chat":     chat,
			"messages": messages,
		},
	})
}

// 流式对话
// 创建流式对话并发起请求
func (c *ChatApiController) CreateStream() {
	var req dto.StreamChatReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, "请求参数错误"+err.Error())
		return
	}

	// 参数校验
	v := validate.Struct(req)
	if !v.Validate() {
		c.ErrorJson(400, v.Errors.One())
		return
	}

	userId := c.GetUserId()

	// 如果 chatId 为空，则创建一个新的对话
	if req.ChatID == "" {
		// 使用 prompt 作为标题，如果 prompt 太长则截取前30个字符
		title := req.Prompt
		titleRunes := []rune(title)
		if len(titleRunes) > 30 {
			title = string(titleRunes[:30]) + "..."
		}

		chatAddReq := dto.ChatAddReq{
			Title: title,
		}

		var err error
		req.ChatID, err = services.Chat.Add(chatAddReq, userId)
		if err != nil {
			logs.Error("创建对话失败: %v", err)
			c.ErrorJson(500, "创建对话失败: "+err.Error())
			return
		}

		logs.Info("已创建新对话，ID: %s, 标题: %s", req.ChatID, title)
	}

	// 验证对话权限
	chat := &models.Chat{ChatID: req.ChatID}
	if err := chat.Get(); err != nil {
		c.ErrorJson(400, "对话不存在")
		return
	}
	if chat.UserId != userId {
		logs.Error("400 无权操作此对话 chat.UserId=", chat.UserId, " authUserId=", userId)
		c.ErrorJson(400, "无权操作此对话")
		return
	}

	// 创建消息ID
	var messageID string

	// 否则创建新的消息ID
	var err error
	messageID, err = services.ChatMessage.Add(
		req.ChatID, req.Prompt, "", req.FileIDs, req.ReportIDs,
		userId)
	if err != nil {
		logs.Error("保存用户消息失败: %v", err)
		c.ErrorJson(500, "保存消息失败")
		return
	}

	// 创建Redis键名，用于存储流式响应
	redisKey, _ := services.GenerateChatRedisKeys(messageID)

	// 检查Redis中是否已有该消息的响应
	exists, err := services.GetRedisClient().Exists(redisKey)
	if err != nil {
		logs.Error("检查Redis键是否存在失败: %v", err)
	}

	// 如果请求中包含 messageID 且 Redis 中已有响应，则不需要重新请求百炼API
	if !exists {
		// 获取最新的一条消息
		// 准备Agent请求
		agentReq := dto.AgentChatReq{
			ChatID: req.ChatID,
			Prompt: req.Prompt,
		}

		// 获取最近的10条消息记录
		historyMessages, err := services.ChatMessage.GetRecentMessages(req.ChatID, 10, userId)
		if err != nil {
			logs.Error("获取历史消息失败: %v", err)
		} else if len(historyMessages) > 0 {
			// 将历史消息添加到请求中
			agentReq.Messages = historyMessages
			logs.Info("添加了 %d 条历史消息到请求中", len(historyMessages))
		}

		// 如果有文件ID，处理文件内容
		if len(req.FileIDs) > 0 {
			fileContents := []string{}
			for _, fileID := range req.FileIDs {
				// 获取文件内容
				file, err := services.FileService.GetDetail(fileID, userId)
				if err != nil {
					logs.Error("获取文件内容失败: %v", err)
					continue
				}
				content := fmt.Sprintf("```\n# %s\n%s\n```", file.FileName, file.Content)
				fileContents = append(fileContents, content)
			}

			// 将文件内容添加到提示中
			if len(fileContents) > 0 {
				agentReq.Prompt = agentReq.Prompt + "\n\n以下是用户上传的文件内容:\n\n---\n\n" + strings.Join(fileContents, "\n\n---\n\n")
			}
		}

		// 如果有报告ID，处理报告内容
		if len(req.ReportIDs) > 0 {
			reportContents := []string{}
			for _, reportId := range req.ReportIDs {
				// 获取文件内容
				report, err := services.Report.GetDetail(reportId, userId)
				if err != nil {
					logs.Error("获取文件内容失败: %v", err)
					continue
				}
				content := fmt.Sprintf("```\n# %s\n%s\n```", report.ReportFileName, report.Content)
				reportContents = append(reportContents, content)
			}

			// 将报告内容添加到提示中
			if len(reportContents) > 0 {
				agentReq.Prompt = req.Prompt + "\n\n以下是用户的诊断报告的内容:\n\n---\n\n" + strings.Join(reportContents, "\n\n---\n\n")
			}
		}

		services.ChatMessage.UpdateRawPrompt(messageID, agentReq.Prompt, userId)

		// 启动异步协程处理百炼API请求
		go func() {
			// 创建一个自定义的ResponseWriter，用于将百炼API的响应写入Redis
			redisWriter := services.NewRedisResponseWriter(messageID)

			// 调用百炼服务获取响应
			res, err := services.BailianService.StreamChat(agentReq, userId, redisWriter)

			if err != nil {
				// 发生错误时，将错误信息写入Redis
				errorMsg := map[string]interface{}{
					"type":    "error",
					"content": "对话处理失败: " + err.Error(),
				}
				jsonData, _ := json.Marshal(errorMsg)
				services.GetRedisClient().RPush(redisKey, string(jsonData))
				logs.Error("百炼API请求失败: %v\n\njsonData: %s", err, string(jsonData))
				return
			} else {
				logs.Debug("百炼API请求成功，响应: %s", res.Content)
			}

			// 请求完成后，更新消息内容和使用统计
			services.ChatMessage.UpdateCompletion(messageID, res.Content, res.ReasoningContent, userId)

			// 更新使用统计
			if res.Stats.Model != "" {
				services.Chat.UpdateUsageStats(req.ChatID, &res.Stats, userId)
			}

			// 检查是否已停止
			if services.IsStreamStopped(messageID) {
				logs.Info("对话已被用户停止，消息ID: %s", messageID)
				return
			}

			logs.Info("百炼API请求完成，消息ID: %s", messageID)
		}()
	}

	// 获取消息详情
	messageDetail, err := services.ChatMessage.GetDetail(messageID, userId)
	if err != nil {
		logs.Error("获取消息详情失败: %v", err)
	}

	// 返回创建成功的响应
	c.JSON(common.JsonRes{
		Code: 0,
		Msg:  "创建成功",
		Data: map[string]interface{}{
			"chat_id":    req.ChatID,
			"message_id": messageID,
			"message":    messageDetail, // 添加消息详情
		},
	})
}

// 流式对话读取响应
func (c *ChatApiController) ReadStream() {
	var req dto.StreamReadReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, "请求参数错误"+err.Error())
		return
	}

	// 参数校验
	v := validate.Struct(req)
	if !v.Validate() {
		c.ErrorJson(400, v.Errors.One())
		return
	}

	userId := c.GetUserId()

	// 验证对话权限
	chat := &models.Chat{ChatID: req.ChatID}
	if err := chat.Get(); err != nil {
		c.ErrorJson(400, "对话不存在")
		return
	}
	if chat.UserId != userId {
		logs.Error("400 无权操作此对话 chat.UserId=", chat.UserId, " authUserId=", userId)
		c.ErrorJson(400, "无权操作此对话")
		return
	}

	// 验证消息是否存在且属于当前对话
	message := &models.ChatMessage{MessageID: req.MessageID}
	if err := message.Get(); err != nil {
		logs.Error("获取消息失败: %v", err)
		c.ErrorJson(400, "消息不存在")
		return
	}

	if message.ChatID != req.ChatID {
		logs.Error("消息不属于当前对话")
		c.ErrorJson(400, "消息不属于当前对话")
		return
	}

	// 设置响应头，支持流式输出
	c.Ctx.ResponseWriter.Header().Set("Content-Type", "text/event-stream")
	c.Ctx.ResponseWriter.Header().Set("Cache-Control", "no-cache")
	c.Ctx.ResponseWriter.Header().Set("Connection", "keep-alive")
	c.Ctx.ResponseWriter.Header().Set("Transfer-Encoding", "chunked")

	// 首先返回初始响应
	sessionJSON, _ := json.Marshal(map[string]string{"type": "start", "chat_id": req.ChatID, "message_id": req.MessageID})
	fmt.Fprintf(c.Ctx.ResponseWriter, "data: %s\n\n", sessionJSON)
	c.Ctx.ResponseWriter.Flush()

	// 获取Redis键名，用于读取流式响应
	redisKey, _ := services.GenerateChatRedisKeys(req.MessageID)

	// 从Redis中读取流式响应并输出到客户端
	timeout := time.After(1800 * time.Second)        // 设置超时时间
	ticker := time.NewTicker(100 * time.Millisecond) // 轮询间隔
	lastIndex := int64(0)                            // 上次读取的索引位置
	isEnded := false

	for {
		select {
		case <-ticker.C:
			// 从Redis中获取新的响应
			responses, err := services.GetRedisClient().LRange(redisKey, int64(lastIndex), -1)
			if err != nil {
				logs.Error("从Redis获取响应失败: %v", err)
				continue
			}

			if len(responses) > 0 {
				// 输出新的响应
				for _, resp := range responses {
					fmt.Fprintf(c.Ctx.ResponseWriter, "data: %s\n\n", resp)
					c.Ctx.ResponseWriter.Flush()

					// 检查是否是结束标记或停止标记
					var respObj map[string]interface{}
					if err := json.Unmarshal([]byte(resp), &respObj); err == nil {
						if respType, ok := respObj["type"].(string); ok && (respType == "end" || respType == "stop") {
							isEnded = true
						}
					}
				}

				// 更新索引
				lastIndex += int64(len(responses))

				// 如果收到结束标记或停止标记，退出循环
				if isEnded {
					ticker.Stop()
					return
				}
			} else if message.Completion != "" {
				reasoningEvent := map[string]string{
					"type":    "reasoning",
					"content": message.Reasoning,
				}
				reasoningJSON, _ := json.Marshal(reasoningEvent)
				fmt.Fprintf(c.Ctx.ResponseWriter, "data: %s\n\n", reasoningJSON)
				answerEvent := map[string]string{
					"type":    "answer",
					"content": message.Completion,
				}
				answerJSON, _ := json.Marshal(answerEvent)
				fmt.Fprintf(c.Ctx.ResponseWriter, "data: %s\n\n", answerJSON)
				endEvent := map[string]string{
					"type":    "end",
					"content": "",
				}
				endJSON, _ := json.Marshal(endEvent)
				fmt.Fprintf(c.Ctx.ResponseWriter, "data: %s\n\n", endJSON)
				c.Ctx.ResponseWriter.Flush()
				ticker.Stop()
				return
			}

		case <-timeout:
			// 超时处理
			logs.Error("流式响应超时")
			timeoutMsg := map[string]interface{}{
				"type":    "error",
				"content": "响应超时",
			}
			jsonData, _ := json.Marshal(timeoutMsg)
			fmt.Fprintf(c.Ctx.ResponseWriter, "%s\n\n", jsonData)
			c.Ctx.ResponseWriter.Flush()
			ticker.Stop()
			return
		}
	}
}

// 停止对话
func (c *ChatApiController) StopStream() {
	var req dto.StopStreamReq
	if err := c.ParseJSON(&req); err != nil {
		c.ErrorJson(400, "请求参数错误"+err.Error())
		return
	}

	// 参数校验
	v := validate.Struct(req)
	if !v.Validate() {
		c.ErrorJson(400, v.Errors.One())
		return
	}

	userId := c.GetUserId()

	// 验证对话权限
	chat := &models.Chat{ChatID: req.ChatID}
	if err := chat.Get(); err != nil {
		c.ErrorJson(400, "对话不存在")
		return
	}
	if chat.UserId != userId {
		c.ErrorJson(400, "无权操作此对话")
		return
	}

	// 验证消息是否存在且属于当前对话
	message := &models.ChatMessage{MessageID: req.MessageID}
	if err := message.Get(); err != nil {
		c.ErrorJson(400, "消息不存在")
		return
	}
	if message.ChatID != req.ChatID {
		c.ErrorJson(400, "消息不属于当前对话")
		return
	}

	// 创建Redis键名，用于存储流式响应和停止标记
	redisKey, stopKey := services.GenerateChatRedisKeys(req.MessageID)

	// 设置停止标记
	err := services.GetRedisClient().HSet(stopKey, "stopped", "true")
	if err != nil {
		logs.Error("设置停止标记失败: %v", err)
		c.ErrorJson(500, "设置停止标记失败")
		return
	}

	// 添加停止消息到流中
	stopEvent := map[string]interface{}{
		"type":    "stop",
		"content": "对话已停止",
	}
	stopJSON, _ := json.Marshal(stopEvent)
	services.GetRedisClient().RPush(redisKey, string(stopJSON))

	// 获取当前已生成的内容
	responses, err := services.GetRedisClient().LRange(redisKey, 0, -1)
	if err != nil {
		logs.Error("获取已生成内容失败: %v", err)
	}

	// 从响应中提取文本内容
	var fullText string
	var reasoning string
	for _, resp := range responses {
		var respObj map[string]interface{}
		if err := json.Unmarshal([]byte(resp), &respObj); err == nil {
			if contentType, ok := respObj["type"].(string); ok {
				if content, ok := respObj["content"].(string); ok {
					if contentType == "reasoning" {
						reasoning += content
					} else if contentType == "answer" {
						fullText += content
					}

				}
			}

		}
	}

	// 更新消息内容到数据库
	if fullText != "" {
		_, err = message.UpdateCompletion(fullText, reasoning)
		if err != nil {
			logs.Error("更新消息内容失败: %v", err)
		}
	}

	c.JSON(common.JsonRes{
		Code: 0,
		Msg:  "已停止对话",
	})
}
