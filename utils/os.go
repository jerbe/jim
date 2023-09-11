package utils

import (
	"net"
	"os"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/9/11 13:49
  @describe :
*/

var (
	_hostname = "localhost"
)

// GetHostname 获取本机名
func GetHostname() string {
	if _hostname != "localhost" {
		return _hostname
	}
	_hostname, _ = os.Hostname()
	return _hostname
}

// GetLocalIP 获取本机IP地址
func GetLocalIP() string {
	// 获取本机 IP 地址
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			// 检查 IP 地址是否为 IPv4 或 IPv6 地址
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					return ipNet.IP.String()
				} else if ipNet.IP.To16() != nil {
					return ipNet.IP.String()
				}
			}
		}
	}
	return ""
}

func GetLocalIPs() []string {
	ips := make([]string, 0)
	// 获取本机 IP 地址
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			// 检查 IP 地址是否为 IPv4 或 IPv6 地址
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ips = append(ips, ipNet.IP.String())
				} else if ipNet.IP.To16() != nil {
					ips = append(ips, ipNet.IP.String())
				}
			}
		}
	}
	return ips
}
