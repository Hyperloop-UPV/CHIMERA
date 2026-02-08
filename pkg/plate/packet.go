package plate

import (
	"bytes"
	"encoding/binary"
)

func (pkt *PacketRuntime) BuildPayload() ([]byte, error) {

	buf := new(bytes.Buffer)

	// First: header with packet ID (uint16)
	err := binary.Write(buf, binary.LittleEndian, pkt.Packet.Id)
	if err != nil {
		return nil, err
	}

	// Second: each mesuearement value, encoded according to its type and enum values
	for _, measure := range pkt.Measurements {

		err := measure.WriteTo(buf)
		if err != nil {
			return nil, err
		}

	}
	return buf.Bytes(), nil
}
