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
	"time"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/plate"
)

// Control server is the auxiliar server that controls Chimera's emulator options

func StartControlServer(port int, boards plate.PlateGenerators) {

	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(port))

	server := NewServer(addr, func(cmd Command, w ResponseWriter) error {
		return handleCommand(cmd, boards, w)
	})

	log.Printf("Control server started in 0.0.0.0:%d", port)

	if err := server.Start(); err != nil {
		log.Fatal(err)
		return
	}

}

type commandHandler func(Command, plate.PlateGenerators, ResponseWriter) error

var controlHandlers = map[string]commandHandler{
	"help": handleHelp,
	"h":    handleHelp,
	"list": handleList,
	"set":  handleSet,
	"test": handleTest,
}

func handleCommand(cmd Command, boards plate.PlateGenerators, w ResponseWriter) error {
	if len(cmd) == 0 {
		return fmt.Errorf("EMPTY")
	}

	handler, ok := controlHandlers[strings.ToLower(cmd[0])]
	if !ok {
		return fmt.Errorf("Unknown order. Use \"h\" to access the help menu")
	}

	return handler(cmd, boards, w)
}

func handleHelp(_ Command, _ plate.PlateGenerators, w ResponseWriter) error {
	return w.WriteLine(`AVAILABLE COMMANDS:
  h|help                     - show this help menu
  list                       - list registered boards and their IPs
  list <BOARD>               - show board packets and measurements
  list <BOARD> packets       - show only packets for a board
  list <BOARD> measurements  - show only measurements for a board
  set <BOARD> <ID> <V>       - set a measurement value on a board
  test TCP-abrupt            - simulate abrupt TCP disconnect for all boards
  test TCP-abrupt <BOARD>    - simulate abrupt TCP disconnect for a specific board
  quit|exit|bye              - close the control session
`)
}

func handleList(cmd Command, boards plate.PlateGenerators, w ResponseWriter) error {
	var buf bytes.Buffer
	writer := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)

	if len(cmd) == 1 {
		fmt.Fprintln(writer, "BOARD NAME\tIP")
		fmt.Fprintln(writer, "----------\t-----------")

		for name, rt := range boards {
			fmt.Fprintf(writer, "%s\t%s\n", name, rt.Board.IP)
		}

		writer.Flush()
		return w.WriteLine(strings.TrimSuffix(buf.String(), "\n"))
	}

	boardName := strings.ToUpper(cmd[1])
	rt, ok := boards[boardName]
	if !ok {
		return w.WriteLine("BOARD NOT FOUND")
	}

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
		showPackets = true
		showMeasurements = true
	}

	var packets []string
	var measurements []string

	if showPackets {
		for _, p := range rt.Packets {
			packets = append(packets, fmt.Sprintf("%s (%s)", p.Packet.Name, p.Packet.Type))
		}
	}

	if showMeasurements && !showPackets {
		for _, m := range rt.Measurements {
			measurements = append(measurements, fmt.Sprintf("%s (%s)", m.Measurement.Id, m.Measurement.Type))
		}
		sort.Strings(measurements)
	}

	if showPackets && showMeasurements {
		fmt.Fprintln(writer, "PACKETS\tMEASUREMENTS")
		fmt.Fprintln(writer, "-------\t-------------")
	} else if showPackets {
		fmt.Fprintln(writer, "PACKETS")
		fmt.Fprintln(writer, "-------")
	} else if showMeasurements {
		fmt.Fprintln(writer, "MEASUREMENTS")
		fmt.Fprintln(writer, "------------")
	}

	if showPackets && !showMeasurements {
		for _, p := range packets {
			fmt.Fprintf(writer, "%s\n", p)
		}
	} else if showMeasurements && !showPackets {
		for _, m := range measurements {
			fmt.Fprintf(writer, "%s\n", m)
		}
	} else {
		for idx, p := range rt.Packets {
			packetStr := fmt.Sprintf("%s (%s)", p.Packet.Name, p.Packet.Type)

			var packetMeasurements []string
			for _, m := range p.Measurements {
				packetMeasurements = append(packetMeasurements, fmt.Sprintf("%s (%s)", m.Measurement.Id, m.Measurement.Type))
			}

			sort.Strings(packetMeasurements)

			if len(packetMeasurements) == 0 {
				fmt.Fprintf(writer, "%s\n", packetStr)
				continue
			}

			for i, m := range packetMeasurements {
				if i == 0 {
					fmt.Fprintf(writer, "%s\t%s\n", packetStr, m)
				} else {
					fmt.Fprintf(writer, "\t%s\n", m)
				}
			}

			if idx < len(packets)-1 {
				// no-op (keeps packet order stable)
			}
		}
	}

	writer.Flush()
	return w.WriteLine(strings.TrimSuffix(buf.String(), "\n"))
}

func handleSet(cmd Command, boards plate.PlateGenerators, w ResponseWriter) error {
	if len(cmd) < 4 {
		return w.WriteLine("ERROR: usage -> set <board> <measurement-id> <value>")
	}

	boardName := strings.ToUpper(cmd[1])
	measID := plate.MeasurementID(cmd[2])
	value := cmd[3]

	rt, ok := boards[boardName]
	if !ok {
		return w.WriteLine("ERROR: board not found")
	}

	measState, ok := rt.Measurements[measID]
	if !ok {
		return w.WriteLine("ERROR: measurement not found")
	}

	if err := measState.SetGenerator(value); err != nil {
		return w.WriteLine("ERROR: " + err.Error())
	}

	return w.WriteLine("SUCCESSFUL")
}

func handleTest(cmd Command, boards plate.PlateGenerators, w ResponseWriter) error {
	if len(cmd) < 2 || len(cmd) > 3 {
		return w.WriteLine("ERROR: usage -> test TCP-abrupt [BOARD]")
	}

	if strings.ToLower(cmd[1]) != "tcp-abrupt" {
		return w.WriteLine("ERROR: unknown test. only TCP-abrupt is supported")
	}

	// If no board is specified, apply the test to all boards
	if len(cmd) == 2 {
		if err := w.WriteLine("Starting TCP-abrupt test for all boards"); err != nil {
			return err
		}

		for boardName, rt := range boards {
			if err := rt.AbruptlyClose(); err != nil {
				_ = w.WriteLine(fmt.Sprintf("ERROR %s: %s", boardName, err.Error()))
				continue
			}
			boards[boardName] = rt
			_ = w.WriteLine(fmt.Sprintf("Connection closed abruptly at %s", boardName))
		}

		if err := w.WriteLine("Waiting 30 seconds before restoring all board IPs..."); err != nil {
			return err
		}
		time.Sleep(30 * time.Second)

		for boardName, rt := range boards {
			if err := rt.RestoreIP(); err != nil {
				_ = w.WriteLine(fmt.Sprintf("ERROR restore %s: %s", boardName, err.Error()))
				continue
			}
			boards[boardName] = rt
			_ = w.WriteLine(fmt.Sprintf("Restored connection for %s", boardName))
		}

		return w.WriteLine("TCP-abrupt test completed")
	}

	// Test a specific board
	boardName := strings.ToUpper(cmd[2])
	rt, ok := boards[boardName]
	if !ok {
		return w.WriteLine("ERROR: board not found")
	}

	if err := w.WriteLine("Starting TCP-abrupt test for " + boardName); err != nil {
		return err
	}

	if err := rt.AbruptlyClose(); err != nil {
		return w.WriteLine("ERROR: " + err.Error())
	}

	w.WriteLine("Connection closed abruptly at " + boardName)

	if err := w.WriteLine("Waiting 30 seconds before restoring board IP..."); err != nil {
		return err
	}
	time.Sleep(30 * time.Second)

	if err := rt.RestoreIP(); err != nil {
		return w.WriteLine("ERROR: " + err.Error())
	}

	w.WriteLine("Restored connection for " + boardName)

	boards[boardName] = rt
	return w.WriteLine("TCP-abrupt test completed")
}
