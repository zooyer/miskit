/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: dns_test.go
 * @Package: dns
 * @Version: 1.0.0
 * @Date: 2022/9/28 16:37
 */

package dns

import (
	"net"
	"testing"

	"golang.org/x/net/dns/dnsmessage"
)

func TestHook(t *testing.T) {
	addrs, err := net.LookupIP("vpn.zb.genkitol.com")
	if err != nil {
		t.Fatal(err)
	}

	for _, addr := range addrs {
		t.Log(addr)
	}
}

func TestHookHosts(t *testing.T) {
	var hosts = map[string][]dnsmessage.Resource{
		"zzy": {
			{
				Header: dnsmessage.ResourceHeader{
					Name:   dnsmessage.MustNewName("zzy"),
					Type:   dnsmessage.TypeA,
					Class:  dnsmessage.ClassINET,
					TTL:    600,
					Length: 0,
				},
				Body: &dnsmessage.AResource{
					A: [4]byte{110, 110, 110, 110},
				},
			},
			{
				Header: dnsmessage.ResourceHeader{
					Name:   dnsmessage.MustNewName("zzy"),
					Type:   dnsmessage.TypeAAAA,
					Class:  dnsmessage.ClassINET,
					TTL:    600,
					Length: 0,
				},
				Body: &dnsmessage.AAAAResource{
					AAAA: [16]byte{0x24, 0x08, 0x82, 0x1b, 0x71, 0x10, 0x19, 0x10, 0x2e, 0xb2, 0x1a, 0xff, 0xfe, 0xca, 0x21, 0x9d},
				},
			},
		},
	}

	if err := HookHosts(hosts); err != nil {
		t.Fatal(err)
	}

	select {}
}

func TestHookHostsByText(t *testing.T) {
	var hosts = map[string][]string{
		"zzy": {
			"110.110.110.110",
			"2408:821b:7110:1910:2eb2:1aff:feca:219d",
		},
	}

	if err := HookHostsByText(hosts); err != nil {
		t.Fatal(err)
	}

	select {}
}
