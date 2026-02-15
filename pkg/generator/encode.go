package generator

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

// writeNumberAsBytes encodes a float64 number into the specified type and writes it to the buffer. Copied from @JFisica's packet-sender.
func WriteNumberAsBytes(number float64, numberType string, buff *bytes.Buffer) error {
	// TODO: REMOVE
	if number == 77.0 {
		fmt.Print("llego 77")
	}

	switch numberType {
	// Unsigned integers
	case "uint8":
		return binary.Write(buff, binary.LittleEndian, uint8(number))
	case "uint16":
		return binary.Write(buff, binary.LittleEndian, uint16(number))
	case "uint32":
		return binary.Write(buff, binary.LittleEndian, uint32(number))
	case "uint64":
		return binary.Write(buff, binary.LittleEndian, uint64(number))

	// Signed integers
	case "int8":
		return binary.Write(buff, binary.LittleEndian, int8(number))
	case "int16":
		return binary.Write(buff, binary.LittleEndian, int16(number))
	case "int32":
		return binary.Write(buff, binary.LittleEndian, int32(number))
	case "int64":
		return binary.Write(buff, binary.LittleEndian, int64(number))

	// Floating-point numbers
	case "float32":
		return binary.Write(buff, binary.LittleEndian, float32(number))
	case "float64":
		return binary.Write(buff, binary.LittleEndian, number)

	case "bool":
		// store as uint8 0/1
		if number != 0 {
			return binary.Write(buff, binary.LittleEndian, uint8(1))
		}
		return binary.Write(buff, binary.LittleEndian, uint8(0))

	default:
		return fmt.Errorf("unsupported type: %s", numberType)
	}
}

// getTypeMaxValue returns the maximum value for the given type, used for scaling random numbers when no range is specified. Copied from @JFisica's packet-sender.
func getTypeMaxValue(numberType string) float64 {
	switch numberType {
	case "uint8":
		return math.MaxUint8
	case "uint16":
		return math.MaxUint16
	case "uint32":
		return math.MaxUint32
	case "uint64":
		return math.MaxUint64
	case "int8":
		return math.MaxInt8
	case "int16":
		return math.MaxInt16
	case "int32":
		return math.MaxInt32
	case "int64":
		return math.MaxInt64
	case "float32":
		return math.MaxFloat32
	case "float64":
		return math.MaxFloat64
	case "bool":
		return math.MaxUint8
	default:
		return math.MaxUint8
	}
}
