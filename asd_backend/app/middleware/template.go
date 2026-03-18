package middleware

import (
	"asd/utils/convert"

	beego "github.com/beego/beego/v2/adapter"
)

func LoadTemplateFunc() {
	beego.AddFuncMap("commaI", convert.CommaInt)
	beego.AddFuncMap("comma", convert.Comma)
	beego.AddFuncMap("mul", mul)
}

// 自定义乘法函数
func mul(a, b float64) float64 {
	return a * b
}
