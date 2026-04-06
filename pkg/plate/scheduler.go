package plate

import (
	"context"
	"log"
	"net"
	"time"
)

// Start starts the plate runtime, which runs a goroutine for each data packet defined in the board. Each goroutine generates and sends packets at the specified period until the context is cancelled.
func (plate *PlateRuntime) Start(ctx context.Context) {

	for _, pkt := range plate.Packets {
		go pkt.Run(ctx, plate.UDPConn)
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
