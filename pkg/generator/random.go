package generator

// random.go contains all the random generetors used by each mesurement of each plate, for enums, numerica and bool

import (
	"bytes"
	"encoding/binary"
	mrand "math/rand"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
)

// All functions of this file are an adaptation of @JFisica's packet sender

// Enum generator
type RandomEnumGenerator struct {
	RNG *mrand.Rand
}

func (g *RandomEnumGenerator) Generate(m adj.Measurement) ([]byte, error) {

	// If the enum list is empty, fall back to 0 (to avoid stopping packet generation).
	if len(m.EnumValues) == 0 {
		buf := new(bytes.Buffer)
		_ = binary.Write(buf, binary.LittleEndian, uint8(0))
		return buf.Bytes(), nil
	}

	//! IMPORTANT: enums are defined as uint8
	// Generate Value among the possible enum values
	val := uint8(g.RNG.Intn(len(m.EnumValues)))

	// Write buffer
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, val)

	return buf.Bytes(), err
}

// Bool random generator
type RandomBoolGenerator struct {
	RNG *mrand.Rand
}

func (g *RandomBoolGenerator) Generate(m adj.Measurement) ([]byte, error) {

	// Select 1 or 0 for eneum
	val := g.RNG.Int31n(2) == 1

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, val)

	return buf.Bytes(), err
}

// Numeric geneator
type RandomNumericGenerator struct {
	RNG *mrand.Rand
}

func (g *RandomNumericGenerator) Generate(m adj.Measurement) ([]byte, error) {

	var number float64

	// Not warining range, completly random
	if len(m.WarningRange) == 0 {

		number = MapNumberToRange(
			g.RNG.Float64(),
			m.WarningRange,
			m.Type,
		)

	} else if m.WarningRange[0] != nil &&
		m.WarningRange[1] != nil { // if there is upper and lower range

		low := *m.WarningRange[0] * 0.8
		high := *m.WarningRange[1] * 1.2

		number = MapNumberToRange(
			g.RNG.Float64(),
			[]*float64{&low, &high},
			m.Type,
		)

	} else {

		number = MapNumberToRange(
			g.RNG.Float64(),
			[]*float64{},
			m.Type,
		)
	}

	buf := new(bytes.Buffer)

	err := WriteNumberAsBytes(number, m.Type, buf)

	return buf.Bytes(), err
}

// mapNumberToRange maps a [0,1) random number to the given range for the specified type. If the range is empty, it maps to [0, max(type)]
func MapNumberToRange(number float64, numberRange []*float64, numberType string) float64 {
	if len(numberRange) == 0 {
		return number * getTypeMaxValue(numberType)
	} else {
		return (number * (*numberRange[1] - *numberRange[0])) + *numberRange[0]
	}
}
