package plate

import (
	"fmt"
	"io"
	"strings"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
	"github.com/Hyperloop-UPV/CHIMERA/pkg/generator"
)

// NewMeasurementState creates a new MeasurementState for the given ADJ measurement. It initializes the generator for the measurement based on its type and range.
func NewMeasurementState(measurement adj.Measurement) *MeasurementState {

	return &MeasurementState{
		Measurement: measurement,
		Generator:   generator.SelectRandomGenerator(measurement),
	}
}

// given a mesurament state and a paylod writes its value in the payload
func (m *MeasurementState) WriteTo(w io.Writer) error {

	// Gets geneartor
	m.mu.RLock()
	gen := m.Generator
	m.mu.RUnlock()

	if gen == nil {
		return fmt.Errorf("generator not configured")
	}

	data, err := gen.Generate(m.Measurement)
	if err != nil {
		return err
	}

	// Write output into payload
	_, err = w.Write(data)
	return err

}

// SetGenerator modifys the generator of the measurement

func (m *MeasurementState) SetGenerator(newG string) error {

	m.mu.Lock()
	defer m.mu.Unlock()

	// Random generator
	if strings.EqualFold(newG, "r") || strings.EqualFold(newG, "random") {
		m.Generator = generator.SelectRandomGenerator(m.Measurement)
		return nil
	}

	val, err := generator.ParseValue(m.Measurement, newG)
	if err != nil {
		return err
	}

	m.Generator = &generator.FixedGenerator{
		Value: val,
	}

	return nil
}
