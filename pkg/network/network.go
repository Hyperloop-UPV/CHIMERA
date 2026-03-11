package network

import (
	"log"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/utils"
)

// Set up Network Configuration

// SetupNetwork configures the network settings for the system, including modifying Linux Kernel
func SetUpNetwork(iface string, ip string) error {

	// Set external network interface
	if err := SetUpExternalInterface(iface, ip); err != nil {
		return err
	}

	// Allow the use of dummy interfaces
	if err := utils.RunCommandSilent("modprobe", "dummy"); err != nil {
		return err
	}

	// Enable IP forwarding
	if err := utils.RunCommandSilent("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return err
	}

	log.Printf("Network configured successfully with interface %s and IP %s", iface, ip)
	return nil
}
