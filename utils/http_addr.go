package utils

import (
	"net"
	"strings"
)

// GetLocalIP get server's IP address
func GetLocalIP(ipClass string) (string, error) {
	netInterfaceAddresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, netInterfaceAddress := range netInterfaceAddresses {
		networkIP, ok := netInterfaceAddress.(*net.IPNet)
		if ok && !networkIP.IP.IsLoopback() && networkIP.IP.To4() != nil {
			ip := networkIP.IP.String()
			if strings.Contains(ip, ipClass) == true {
				return ip, nil
			}
		}
	}
	return "", nil
}
