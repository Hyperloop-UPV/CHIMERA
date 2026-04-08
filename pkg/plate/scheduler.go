package plate

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

const purple = "\033[35m"
const reset = "\033[0m"

// Start starts the plate runtime, which runs a goroutine for each data packet defined in the board. Each goroutine generates and sends packets at the specified period until the context is cancelled.
func (plate *PlateRuntime) Start(ctx context.Context) {

	for _, pkt := range plate.Packets {
		go pkt.Run(ctx, plate.UDPConn)
	}

	if plate.TCPListener != nil {
		go plate.acceptTCP(ctx)
	}
}

func (plate *PlateRuntime) acceptTCP(ctx context.Context) {
	for {
		conn, err := plate.TCPListener.AcceptTCP()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("board %s TCP accept error: %v", plate.Board.Name, err)
				return
			}
		}
		select {
		case plate.EventCh <- fmt.Sprintf("[TCP] Board %s: connection from %s", plate.Board.Name, conn.RemoteAddr()):
		default:
		}
		go func(c *net.TCPConn) {
			defer func() {
				select {
				case plate.EventCh <- fmt.Sprintf("[TCP] Board %s: disconnected %s", plate.Board.Name, c.RemoteAddr()):
				default:
				}
				c.Close()
			}()
			buf := make([]byte, 256)
			for {
				if _, err := c.Read(buf); err != nil {
					return
				}
			}
		}(conn)
	}
}

func (pkt *PacketRuntime) Run(ctx context.Context, conn *net.UDPConn) {

	// Use a ticker to generate packets at the specified period
	pkt.mu.RLock()
	period := pkt.Period
	pkt.mu.RUnlock()

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:

			payload, err := pkt.BuildPayload()
			if err != nil {
				// Log the error so that we can see why packet generation stopped.
				// It is better to keep trying than stop sending entirely.
				log.Printf("packet %d build error: %v", pkt.Packet.Id, err)
				continue
			}

			conn.Write(payload)
		}
	}
}
