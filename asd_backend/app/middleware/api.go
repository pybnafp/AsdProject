/**
 * 登录验证中间件
 * @author Evotrek 研发团队
 * @since 2024/9/24
 * @File : checkauth
 */
package middleware

import (
	"asd/app/models"
	"asd/conf"
	"asd/utils/common"
	"asd/utils/gconv"
	"fmt"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/adapter/context"
)

func CheckApi() {

	//登录认证中间件过滤器
	var login = func(ctx *context.Context) {
		// 添加路径白名单
		if strings.HasPrefix(ctx.Request.URL.Path, "/api/login") {
			return
		}

		if strings.HasPrefix(ctx.Request.URL.Path, "/api/admin") {
			// 从请求头中获取 Authorization
			authHeader := ctx.Input.Header("Authorization")
			if authHeader != "Bearer fTrytyIiFuPqDXYWFc4GwFbu2nH7Wh9K" {
				ctx.Output.SetStatus(401)
				ctx.Output.Body([]byte("Unauthorized: Missing Authorization Header"))
				return
			} else {
				return
			}
		}

		// 设置
		if strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			if !IsLogin(ctx) {
				ctx.Output.SetStatus(401)
				ctx.Output.Body([]byte("Unauthorized: Invalid login status"))
				return
			}
		}
	}

	// 登录过滤器
	beego.InsertFilter("/api/*", beego.BeforeRouter, login)
}

func IsLogin(ctx *context.Context) bool {
	userId := ctx.Input.Session(conf.USER_ID)
	// 将 user_id 存储到上下文中，供后续使用
	ctx.Input.SetData("user_id", gconv.Int(userId))
	return userId != nil
}

func CheckApiAuth(ctx *context.Context) bool {
	// 从请求头中获取 Authorization
	authHeader := ctx.Input.Header("Authorization")
	if authHeader == "" {
		ctx.Output.SetStatus(401)
		ctx.Output.Body([]byte("Unauthorized: Missing Authorization Header"))
		return false
	}

	// 检查 Bearer 格式
	if !strings.HasPrefix(authHeader, "Bearer ") {
		ctx.Output.SetStatus(401)
		ctx.Output.Body([]byte("Unauthorized: Invalid Authorization Header"))
		return false
	}

	// 提取令牌部分
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// 解析并验证 JWT
	claims, err := common.ParseJWT(tokenString)

	if err != nil {
		ctx.Output.SetStatus(401)
		ctx.Output.Body([]byte("Unauthorized: Invalid Token"))
		return false
	}

	// 验证过期时间
	if exp, ok := (*claims)["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			ctx.Output.SetStatus(401)
			ctx.Output.Body([]byte("Unauthorized: Token Expired"))
			return false
		}
	}

	// 从 claims 中提取用户信息（例如 user_id）
	if userID, ok := (*claims)["user_id"].(float64); ok {
		// 将 user_id 存储到上下文中，供后续使用
		ctx.Input.SetData("user_id", gconv.Int(userID))
		fmt.Printf("Authenticated user_id: %f\n", userID)

		// 查询用户
		user := &models.User{Id: gconv.Int(userID)}
		err = user.Get()
		if err != nil {
			ctx.Output.SetStatus(401)
			ctx.Output.Body([]byte("Unauthorized: User not found"))
			return false
		}

		// 验证JWT是否存在
		if user.Jwt == "" || user.Jwt != tokenString {
			ctx.Output.SetStatus(401)
			ctx.Output.Body([]byte("Unauthorized: Invalid login status"))
			return false
		}
	} else {
		ctx.Output.SetStatus(401)
		ctx.Output.Body([]byte("Unauthorized: Invalid Token Claims"))
		return false
	}

	// 验证通过
	return true
}
