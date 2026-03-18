package main

import (
	_ "asd/boot/config"
	_ "asd/boot/init" // 确保最先导入
	_ "asd/boot/postgres"
	_ "asd/boot/session"
	"asd/conf"
	_ "asd/routers"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	// 配置全局日志输出到文件
	err := logs.SetLogger(logs.AdapterFile, `{"filename":"global.log"}`)
	if err != nil {
		panic(err)
	}

	// 设置全局日志级别
	if beego.BConfig.RunMode == "dev" {
		logs.SetLevel(logs.LevelDebug)
	} else {
		logs.SetLevel(logs.LevelInfo)
	}
	logs.Info("run at mode=%v, port=%v", conf.RunMode, beego.BConfig.Listen.HTTPPort)
	beego.Run()
}
