/**
 * 操作日志中间件
 * @author Evotrek 研发团队
 * @since 2024/9/24
 */
package middleware

import (
	"asd/app/dto"
	"asd/app/services"
	"asd/conf"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/adapter/context"
	"github.com/beego/beego/v2/core/logs"
)

// 操作日志中间件
func OperLog() {
	var operlog = func(ctx *context.Context) {
		logs.Info("--------add operlog form" + ctx.Request.RequestURI)
		// 不记录静态资源
		if strings.Contains(ctx.Request.URL.Path, "resource") ||
			strings.Contains(ctx.Request.URL.Path, "captcha") {
			return
		}
		// 过滤掉非修改操作的请求
		if ctx.Request.Method == http.MethodGet {
			logs.Info("skip from get method")
			return
		}

		// 过滤掉一些不需要记录的路径
		excludePaths := []string{
			"/login",
			"/logout",
			"/index",
			"/main",
			"/userInfo",
		}
		for _, path := range excludePaths {
			if strings.HasPrefix(ctx.Request.URL.Path, path) {
				logs.Info("skip from excludpath" + path)
				return
			}
		}

		// Check if path contains allowed operations
		path := strings.ToLower(ctx.Request.URL.Path)
		allowedOps := []string{"/update", "/delete", "/transfer", "/add"}
		shouldLog := false
		for _, op := range allowedOps {
			if strings.Contains(path, op) {
				shouldLog = true
				break
			}
		}

		//shouldLog = true
		if !shouldLog {
			logs.Info("skip - not an allowed operation path")
			return
		}

		logs.Info("add operlog form" + ctx.Request.RequestURI)

		//url, _ := json.Marshal(ctx.Input.Data()["RouterPattern"])
		params, _ := json.Marshal(ctx.Request.Form)
		outputBytes, _ := json.Marshal(ctx.Input.Data()["json"])

		// Get response status code from context

		// 获取当前登录用户
		userId := ctx.Input.Session(conf.USER_ID).(int)

		// 操作日志实体
		var operLog dto.OperLogAddReq
		operLog.UserId = userId
		// Get first segment of URL path as model
		pathParts := strings.Split(strings.Trim(ctx.Request.URL.Path, "/"), "/")
		if len(pathParts) > 0 {
			operLog.Model = pathParts[0]
		}
		operLog.OperMethod = ctx.Request.Method
		operLog.OperUrl = ctx.Request.URL.Path
		operLog.OperIp = ctx.Input.IP()

		// 请求参数
		operLog.RequestParam = string(params)
		// 响应结果
		operLog.Result = string(outputBytes)

		statusCode := ctx.Output.Status

		// Define operation type mapping
		operTypeMap := map[string]int{
			"/add":    1, // New addition
			"/update": 2, // Modification
			"/delete": 3, // Deletion
		}

		// Set OperType based on URL path
		operLog.OperType = 0 // Default to "other"
		for pathKey, typeValue := range operTypeMap {
			if strings.Contains(path, pathKey) {
				operLog.OperType = typeValue
				break
			}
		}

		fmt.Println(statusCode)

		fmt.Println(ctx.Output.IsSuccessful())

		operLog.Status = 1

		operLog.UserAgent = ctx.Request.UserAgent()

		// 调用服务创建操作日志
		services.OperLog.Add(operLog, 0)
	}

	beego.InsertFilter("*", beego.AfterExec, operlog, false)
}

// 自定义ResponseWriter
type responseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = b
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
