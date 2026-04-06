package plate

import (
	"net"
	"sync"
	"time"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/generator"
)

// Define MeasurementID as an string
type MeasurementID string

type PlateGenerators map[string]PlateRuntime

// PlateRuntime is the main struct for the plate runtime. It contains the board and the connection to the backend

type PlateRuntime struct {
	Board   adj.Board
	UDPConn *net.UDPConn

	BoardInterfaceName string

	Packets      []*PacketRuntime
	Measurements map[MeasurementID]*MeasurementState // Map of measurement name to its state, for easy access and updates
}

type PacketRuntime struct {
	Packet adj.Packet
	Period time.Duration

	Measurements []*MeasurementState

	mu sync.RWMutex
}

type MeasurementState struct {
	Measurement adj.Measurement
	Generator   generator.Generator

	mu sync.RWMutex
}
