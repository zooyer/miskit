/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: interface_darwin.go
 * @Package: device
 * @Version: 1.0.0
 * @Date: 2022/10/5 15:18
 */

package device

import (
	"encoding/json"
	"syscall"

	"golang.org/x/net/route"
)

func interfaceMessages(index int) ([]route.Message, error) {
	rib, err := route.FetchRIB(syscall.AF_UNSPEC, syscall.NET_RT_IFLIST, index)
	if err != nil {
		return nil, err
	}
	return route.ParseRIB(syscall.NET_RT_IFLIST, rib)
}

func isPhysicalEthernet(index int) (is bool, err error) {
	msg, err := interfaceMessages(index)
	if err != nil {
		return
	}

	for _, m := range msg {
		m, ok := m.(*route.InterfaceMessage)
		if !ok {
			continue
		}

		if m.Index != index {
			continue
		}

		if sys := m.Sys(); len(sys) == 1 {
			if im, ok := sys[0].(*route.InterfaceMetrics); ok && im.Type == 6 {
				//fmt.Println(fmt.Sprintf("name: %s, flags:%b", m.Name, m.Flags))
				//fmt.Println("name:", m.Name, "version:", m.Version, "type:", m.Type, "flags:", m.Flags, "index:", m.Index, "sys:", marshalJSON(m.Sys()))
				_ = im
				return true, nil
			}
		}
	}

	return
}

func marshalJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
