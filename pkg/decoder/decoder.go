// Package decoder turns raw TCP payloads sent by the backend into a
// human-readable representation of the order so the TUI (local and remote)
// can print which command each board has received.
//
// Wire format (Add/Remove State Orders):
//
//	bytes 0..1 : message id (uint16 LE)         — add_state_order / remove_state_order
//	bytes 2..3 : count       (uint16 LE)         — number of order ids
//	bytes 4..  : order ids   (uint16 LE * count) — each one references an order packet
package decoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
)

// Decoder knows the order packets of a single board (indexed by id) and the
// global "add/remove state order" message ids, so it can resolve an incoming
// TCP payload into a human-readable command.
type Decoder struct {
	orders         map[uint16]adj.Packet
	addStateID     uint16
	removeStateID  uint16
	hasAddState    bool
	hasRemoveState bool
}

// New builds a Decoder for board, resolving the add/remove state order ids
// from messageIDs (typically adj.Info.MessageIds).
func New(board adj.Board, messageIDs map[string]uint16) *Decoder {
	orders := make(map[uint16]adj.Packet, len(board.Packets))
	for _, p := range board.Packets {
		if p.Type == "order" {
			orders[p.Id] = p
		}
	}

	d := &Decoder{orders: orders}
	d.addStateID, d.hasAddState = messageIDs["add_state_order"]
	d.removeStateID, d.hasRemoveState = messageIDs["remove_state_order"]
	return d
}

// Decode returns a human-readable representation of payload.
func (d *Decoder) Decode(payload []byte) (string, error) {
	r := bytes.NewReader(payload)

	var msgID uint16
	if err := binary.Read(r, binary.LittleEndian, &msgID); err != nil {
		return "", fmt.Errorf("read message id: %w", err)
	}

	action, ok := d.actionFor(msgID)
	if !ok {
		return "", fmt.Errorf("unknown message id %d", msgID)
	}

	var count uint16
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return "", fmt.Errorf("read count: %w", err)
	}

	names := make([]string, 0, count)
	for i := uint16(0); i < count; i++ {
		var id uint16
		if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
			return "", fmt.Errorf("read order id %d/%d: %w", i+1, count, err)
		}
		names = append(names, d.orderName(id))
	}

	return fmt.Sprintf("%s count=%d orders=[%s]", action, count, strings.Join(names, ", ")), nil
}

func (d *Decoder) actionFor(msgID uint16) (string, bool) {
	switch {
	case d.hasAddState && msgID == d.addStateID:
		return "ADD_STATE_ORDER", true
	case d.hasRemoveState && msgID == d.removeStateID:
		return "REMOVE_STATE_ORDER", true
	}
	return "", false
}

func (d *Decoder) orderName(id uint16) string {
	if pkt, ok := d.orders[id]; ok {
		return fmt.Sprintf("%s(%d)", pkt.Name, id)
	}
	return fmt.Sprintf("<unknown>(%d)", id)
}
