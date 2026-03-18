/**
 * 文件处理类
 * @author Evotrek 研发团队
 * @since 2021/11/15
 * @File : gfile
 */
package gfile

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Remove(path string) error {
	return os.RemoveAll(path)
}

func Create(path string) (*os.File, error) {
	dir := Dir(path)
	if !Exists(dir) {
		if err := Mkdir(dir); err != nil {
			return nil, err
		}
	}
	return os.Create(path)
}

func Mkdir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func Dir(path string) string {
	if path == "." {
		return filepath.Dir(RealPath(path))
	}
	return filepath.Dir(path)
}

func RealPath(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	if !Exists(p) {
		return ""
	}
	return p
}

func Exists(path string) bool {
	if stat, err := os.Stat(path); stat != nil && !os.IsNotExist(err) {
		return true
	}
	return false
}

// 获取完整的扩展名（处理复合扩展名）
func GetFullExt(fileName string) string {
	knownExts := []string{".tar.gz", ".tar.bz2", ".tar.xz", ".zip", ".rar", ".7z"}
	lowerFileName := strings.ToLower(fileName)

	// 检查是否有已知的复合扩展名
	for _, ext := range knownExts {
		if strings.HasSuffix(lowerFileName, ext) {
			return ext
		}
	}

	// 如果没有复合扩展名，则返回最后的单一扩展名
	return filepath.Ext(fileName)
}

// downloadFile 下载文件到指定路径
func DownloadUrlFile(url string, filepath string) error {
	// 创建HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建文件
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 写入文件
	_, err = io.Copy(out, resp.Body)
	return err
}
