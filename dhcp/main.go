/**
 * @Author: zzy
 * @Email: zhangzhongyuan@didiglobal.com
 * @Description:
 * @File: main.go
 * @Package: dhcp4
 * @Version: 1.0.0
 * @Date: 2022/11/9 23:36
 */

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/krolaw/dhcp4"
	"log"
	"math/rand"
	"net"
	"time"
)

// ExampleHandler using DHCP with a single network interface device
func ExampleHandler() {
	serverIP := net.IP{192, 168, 1, 1}
	handler := &DHCPHandler{
		ip:            serverIP,
		leaseDuration: 2 * time.Hour,
		start:         net.IP{192, 168, 1, 2},
		leaseRange:    50,
		leases:        make(map[int]lease, 10),
		options: dhcp4.Options{
			dhcp4.OptionSubnetMask:       []byte{255, 255, 255, 0},
			dhcp4.OptionRouter:           []byte(serverIP), // Presuming Server is also your router
			dhcp4.OptionDomainNameServer: []byte(serverIP), // Presuming Server is also your DNS server
		},
	}
	log.Fatal(dhcp4.ListenAndServe(handler))
	// log.Fatal(dhcp4.Serve(dhcp4.NewUDP4BoundListener("eth0",":67"), handler)) // Select interface on multi interface device - just linux for now
	// log.Fatal(dhcp4.Serve(dhcp4.NewUDP4FilterListener("en0",":67"), handler)) // Work around for other OSes
}

type lease struct {
	nic    string    // Client's CHAddr
	expiry time.Time // When the lease expires
}

type DHCPHandler struct {
	ip            net.IP        // Server IP to use
	options       dhcp4.Options // Options to send to DHCP Clients
	start         net.IP        // Start of IP range to distribute
	leaseRange    int           // Number of IPs to distribute (starting from start)
	leaseDuration time.Duration // Lease period
	leases        map[int]lease // Map to keep track of leases
}

func (h *DHCPHandler) ServeDHCP(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) (d dhcp4.Packet) {
	var opt = make(map[string]interface{})
	for key, val := range options {
		switch key {
		case dhcp4.OptionDHCPMessageType:
			u, err := binary.ReadUvarint(bytes.NewReader(val))
			if err != nil {
				opt[key.String()+"Error"] = err.Error()
				continue
			}
			opt[key.String()] = u
		case dhcp4.OptionParameterRequestList:
			fallthrough
		case dhcp4.OptionRequestedIPAddress:
			opt[key.String()] = net.IP{val[0], val[1], val[2], val[3]}.String()
		default:
			opt[key.String()] = string(val)
		}
	}
	data, err := json.Marshal(opt)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))

	switch msgType {

	case dhcp4.Discover:
		free, nic := -1, p.CHAddr().String()
		for i, v := range h.leases { // Find previous lease
			if v.nic == nic {
				free = i
				goto reply
			}
		}
		if free = h.freeLease(); free == -1 {
			return
		}
	reply:
		return dhcp4.ReplyPacket(p, dhcp4.Offer, h.ip, dhcp4.IPAdd(h.start, free), h.leaseDuration,
			h.options.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]))

	case dhcp4.Request:
		if server, ok := options[dhcp4.OptionServerIdentifier]; ok && !net.IP(server).Equal(h.ip) {
			return nil // Message not for this dhcp4 server
		}
		reqIP := net.IP(options[dhcp4.OptionRequestedIPAddress])
		if reqIP == nil {
			reqIP = net.IP(p.CIAddr())
		}

		if len(reqIP) == 4 && !reqIP.Equal(net.IPv4zero) {
			if leaseNum := dhcp4.IPRange(h.start, reqIP) - 1; leaseNum >= 0 && leaseNum < h.leaseRange {
				if l, exists := h.leases[leaseNum]; !exists || l.nic == p.CHAddr().String() {
					h.leases[leaseNum] = lease{nic: p.CHAddr().String(), expiry: time.Now().Add(h.leaseDuration)}
					return dhcp4.ReplyPacket(p, dhcp4.ACK, h.ip, reqIP, h.leaseDuration,
						h.options.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]))
				}
			}
		}
		return dhcp4.ReplyPacket(p, dhcp4.NAK, h.ip, nil, 0, nil)

	case dhcp4.Release, dhcp4.Decline:
		nic := p.CHAddr().String()
		for i, v := range h.leases {
			if v.nic == nic {
				delete(h.leases, i)
				break
			}
		}
	}
	return nil
}

func (h *DHCPHandler) freeLease() int {
	now := time.Now()
	b := rand.Intn(h.leaseRange) // Try random first
	for _, v := range [][]int{[]int{b, h.leaseRange}, []int{0, b}} {
		for i := v[0]; i < v[1]; i++ {
			if l, ok := h.leases[i]; !ok || l.expiry.Before(now) {
				return i
			}
		}
	}
	return -1
}

func main() {
	ExampleHandler()
}
