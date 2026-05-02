// Package decoder turns raw TCP payloads sent by the backend into a
// human-readable representation of the order so the TUI (local and remote)
// can print which command each board has received.
//
// Wire format of an order packet:
//
//	bytes 0..1 : packet id (uint16 LE) — id of the order in the board's ADJ
//	bytes 2..  : variables encoded back-to-back according to each measurement
//	             type (mirror of pkg/generator/encode.go)
package decoder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
)

// Decoder knows the order packets of a single board (indexed by id) so it can
// resolve an incoming TCP payload into a human-readable command.
type Decoder struct {
	orders map[uint16]adj.Packet
}

// New builds a Decoder from every order packet of board.
func New(board adj.Board, _ map[string]uint16) *Decoder {
	orders := make(map[uint16]adj.Packet, len(board.Packets))
	for _, p := range board.Packets {
		if p.Type == "order" {
			orders[p.Id] = p
		}
	}
	return &Decoder{orders: orders}
}

// Decode returns a human-readable representation of payload.
func (d *Decoder) Decode(payload []byte) (string, error) {
	r := bytes.NewReader(payload)

	var id uint16
	if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
		return "", fmt.Errorf("read packet id: %w", err)
	}

	pkt, ok := d.orders[id]
	if !ok {
		return "", fmt.Errorf("unknown order id %d", id)
	}

	values := make([]string, 0, len(pkt.Variables))
	for _, m := range pkt.Variables {
		v, err := readValue(r, m)
		if err != nil {
			return "", fmt.Errorf("variable %s (%s): %w", m.Id, m.Type, err)
		}
		values = append(values, fmt.Sprintf("%s=%s", m.Id, v))
	}

	if len(values) == 0 {
		return fmt.Sprintf("%s(%d)", pkt.Name, id), nil
	}
	return fmt.Sprintf("%s(%d) {%s}", pkt.Name, id, strings.Join(values, ", ")), nil
}

func readValue(r *bytes.Reader, m adj.Measurement) (string, error) {
	if len(m.EnumValues) > 0 {
		var v uint8
		if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
			return "", err
		}
		if int(v) < len(m.EnumValues) {
			return fmt.Sprintf("%s(%d)", m.EnumValues[v], v), nil
		}
		return fmt.Sprintf("<invalid enum %d>", v), nil
	}

	switch m.Type {
	case "bool":
		var v uint8
		if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
			return "", err
		}
		return fmt.Sprintf("%t", v != 0), nil
	case "uint8":
		var v uint8
		err := binary.Read(r, binary.LittleEndian, &v)
		return fmt.Sprintf("%d", v), err
	case "uint16":
		var v uint16
		err := binary.Read(r, binary.LittleEndian, &v)
		return fmt.Sprintf("%d", v), err
	case "uint32":
		var v uint32
		err := binary.Read(r, binary.LittleEndian, &v)
		return fmt.Sprintf("%d", v), err
	case "uint64":
		var v uint64
		err := binary.Read(r, binary.LittleEndian, &v)
		return fmt.Sprintf("%d", v), err
	case "int8":
		var v int8
		err := binary.Read(r, binary.LittleEndian, &v)
		return fmt.Sprintf("%d", v), err
	case "int16":
		var v int16
		err := binary.Read(r, binary.LittleEndian, &v)
		return fmt.Sprintf("%d", v), err
	case "int32":
		var v int32
		err := binary.Read(r, binary.LittleEndian, &v)
		return fmt.Sprintf("%d", v), err
	case "int64":
		var v int64
		err := binary.Read(r, binary.LittleEndian, &v)
		return fmt.Sprintf("%d", v), err
	case "float32":
		var v float32
		err := binary.Read(r, binary.LittleEndian, &v)
		return fmt.Sprintf("%g", v), err
	case "float64":
		var v float64
		err := binary.Read(r, binary.LittleEndian, &v)
		return fmt.Sprintf("%g", v), err
	}

	return "", fmt.Errorf("unsupported type %q", m.Type)
}
