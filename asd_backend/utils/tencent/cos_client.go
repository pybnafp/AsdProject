package tencent

import (
	"asd/conf"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/tencentyun/cos-go-sdk-v5"
)

func NewCOSClient() *cos.Client {
	// 替换为您的存储桶 URL 和地域
	u, _ := url.Parse("https://asd-1306205170.cos.ap-nanjing.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u}

	// 初始化 COS 客户端
	return cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  conf.CONFIG.TencentConfig.SecretId,  // 从配置中获取 SecretID
			SecretKey: conf.CONFIG.TencentConfig.SecretKey, // 从配置中获取 SecretKey
		},
	})
}

func UploadFileToCOS(fileName string, filePath string) (string, error) {
	// 创建一个 COS 客户端
	client := NewCOSClient()

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		logs.Error("打开文件失败: %v", err)
		return "", err
	}
	defer file.Close()

	// 上传文件到 COS
	_, err = client.Object.Put(context.Background(), fileName, file, nil)
	if err != nil {
		logs.Error("文件上传失败: %v", err)
		return "", err
	}

	// 返回文件的存储路径
	fileURL := fileName

	return fileURL, nil
}

// GetDownloadURL 生成文件的预签名下载 URL
func GetDownloadURL(fileName string) (string, error) {
	if fileName == "" {
		return "", nil
	}

	expireTime := time.Second * time.Duration(conf.CONFIG.TencentConfig.ExpireTime)

	// 创建一个 COS 客户端
	client := NewCOSClient()

	// 生成带有签名的预签名下载 URL
	presignedURL, err := client.Object.GetPresignedURL(context.Background(), http.MethodGet, fileName, conf.CONFIG.TencentConfig.SecretId, conf.CONFIG.TencentConfig.SecretKey, expireTime, nil)
	if err != nil {
		logs.Error("生成预签名下载 URL 失败: %v", err)
		return "", err
	}

	// 返回预签名的下载链接
	return presignedURL.String(), nil
}

// GetFileSize 获取 COS 文件的大小
func GetFileSize(fileName string) (int64, error) {
	if fileName == "" {
		return 0, nil
	}

	// 创建一个 COS 客户端
	client := NewCOSClient()

	// 获取文件属性
	resp, err := client.Object.Head(context.Background(), fileName, nil)
	fmt.Println(resp)
	if err != nil {
		logs.Error("获取文件属性失败: %v", err)
		return 0, err
	}

	// 从响应头中获取文件大小
	contentLength := resp.ContentLength

	return contentLength, nil
}

// CopyFile 复制 COS 中的文件
func CopyFile(srcFileName string, destFileName string) error {
	if srcFileName == "" || destFileName == "" {
		return nil
	}

	// 创建一个 COS 客户端
	client := NewCOSClient()

	// 构建源文件的完整路径
	sourceURL := fmt.Sprintf("amppl-1306205170.cos.ap-guangzhou.myqcloud.com/%s", srcFileName)

	// 复制对象
	_, _, err := client.Object.Copy(context.Background(), destFileName, sourceURL, nil)
	if err != nil {
		logs.Error("复制文件失败: %v", err)
		return err
	}

	return nil
}
