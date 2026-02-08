package plate

import (
	"fmt"
	"net"

	"github.com/Hyperloop-UPV/NATSOS/pkg/adj"
)

// NewPlateRuntime creates a new PlateRuntime for the given board and remote address. It resolves the local address based on the board's IP and creates a UDP connection to the backend. The local address is created as a dummy IP before, so it doesn't need to be actually assigned to an interface. The backend will receive the packets sent by the plate runtime and forward them to the decodification
func NewPlateRuntime(board adj.Board, remoteAddr *net.UDPAddr) (*PlateRuntime, error) {

	// Resolve the local address for the board the IP of the board must have been created as a dummy IP before
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
	return &PlateRuntime{
		Board: board,
		Conn:  conn,
	}, nil
}
