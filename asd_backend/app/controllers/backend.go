package controllers

import (
	"asd/utils/common"
	"encoding/json"

	beego "github.com/beego/beego/v2/adapter"
)

// 基类结构体
type BaseController struct {
	beego.Controller
}

type JsonResult struct {
	Code  int         `json:"code"`  // 响应编码：0成功 401请登录 403无权限 500错误
	Msg   string      `json:"msg"`   // 消息提示语
	Data  interface{} `json:"data"`  // 数据对象
	Count int64       `json:"count"` // 记录总数
}

func (ctrl *BaseController) JSON(obj interface{}) {
	ctrl.Data["json"] = obj
	//对json进行序列化输出
	ctrl.ServeJSON()
	ctrl.Ctx.Output.SetStatus(200)
	ctrl.StopRun()
	//ctrl.Render()
}

func (ctrl *BaseController) ErrorJson(code int, msg string) {
	o := common.JsonResult{Code: code, Msg: msg}

	ctrl.Data["json"] = o
	//对json进行序列化输出
	ctrl.ServeJSON()
	ctrl.Ctx.Output.SetStatus(200)
	ctrl.StopRun()
	//ctrl.Render()
}

func (ctrl *BaseController) ServerErrorJson(msg string) {
	ctrl.ErrorJson(500, msg)
}

func (ctrl *BaseController) ParamsErrorJson(msg string) {
	ctrl.ErrorJson(400, msg)
}

func (c *BaseController) ParseJSON(v interface{}) error {
	// 检查请求体是否为空
	if c.Ctx.Request.ContentLength == 0 {
		// 如果请求体为空，则创建一个空的JSON对象 "{}"
		emptyJSON := []byte("{}")
		return json.Unmarshal(emptyJSON, v)
	}

	// 正常解析JSON
	return json.NewDecoder(c.Ctx.Request.Body).Decode(v)
}

func (ctrl *BaseController) GetUserId() int {
	// 从上下文中获取 "user_id"
	userID := ctrl.Ctx.Input.GetData("user_id")

	// 检查是否为 nil，避免空指针异常
	if userID == nil {
		return 0 // 如果 user_id 不存在，返回默认值 0
	}

	// 类型断言，将 interface{} 转为 int
	if id, ok := userID.(int); ok {
		return id
	}

	// 如果类型不匹配，返回默认值 0
	return 0
}

//func (ctl *BaseController) Html(params ...string) {
//	ctl.Ctx.WriteString(params[0])
//}
//
//func (ctl *BaseController) Render(params ...string) {
//}
