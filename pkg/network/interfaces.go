package network

import (
	"fmt"
	"net"
	"strings"

	"github.com/Hyperloop-UPV/NATSOS/pkg/utils"
)

// External Interfaces

// GetExternalInterface returns a list of available network interfaces on the system.
func GetExternalInterface() ([]string, error) {

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(ifaces))

	for _, iface := range ifaces {
		names = append(names, iface.Name)
	}

	return names, nil
}

// doesInterfaceExist checks if the specified network interface exists on the system.
func doesInterfaceExist(iface string) (bool, error) {

	ifaces, err := GetExternalInterface()
	if err != nil {
		return false, err
	}

	return utils.Contains(ifaces, iface), nil
}

// SetupExternalInterface configures the specified network interface with the given IP address and brings it up.
func SetUpExternalInterface(iface string, ip string) error {

	// Check if the interface exists
	exists, err := doesInterfaceExist(iface)
	if err != nil {
		return fmt.Errorf("failed to check if interface exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("interface %s does not exist", iface)
	}

	// Remove any existing IP addresses from the interface
	if err := utils.RunCommand("ip", "addr", "flush", "dev", iface); err != nil {
		return err
	}

	// Add the new IP address to the interface
	if err := utils.RunCommand("ip", "addr", "add", ip, "dev", iface); err != nil {
		return err
	}

	// Bring the interface up
	if err := utils.RunCommand("ip", "link", "set", iface, "up"); err != nil {
		return err
	}

	return nil
}

// Dummy Interfaces

// SetupDummyInterface creates a dummy network interface with the specified name and IP address, and brings it up.
func SetUpDummyInterface(name string, ip string) error {

	dummyName := generateDummyInterfaceName(name)

	// Check if the ip address is valid
	if !IsValidIPv4(ip) {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	dummyIP := AddSubnetMask(ip, 16)

	// Create the dummy interface
	if err := utils.RunCommand("ip", "link", "add", dummyName, "type", "dummy"); err != nil {
		return err
	}

	// Add the IP address to the dummy interface
	if err := utils.RunCommand("ip", "addr", "add", dummyIP, "dev", dummyName); err != nil {
		return err
	}

	// Bring the dummy interface up
	if err := utils.RunCommand("ip", "link", "set", dummyName, "up"); err != nil {
		return err
	}

	return nil
}

func generateDummyInterfaceName(boardName string) string {

	// remmove spaces and special characters from the board name and convert it to uppercase
	boardName = strings.ReplaceAll(boardName, " ", "")
	boardName = strings.ReplaceAll(boardName, "-", "")
	boardName = strings.ReplaceAll(boardName, "_", "")
	boardName = strings.ToUpper(boardName)

	// maximum length of a network interface name is 15 characters, so we need to truncate the board name if it's too long (we also need to reserve some characters for the "dummy" prefix, so we limit it to 10 characters)
	if len(boardName) > 10 {
		boardName = boardName[:10]
	}

	// Prexis the dummy prefix to the board name to create the dummy interface name
	return fmt.Sprintf("dummy%s", boardName)
}
