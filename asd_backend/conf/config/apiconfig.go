package config

// Api配置
type ApiConfig struct {
	JWTExp                      string   `mapstructure:"jwt_exp" json:"jwt_exp" yaml:"jwt_exp"`
	UseRag                      bool     `mapstructure:"use_rag" json:"use_rag" yaml:"use_rag"`
	RagUrls                     []string `mapstructure:"rag_urls" json:"rag_urls" yaml:"rag_urls"`
	SecretKey                   string   `mapstructure:"secret_key" json:"secret_key" yaml:"secret_key"`
	FileCachePath               string   `mapstructure:"file_cache_path" json:"file_cache_path" yaml:"file_cache_path"`
	AlibabaCloudAccessKeyId     string   `mapstructure:"alibaba_cloud_access_key_id" json:"alibaba_cloud_access_key_id" yaml:"alibaba_cloud_access_key_id"`
	AlibabaCloudAccessKeySecret string   `mapstructure:"alibaba_cloud_access_key_secret" json:"alibaba_cloud_access_key_secret" yaml:"alibaba_cloud_access_key_secret"`
	AlibabaBailianAppId         string   `mapstructure:"alibaba_bailian_app_id" json:"alibaba_bailian_app_id" yaml:"alibaba_bailian_app_id"`
	AlibabaBailianApiKey        string   `mapstructure:"alibaba_bailian_api_key" json:"alibaba_bailian_api_key" yaml:"alibaba_bailian_api_key"`
	AlibabaBailianWorkspaceId   string   `mapstructure:"alibaba_bailian_workspace_id" json:"alibaba_bailian_workspace_id" yaml:"alibaba_bailian_workspace_id"`
	AlibabaSmsSignName          string   `mapstructure:"alibaba_sms_sign_name" json:"alibaba_sms_sign_name" yaml:"alibaba_sms_sign_name"`
	AlibabaSmsTemplateCode      string   `mapstructure:"alibaba_sms_template_code" json:"alibaba_sms_template_code" yaml:"alibaba_sms_template_code"`
	OpenRouterApiKey            string   `mapstructure:"openrouter_api_key" json:"openrouter_api_key" yaml:"openrouter_api_key"`
	OpenRouterBaseUrl           string   `mapstructure:"openrouter_base_url" json:"openrouter_base_url" yaml:"openrouter_base_url"`
	AlgorithmBaseUrl            string   `mapstructure:"algorithm_base_url" json:"algorithm_base_url" yaml:"algorithm_base_url"`
}
