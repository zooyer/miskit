/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: device.go
 * @Package: device
 * @Version: 1.0.0
 * @Date: 2022/10/4 14:05
 */

package device

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/yumaojun03/dmidecode"
	"github.com/zooyer/embed/log"
)

func containsAny(value string, any ...string) bool {
	for _, v := range any {
		if strings.Contains(value, v) {
			return true
		}
	}

	return false
}

func execCommand(name string, arg ...string) (output string, err error) {
	out, err := exec.Command(name, arg...).CombinedOutput()
	if err != nil {
		return
	}

	return string(out), nil
}

func wmicByKey(t, key string) string {
	output, err := execCommand("wmic", t, "get", key)
	if err != nil {
		log.ZError("wmicByKey execCommand error:", err.Error())
		return ""
	}

	// 去掉前缀
	output = strings.TrimSpace(output)
	output = strings.TrimPrefix(output, key)
	output = strings.TrimSpace(output)

	// 多个结果以\n分隔
	lines := strings.Split(output, "\n")
	outs := make([]string, 0, len(lines))
	for _, line := range lines {
		if line = strings.TrimSpace(line); line == "" {
			continue
		}
		outs = append(outs, line)
	}
	output = strings.Join(outs, "\n")

	// 去掉无效值
	output = strings.TrimPrefix(output, "None")
	output = strings.TrimPrefix(output, "Default string")

	return output
}

func dmiDecode(t string) map[string]string {
	output, err := execCommand("dmidecode", "-q", "-t", t)
	if err != nil {
		log.ZError("dmiDecode execCommand error:", err.Error())
		return nil
	}

	var info = make(map[string]string)
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		lines = lines[1:]
	}

	for _, line := range lines {
		if line = strings.TrimSpace(line); line == "" {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			continue
		}

		name := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])
		if name == "" {
			continue
		}

		info[name] = value
	}

	return info
}

func dmiDecodeByKey(t, key string) string {
	return dmiDecode(t)[key]
}

func getMacAddress() (addr []string, err error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}

	defer sort.Strings(addr)

	for _, i := range interfaces {
		is, err := isPhysicalEthernet(i.Index)
		if err != nil {
			return nil, err
		}

		if !is || i.Flags&net.FlagLoopback != 0 {
			continue
		}

		addr = append(addr, fmt.Sprintf("%s", i.HardwareAddr.String()))

		//switch runtime.GOOS {
		//case "windows":
		//case "darwin":
		//	fallthrough
		//case "linux":
		//	if !containsAny(i.Name, "en", "eno", "ens", "enp2s", "eth") {
		//		continue
		//	}
		//
		//	if containsAny(i.Name, "bridge", "vmenet", "utun") {
		//		continue
		//	}
		//}
	}

	return
}

func getBaseboardIDByWindows() string {
	return wmicByKey("baseboard", "SerialNumber")
}

func getBaseboardIDByDarwin() []string {
	return nil
	panic("implement")
}

func getBaseboardIDByLinux() string {
	return dmiDecodeByKey("2", "Serial Number")
}

func getBiosSNByWindows() string {
	return strings.ReplaceAll(wmicByKey("bios", "SerialNumber"), " ", "")
}

func getBiosSNByDarwin() []string {
	output, err := execCommand("system_profiler", "SPHardwareDataType")
	if err != nil {
		log.ZError("getBiosSNByDarwin execCommand error:", err.Error())
		return nil
	}

	var sn []string
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Serial Number (system)") && strings.Contains(line, ":") {
			sn = append(sn, strings.TrimSpace(strings.Split(line, ":")[1]))
		}
	}

	return sn
}

func getBiosSNByLinux() string {
	return strings.ReplaceAll(dmiDecodeByKey("1", "Serial Number"), "System Serial Number", "")
}

func getCPUIDByWindows() []string {
	return strings.Split(wmicByKey("cpu", "ProcessorId"), "\n")
}

func getCPUIDByDarwin() []string {
	return nil
	panic("implement")
}

func getCPUIDByLinux() []string {
	return strings.Split(dmiDecodeByKey("4", "ID"), " ")
}

func getOSUUIDByWindows() string {
	return wmicByKey("csproduct", "UUID")
}

func getOSUUIDByDarwin() []string {
	return nil
	panic("implement")
}

func getOSUUIDByLinux() string {
	output, err := execCommand("dmidecode", "-s", "system-uuid")
	if err != nil {
		log.ZError("getOSUUIDByLinux execCommand error:", err.Error())
		return ""
	}

	output = strings.TrimSpace(output)

	return output
}

func BaseboardSN() []string {
	switch runtime.GOOS {
	case "darwin":
		return getBaseboardIDByDarwin()
	case "linux":
		fallthrough
	case "windows":
		dmi, err := dmidecode.New()
		if err != nil {
			return nil
		}

		boards, err := dmi.BaseBoard()
		if err != nil {
			return nil
		}

		var sn = make([]string, 0, len(boards))
		for _, board := range boards {
			sn = append(sn, fmt.Sprintf("%s %s", board.Manufacturer, board.SerialNumber))
		}

		return sn
	}

	return nil
}

func BiosSN() []string {
	switch runtime.GOOS {
	case "darwin":
		return getBiosSNByDarwin()
	case "linux":
		fallthrough
	case "windows":
		dmi, err := dmidecode.New()
		if err != nil {
			return nil
		}

		bios, err := dmi.BIOS()
		if err != nil {
			return nil
		}

		var sn = make([]string, len(bios))
		for _, b := range bios {
			sn = append(sn, fmt.Sprintf("%s %s %v", strings.TrimSpace(b.Vendor), strings.TrimSpace(b.BIOSVersion), b.Characteristics))
		}

		return sn
	}

	return nil
}

func ProcessorSN() []string {
	switch runtime.GOOS {
	case "darwin":
		return getCPUIDByDarwin()
	case "linux":
		fallthrough
	case "windows":
		dmi, err := dmidecode.New()
		if err != nil {
			return nil
		}

		processor, err := dmi.Processor()
		if err != nil {
			return nil
		}

		var id = make([]string, 0, len(processor))
		for _, p := range processor {
			id = append(id, fmt.Sprintf("%s %s", strings.TrimSpace(p.Version), p.ID))
		}

		return id
	}

	return nil
}

func MAC() []string {
	addr, _ := getMacAddress()
	return addr
}

func SystemID() []string {
	switch runtime.GOOS {
	case "darwin":
		return getOSUUIDByDarwin()
	case "linux":
		fallthrough
	case "windows":
		dmi, err := dmidecode.New()
		if err != nil {
			return nil
		}

		system, err := dmi.System()
		if err != nil {
			return nil
		}

		var id = make([]string, 0, len(system))
		for _, s := range system {
			id = append(id, fmt.Sprintf("%s %s %s", s.Manufacturer, s.ProductName, s.SerialNumber))
		}

		return id
	}

	return nil
}
