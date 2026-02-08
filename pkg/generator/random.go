package generator

import (
	"bytes"
	"encoding/binary"
	mrand "math/rand"
	"strings"
	"time"

	"github.com/Hyperloop-UPV/NATSOS/pkg/adj"
)

type RandomGenerator struct {
	r *mrand.Rand
}

// NewRandomGenerator creates a new RandomGenerator with a random seed based on the current time.
func NewRandomGenerator() *RandomGenerator {
	seed := time.Now().UnixNano()
	return &RandomGenerator{r: mrand.New(mrand.NewSource(seed))}
}

// Generate creates a random packet based on the given board name and packet definition. It encodes the packet ID and each measurement value according to their types and ranges. Based on @JFisica's packet-sender
func (gene *RandomGenerator) Generate(boardName string, pkt adj.Packet) ([]byte, error) {
	_ = boardName // reserved for future per-board logic

	buf := new(bytes.Buffer)

	// First: header with packet ID (uint16)
	err := binary.Write(buf, binary.LittleEndian, pkt.Id)
	if err != nil {
		return nil, err
	}

	// Second: each mesuearement value, encoded according to its type and enum values
	for _, meas := range pkt.Variables {

		// For enums
		if strings.Contains(meas.Type, "enum") {
			// Enum: pick index (0..len-1) or value; here we encode as uint8 index
			err := binary.Write(buf, binary.LittleEndian, uint8(gene.r.Intn(len(meas.EnumValues))))
			if err != nil {
				return nil, err
			}

		} else if meas.Type == "bool" {
			err := binary.Write(buf, binary.LittleEndian, gene.r.Int31n(2) == 1)
			if err != nil {
				return nil, err
			}
		} else if meas.Type != "string" {

			var number float64

			// For numbers, we consider the warning range as the most interesting to generate values around, so we use it as the main range. If it's not specified, we fallback to the full type range.
			if len(meas.WarningRange) == 0 {

				number = MapNumberToRange(gene.r.Float64(), meas.WarningRange, meas.Type)

			} else if meas.WarningRange[0] != nil && meas.WarningRange[1] != nil {
				low := *meas.WarningRange[0] * 0.8
				high := *meas.WarningRange[1] * 1.2
				number = MapNumberToRange(gene.r.Float64(), []*float64{&low, &high}, meas.Type)

			} else {
				// Fallback if any bound is nil
				number = MapNumberToRange(gene.r.Float64(), []*float64{}, meas.Type)
			}

			err = WriteNumberAsBytes(number, meas.Type, buf)
			if err != nil {
				return nil, err
			}

		}

	}
	return buf.Bytes(), nil
}

// mapNumberToRange maps a [0,1) random number to the given range for the specified type. If the range is empty, it maps to [0, max(type)]. Based on @JFisica's packet-sender.
func MapNumberToRange(number float64, numberRange []*float64, numberType string) float64 {
	if len(numberRange) == 0 {
		return number * getTypeMaxValue(numberType)
	} else {
		return (number * (*numberRange[1] - *numberRange[0])) + *numberRange[0]
	}
}
