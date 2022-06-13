/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: time.go
 * @Package: utils
 * @Version: 1.0.0
 * @Date: 2022/6/13 15:59
 */

package utils

import "time"

// TimestampFormat 序列化时间戳
func TimestampFormat(seconds int64, format string) string {
	return time.Unix(seconds, 0).Format(format)
}
