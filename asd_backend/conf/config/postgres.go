package config

// 数据库结构体
type Postgres struct {
	Default DBConfig `mapstructure:"default" json:"default" yaml:"default"`
}
