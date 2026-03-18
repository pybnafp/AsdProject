package utils

import "fmt"

// 任务相关的 Redis key
func GetTaskKey(taskID string) string {
	return fmt.Sprintf("task:%s", taskID)
}

// 获取任务状态的 Redis key
func GetTaskStatusKey(taskID string) string {
	return fmt.Sprintf("task:status:%s", taskID)
}

// 获取任务日志的 Redis key
func GetTaskLogKey(taskID string) string {
	return fmt.Sprintf("task:log:%s", taskID)
}

// GetTaskCOSPathsKey 获取任务 COS 路径的 Redis 键名
func GetTaskCOSPathsKey(taskID string) string {
	return fmt.Sprintf("task:cospaths:%s", taskID)
}

// GetTaskResultsKey 获取任务结果的 Redis 键名
func GetTaskResultsKey(taskID string) string {
	return fmt.Sprintf("task:results:%s", taskID)
}
