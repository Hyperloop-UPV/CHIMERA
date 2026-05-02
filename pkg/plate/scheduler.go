package plate

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

// Start starts the plate runtime, which runs a goroutine for each data packet defined in the board. Each goroutine generates and sends packets at the specified period until the context is cancelled.
func (plate *PlateRuntime) Start(ctx context.Context) {

	for _, pkt := range plate.Packets {
		go pkt.Run(ctx, plate.UDPConn)
	}

	if plate.TCPListener != nil {
		go plate.acceptTCP(ctx)
	}
}

const tcpReadBufferSize = 256

// acceptTCP loops accepting incoming TCP connections and dispatches each one
// to its own goroutine. It exits when the listener is closed (typically on
// ctx cancellation).
func (plate *PlateRuntime) acceptTCP(ctx context.Context) {
	for {
		conn, err := plate.TCPListener.AcceptTCP()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("board %s TCP accept error: %v", plate.Board.Name, err)
			return
		}

		plate.emitEvent(EventTCP, "[TCP] Board %s: connection from %s", plate.Board.Name, conn.RemoteAddr())
		go plate.handleTCPConnection(conn)
	}
}

// handleTCPConnection reads orders from a single TCP connection until it is
// closed or errors out, decoding and forwarding each one to the event stream.
func (plate *PlateRuntime) handleTCPConnection(conn *net.TCPConn) {
	defer func() {
		plate.emitEvent(EventTCP, "[TCP] Board %s: disconnected %s", plate.Board.Name, conn.RemoteAddr())
		conn.Close()
	}()

	buf := make([]byte, tcpReadBufferSize)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}

		plate.emitEvent(EventOrder, "[ORDER] Board %s: %s", plate.Board.Name, plate.decodeOrder(buf[:n]))
	}
}

// decodeOrder turns a raw TCP payload into a human-readable string, falling
// back to a hex dump if the decoder is unavailable or fails.
func (plate *PlateRuntime) decodeOrder(payload []byte) string {
	if plate.Decoder == nil {
		return fmt.Sprintf("<no decoder> %x", payload)
	}

	decoded, err := plate.Decoder.Decode(payload)
	if err != nil {
		return fmt.Sprintf("decode error: %v", err)
	}
	return decoded
}

// emitEvent pushes a categorized event onto the event channel without blocking.
func (plate *PlateRuntime) emitEvent(kind EventKind, format string, args ...any) {
	select {
	case plate.EventCh <- Event{Kind: kind, Message: fmt.Sprintf(format, args...)}:
	default:
	}
}

func (pkt *PacketRuntime) Run(ctx context.Context, conn *net.UDPConn) {

	// Use a ticker to generate packets at the specified period
	pkt.mu.RLock()
	period := pkt.Period
	pkt.mu.RUnlock()

	// creates a ticker that ticks at the specified period.
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	// Loop that generates and sends packets at each tick until the context is cancelled.
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
