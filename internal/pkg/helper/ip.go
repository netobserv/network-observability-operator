package helper

import "net"

// IPToCIDR converts an IP address to a CIDR with proper prefix length
// IPv4 addresses get /32, IPv6 addresses get /128
func IPToCIDR(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}

	// Check if it's IPv4 (net.IP.To4() returns nil for IPv6)
	if ip.To4() != nil {
		return ipStr + "/32"
	}
	// IPv6
	return ipStr + "/128"
}
