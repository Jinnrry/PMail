package id

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

var ip_instance string

func getLocalIP() string {
	if ip_instance != "" {
		return ip_instance
	}

	ip := "127.0.0.1"
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		ip_instance = ip
		return ip
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				break
			}
		}
	}
	ip_instance = ip
	return ip
}

func GenLogID() string {
	r := rand.New(rand.NewSource(time.Now().UnixMicro()))

	ip := getLocalIP()

	now := time.Now()
	timestamp := uint32(now.Unix())
	timeNano := now.UnixNano()
	pid := os.Getpid()
	b := bytes.Buffer{}

	b.WriteString(hex.EncodeToString(net.ParseIP(ip).To4()))
	b.WriteString(fmt.Sprintf("%x", timestamp&0xffffffff))
	b.WriteString(fmt.Sprintf("%04x", timeNano&0xffff))
	b.WriteString(fmt.Sprintf("%04x", pid&0xffff))
	b.WriteString(fmt.Sprintf("%06x", r.Int31n(1<<24)))
	b.WriteString("b0")

	return b.String()
}
