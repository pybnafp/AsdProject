package common

import (
	"errors"
	"reflect"

	"github.com/beego/beego/v2/server/web/context"
)

// 添加数据过滤
func AddDataFitler(ctx *context.Context, req interface{}) error {
	// 从上下文中获取当前用户的 ID
	userId, ok := ctx.Input.GetData("userId").(int)
	if !ok {
		return errors.New("userId 不存在")
	}

	// 使用反射获取 req 的值
	v := reflect.ValueOf(req)

	// 检查 req 是否是指针类型，并且解引用之后是结构体
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("req 不是指向结构体的指针")
	}

	// 获取结构体的值
	v = v.Elem()

	// 获取结构体的类型
	t := v.Type()

	// 遍历结构体的字段，找到名为 CreateUser 的字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 检查字段名称是否是 CreateUser，并且字段类型是 int
		if field.Name == "CreateUser" && v.Field(i).Kind() == reflect.Int {
			// 设置 CreateUser 字段的值为 userId
			v.Field(i).SetInt(int64(userId))
			return nil // 设置成功，返回 nil 表示没有错误
		}
	}

	// 如果没有找到 CreateUser 字段，返回错误
	return errors.New("未找到 CreateUser 字段")

}
