package plate

import (
	"bytes"
	"encoding/binary"
)

// BuildPayload builds the payload of the packet by encoding the packet ID and the measurement values according to their types and enum values. The payload is built as follows:
// - First, the packet ID is encoded as a uint16 in little endian format.
// - Then, for each measurement in the packet, its value is generated and encoded according to its type and enum values. The encoding is done by the WriteTo method of the MeasurementState struct, which handles the different types and enums.
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
