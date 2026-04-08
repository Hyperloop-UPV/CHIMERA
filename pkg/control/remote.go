package control

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

type remoteTUIServer struct {
	conn           net.Conn
	reader         *bufio.Reader
	boardNames     []prompt.Suggest
	measurements   map[string][]prompt.Suggest
	measurementMap map[string]map[string]string
}

func StartRemoteTUI(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to CHIMERA daemon at %s: %w", addr, err)
	}
	defer conn.Close()

	tui := &remoteTUIServer{
		conn:           conn,
		reader:         bufio.NewReader(conn),
		boardNames:     []prompt.Suggest{},
		measurements:   make(map[string][]prompt.Suggest),
		measurementMap: make(map[string]map[string]string),
	}

	if err := tui.syncBoards(); err != nil {
		return fmt.Errorf("failed to sync board metadata: %w", err)
	}

	fmt.Println("Connected to CHIMERA daemon. Type 'quit' to exit.")

	p := prompt.New(
		tui.executor,
		tui.completer,
		prompt.OptionPrefix("CHIMERA> "),
		prompt.OptionTitle("CHIMERA Remote TUI"),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionSelectedSuggestionBGColor(prompt.Blue),
		prompt.OptionSelectedSuggestionTextColor(prompt.White),
	)

	p.Run()
	return nil
}

func (t *remoteTUIServer) executor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	if strings.EqualFold(input, "quit") || strings.EqualFold(input, "exit") || strings.EqualFold(input, "bye") {
		fmt.Fprintln(t.conn, input)
		os.Exit(0)
	}

	if err := t.sendCommand(input); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	if strings.HasPrefix(strings.ToLower(input), "list") || strings.HasPrefix(strings.ToLower(input), "set") {
		_ = t.syncBoards()
	}
}

func (t *remoteTUIServer) sendCommand(command string) error {
	if _, err := fmt.Fprintln(t.conn, command); err != nil {
		return err
	}

	for {
		line, err := t.reader.ReadString('\n')
		if err != nil {
			return err
		}

		line = strings.TrimSuffix(line, "\n")
		if line == controlEndMarker {
			return nil
		}

		fmt.Println(line)
	}
}

func (t *remoteTUIServer) completer(d prompt.Document) []prompt.Suggest {
	text := d.TextBeforeCursor()
	if text == "" {
		return prompt.FilterHasPrefix(commandSuggestions, "", true)
	}

	if strings.HasSuffix(text, " ") {
		text += " "
	}

	args := splitArgs(text)
	if len(args) == 0 {
		return prompt.FilterHasPrefix(commandSuggestions, text, true)
	}

	switch strings.ToLower(args[0]) {
	case "list":
		if len(args) == 1 {
			return prompt.FilterHasPrefix(t.boardNames, "", true)
		}
		if len(args) == 2 {
			return prompt.FilterHasPrefix(t.boardNames, args[1], true)
		}
		if len(args) == 3 {
			return prompt.FilterHasPrefix(listSubSuggestions, args[2], true)
		}
	case "set":
		if len(args) == 1 {
			return prompt.FilterHasPrefix(t.boardNames, "", true)
		}
		if len(args) == 2 {
			return prompt.FilterHasPrefix(t.boardNames, args[1], true)
		}
		if len(args) == 3 {
			boardName := strings.ToUpper(args[1])
			if measurements, ok := t.measurements[boardName]; ok {
				return prompt.FilterHasPrefix(measurements, args[2], true)
			}
		}
		if len(args) == 4 {
			boardName := strings.ToUpper(args[1])
			measurementId := args[2]
			if measurementTypes, ok := t.measurementMap[boardName]; ok {
				if measurementType, exists := measurementTypes[measurementId]; exists {
					return []prompt.Suggest{{
						Text:        args[3],
						Description: fmt.Sprintf("Type: %s", measurementType),
					}}
				}
			}
		}
	}

	return prompt.FilterHasPrefix(commandSuggestions, args[0], true)
}

func (t *remoteTUIServer) syncBoards() error {
	response, err := t.requestRemote("list")
	if err != nil {
		return err
	}

	names := parseBoardNames(response)
	if len(names) == 0 {
		return nil
	}

	t.boardNames = make([]prompt.Suggest, len(names))
	for i, n := range names {
		t.boardNames[i] = prompt.Suggest{Text: n, Description: "Board"}
	}

	t.measurements = make(map[string][]prompt.Suggest)
	t.measurementMap = make(map[string]map[string]string)

	for _, board := range names {
		response, err := t.requestRemote(fmt.Sprintf("list %s measurements", board))
		if err != nil {
			continue
		}
		info := parseMeasurementInfo(response)
		boardMeasurements := make([]prompt.Suggest, 0, len(info))
		measurementTypes := make(map[string]string)
		for id, typ := range info {
			boardMeasurements = append(boardMeasurements, prompt.Suggest{Text: id, Description: typ})
			measurementTypes[id] = typ
		}
		sort.Slice(boardMeasurements, func(i, j int) bool {
			return boardMeasurements[i].Text < boardMeasurements[j].Text
		})
		t.measurements[strings.ToUpper(board)] = boardMeasurements
		t.measurementMap[strings.ToUpper(board)] = measurementTypes
	}

	return nil
}

func (t *remoteTUIServer) requestRemote(command string) (string, error) {
	if _, err := fmt.Fprintln(t.conn, command); err != nil {
		return "", err
	}

	var builder strings.Builder
	for {
		line, err := t.reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		line = strings.TrimSuffix(line, "\n")
		if line == controlEndMarker {
			break
		}
		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return strings.TrimSuffix(builder.String(), "\n"), nil
}

func parseMeasurementInfo(response string) map[string]string {
	info := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(response), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "MEASUREMENTS") || strings.HasPrefix(line, "PACKETS") || strings.HasPrefix(line, "-------") {
			continue
		}

		idx := strings.Index(line, "(")
		if idx < 0 {
			continue
		}

		id := strings.TrimSpace(line[:idx])
		typ := strings.TrimSpace(strings.TrimSuffix(line[idx+1:], ")"))
		info[id] = typ
	}
	return info
}
