package config

// Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"redis_host" json:"redis_host" yaml:"redis_host"`
	Port     int    `mapstructure:"redis_port" json:"redis_port" yaml:"redis_port"`
	Password string `mapstructure:"redis_password" json:"redis_password" yaml:"redis_password"`
	DB       int    `mapstructure:"redis_db" json:"redis_db" yaml:"redis_db"`
}
