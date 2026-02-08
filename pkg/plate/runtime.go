package plate

import (
	"net"

	"github.com/Hyperloop-UPV/NATSOS/pkg/adj"
)

// PlateRuntime is the main struct for the plate runtime. It contains the board and the connection to the backend

type PlateRuntime struct {
	Board adj.Board
	Conn  *net.UDPConn
}
