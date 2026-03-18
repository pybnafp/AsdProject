package middleware

import (
	"strings"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/adapter/context"
)

func AddDataFitler() {
	beego.InsertFilter("/*", beego.BeforeRouter, DataPermissionFilter)

}

// 数据权限过滤器
var DataPermissionFilter = func(ctx *context.Context) {
	// 获取请求的路径
	path := ctx.Request.URL.Path

	// 如果路径是 /openapi/ 开头的，直接返回，不执行过滤器逻辑
	if strings.HasPrefix(path, "/openapi/") {
		return
	}

	// 假设通过 Session 获取当前登录用户的 ID
	userId, ok := ctx.Input.Session("userId").(int)
	if ok {
		// 将当前用户的 ID 注入到上下文中，供后续控制器使用
		ctx.Input.SetData("userId", userId)
	}
}
