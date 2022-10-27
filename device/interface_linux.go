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
	"net"
	"os"
	"strings"
)

func isPhysicalEthernet(index int) (is bool, err error) {
	inf, err := net.InterfaceByIndex(index)
	if err != nil {
		return
	}

	// 读取网卡软连
	link, err := os.Readlink("/sys/class/net/" + inf.Name)
	if err != nil {
		return
	}

	// 不包含USB网卡
	if !strings.Contains(link, "/devices/pci") {
		return
	}

	// 过滤掉虚拟网卡
	if !strings.Contains(link, "/devices/virtual/net/") {
		return true, nil
	}

	return
}
