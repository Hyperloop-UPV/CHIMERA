package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the base configuration structure for the application.
// Extend this struct with your specific configuration fields.
type Config struct {
	// Application settings
	ADJBranch string `json:"adj-branch"`
}

// LoadConfig reads and parses a JSON configuration file.
// It takes the file path as a parameter and returns a Config struct or an error.
func LoadConfig(filePath string) (*Config, error) {
	// Read the configuration file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the JSON data into the Config struct
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// String returns a string representation of the configuration for logging.
func (c *Config) String() string {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Sprintf("Config{Error: %v}", err)
	}
	return string(data)
}
