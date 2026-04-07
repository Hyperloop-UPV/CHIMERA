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
	//ADJ data
	Board adj.Board

	status PlateStatus

	//Connection data
	UDPConn     *net.UDPConn
	TCPListener *net.TCPListener

	//Interface name of the dummy interface created for the board, used for cleanup
	boardInterfaceName string
	ipAddressCIDR      string

	// Runtime data
	Packets      []*PacketRuntime
	Measurements map[MeasurementID]*MeasurementState // Map of measurement name to its state, for easy access and updates
}

// status of the plate runtime

type PlateStatus int

const (
	StatusOK PlateStatus = iota
	StatusUnavailable
)

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
