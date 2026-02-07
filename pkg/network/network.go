package network

import "github.com/Hyperloop-UPV/NATSOS/pkg/utils"

// Set up Network Configuration

// SetupNetwork configures the network settings for the system, including modifying Linux Kernel
func SetupNetwork() error {

	// Allow the use of dummy interfaces
	if err := utils.RunCommand("modprobe", "dummy"); err != nil {
		return err
	}

	// Enable IP forwarding
	if err := utils.RunCommand("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return err
	}

	return nil
}
