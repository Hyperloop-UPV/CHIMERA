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
	"github.com/Hyperloop-UPV/NATSOS/pkg/control"
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
	if err := network.SetUpNetwork(cfg.Network.Interface, cfg.Network.IP); err != nil {
		log.Fatalf("Failed to setup network: %v", err)
	}

	// Define context for the plate runtimes
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure the boards and create plate runtimes
	boardGenerator, err := configureBoards(adj, *cfg, ctx)
	if err != nil {
		log.Fatalf("Failed to configure boards: %v", err)
	}

	go control.StartControlServer(cfg.Network.ControlPort, boardGenerator)

	// Block forever
	select {}

}

func configureBoards(adj adj.ADJ, cfg config.Config, ctx context.Context) (plate.PlateGenerators, error) {

	// Obtain backend address from configuration
	backendAddr, err := net.ResolveUDPAddr("udp", network.FormatIP(adj.Info.Addresses["backend"], int(adj.Info.Ports["UDP"])))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve backend address: %v", err)
	}

	// generator runtime

	runtimeGenerators := make(map[string]plate.PlateRuntime)

	// define period
	period := time.Duration(cfg.InitialPeriod) * time.Millisecond

	// For each board
	for _, board := range adj.Boards {

		// Set up a dummy interface for the board
		err := network.SetUpDummyInterface(board.Name, board.IP)
		if err != nil {
			return nil, fmt.Errorf("failed to set up dummy interface for board %s: %v", board.Name, err)
		}

		// Create a plate runtime for the board
		plateRuntime, err := plate.NewPlateRuntime(board, backendAddr, period)
		if err != nil {
			return nil, fmt.Errorf("failed to create plate runtime for board %s: %v", board.Name, err)
		}

		// Start the plate runtime
		plateRuntime.Start(ctx)
		log.Printf("Plate runtime created for board %s", plateRuntime.Board.Name)

		// Store board

		runtimeGenerators[board.Name] = *plateRuntime
	}

	return runtimeGenerators, nil

}
