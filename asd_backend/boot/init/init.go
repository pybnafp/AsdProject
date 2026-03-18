package init

import (
	"asd/conf"
	"flag"

	beego "github.com/beego/beego/v2/server/web"
)

var (
	mode = flag.String("mode", "api", "运行模式: api, job, console")
	port *int // 添加全局 port 变量
)

func init() {
	// 从配置文件获取默认端口
	defaultPort, _ := beego.AppConfig.Int("httpport")
	if defaultPort == 0 {
		defaultPort = 8080
	}

	// 使用配置文件中的端口作为默认值
	port = flag.Int("port", defaultPort, "HTTP服务端口")

	flag.Parse()
	conf.RunMode = *mode
	beego.BConfig.Listen.HTTPPort = *port
}

// GetPort 获取端口号
func GetPort() int {
	return *port
}
