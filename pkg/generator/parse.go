package generator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/adj"
)

// Parse Value given astring parses it
func ParseValue(m adj.Measurement, input string) (float64, error) {

	fmt.Print(input, "  ", m.Type, "  ", m.Name)

	// ---------- ENUM (always uint8) ----------
	if len(m.EnumValues) > 0 {

		for i, val := range m.EnumValues {
			if strings.EqualFold(val, input) {
				return float64(uint8(i)), nil
			}
		}

		return 0, fmt.Errorf("invalid enum value")
	}

	// ---------- BOOL ----------
	if m.Type == "bool" {

		switch strings.ToLower(input) {
		case "true", "1":
			return 1, nil

		case "false", "0":
			return 0, nil

		default:
			return 0, fmt.Errorf("invalid bool value")
		}
	}

	// ---------- INTEGER TYPES ----------
	switch m.Type {

	case "uint8":
		v, err := strconv.ParseUint(input, 10, 8)
		return float64(uint8(v)), err

	case "uint16":
		v, err := strconv.ParseUint(input, 10, 16)
		return float64(uint16(v)), err

	case "uint32":
		v, err := strconv.ParseUint(input, 10, 32)
		return float64(uint32(v)), err

	case "uint64":
		v, err := strconv.ParseUint(input, 10, 64)
		return float64(v), err

	case "int8":
		v, err := strconv.ParseInt(input, 10, 8)
		return float64(int8(v)), err

	case "int16":
		v, err := strconv.ParseInt(input, 10, 16)
		return float64(int16(v)), err

	case "int32":
		v, err := strconv.ParseInt(input, 10, 32)
		return float64(int32(v)), err

	case "int64":
		v, err := strconv.ParseInt(input, 10, 64)
		return float64(v), err
	}

	// ---------- FLOAT ----------
	switch m.Type {

	case "float32":
		v, err := strconv.ParseFloat(input, 32)
		return float64(float32(v)), err

	case "float64":
		return strconv.ParseFloat(input, 64)
	}

	return 0, fmt.Errorf("unsupported measurement type")
}
