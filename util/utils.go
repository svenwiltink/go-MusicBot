package util

import (
	"errors"
	"fmt"
	"net"
	"time"
)

func FormatDuration(duration time.Duration) (form string) {
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) - (minutes * 60)

	if minutes > 60 {
		hours := minutes / 60
		minutes -= hours * 60
		form = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		return
	}
	form = fmt.Sprintf("%02d:%02d", minutes, seconds)
	return
}

func GetExternalIPs() (ips []net.IP, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		var addrs []net.Addr
		addrs, err = iface.Addrs()
		if err != nil {
			return
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			ips = append(ips, ip)
		}
	}
	if len(ips) == 0 {
		err = errors.New("are you connected to the network?")
	}
	return
}
