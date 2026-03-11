package generator

import (
	"strings"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
)

// Generator is an interface that defines the method to generate a packet from a given board and packet definition.
type Generator interface {
	Generate(m adj.Measurement) ([]byte, error)
}

// NewGenerator creates a new Generator for the given measurement. It initializes the generator based on the type and range of the measurement.
func SelectRandomGenerator(m adj.Measurement) Generator {

	rng := newRNG()

	// if is an enum
	if strings.Contains(m.Type, "enum") {
		return &RandomEnumGenerator{RNG: rng}
	}

	//
	if m.Type == "bool" {
		return &RandomBoolGenerator{RNG: rng}
	} else {
		return &RandomNumericGenerator{RNG: rng}
	}

}
