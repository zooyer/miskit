/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: dns.go
 * @Package: dns
 * @Version: 1.0.0
 * @Date: 2022/9/28 16:34
 */

package dns

import (
	"net"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

var (
	ttl  uint32 = 600
	conn *net.UDPConn
)

func setForwardTTL(t uint32) {
	ttl = t
}

func SetForwardTTL(ttl uint32) {
	setForwardTTL(ttl)
}

func handleHook(conn *net.UDPConn, addr *net.UDPAddr, msg dnsmessage.Message, hosts map[string][]dnsmessage.Resource) {
	if conn == nil || addr == nil || len(msg.Questions) < 1 {
		return
	}

	var questions = make([]dnsmessage.Question, 0, len(msg.Questions))
	for _, question := range msg.Questions {
		var hit bool

		name := strings.TrimRight(question.Name.String(), ".")
		for _, host := range hosts[name] {
			var (
				ok  bool
				res dnsmessage.ResourceBody
			)

			if host.Header.Class != question.Class {
				continue
			}

			switch question.Type {
			case dnsmessage.TypeA:
				res, ok = host.Body.(*dnsmessage.AResource)
			case dnsmessage.TypeAAAA:
				res, ok = host.Body.(*dnsmessage.AAAAResource)
			}

			if !ok {
				continue
			}

			hit = true

			msg.Answers = append(msg.Answers, dnsmessage.Resource{
				Header: dnsmessage.ResourceHeader{
					Name:  question.Name,
					Class: question.Class,
					TTL:   host.Header.TTL,
				},
				Body: res,
			})
		}

		if !hit {
			questions = append(questions, question)
		}
	}

	if len(questions) > 0 {
		for _, question := range questions {
			ips, err := net.LookupIP(question.Name.String())
			if err != nil {
				continue
			}

			for _, ip := range ips {
				var resource = dnsmessage.Resource{
					Header: dnsmessage.ResourceHeader{
						Name:  question.Name,
						Class: question.Class,
						TTL:   ttl,
					},
					Body: nil,
				}

				if ipv4 := ip.To4(); len(ipv4) == net.IPv4len {
					if question.Type != dnsmessage.TypeA {
						continue
					}
					var v4 dnsmessage.AResource
					copy(v4.A[:], ipv4[:net.IPv4len])
					resource.Body = &v4
					resource.Header.Type = dnsmessage.TypeA
				} else if ipv6 := ip.To16(); len(ipv6) == net.IPv6len {
					if question.Type != dnsmessage.TypeAAAA {
						continue
					}
					var v6 dnsmessage.AAAAResource
					copy(v6.AAAA[:], ipv6[:net.IPv6len])
					resource.Body = &v6
					resource.Header.Type = dnsmessage.TypeAAAA
				}

				msg.Answers = append(msg.Answers, resource)
			}
		}
	}

	if len(msg.Answers) > 0 {
		msg.Response = true
	}

	pkg, err := msg.Pack()
	if err != nil {
		return
	}

	if _, err = conn.WriteToUDP(pkg, addr); err != nil {
		return
	}
}

func HookHosts(hosts map[string][]dnsmessage.Resource) (err error) {
	conn, err = net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	if err != nil {
		return
	}

	var buf = make([]byte, 512)
	for {
		_, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		var msg dnsmessage.Message
		if err = msg.Unpack(buf); err != nil {
			continue
		}

		go handleHook(conn, addr, msg, hosts)
	}
}

func HookHostsByText(hosts map[string][]string) (err error) {
	var resources = make(map[string][]dnsmessage.Resource)
	for name, hosts := range hosts {
		for _, host := range hosts {
			var resource = dnsmessage.Resource{
				Header: dnsmessage.ResourceHeader{
					Name:   dnsmessage.MustNewName(name),
					Type:   0,
					Class:  dnsmessage.ClassINET,
					TTL:    ttl,
					Length: 0,
				},
				Body: nil,
			}

			ip := net.ParseIP(host)

			if ipv4 := ip.To4(); len(ipv4) == net.IPv4len {
				var v4 dnsmessage.AResource
				copy(v4.A[:], ipv4[:net.IPv4len])
				resource.Body = &v4
				resource.Header.Type = dnsmessage.TypeA
			} else if ipv6 := ip.To16(); len(ipv6) == net.IPv6len {
				var v6 dnsmessage.AAAAResource
				copy(v6.AAAA[:], ipv6[:net.IPv6len])
				resource.Body = &v6
				resource.Header.Type = dnsmessage.TypeAAAA
			}

			resources[name] = append(resources[name], resource)
		}
	}

	return HookHosts(resources)
}

func HookHostsByLocal(names ...string) (err error) {
	address, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	var hosts = make(map[string][]string)
	for _, addr := range address {
		if addr, ok := addr.(*net.IPNet); ok && addr.IP.IsGlobalUnicast() {
			for _, name := range names {
				hosts[name] = append(hosts[name], addr.IP.String())
			}
		}
	}

	return HookHostsByText(hosts)
}
