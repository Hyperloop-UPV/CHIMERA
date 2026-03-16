package plate

import (
	"fmt"
	"net"
	"time"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/network"
)

// NewPlateRuntime creates a new PlateRuntime for the given board and remote address. It resolves the local address based on the board's IP and creates a UDP connection to the backend. The local address is created as a dummy IP before, so it doesn't need to be actually assigned to an interface. The backend will receive the packets sent by the plate runtime and forward them to the decodification
func NewPlateRuntime(board adj.Board, remoteAddr *net.UDPAddr, period time.Duration) (*PlateRuntime, error) {

	// Set up a dummy interface for the board
	interfaceName, err := network.SetUpDummyInterface(board.Name, board.IP)
	if err != nil {
		return nil, fmt.Errorf("failed to set up dummy interface for board %s: %v", board.Name, err)
	}

	// Resolve the local address for the board the IP of the board must have been created as a dummy IP before
	// :0 selects a random port
	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:0", board.IP))
	if err != nil {
		return nil, fmt.Errorf("error resolving local address: %v", err)
	}

	// Create the UDP connection to the backend
	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("error dialing UDP connection: %v", err)
	}

	// Return the plate runtime
	plate := &PlateRuntime{
		Board:              board,
		Conn:               conn,
		BoardInterfaceName: interfaceName,
	}

	plate.applyADJBoardConfig(period) // Default period of 1 second for all packets, can be customized later

	return plate, nil
}

// applyADJBoardConfig applies the configuration from the ADJ board to the plate runtime. It initializes the packets and measurements based on the ADJ board configuration.
func (plate *PlateRuntime) applyADJBoardConfig(period time.Duration) {

	// Initialize measurements
	plate.Measurements = make(map[MeasurementID]*MeasurementState)

	// Define each board
	for _, measure := range plate.Board.Measurements {
		plate.Measurements[MeasurementID(measure.Id)] = NewMeasurementState(measure)

	}

	// Initialize packets
	for _, pkt := range plate.Board.Packets {

		// Add only packets that are data packets
		if pkt.Type != "data" {
			continue
		}

		var measStates []*MeasurementState

		// For each variable in the packet, find the corresponding measurement state and add it to the packet runtime
		for _, measure := range pkt.Variables {
			if meas, exists := plate.Measurements[MeasurementID(measure.Id)]; exists {
				measStates = append(measStates, meas)
			}
		}

		plate.Packets = append(plate.Packets, &PacketRuntime{
			Packet:       pkt,
			Period:       period,
			Measurements: measStates,
		})
	}
}

func (plate *PlateRuntime) Delete() error {

	// Closes connection
	if err := plate.Conn.Close(); err != nil {
		return err
	}

	if err := network.DeleteInterface(plate.BoardInterfaceName); err != nil {
		return err
	}

	return nil
}
