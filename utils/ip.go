/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: ip.go
 * @Package: utils
 * @Version: 1.0.0
 * @Date: 2022/6/13 15:57
 */

package utils

import "net"

// IsPrivateIP 判断是否是内网ip
func IsPrivateIP(ip string) bool {
	return !IsPublicIP(ip)
}

// IsPublicIP 判断是否是公网IP
func IsPublicIP(ipaddr string) bool {
	var ip = net.ParseIP(ipaddr)
	if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		}
		return true
	}
	return false
}
