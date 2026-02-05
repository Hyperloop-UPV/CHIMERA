package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Hyperloop-UPV/NATSOS/pkg/config"
)

func main() {

	// Get the configuration file path from command line arguments
	configFile := flag.String("config", "config.json", "path to the configuration file")
	flag.Parse()

	// Load the configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Configuration loaded successfully:\n%s\n", cfg.ADJBranch)
}
