package network

import (
	"net"
	"strconv"
)

// IsValidIPv4 checks if the provided string is a valid IPv4 address.
func IsValidIPv4(ip string) bool {

	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}

	return parsed.To4() != nil
}

// AddSubnetMask adds a subnet mask to an IP address if it doesn't already have one, using a default mask of /24 if the provided mask is invalid.
func AddSubnetMask(ip string, mask int) string {

	// check that mask is between 0 and 32º
	if mask < 0 || mask > 32 {
		mask = 24
	}

	return ip + "/" + strconv.Itoa(mask)
}

// IsValidPort checks if the provided port number is valid (between 1 and 65535).
func IsValidPort(port int) bool {

	return port > 0 && port < 65536
}

// FormatIP formats an IP address and port into a string suitable for network operations, using the provided IP and port. If the port is invalid, it defaults to 0 (which means any available port).
func FormatIP(ip string, port int) string {

	if !IsValidPort(port) {
		port = 0
	}

	return net.JoinHostPort(ip, strconv.Itoa(port))
}
