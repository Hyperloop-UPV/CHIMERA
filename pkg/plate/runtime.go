package plate

import (
	"net"
	"sync"
	"time"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/decoder"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/generator"
)

// Define MeasurementID as an string
type MeasurementID string

type PlateGenerators map[string]*PlateRuntime

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

	// EventCh receives categorized events (TCP connection lifecycle, decoded
	// orders, ...) for display in the TUI.
	EventCh chan Event

	// Decoder turns raw TCP payloads into human-readable orders for the TUI
	Decoder *decoder.Decoder
}

// EventKind classifies plate events so consumers (TUI, remote) can render
// them with the appropriate style.
type EventKind int

const (
	EventTCP   EventKind = iota // TCP connection lifecycle
	EventOrder                  // decoded order received from the backend
)

// Event is a single categorized message produced by a PlateRuntime.
type Event struct {
	Kind    EventKind
	Message string
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
