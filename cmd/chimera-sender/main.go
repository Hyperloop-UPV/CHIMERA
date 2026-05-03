package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/config"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/control"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/network"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/plate"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/utils"
)

// chimera-sender is the desktop TUI variant of CHIMERA. It creates dummy
// interfaces for each board and emits packets, but does NOT alter the host
// network IP. Only the local TUI mode is available — no daemon, no remote.
func main() {
	configFile := flag.String("config", "config-sender.json", "path to the configuration file")
	verboseFlag := flag.Bool("verbose", false, "log every shell command executed and its output")
	flag.Parse()

	utils.SetVerbose(*verboseFlag)

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	adjData, err := adj.NewADJ(cfg.ADJBranch, cfg.ADJPath)
	if err != nil {
		log.Fatalf("Failed to initialize ADJ: %v at %s", err, cfg.ADJPath)
	}

	log.Printf("ADJ branch: %s, commit: %s", adjData.Branch, adjData.CommitHash)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	boardGenerator, err := configureBoards(adjData, *cfg, ctx)
	if err != nil {
		log.Fatalf("Failed to configure boards: %v", err)
	}

	log.Println("Starting CHIMERA sender (TUI mode)")
	control.StartControlServer(cfg.Network.ControlPort, boardGenerator, adjData.Branch, adjData.CommitHash)
}

func configureBoards(adjData adj.ADJ, cfg config.Config, ctx context.Context) (plate.PlateGenerators, error) {
	backendAddrUDP, err := net.ResolveUDPAddr("udp", network.FormatIP(adjData.Info.Addresses["backend"], int(adjData.Info.Ports["UDP"])))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve backend UDP address: %v", err)
	}

	portTCP := adjData.Info.Ports["TCP_SERVER"]
	runtimeGenerators := make(plate.PlateGenerators)
	period := time.Duration(cfg.InitialPeriod) * time.Millisecond

	for _, board := range adjData.Boards {
		plateRuntime, err := plate.NewPlateRuntime(board, backendAddrUDP, portTCP, period, adjData.Info.MessageIds)
		if err != nil {
			return nil, fmt.Errorf("failed to create plate runtime for board %s: %v", board.Name, err)
		}
		plateRuntime.Start(ctx)
		log.Printf("Plate runtime created for board %s", plateRuntime.Board.Name)
		runtimeGenerators[board.Name] = plateRuntime
	}

	return runtimeGenerators, nil
}
