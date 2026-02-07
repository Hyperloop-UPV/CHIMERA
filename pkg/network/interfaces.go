package network

import (
	"fmt"
	"net"
	"strings"

	"github.com/Hyperloop-UPV/NATSOS/pkg/utils"
)

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
func SetupExternalInterface(iface string, ip string) error {

	// Check if the interface exists
	exists, err := doesInterfaceExist(iface)
	if err != nil {
		return fmt.Errorf("failed to check if interface exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("interface %s does not exist", iface)
	}

	if err := utils.RunCommand("ip", "addr", "flush", "dev", iface); err != nil {
		return err
	}

	if err := utils.RunCommand("ip", "addr", "add", ip, "dev", iface); err != nil {
		return err
	}

	if err := utils.RunCommand("ip", "link", "set", iface, "up"); err != nil {
		return err
	}

	return nil
}

// SetupDummyInterface creates a dummy network interface with the specified name and IP address, and brings it up.
func SetupDummyInterface(name string, ip string) error {

	dummyName := generateDummyInterfaceName(name)

	// Check if the ip address is valid
	if !IsValidIPv4(ip) {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	dummyIP := AddSubnetMask(ip, 16)

	if err := utils.RunCommand("ip", "link", "add", dummyName, "type", "dummy"); err != nil {
		return err
	}

	if err := utils.RunCommand("ip", "addr", "add", dummyIP, "dev", dummyName); err != nil {
		return err
	}

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

	// maximum length of a network interface name is 15 characters, so we need to truncate the board name if it's too long
	if len(boardName) > 10 {
		boardName = boardName[:10]
	}

	return fmt.Sprintf("dummy%s", boardName)
}
