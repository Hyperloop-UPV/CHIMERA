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
		Board:              board,
		boardInterfaceName: network.GenerateDummyInterfaceName(board.Name),
		ipAddressCIDR:      network.AddSubnetMask(board.IP, 24),
		status:             StatusOK,
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

/*****************
* Public methods *
*****************/

// AbruptlyClose forcefully closes the plate runtime connection by removing its IP.
// Due to kernel limitations if you set down the iface while the connection is active, the connection will still be active
func (plate *PlateRuntime) AbruptlyClose() error {

	// Remove the IP address from the interface to forcefully close the connection
	if err := network.DeleteIPFromInterface(plate.boardInterfaceName, plate.ipAddressCIDR); err != nil {
		return fmt.Errorf("failed to delete IP from interface: %w", err)
	}
	plate.status = StatusUnavailable
	return nil
}

// RestoreIP restores the IP address to the dummy interface, allowing the plate runtime to re-establish the connection if needed.
func (plate *PlateRuntime) RestoreIP() error {

	// Add the IP address back to the interface to restore the connection
	if err := network.AddIPToInterface(plate.boardInterfaceName, plate.ipAddressCIDR); err != nil {
		return fmt.Errorf("failed to add IP to interface: %w", err)
	}

	plate.status = StatusOK

	return nil
}

// Delete cleans up the plate runtime by closing connections and deleting the dummy interface.
func (plate *PlateRuntime) Delete() error {
	var errs []error

	if plate.UDPConn != nil {
		if err := plate.UDPConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("udp close failed: %w", err))
		}
		plate.UDPConn = nil
	}

	if plate.TCPListener != nil {
		if err := plate.TCPListener.Close(); err != nil {
			errs = append(errs, fmt.Errorf("tcp close failed: %w", err))
		}
		plate.TCPListener = nil
	}

	if plate.boardInterfaceName != "" {
		if err := network.DeleteInterface(plate.boardInterfaceName); err != nil {
			errs = append(errs, fmt.Errorf("delete interface failed: %w", err))
		}
		plate.boardInterfaceName = ""
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return fmt.Errorf("delete encountered multiple errors: %v", errs)
	}
}

// GetStatus returns the current status of the plate runtime, indicating whether it is operational or unavailable.
func (plate *PlateRuntime) GetStatus() PlateStatus {
	return plate.status
}

/******************
* Private methods *
******************/

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
	err := network.SetUpDummyInterface(plate.boardInterfaceName, plate.Board.IP)
	if err != nil {
		return fmt.Errorf("failed to set up dummy interface for board %s: %v", plate.Board.Name, err)
	}

	return nil
}
