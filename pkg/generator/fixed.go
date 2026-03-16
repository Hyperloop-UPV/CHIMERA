package generator

import (
	"bytes"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
)

type FixedGenerator struct {
	Value float64
}

// fixed values

func (g *FixedGenerator) Generate(m adj.Measurement) ([]byte, error) {

	buf := new(bytes.Buffer)

	err := WriteNumberAsBytes(g.Value, m.Type, buf)

	return buf.Bytes(), err
}
