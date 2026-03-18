package sms

import (
	"asd/conf"
	"encoding/json"
	"regexp"
	"time"

	"asd/app/services"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

func IsValidMobile(mobile string) bool {
	re := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return re.MatchString(mobile)
}

func SendSmsCode(phone string, code string) error {
	// Save code to Redis with 5 minutes expiration
	redisClient := services.GetRedisClient()
	key := "sms:code:" + phone
	err := redisClient.SetEx(key, code, 5*time.Minute)
	if err != nil {
		return err
	}

	client, err := dysmsapi.NewClientWithAccessKey(
		"cn-hangzhou",
		conf.CONFIG.ApiConfig.AlibabaCloudAccessKeyId,
		conf.CONFIG.ApiConfig.AlibabaCloudAccessKeySecret,
	)
	if err != nil {
		return err
	}

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = phone
	request.SignName = conf.CONFIG.ApiConfig.AlibabaSmsSignName
	request.TemplateCode = conf.CONFIG.ApiConfig.AlibabaSmsTemplateCode

	param := map[string]string{
		"code": code,
	}
	paramBytes, _ := json.Marshal(param)
	request.TemplateParam = string(paramBytes)
	_, err = client.SendSms(request)
	if err != nil {
		return err
	}
	return nil
}

func VerifySmsCode(phone string, code string) bool {
	redisClient := services.GetRedisClient()
	key := "sms:code:" + phone
	storedCode, err := redisClient.Get(key)
	if err != nil {
		return false
	}
	return storedCode == code
}
