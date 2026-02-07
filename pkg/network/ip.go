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
