package plate

import (
	"fmt"
	"net"
	"time"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/network"
)

// NewPlateRuntime creates a new PlateRuntime for the given board and remote address. It resolves the local address based on the board's IP and creates a UDP connection to the backend. The local address is created as a dummy IP before, so it doesn't need to be actually assigned to an interface. The backend will receive the packets sent by the plate runtime and forward them to the decodification
func NewPlateRuntime(board adj.Board, remoteAddrUDP *net.UDPAddr, portTCP uint16, period time.Duration) (*PlateRuntime, error) {

	// Create plate runtime
	plate := &PlateRuntime{
		Board: board,
	}

	// Create dummy interface
	if err := plate.createInterface(); err != nil {
		return nil, fmt.Errorf("failed to create dummy interface for board %s: %v", board.Name, err)
	}

	// UDP
	if err := plate.setupUDPConnection(remoteAddrUDP); err != nil {
		plate.Delete() // Clean up the created interface
		return nil, fmt.Errorf("failed to set up UDP connection for board %s: %v", board.Name, err)
	}

	// TCP
	if err := plate.setupTCPConnection(portTCP); err != nil {
		plate.Delete() // Clean up the created interface
		return nil, fmt.Errorf("failed to set up TCP connection for board %s: %v", board.Name, err)
	}

	// Apply ADJ board configuration to the plate runtime
	plate.applyADJBoardConfig(period) // Default period of 1 second for all packets, can be customized later

	return plate, nil
}

// setupUDPConnection sets up a UDP connection to the backend
func (plate *PlateRuntime) setupUDPConnection(remoteAddr *net.UDPAddr) error {
	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:0", plate.Board.IP))
	if err != nil {
		return fmt.Errorf("error resolving local address: %v", err)
	}

	conn, err := net.DialUDP("udp", localAddr, remoteAddr)
	if err != nil {
		return fmt.Errorf("error dialing UDP connection: %v", err)
	}

	plate.UDPConn = conn
	return nil
}

// setupTCPConnection creates a TCP listener on the board IP and port
func (plate *PlateRuntime) setupTCPConnection(portTCP uint16) error {

	// TCP
	plateAddrTCP, err := net.ResolveTCPAddr("tcp", network.FormatIP(plate.Board.IP, int(portTCP)))
	if err != nil {
		return fmt.Errorf("failed to resolve TCP listen address: %v", err)
	}

	listener, err := net.ListenTCP("tcp", plateAddrTCP)
	if err != nil {
		return fmt.Errorf("error creating TCP listener: %v", err)
	}

	plate.TCPListener = listener
	return nil
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

// createInterface creates a dummy interface
func (plate *PlateRuntime) createInterface() error {
	interfaceName, err := network.SetUpDummyInterface(plate.Board.Name, plate.Board.IP)
	if err != nil {
		return fmt.Errorf("failed to set up dummy interface for board %s: %v", plate.Board.Name, err)
	}

	plate.BoardInterfaceName = interfaceName
	return nil
}

// Delete cleans up the plate runtime by closing connections and deleting the dummy interface.
func (plate *PlateRuntime) Delete() error {

	// Close UDP connection
	if plate.UDPConn != nil {
		if err := plate.UDPConn.Close(); err != nil {
			return err
		}
	}

	// Close TCP listener if created
	if plate.TCPListener != nil {
		if err := plate.TCPListener.Close(); err != nil {
			return err
		}
	}

	if err := network.DeleteInterface(plate.BoardInterfaceName); err != nil {
		return err
	}

	return nil
}
