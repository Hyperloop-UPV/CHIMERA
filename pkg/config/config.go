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
	ADJBranch     string  `json:"adj-branch"`
	ADJPath       string  `json:"adj-path"` // Path to the ADJ repository empty for download usingg git
	Network       Network `json:"network"`
	InitialPeriod uint64  `json:"initial-period-ms"` // Initial period for packet generation in milliseconds
}

type Network struct {
	Interface   string `json:"interface"`
	ControlPort int    `json:"control-port"`
	IP          string `json:"ip"` // ip of the device
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

	// Perform additional checks on the configuration values
	err = additionalChecks(&cfg)
	if err != nil {
		return nil, err
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

// additional Checks ensure that config.json match specifications
func additionalChecks(cfg *Config) error {

	// Initial period must be positive
	if cfg.InitialPeriod <= 0 {
		return fmt.Errorf(
			"Config{Error: initial-period-ms must be a positive integer higher than 0 and %d is not}",
			cfg.InitialPeriod,
		)
	}

	return nil
}
