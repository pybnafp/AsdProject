/**
 * 登录验证中间件
 * @author Evotrek 研发团队
 * @since 2024/9/24
 * @File : checkauth
 */
package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	beego "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/adapter/context"
	"github.com/beego/beego/v2/core/logs"
)

var appSecretMap = map[string]string{
	"XL3JIVWF": "O215ZAIRPJE36LXY",
}

func CheckOpenApi() {

	//登录认证中间件过滤器
	var login = func(ctx *context.Context) {
		// 设置
		if strings.HasPrefix(ctx.Request.URL.Path, "/openapi/") {
			if !CheckOpenApiAuth(ctx) {
				return
			}
		}
	}

	// 登录过滤器
	beego.InsertFilter("/openapi/*", beego.BeforeRouter, login)
}

func CheckOpenApiAuth(ctx *context.Context) bool {
	// 获取基础参数
	appId := ctx.Input.Query("appId")
	nonce := ctx.Input.Query("nonce")
	sign := ctx.Input.Query("sign")
	timestamp := ctx.Input.Query("timestamp")

	// 检查基础参数是否存在
	if appId == "" || nonce == "" || sign == "" || timestamp == "" {
		ctx.Output.SetStatus(400)
		ctx.Output.Body([]byte("Missing required parameters"))
		return false
	}

	// 获取 appSecret
	appSecret, isOk := appSecretMap[appId]
	if !isOk {
		ctx.Output.SetStatus(400)
		ctx.Output.Body([]byte("Invalid appId or appSecret"))
		return false
	}

	// 构建参数映射（去除 sign 字段）
	data := map[string]string{}

	// 获取 GET 和 POST 参数
	// 获取 GET 参数并排除 sign
	for k, v := range ctx.Input.Context.Request.URL.Query() {
		if k == "sign" {
			continue
		}
		// 只取第一个值，如果有多个值
		if len(v) > 0 {
			data[k] = strings.TrimSpace(v[0])
		}
	}

	// 获取 POST 表单中的所有参数 (使用 GetPostForm)
	for k := range ctx.Input.Context.Request.PostForm {
		if k == "sign" {
			continue
		}
		data[k] = strings.TrimSpace(ctx.Input.Context.Request.PostFormValue(k))
	}

	// 获取并排序参数的键
	keyArray := make([]string, 0, len(data))
	for k := range data {
		keyArray = append(keyArray, k)
	}
	sort.Strings(keyArray)

	// 拼接参数
	var sb strings.Builder
	for _, k := range keyArray {
		value := data[k]
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(url.QueryEscape(value)) // URL encode values
		sb.WriteString("&")
	}

	// 拼接 appSecret 作为最后一个参数
	sb.WriteString("appSecret=")
	sb.WriteString(appSecret)

	// 打印拼接后的参数（可选，用于调试）
	logs.Info("openapi 入参：%v", data)

	// 生成签名（使用 SHA-256）
	hash := sha256.New()
	hash.Write([]byte(sb.String()))
	expectedSign := hex.EncodeToString(hash.Sum(nil))

	// 验证签名
	if expectedSign != sign {
		ctx.Output.SetStatus(403)
		ctx.Output.Body([]byte("Invalid signature"))
		return false
	}

	// 签名验证通过
	return true
}

func CheckOpenApiAuth2(ctx *context.Context) bool {
	// 复制请求 Body
	body := ctx.Input.CopyBody(beego.BConfig.MaxMemory)

	// 解析 JSON 请求体
	var postData map[string]interface{}
	if err := json.Unmarshal(body, &postData); err != nil {
		ctx.Output.SetStatus(400)
		ctx.Output.Body([]byte("Invalid JSON format"))
		return false
	}

	// 获取基础参数
	appId, ok := postData["appId"].(string)
	nonce, nonceOk := postData["nonce"].(string)
	sign, signOk := postData["sign"].(string)
	timestamp, timestampOk := postData["timestamp"].(string)

	logs.Info(ctx.Request.URL.Path)
	// 检查基础参数是否存在
	if !ok || !nonceOk || !signOk || !timestampOk || appId == "" || nonce == "" || sign == "" || timestamp == "" {
		ctx.Output.SetStatus(400)
		ctx.Output.Body([]byte("Missing required parameters"))
		return false
	}

	// 获取 appSecret
	appSecret, isOk := appSecretMap[appId]
	if !isOk {
		ctx.Output.SetStatus(400)
		ctx.Output.Body([]byte("Invalid appId or appSecret"))
		return false
	}

	// 构建参数映射（去除 sign 字段并移除空值）
	data := map[string]string{}
	for k, v := range postData {
		if k == "sign" {
			continue
		}
		// 只处理字符串和数字类型的值，移除空值
		switch value := v.(type) {
		case string:
			trimmedValue := strings.TrimSpace(value)
			data[k] = trimmedValue
		case float64:
			// 将数字类型格式化为字符串，移除 0 值
			data[k] = fmt.Sprintf("%.0f", value)
		}
	}

	// 获取并排序参数的键
	keyArray := make([]string, 0, len(data))
	for k := range data {
		keyArray = append(keyArray, k)
	}
	sort.Strings(keyArray)

	// 拼接参数
	var sb strings.Builder
	for _, k := range keyArray {
		value := data[k]
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(url.QueryEscape(value)) // 对值进行 URL 编码
		sb.WriteString("&")
	}

	// 拼接 appSecret 作为最后一个参数
	sb.WriteString("appSecret=")
	sb.WriteString(appSecret)

	// 打印拼接后的参数（可选，用于调试）
	fmt.Println("拼接后的参数：" + sb.String())

	// 生成签名（使用 SHA-256）
	hash := sha256.New()
	hash.Write([]byte(sb.String()))
	expectedSign := hex.EncodeToString(hash.Sum(nil))

	// 验证签名
	if expectedSign != sign {
		ctx.Output.SetStatus(403)
		ctx.Output.Body([]byte("Invalid signature"))
		return false
	}

	// 签名验证通过
	return true
}
