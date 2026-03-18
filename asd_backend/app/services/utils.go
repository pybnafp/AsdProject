package services

import "fmt"

// GenerateChatRedisKeys 生成对话相关的Redis键名
func GenerateChatRedisKeys(messageID string) (redisKey string, stopKey string) {
	redisKey = fmt.Sprintf("chat:stream:%s", messageID)
	stopKey = fmt.Sprintf("chat:stream:stop:%s", messageID)
	return
}

// IsStreamStopped 检查对话是否已停止
func IsStreamStopped(messageID string) bool {
	_, stopKey := GenerateChatRedisKeys(messageID)
	stopped, _ := GetRedisClient().HGet(stopKey, "stopped")
	return stopped == "true"
}
