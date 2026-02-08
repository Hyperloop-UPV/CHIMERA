package plate

import (
	"context"
	"net"
	"time"
)

// Start starts the plate runtime, which runs a goroutine for each data packet defined in the board. Each goroutine generates and sends packets at the specified period until the context is cancelled.
func (plate *PlateRuntime) Start(ctx context.Context) {

	for _, pkt := range plate.packets {
		go pkt.Run(ctx, plate.Conn)
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
				continue
			}

			conn.Write(payload)
		}
	}
}
