package generator

import (
	"bytes"
	"fmt"

	"github.com/Hyperloop-UPV/NATSOS/pkg/adj"
)

type FixedGenerator struct {
	Value float64
}

// fixed values

func (g *FixedGenerator) Generate(m adj.Measurement) ([]byte, error) {

	buf := new(bytes.Buffer)
	// TODO: REMOVE
	fmt.Println(m.Name, g.Value)

	err := WriteNumberAsBytes(g.Value, m.Type, buf)

	return buf.Bytes(), err
}
