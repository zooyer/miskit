//go:build !windows && !darwin && !linux

/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: interface_other.go
 * @Package: device
 * @Version: 1.0.0
 * @Date: 2022/10/5 15:07
 */

package device

import (
	"fmt"
	"runtime"
)

func isPhysicalEthernet(index int) (is bool, err error) {
	return false, fmt.Errorf("os %s not implemented", runtime.GOOS)
}
