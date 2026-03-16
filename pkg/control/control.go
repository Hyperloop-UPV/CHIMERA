package control

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sort"
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

		// Determine if the user wants a subset
		var showPackets, showMeasurements bool
		switch {
		case len(cmd) == 2:
			showPackets = true
			showMeasurements = true
		case strings.ToLower(cmd[2]) == "packets":
			showPackets = true
		case strings.ToLower(cmd[2]) == "measurements":
			showMeasurements = true
		default:
			// Unrecognized argument -> show both
			showPackets = true
			showMeasurements = true
		}

		// Build lists
		var packets []string
		var measurements []string

		if showPackets {
			for _, p := range plate.Packets {
				packets = append(packets, fmt.Sprintf("%s (%s)", p.Packet.Name, p.Packet.Type))
			}
		}

		// When only measurements are requested without packet context, build the global list.
		if showMeasurements && !showPackets {
			for _, m := range plate.Measurements {
				measurements = append(measurements, fmt.Sprintf("%s (%s)", m.Measurement.Id, m.Measurement.Type))
			}

			// Sort measurements alphabetically (short, human-friendly ordering)
			sort.Strings(measurements)
		}

		// Cabecera (headers match the columns being printed)
		if showPackets && showMeasurements {
			fmt.Fprintln(w, "PACKETS\tMEASUREMENTS")
			fmt.Fprintln(w, "-------\t-------------")
		} else if showPackets {
			fmt.Fprintln(w, "PACKETS")
			fmt.Fprintln(w, "-------")
		} else if showMeasurements {
			fmt.Fprintln(w, "MEASUREMENTS")
			fmt.Fprintln(w, "------------")
		}

		// If only packets or measurements are requested, print in a single column
		if showPackets && !showMeasurements {
			for _, p := range packets {
				fmt.Fprintf(w, "%s\n", p)
			}
		} else if showMeasurements && !showPackets {
			for _, m := range measurements {
				fmt.Fprintf(w, "%s\n", m)
			}
		} else {
			// Both: for each packet list its own measurements (sorted)
			for idx, p := range plate.Packets {
				packetStr := fmt.Sprintf("%s (%s)", p.Packet.Name, p.Packet.Type)

				// Collect measurements for this packet
				var packetMeasurements []string
				for _, m := range p.Measurements {
					packetMeasurements = append(packetMeasurements, fmt.Sprintf("%s (%s)", m.Measurement.Id, m.Measurement.Type))
				}

				// Sort measurements for this packet (human-friendly ordering)
				sort.Strings(packetMeasurements)

				if len(packetMeasurements) == 0 {
					fmt.Fprintf(w, "%s\n", packetStr)
					continue
				}

				for i, m := range packetMeasurements {
					if i == 0 {
						fmt.Fprintf(w, "%s\t%s\n", packetStr, m)
					} else {
						fmt.Fprintf(w, "\t%s\n", m)
					}
				}
				// Keep order consistent with packets list
				if idx < len(packets)-1 {
					// no-op (keeps packet order stable)
				}
			}
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

	helpText := `AVAILABLE COMMANDS:
  h|help                     - show this help menu
  list                       - list registered boards and their IPs
  list <BOARD>               - show board packets and measurements
  list <BOARD> packets       - show only packets for a board
  list <BOARD> measurements  - show only measurements for a board
  set <BOARD> <ID> <V>       - set a measurement value on a board
  quit|exit|bye              - close the control session
`

	switch strings.ToLower(cmd[0]) {
	case "help", "h":
		out = helpText

	case "list":
		out = showList()
	case "set":
		out = set()

	default:
		out = "Unknown order. Use \"h\" to access the help menu"
	}

	return out
}
