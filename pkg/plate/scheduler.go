package plate

import (
	"context"
	"fmt"
	"time"

	"github.com/Hyperloop-UPV/NATSOS/pkg/adj"
	"github.com/Hyperloop-UPV/NATSOS/pkg/generator"
)

// Start starts the plate runtime, which runs a goroutine for each data packet defined in the board. Each goroutine generates and sends packets at the specified period until the context is cancelled.
func (p *PlateRuntime) Start(ctx context.Context) {

	for _, pkt := range p.Board.Packets {

		if pkt.Type != "data" {
			continue
		}

		go p.runPacket(ctx, pkt)
	}
}

// runPacket runs a goroutine that generates and sends packets at the specified period until the context is cancelled. Uses a ticker.
func (p *PlateRuntime) runPacket(ctx context.Context, packet adj.Packet) {

	ticker := time.NewTicker(time.Second) //TODO: use cfg period, currently hardcoded to 1 second for testing
	defer ticker.Stop()

	fmt.Printf("Starting packet %s with period %d ms\n", packet.Name, 7)

	generator := generator.NewRandomGenerator() // TODO: allow configuring different generators, maybe per-board or per-packet

	for {

		select {

		case <-ctx.Done():
			return

		case <-ticker.C:

			// Generate the packet following ADJ convenctions and using the generator
			packetBytes, err := generator.Generate(p.Board.Name, packet)
			if err != nil {
				continue
			}

			// Send the packet
			p.Conn.Write(packetBytes)
		}
	}
}
