package services

import (
	"fmt"
	"net/http"
)

// RedisResponseWriter 是一个自定义的ResponseWriter，用于将响应写入Redis
type RedisResponseWriter struct {
	messageID string
	redisKey  string
}

// NewRedisResponseWriter 创建一个新的RedisResponseWriter
func NewRedisResponseWriter(messageID string) *RedisResponseWriter {
	return &RedisResponseWriter{
		messageID: messageID,
		redisKey:  fmt.Sprintf("chat:stream:%s", messageID),
	}
}

// GetMessageID 获取消息ID
func (w *RedisResponseWriter) GetMessageID() string {
	return w.messageID
}

// Header 实现http.ResponseWriter接口
func (w *RedisResponseWriter) Header() http.Header {
	return http.Header{}
}

// Write 实现http.ResponseWriter接口，将数据写入Redis
func (w *RedisResponseWriter) Write(data []byte) (int, error) {
	// 将数据写入Redis
	redisClient := GetRedisClient()
	err := redisClient.RPush(w.redisKey, string(data))
	if err != nil {
		return 0, err
	}

	// 设置过期时间，避免Redis中存储过多数据
	redisClient.Expire(w.redisKey, 3600) // 1小时后过期
	return len(data), nil
}

// WriteHeader 实现http.ResponseWriter接口
func (w *RedisResponseWriter) WriteHeader(statusCode int) {
	// 不需要实现
}

// Flush 实现http.Flusher接口
func (w *RedisResponseWriter) Flush() {
	// Redis操作是立即执行的，不需要额外的刷新操作
}
