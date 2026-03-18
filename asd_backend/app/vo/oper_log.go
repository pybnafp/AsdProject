/**
 * 操作日志记录Vo
 * @author Evotrek 研发团队
 * @since 2024-11-10
 * @File : oper_log
 */
package vo

import "asd/app/models"

// 操作日志记录信息Vo
type OperLogInfoVo struct {
	models.OperLog

	OperTypeName int `json:"operTypeName"` // 操作类型名称
	StatusName   int `json:"statusName"`   // 日志状态名称
}
