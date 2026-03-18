package config

// 系统配置
type TencentConfig struct {
	SecretId          string `mapstructure:"secretid" json:"secretid" yaml:"secretid"`
	SecretKey         string `mapstructure:"secretkey" json:"secretkey" yaml:"secretkey"`
	ExpireTime        int    `mapstructure:"expiretime" json:"expiretime" yaml:"expiretime"` // // 以秒为单位的过期时间
	WechatAppID       string `mapstructure:"wechat_app_id" json:"wechat_app_id" yaml:"wechat_app_id"`
	WechatAppSecret   string `mapstructure:"wechat_app_secret" json:"wechat_app_secret" yaml:"wechat_app_secret"`
	WechatRedirectURI string `mapstructure:"wechat_redirect_uri" json:"wechat_redirect_uri" yaml:"wechat_redirect_uri"`
}
