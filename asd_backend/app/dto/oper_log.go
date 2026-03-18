/**
 * 操作日志记录Dto
 * @author Evotrek 研发团队
 * @since 2024-11-10
 * @File : oper_log
 */
package dto

import "github.com/gookit/validate"

// 分页查询
type OperLogPageReq struct {
	Page  int `form:"page"`  // 页码
	Limit int `form:"limit"` // 每页数

	OperType int `form:"operType"` // 操作类型：0其它 1新增 2修改 3删除 4查询 5设置状态 6导入 7导出 8设置权限 9设置密码

	Status int `form:"status"` // 日志状态：0正常日志 1错误日志

}

// 添加操作日志记录
type OperLogAddReq struct {
	UserId       int    `form:"userId"`                           //  用户ID
	Model        string `form:"model" validate:"required"`        // 操作模块
	OperType     int    `form:"operType" validate:"int"`          // 操作类型：0其它 1新增 2修改 3删除 4查询 5设置状态 6导入 7导出 8设置权限 9设置密码
	OperMethod   string `form:"operMethod" validate:"required"`   // 操作方法
	Username     string `form:"username" validate:"required"`     // 操作账号
	OperName     string `form:"operName" validate:"required"`     // 操作用户
	OperUrl      string `form:"operUrl" validate:"required"`      // 请求URL
	OperIp       string `form:"operIp" validate:"required"`       // 主机地址
	OperLocation string `form:"operLocation" validate:"required"` // 操作地点
	RequestParam string `form:"requestParam" validate:"required"` // 请求参数
	Result       string `form:"result" validate:"required"`       // 返回参数
	Status       int    `form:"status" validate:"int"`            // 日志状态：0正常日志 1错误日志
	UserAgent    string `form:"userAgent" validate:"required"`    // 代理信息
	Note         string `form:"note" validate:"required"`         // 备注
}

// 添加表单验证
func (v OperLogAddReq) Messages() map[string]string {
	return validate.MS{

		"Model.required": "操作模块不能为空.", // 操作模块

		"OperType.int": "请选择操作类型.", // 操作类型：0其它 1新增 2修改 3删除 4查询 5设置状态 6导入 7导出 8设置权限 9设置密码

		"OperMethod.required": "操作方法不能为空.", // 操作方法

		"Username.required": "操作账号不能为空.", // 操作账号

		"OperName.required": "操作用户不能为空.", // 操作用户

		"OperUrl.required": "请求URL不能为空.", // 请求URL

		"OperIp.required": "主机地址不能为空.", // 主机地址

		"OperLocation.required": "操作地点不能为空.", // 操作地点

		"RequestParam.required": "请求参数不能为空.", // 请求参数

		"Result.required": "返回参数不能为空.", // 返回参数

		"Status.int": "请选择日志状态.", // 日志状态：0正常日志 1错误日志

		"UserAgent.required": "代理信息不能为空.", // 代理信息

		"Note.required": "备注不能为空.", // 备注

	}
}

// 编辑操作日志记录
type OperLogUpdateReq struct {
	Id int `form:"id" validate:"int"`

	Model string `form:"model" validate:"required"` // 操作模块

	OperType int `form:"operType" validate:"int"` // 操作类型：0其它 1新增 2修改 3删除 4查询 5设置状态 6导入 7导出 8设置权限 9设置密码

	OperMethod string `form:"operMethod" validate:"required"` // 操作方法

	Username string `form:"username" validate:"required"` // 操作账号

	OperName string `form:"operName" validate:"required"` // 操作用户

	OperUrl string `form:"operUrl" validate:"required"` // 请求URL

	OperIp string `form:"operIp" validate:"required"` // 主机地址

	OperLocation string `form:"operLocation" validate:"required"` // 操作地点

	RequestParam string `form:"requestParam" validate:"required"` // 请求参数

	Result string `form:"result" validate:"required"` // 返回参数

	Status int `form:"status" validate:"int"` // 日志状态：0正常日志 1错误日志

	UserAgent string `form:"userAgent" validate:"required"` // 代理信息

	Note string `form:"note" validate:"required"` // 备注

}

// 更新表单验证
func (v OperLogUpdateReq) Messages() map[string]string {
	return validate.MS{
		"Id.int": "记录ID不能为空.",

		"Model.required": "操作模块不能为空.", // 操作模块

		"OperType.int": "请选择操作类型.", // 操作类型：0其它 1新增 2修改 3删除 4查询 5设置状态 6导入 7导出 8设置权限 9设置密码

		"OperMethod.required": "操作方法不能为空.", // 操作方法

		"Username.required": "操作账号不能为空.", // 操作账号

		"OperName.required": "操作用户不能为空.", // 操作用户

		"OperUrl.required": "请求URL不能为空.", // 请求URL

		"OperIp.required": "主机地址不能为空.", // 主机地址

		"OperLocation.required": "操作地点不能为空.", // 操作地点

		"RequestParam.required": "请求参数不能为空.", // 请求参数

		"Result.required": "返回参数不能为空.", // 返回参数

		"Status.int": "请选择日志状态.", // 日志状态：0正常日志 1错误日志

		"UserAgent.required": "代理信息不能为空.", // 代理信息

		"Note.required": "备注不能为空.", // 备注

	}
}

// 设置日志状态
type OperLogStatusReq struct {
	Id     int `form:"id" validate:"int"`
	Status int `form:"status" validate:"int"`
}

// 设置状态参数验证
func (v OperLogStatusReq) Messages() map[string]string {
	return validate.MS{
		"Id.int":     "记录ID不能为空.",
		"Status.int": "请选择日志状态：0正常日志 1错误日志.",
	}
}
