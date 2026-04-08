package control

import "strings"

// Each command is an array of words
type Command []string

// ParseCommand given a string line returns its command
func ParseCommand(line string) Command {
	return strings.Fields(strings.TrimSpace(line))
}

// ParseCommands splits a line into multiple commands separated by semicolons.
func ParseCommands(line string) []Command {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	parts := strings.Split(line, ";")
	commands := make([]Command, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			commands = append(commands, ParseCommand(trimmed))
		}
	}

	return commands
}
