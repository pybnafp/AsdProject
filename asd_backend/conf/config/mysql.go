package config

// 数据库结构体
type MySQL struct {
	Default      DBConfig `mapstructure:"default" json:"default" yaml:"default"`
}

// 单个数据库配置
type DBConfig struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Database string `mapstructure:"database" json:"database" yaml:"database"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
}
