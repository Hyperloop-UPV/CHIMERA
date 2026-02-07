package network

import (
	"fmt"
	"os/exec"
)

// Runs a command and prints its output. Returns an error if the command fails.
func RunCommand(name string, args ...string) error {

	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()

	fmt.Printf("Running: %s %v\n", name, args)
	fmt.Println(string(output))

	return err
}
