/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: interface_windows.go
 * @Package: device
 * @Version: 1.0.0
 * @Date: 2022/10/5 14:43
 */

package device

import (
	"fmt"
	"strings"
	_ "unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

//go:linkname adapterAddresses net.adapterAddresses
func adapterAddresses() ([]*windows.IpAdapterAddresses, error)

func regString(key registry.Key, path, name string) (value string, err error) {
	subKey, err := registry.OpenKey(key, path, registry.READ)
	if err != nil {
		return
	}

	defer subKey.Close()

	if value, _, err = subKey.GetStringValue(name); err != nil {
		return
	}

	return
}

func regInteger(key registry.Key, path, name string) (value uint64, err error) {
	subKey, err := registry.OpenKey(key, path, registry.READ)
	if err != nil {
		return
	}

	defer subKey.Close()

	if value, _, err = subKey.GetIntegerValue(name); err != nil {
		return
	}

	return
}

// isPhysical1 判断设备前缀
func isPhysical1(name string) (is bool, err error) {
	var path = fmt.Sprintf("SYSTEM\\CurrentControlSet\\Control\\Network\\{4D36E972-E325-11CE-BFC1-08002BE10318}\\%s\\Connection", name)

	instance, err := regString(registry.LOCAL_MACHINE, path, "PnPInstanceId")
	if err != nil {
		return
	}

	// 不包含USB网卡
	return strings.HasPrefix(instance, "PCI"), nil
}

// isPhysical2 判断设备前缀和属性
func isPhysical2(name string) (is bool, err error) {
	const path = "SYSTEM\\CurrentControlSet\\Control\\Class\\{4D36E972-E325-11CE-BFC1-08002bE10318}"
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.READ)
	if err != nil {
		return
	}

	defer key.Close()

	stat, err := key.Stat()
	if err != nil {
		return
	}

	subs, err := key.ReadSubKeyNames(int(stat.SubKeyCount))
	if err != nil {
		return
	}

	for _, sub := range subs {
		instance, err := regString(key, sub, "NetCfgInstanceId")
		if err == registry.ErrNotExist || err != nil || instance != name {
			continue
		}

		// 不包含USB网卡
		device, err := regString(key, sub, "DeviceInstanceID")
		if err == registry.ErrNotExist || err != nil || !strings.HasPrefix(device, "PCI") {
			continue
		}

		// ox4 NCF_PHYSICAL 说明组件是一个物理适配器
		value, err := regInteger(key, sub, "Characteristics")
		if err == nil && value&0x4 == 0x4 {
			return true, nil
		}
	}

	return
}

// isPhysical3 isPhysical1增强版本，通过bus方式判断：https://blog.csdn.net/whf727/article/details/6660815
func isPhysical3(name string) (is bool, err error) {
	var path = fmt.Sprintf("SYSTEM\\CurrentControlSet\\Control\\Network\\{4D36E972-E325-11CE-BFC1-08002BE10318}\\%s\\Connection", name)

	instance, err := regString(registry.LOCAL_MACHINE, path, "PnPInstanceId")
	if err != nil {
		return
	}

	// 不包含USB网卡
	if !strings.HasPrefix(instance, "PCI") {
		return
	}

	guid, err := windows.GUIDFromString("{4D36E972-E325-11CE-BFC1-08002BE10318}")
	if err != nil {
		return
	}

	dev, err := windows.SetupDiGetClassDevsEx(&guid, "", 0, windows.DIGCF_PRESENT, 0, "")
	if err != nil {
		return
	}

	defer dev.Close()

	for i := 0; ; i++ {
		info, err := dev.EnumDeviceInfo(i)
		if err != nil {
			if err == windows.ERROR_NO_MORE_ITEMS {
				break
			}
			continue
		}

		id, err := dev.DeviceInstanceID(info)
		if err != nil {
			continue
		}

		if !strings.EqualFold(id, instance) {
			continue
		}

		value, err := dev.DeviceRegistryProperty(info, windows.SPDRP_BUSNUMBER)
		if err != nil {
			continue
		}

		if num, ok := value.(uint32); ok && int32(num) != -1 {
			return true, nil
		}
	}

	return false, nil
}

func isPhysicalEthernet(index int) (is bool, err error) {
	aas, err := adapterAddresses()
	if err != nil {
		return
	}

	for _, aa := range aas {
		if int(aa.IfIndex) != index {
			continue
		}

		// 有线网卡和无线网卡
		if aa.IfType == windows.IF_TYPE_ETHERNET_CSMACD || aa.IfType == windows.IF_TYPE_IEEE80211 {
			return isPhysical3(windows.BytePtrToString(aa.AdapterName))
		}
	}

	return
}
