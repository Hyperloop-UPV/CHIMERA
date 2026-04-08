package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/config"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/control"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/network"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/plate"
)

func main() {

	// Get the configuration file path from command line arguments
	configFile := flag.String("config", "config.json", "path to the configuration file")
	modeFlag := flag.String("mode", "daemon", "run mode: daemon or tui")
	flag.Parse()

	mode := strings.ToLower(strings.TrimSpace(*modeFlag))
	if flag.NArg() > 0 {
		mode = strings.ToLower(strings.TrimSpace(flag.Arg(0)))
	}

	if mode != "daemon" && mode != "tui" {
		log.Fatalf("Unknown mode %q. Use 'daemon' or 'tui'", mode)
	}

	// Load the configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if mode == "tui" {
		log.Println("Connecting to CHIMERA daemon for remote TUI")
		if err := control.StartRemoteTUI(cfg.Network.ControlPort); err != nil {
			log.Fatalf("Failed to start TUI client: %v", err)
		}
		return
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
	// Context that cancels on Ctrl+C or SIGTERM
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Configure the boards and create plate runtimes
	boardGenerator, err := configureBoards(adj, *cfg, ctx)
	if err != nil {
		log.Fatalf("Failed to configure boards: %v", err)
	}

	log.Println("Starting CHIMERA in daemon mode")
	go func() {
		if err := control.StartControlDaemon(cfg.Network.ControlPort, boardGenerator, ctx.Done()); err != nil {
			log.Fatalf("Control daemon failed: %v", err)
		}
	}()

	// Wait until Ctrl+C
	<-ctx.Done()

	log.Println("Shutting down")

	for _, plate := range boardGenerator {
		plate.Delete()
	}

	log.Println("Shutdown complete")
}

func configureBoards(adj adj.ADJ, cfg config.Config, ctx context.Context) (plate.PlateGenerators, error) {

	// Obtain backend address from configuration
	backendAddrUDP, err := net.ResolveUDPAddr("udp", network.FormatIP(adj.Info.Addresses["backend"], int(adj.Info.Ports["UDP"])))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve backend UDP address: %v", err)
	}

	portTCP := adj.Info.Ports["TCP_SERVER"]

	// generator runtime

	runtimeGenerators := make(plate.PlateGenerators)

	// define period
	period := time.Duration(cfg.InitialPeriod) * time.Millisecond

	// For each board
	for _, board := range adj.Boards {

		// Create a plate
		plateRuntime, err := plate.NewPlateRuntime(board, backendAddrUDP, portTCP, period)
		if err != nil {
			return nil, fmt.Errorf("failed to create plate runtime for board %s: %v", board.Name, err)
		}

		// Start the plate runtime
		plateRuntime.Start(ctx)
		log.Printf("Plate runtime created for board %s", plateRuntime.Board.Name)

		// Store board

		runtimeGenerators[board.Name] = plateRuntime
	}

	return runtimeGenerators, nil

}
