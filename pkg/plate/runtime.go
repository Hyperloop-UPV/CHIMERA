package plate

import (
	mrand "math/rand"
	"net"
	"sync"
	"time"

	"github.com/Hyperloop-UPV/NATSOS/pkg/adj"
)

// Define MeasurementID as an string
type MeasurementID string

// PlateRuntime is the main struct for the plate runtime. It contains the board and the connection to the backend

type PlateRuntime struct {
	Board adj.Board
	Conn  *net.UDPConn

	packets      []*PacketRuntime
	measurements map[MeasurementID]*MeasurementState // Map of measurement name to its state, for easy access and updates
}

type PacketRuntime struct {
	Packet adj.Packet
	Period time.Duration

	Measurements []*MeasurementState

	mu sync.RWMutex
}

type MeasurementState struct {
	Measurement adj.Measurement
	Generator   *mrand.Rand

	mu sync.RWMutex
}
