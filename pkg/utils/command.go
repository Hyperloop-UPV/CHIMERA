package utils

import (
	"fmt"
	"os/exec"
)

// runCaommands runs a command in the terminal
func runCommand(log bool, name string, args ...string) error {

	// execute command
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()

	// log if required
	if log {
		fmt.Printf("Running: %s %v\n", name, args)
		fmt.Println(string(output))
	}

	// if fails returns an error
	if err != nil {
		return fmt.Errorf("command failed: %s %v\n%s: %w",
			name,
			args,
			string(output),
			err,
		)
	}

	return nil
}

// RunCommand executes a command printing it
func RunCommand(name string, args ...string) error {
	return runCommand(true, name, args...)
}

// RunCommandSilent executes a command
func RunCommandSilent(name string, args ...string) error {
	return runCommand(false, name, args...)
}
