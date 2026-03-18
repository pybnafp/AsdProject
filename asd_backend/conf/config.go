package conf

import "asd/conf/config"

// 全局配置结构体
type Config struct {
	Mysql         config.MySQL         `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Postgres      config.Postgres      `mapstructure:"postgres" json:"postgres" yaml:"postgres"`
	Redis         config.RedisConfig   `mapstructure:"redis" json:"redis" yaml:"redis"`
	Attachment    config.Attachment    `mapstructure:"attachment" json:"attachment" yaml:"attachment"`
	SystemConfig  config.SystemConfig  `mapstructure:"systemconfig" json:"systemconfig" yaml:"systemconfig"`
	TencentConfig config.TencentConfig `mapstructure:"tencentconfig" json:"tencentconfig" yaml:"tencentconfig"`
	ApiConfig     config.ApiConfig     `mapstructure:"apiconfig" json:"apiconfig" yaml:"apiconfig"`
}
