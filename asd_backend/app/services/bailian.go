package services

import (
	"asd/app/dto"
	"asd/conf"
	"asd/utils/gstr"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	bailian20231229 "github.com/alibabacloud-go/bailian-20231229/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/beego/beego/v2/core/logs"
)

var BailianService = new(bailianService)

type bailianService struct{}

// NewbailianService 创建百炼服务实例
func NewbailianService() *bailianService {
	return &bailianService{}
}

// CreateMemory 创建记忆体
func (s *bailianService) CreateMemory() (string, error) {
	client, err := s.createClient()
	if err != nil {
		return "", err
	}

	createMemoryRequest := &bailian20231229.CreateMemoryRequest{}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)

	// 发送请求
	resp, err := client.CreateMemoryWithOptions(tea.String(conf.CONFIG.ApiConfig.AlibabaBailianWorkspaceId), createMemoryRequest, headers, runtime)
	if err != nil {
		return "", err
	}

	// 从响应中获取记忆体ID
	if resp != nil && resp.Body != nil {
		// 直接使用resp.Body，它已经是正确的类型，不需要使用tea.StringValue
		bodyBytes, err := json.Marshal(resp.Body)
		if err != nil {
			return "", err
		}

		bodyMap := make(map[string]interface{})
		err = json.Unmarshal(bodyBytes, &bodyMap)
		if err != nil {
			return "", err
		}

		if memoryID, ok := bodyMap["memoryId"].(string); ok {
			fmt.Printf("成功创建记忆体，ID: %s\n", memoryID)
			return memoryID, nil
		}
	}

	return "", fmt.Errorf("创建记忆体失败：无法从响应中提取记忆体ID")
}

// 初始化客户端
func (s *bailianService) createClient() (*bailian20231229.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(conf.CONFIG.ApiConfig.AlibabaCloudAccessKeyId),
		AccessKeySecret: tea.String(conf.CONFIG.ApiConfig.AlibabaCloudAccessKeySecret),
	}
	config.Endpoint = tea.String("bailian.cn-beijing.aliyuncs.com")
	return bailian20231229.NewClient(config)
}

type StreamChatReturn struct {
	Content          string
	ReasoningContent string
	Stats            dto.UsageStats
}

// StreamChat 流式对话
func (s *bailianService) StreamChat(req dto.AgentChatReq, userId int, w http.ResponseWriter) (*StreamChatReturn, error) {
	startTime := time.Now()

	// 初始返回值
	retval := &StreamChatReturn{}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 确保函数退出时取消上下文

	// 获取消息ID
	var messageID string
	if customWriter, ok := w.(*RedisResponseWriter); ok {
		messageID = customWriter.GetMessageID()
	}

	// 创建一个通道用于通知主goroutine请求已被取消
	stopChan := make(chan struct{})

	// 创建停止检查的goroutine
	if messageID != "" {
		go func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if IsStreamStopped(messageID) {
						logs.Info("请求已被取消，停止百炼API请求:" + messageID)
						cancel()               // 取消请求
						stopChan <- struct{}{} // 通知主goroutine请求已被取消
					}
				case <-ctx.Done():
					logs.Info("请求上下文 Done:" + messageID)
					return
				}
			}
		}()
	}

	// ===========================
	// 新实现：调用 asd_algorithm 中的 DashScope 服务替代 OpenRouter
	// ===========================

	// 在发送请求前先检查是否已取消
	select {
	case <-stopChan:
		// 请求已被取消，直接发送停止消息并返回
		sendStopMessage(w)
		return retval, nil
	default:
		// 继续处理
	}

	// 将历史对话转换为 Python 侧 /api/chat 使用的 history 结构：
	// [{"prompt": "...", "completion": "..."}, ...]
	type historyItem struct {
		Prompt     string `json:"prompt"`
		Completion string `json:"completion"`
	}
	var history []historyItem
	for i := 0; i+1 < len(req.Messages); i += 2 {
		userMsg := req.Messages[i]
		assistantMsg := req.Messages[i+1]
		if userMsg["role"] == "user" && assistantMsg["role"] == "assistant" {
			history = append(history, historyItem{
				Prompt:     userMsg["content"],
				Completion: assistantMsg["content"],
			})
		}
	}
	historyJSON, _ := json.Marshal(history)

	// 组装发往 asd_algorithm 的请求体
	payload := map[string]interface{}{
		"content": req.Prompt,
		"history": string(historyJSON),
	}
	body, _ := json.Marshal(payload)

	// TODO: 如需从配置中读取地址，可替换为 conf.CONFIG.ApiConfig.AlgorithmBaseUrl
	algorithmURL := conf.CONFIG.ApiConfig.AlgorithmBaseUrl //"http://localhost:5003/api/chat"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, algorithmURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("dashscope api error: %s, body=%s", resp.Status, string(respBody))
	}

	// 对应 asd_algorithm/main.py 中 /api/chat 的返回结构
	var dsResp struct {
		Success bool `json:"success"`
		Result  struct {
			Intention string `json:"intention"`
			Answer    string `json:"answer"`
		} `json:"result"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dsResp); err != nil {
		return nil, err
	}
	if !dsResp.Success {
		return nil, fmt.Errorf("dashscope failed: %s", dsResp.Error)
	}

	answer := gstr.FixPubMedLinks(dsResp.Result.Answer)
	retval.Content = answer
	retval.ReasoningContent = ""

	// 将结果写入流（通过 RedisResponseWriter）以兼容现有前端 SSE 协议
	if w != nil {
		// 一次性发送 answer 事件
		answerEvent := map[string]string{
			"type":    "answer",
			"content": answer,
		}
		answerJSON, _ := json.Marshal(answerEvent)
		fmt.Fprintf(w, "%s\n\n", answerJSON)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		// 然后发送 end 事件
		endEvent := map[string]any{
			"type":    "end",
			"content": answer,
		}
		endJSON, _ := json.Marshal(endEvent)
		fmt.Fprintf(w, "%s\n\n", endJSON)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	// ===========================
	// 旧实现：基于 OpenRouter 的流式对话（已废弃，仅保留注释以供参考）
	// ===========================
	/*
			config := openai.DefaultConfig(conf.CONFIG.ApiConfig.OpenRouterApiKey)
			config.BaseURL = conf.CONFIG.ApiConfig.OpenRouterBaseUrl
			chatClient := openai.NewClientWithConfig(config)
			chatReq := openai.ChatCompletionRequest{
				Model:       "qwen/qwen3.5-35b-a3b",
				MaxTokens:   128000,
				Stream:      true,
				Temperature: 0.6,
			}

			if conf.CONFIG.ApiConfig.UseRag {
				ragMsgs, err := RagService.GetMessagesFromRag(ctx, req.Prompt, req.Messages, req.EnableGuideLines, req.EnableResearches)
				if err != nil {
					return nil, err
				}
				chatReq.Messages = append(chatReq.Messages, ragMsgs...)
			} else {
				for _, msg := range req.Messages {
					chatReq.Messages = append(chatReq.Messages, openai.ChatCompletionMessage{
						Role:    msg["role"],
						Content: msg["content"],
					})
				}
				chatReq.Messages = append(chatReq.Messages, openai.ChatCompletionMessage{
					Role:    "user",
					Content: req.Prompt,
				})
			}

			// 在发送Stream请求前先检查是否已取消
			select {
			case <-stopChan:
				// 请求已被取消，直接发送停止消息并返回
				sendStopMessage(w)
				// 此时 responseText 还未定义，应该返回空字符串
				return retval, nil
			default:
				// 继续处理
			}

			stream, err := chatClient.CreateChatCompletionStream(ctx, chatReq)
			if err != nil {
				return nil, err
			}
			defer stream.Close()

			// 异步发送请求
			respChan := make(chan *openai.ChatCompletionStreamResponse)
			errChan := make(chan error)

			// 处理流式响应
			var responseText strings.Builder
			var reasoningText strings.Builder

			// 如果提供了ResponseWriter，设置SSE头部
			if w != nil {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.Header().Set("Access-Control-Allow-Origin", "*")
				_, ok := w.(http.Flusher)
				if !ok {
					return nil, fmt.Errorf("流式输出不支持")
				}
			}

		stream_chat:
			for {
				go func() {
					response, err := stream.Recv()
					if err != nil {
						errChan <- err
					} else {
						respChan <- &response
					}
				}()

				select {
				case chunk := <-respChan:
					if chunk.Usage != nil {
						retval.Stats.Model = chunk.Model
						retval.Stats.CompletionTokens += chunk.Usage.CompletionTokens
						retval.Stats.PromptTokens += chunk.Usage.PromptTokens
						retval.Stats.TotalTokens += chunk.Usage.TotalTokens
						if chunk.Usage.CompletionTokensDetails != nil {
							retval.Stats.ReasoningTokens += chunk.Usage.CompletionTokensDetails.ReasoningTokens
						}
					}
					if len(chunk.Choices) == 0 {
						continue stream_chat
					}

					delta := chunk.Choices[0].Delta

					event := map[string]string{}
					if delta.ReasoningContent != "" {
						event["type"] = "reasoning"
						event["content"] = delta.ReasoningContent
						reasoningText.WriteString(delta.ReasoningContent)
					} else if delta.Content != "" {
						event["type"] = "answer"
						event["content"] = delta.Content
						responseText.WriteString(delta.Content)
					}
					// 向HTTP客户端发送内容
					if w != nil && len(event) > 0 {
						eventJSON, _ := json.Marshal(event)
						fmt.Fprintf(w, "%s\n\n", eventJSON)
						flusher, _ := w.(http.Flusher)
						flusher.Flush()
					}
				case err = <-errChan:
					if errors.Is(err, io.EOF) {
						break stream_chat
					} else if ctx.Err() != nil {
						// 如果是因为取消导致的错误，发送停止消息
						sendStopMessage(w)
						retval.Content = responseText.String()
						retval.ReasoningContent = reasoningText.String()
						return retval, nil
					} else {
						retval.Content = responseText.String()
						retval.ReasoningContent = reasoningText.String()
						return retval, err
					}
				case <-stopChan:
					// 请求被取消
					logs.Info("百炼API请求被取消:" + messageID)
					sendStopMessage(w)
					retval.Content = responseText.String()
					retval.ReasoningContent = reasoningText.String()
					return retval, nil
				}
			}
			retval.Content = gstr.FixPubMedLinks(responseText.String())
			retval.ReasoningContent = reasoningText.String()
	*/

	duration := time.Since(startTime)
	logs.Debug("\n\n[对话耗时: %.2f 秒]\n", duration.Seconds())

	// 发送结束标记
	if w != nil {
		endEvent := map[string]any{
			"type":    "end",
			"content": retval.Content,
			"usage": map[string]int{
				"prompt_tokens":     retval.Stats.PromptTokens,
				"completion_tokens": retval.Stats.CompletionTokens,
				"total_tokens":      retval.Stats.TotalTokens,
			},
		}
		endJSON, _ := json.Marshal(endEvent)
		fmt.Fprintf(w, "%s\n\n", endJSON)
		flusher, _ := w.(http.Flusher)
		flusher.Flush()
	}

	return retval, nil
}

// 发送停止消息到客户端
func sendStopMessage(w http.ResponseWriter) {
	if w == nil {
		return
	}

	stopEvent := map[string]interface{}{
		"type":    "stop",
		"content": "对话已被用户停止111",
	}
	stopJSON, _ := json.Marshal(stopEvent)
	fmt.Fprintf(w, "%s\n\n", stopJSON)

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}
