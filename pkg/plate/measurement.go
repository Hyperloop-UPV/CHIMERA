package plate

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/Hyperloop-UPV/NATSOS/pkg/generator"
)

func (m *MeasurementState) WriteTo(w io.Writer) error {

	//! TEMPORALY

	m.mu.RLock()
	defer m.mu.RUnlock()

	// For enums, we store the index of the enum value as uint8
	if strings.Contains(m.Measurement.Type, "enum") {

		if len(m.Measurement.EnumValues) == 0 {
			return fmt.Errorf("enum without values")
		}

		val := uint8(m.Generator.Intn(len(m.Measurement.EnumValues)))
		return binary.Write(w, binary.LittleEndian, val)
	}

	// For bools, we store as uint8 0/1
	if m.Measurement.Type == "bool" {

		val := m.Generator.Int31n(2) == 1
		return binary.Write(w, binary.LittleEndian, val)
	}

	if m.Measurement.Type != "string" {

		var number float64

		if len(m.Measurement.WarningRange) == 0 {

			number = generator.MapNumberToRange(
				m.Generator.Float64(),
				m.Measurement.WarningRange,
				m.Measurement.Type,
			)

		} else if m.Measurement.WarningRange[0] != nil &&
			m.Measurement.WarningRange[1] != nil {

			low := *m.Measurement.WarningRange[0] * 0.8
			high := *m.Measurement.WarningRange[1] * 1.2

			number = generator.MapNumberToRange(
				m.Generator.Float64(),
				[]*float64{&low, &high},
				m.Measurement.Type,
			)

		} else {

			number = generator.MapNumberToRange(
				m.Generator.Float64(),
				[]*float64{},
				m.Measurement.Type,
			)
		}

		buf, ok := w.(*bytes.Buffer)
		if !ok {
			return fmt.Errorf("writer must be a *bytes.Buffer")
		}
		return generator.WriteNumberAsBytes(number, m.Measurement.Type, buf)
	}

	// strings
	return binary.Write(w, binary.LittleEndian, "")
}
