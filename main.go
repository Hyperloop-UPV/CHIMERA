package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Hyperloop-UPV/NATSOS/pkg/adj"
	"github.com/Hyperloop-UPV/NATSOS/pkg/config"
	"github.com/Hyperloop-UPV/NATSOS/pkg/network"
	"github.com/Hyperloop-UPV/NATSOS/pkg/plate"
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

	// get the ADJ branch from the configuration and print it
	adj, err := adj.NewADJ(cfg.ADJBranch, cfg.ADJPath)
	if err != nil {
		log.Fatalf("Failed to initialize ADJ: %v at %s", err, cfg.ADJPath)
	}

	// Set up the network configuration
	if err := network.SetUpNetwork(cfg.Network.Interface, "192.168.0.1/16"); err != nil {
		log.Fatalf("Failed to setup network: %v", err)
	}

	// Define context for the plate runtimes
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure the boards and create plate runtimes
	err = configureBoards(adj, ctx)
	if err != nil {
		log.Fatalf("Failed to configure boards: %v", err)
	}

	// Block forever
	select {}

}

func configureBoards(adj adj.ADJ, ctx context.Context) error {

	// Obtain backend address from configuration
	backendAddr, err := net.ResolveUDPAddr("udp", network.FormatIP(adj.Info.Addresses["backend"], int(adj.Info.Ports["UDP"])))
	if err != nil {
		return fmt.Errorf("failed to resolve backend address: %v", err)
	}

	// For each board
	for _, board := range adj.Boards {

		// Set up a dummy interface for the board
		err := network.SetUpDummyInterface(board.Name, board.IP)
		if err != nil {
			return fmt.Errorf("failed to set up dummy interface for board %s: %v", board.Name, err)
		}

		// Create a plate runtime for the board
		plateRuntime, err := plate.NewPlateRuntime(board, backendAddr, 100*time.Millisecond) // Default period of 100ms for all packets, can be customized later
		if err != nil {
			return fmt.Errorf("failed to create plate runtime for board %s: %v", board.Name, err)
		}

		// Start the plate runtime
		plateRuntime.Start(ctx)
		log.Printf("Plate runtime created for board %s", plateRuntime.Board.Name)
	}

	return nil

}
