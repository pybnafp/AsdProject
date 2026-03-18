package routers

import (
	"asd/app/middleware"
	"fmt"
)

func init() {

	// 添加操作日志

	middleware.AddDataFitler()

	fmt.Println("api 模块路由初始化")
	// API 验证中间件
	middleware.CheckApi()

	// 初始化自定义的模板函数
	middleware.LoadTemplateFunc()
}
