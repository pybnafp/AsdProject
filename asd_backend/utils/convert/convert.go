/**
 * 转换函数库
 * @author Evotrek 研发团队
 * @since 2021/7/14
 * @File : convert
 */
package convert

import (
	"asd/utils/gconv"
	"strings"

	"github.com/dustin/go-humanize"
)

// 带分隔符的字符串分裂成int64数组
func ToInt64Array(str, split string) []int64 {
	result := make([]int64, 0)
	if str == "" {
		return result
	}
	arr := strings.Split(str, split)
	if len(arr) > 0 {
		for i := range arr {
			if arr[i] != "" {
				result = append(result, gconv.Int64(arr[i]))
			}
		}
	}
	return result

}

func CommaInt(num int) string {
	return humanize.Comma(int64(num))
}
func Comma(num int64) string {
	return humanize.Comma(num)
}
