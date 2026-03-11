package control

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/plate"
)

// Control server is the auxiliar server that controls Chimera's emulator options

func StartControlServer(port int, boards plate.PlateGenerators) {

	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(port))

	server := NewServer(addr, func(cmd Command) string {
		return handleCommand(cmd, boards)
	})

	log.Printf("Control server started in 0.0.0.0:%d", port)

	if err := server.Start(); err != nil {
		log.Fatal(err)
		return
	}

}

func handleCommand(cmd Command, boards plate.PlateGenerators) string {

	/**
	* Sub functions
	**/

	showList := func() string {

		var buf bytes.Buffer
		w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)

		// -------- LISTA GENERAL --------
		if len(cmd) == 1 {

			fmt.Fprintln(w, "BOARD NAME\tIP")
			fmt.Fprintln(w, "----------\t-----------")

			for name, plate := range boards {
				fmt.Fprintf(w, "%s\t%s\n", name, plate.Board.IP)
			}

			w.Flush()
			return buf.String()
		}

		// -------- LISTA DETALLADA --------

		boardName := strings.ToUpper(cmd[1])

		plate, ok := boards[boardName]
		if !ok {
			return "BOARD NOT FOUND"
		}

		// Construir listas
		var packets []string
		var measurements []string

		// PACKETS
		for _, p := range plate.Packets {
			packets = append(packets, fmt.Sprintf("%s (%s)", p.Packet.Name, p.Packet.Type))
		}

		// MEASUREMENTS
		for _, m := range plate.Measurements {
			measurements = append(measurements, fmt.Sprintf("%s (%s)", m.Measurement.Id, m.Measurement.Type))
		}

		// Cabecera
		fmt.Fprintln(w, "PACKETS\tMEASUREMENTS")
		fmt.Fprintln(w, "-------\t-------------")

		// Número máximo de filas
		maxLen := max(len(packets), len(measurements))

		for i := 0; i < maxLen; i++ {

			var p string
			var m string

			if i < len(packets) {
				p = packets[i]
			}

			if i < len(measurements) {
				m = measurements[i]
			}

			fmt.Fprintf(w, "%s\t%s\n", p, m)
		}

		w.Flush()
		return buf.String()
	}

	set := func() string {

		// Validate command format
		// Expected: set <board> <measurement-id> <value>
		if len(cmd) < 4 {
			return "ERROR: usage -> set <board> <measurement-id> <value>"
		}

		boardName := strings.ToUpper(cmd[1])
		measID := plate.MeasurementID(cmd[2])
		value := cmd[3]

		// Lookup target board runtime
		plate, ok := boards[boardName]
		if !ok {
			return "ERROR: board not found"
		}

		// Lookup measurement runtime inside the selected board
		measState, ok := plate.Measurements[measID]
		if !ok {
			return "ERROR: measurement not found"
		}

		// Update generator value
		// Assumes generator implements a Set(string) error method
		if err := measState.SetGenerator(value); err != nil {
			return "ERROR: " + err.Error()
		}

		return "SUCCESSFUL"
	}

	var out string

	switch cmd[0] {
	case "help":
	case "h":
		out = "HELP MENÚ"
		break

	case "list":
		out = showList()
	case "set":
		out = set()

		break
	default:
		out = "Unknown order. Use \"h\" to access the help menu"
	}

	return out

}
