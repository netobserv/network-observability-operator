package utils

import (
	"fmt"
	"net"
)

// GetSocket returns socket string in the correct format based on address family
func GetSocket(hostIP string, hostPort int) string {
	socket := fmt.Sprintf("%s:%d", hostIP, hostPort)
	ipAddr := net.ParseIP(hostIP)
	if ipAddr != nil && ipAddr.To4() == nil {
		socket = fmt.Sprintf("[%s]:%d", hostIP, hostPort)
	}
	return socket
}
