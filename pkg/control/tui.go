package control

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Hyperloop-UPV/CHIMERA/pkg/plate"
	prompt "github.com/c-bata/go-prompt"
)

type TUIServer struct {
	boards             plate.PlateGenerators
	boardNames         []prompt.Suggest
	measurements       map[string][]prompt.Suggest            // boardName -> measurements
	measurementMap     map[string]map[string]string           // boardName -> measurementId -> type
	measurementOptions map[string]map[string][]prompt.Suggest // boardName -> measurementId -> possible values
}

type bufferWriter struct {
	buf *bytes.Buffer
}

type stdoutWriter struct{}

func (w *bufferWriter) WriteLine(msg string) error {
	_, err := fmt.Fprintln(w.buf, msg)
	return err
}

func (w *stdoutWriter) WriteLine(msg string) error {
	_, err := fmt.Fprintln(os.Stdout, msg)
	return err
}

// NewTUIServer creates a new TUI control interface
func NewTUIServer(boards plate.PlateGenerators) *TUIServer {
	return &TUIServer{
		boards:             boards,
		boardNames:         []prompt.Suggest{},
		measurements:       make(map[string][]prompt.Suggest),
		measurementMap:     make(map[string]map[string]string),
		measurementOptions: make(map[string]map[string][]prompt.Suggest),
	}
}

// Start runs the TUI interface
func (t *TUIServer) Start() error {
	t.refreshBoardNames()

	fmt.Println("\n╔═══════════════════════════════════════╗")
	fmt.Println("║     CHIMERA Control TUI Interface     ║")
	fmt.Println("╚═══════════════════════════════════════╝")
	fmt.Println("Type 'h' for help or 'quit' to exit")

	p := prompt.New(
		t.executor,
		t.completer,
		prompt.OptionPrefix("CHIMERA> "),
		prompt.OptionTitle("CHIMERA Control"),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionSelectedSuggestionBGColor(prompt.Blue),
		prompt.OptionSelectedSuggestionTextColor(prompt.White),
		prompt.OptionAddKeyBind(prompt.KeyBind{Key: prompt.ControlC, Fn: func(buf *prompt.Buffer) {
			fmt.Println()
			t.cleanupBoards()
			fmt.Println("Bye")
			os.Exit(0)
		}}),
	)

	p.Run()
	return nil
}

func (t *TUIServer) cleanupBoards() {
	if t.boards == nil {
		return
	}

	for boardName, rt := range t.boards {
		if err := rt.Delete(); err != nil {
			fmt.Printf("WARN: cleanup %s failed: %v\n", boardName, err)
		} else {
			fmt.Printf("Cleaned up board %s\n", boardName)
		}
	}
}

func (t *TUIServer) executor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	lower := strings.ToLower(input)
	switch lower {
	case "quit", "exit", "bye":
		t.cleanupBoards()
		fmt.Println("Bye")
		os.Exit(0)
	}

	// Parse and execute command
	cmd := ParseCommand(input)
	if len(cmd) == 0 {
		fmt.Println("EMPTY")
		return
	}

	buf := &bytes.Buffer{}
	var writer ResponseWriter = &bufferWriter{buf: buf}
	if strings.ToLower(cmd[0]) == "test" {
		writer = &stdoutWriter{}
	}

	if err := handleControlCommand(cmd, t.boards, writer); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	if _, ok := writer.(*bufferWriter); ok {
		fmt.Print(buf.String())
	}

	// Refresh board names after list or set commands
	if strings.HasPrefix(lower, "list") || strings.HasPrefix(lower, "set") {
		t.refreshBoardNames()
	}
}

func (t *TUIServer) completer(d prompt.Document) []prompt.Suggest {
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
			// Autocomplete measurements for the selected board
			boardName := strings.ToUpper(args[1])
			if measurements, ok := t.measurements[boardName]; ok {
				return prompt.FilterHasPrefix(measurements, args[2], true)
			}
		}
		if len(args) == 4 {
			// Show type hint for the selected measurement or enum/bool options
			boardName := strings.ToUpper(args[1])
			measurementId := args[2]
			if measurementOpts, ok := t.measurementOptions[boardName]; ok {
				if options, exists := measurementOpts[measurementId]; exists && len(options) > 0 {
					// Show enum/bool options
					return prompt.FilterHasPrefix(options, args[3], true)
				}
			}
			// Fall back to type hint if no options available
			if measurementTypes, ok := t.measurementMap[boardName]; ok {
				if measurementType, exists := measurementTypes[measurementId]; exists {
					return []prompt.Suggest{{
						Text:        args[3], // Keep current input
						Description: fmt.Sprintf("Type: %s", measurementType),
					}}
				}
			}
		}
	case "test":
		if len(args) == 1 {
			return prompt.FilterHasPrefix([]prompt.Suggest{{Text: "TCP-abrupt"}}, "", true)
		}
		if len(args) == 2 {
			return prompt.FilterHasPrefix([]prompt.Suggest{{Text: "TCP-abrupt"}}, args[1], true)
		}
		if len(args) == 3 {
			return prompt.FilterHasPrefix(t.boardNames, args[2], true)
		}
	}

	return prompt.FilterHasPrefix(commandSuggestions, args[0], true)
}

func (t *TUIServer) refreshBoardNames() {
	if t.boards == nil || len(t.boards) == 0 {
		return
	}

	cmd := ParseCommand("list")
	buf := &bytes.Buffer{}
	writer := &bufferWriter{buf: buf}

	if err := handleControlCommand(cmd, t.boards, writer); err != nil {
		return
	}

	names := parseBoardNames(buf.String())
	if len(names) == 0 {
		return
	}

	t.boardNames = make([]prompt.Suggest, len(names))
	for i, n := range names {
		t.boardNames[i] = prompt.Suggest{Text: n, Description: "Board"}
	}
	sort.Slice(t.boardNames, func(i, j int) bool {
		return t.boardNames[i].Text < t.boardNames[j].Text
	})

	// Refresh measurements for each board
	t.measurements = make(map[string][]prompt.Suggest)
	t.measurementMap = make(map[string]map[string]string)
	t.measurementOptions = make(map[string]map[string][]prompt.Suggest)

	for boardName, rt := range t.boards {
		boardMeasurements := make([]prompt.Suggest, 0, len(rt.Measurements))
		measurementTypes := make(map[string]string)
		measurementOpts := make(map[string][]prompt.Suggest)

		for _, measurement := range rt.Measurements {
			boardMeasurements = append(boardMeasurements, prompt.Suggest{
				Text:        string(measurement.Measurement.Id),
				Description: measurement.Measurement.Type,
			})
			measurementTypes[string(measurement.Measurement.Id)] = measurement.Measurement.Type

			// Generate options for enum and bool types
			var options []prompt.Suggest
			if strings.Contains(measurement.Measurement.Type, "enum") && len(measurement.Measurement.EnumValues) > 0 {
				for _, enumVal := range measurement.Measurement.EnumValues {
					options = append(options, prompt.Suggest{
						Text:        enumVal,
						Description: "enum value",
					})
				}
			} else if measurement.Measurement.Type == "bool" {
				options = []prompt.Suggest{
					{Text: "true", Description: "boolean value"},
					{Text: "false", Description: "boolean value"},
				}
			}
			measurementOpts[string(measurement.Measurement.Id)] = options
		}

		sort.Slice(boardMeasurements, func(i, j int) bool {
			return boardMeasurements[i].Text < boardMeasurements[j].Text
		})

		t.measurements[boardName] = boardMeasurements
		t.measurementMap[boardName] = measurementTypes
		t.measurementOptions[boardName] = measurementOpts
	}
}

var commandSuggestions = []prompt.Suggest{
	{Text: "help", Description: "Show help menu"},
	{Text: "h", Description: "Show help menu"},
	{Text: "list", Description: "List registered boards"},
	{Text: "set", Description: "Set measurement value"},
	{Text: "test", Description: "Run a test command"},
	{Text: "quit", Description: "Exit the TUI"},
	{Text: "exit", Description: "Exit the TUI"},
	{Text: "bye", Description: "Exit the TUI"},
}

var listSubSuggestions = []prompt.Suggest{
	{Text: "packets", Description: "Show packets only"},
	{Text: "measurements", Description: "Show measurements only"},
}

func splitArgs(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	parts := strings.Fields(text)
	if strings.HasSuffix(text, " ") {
		parts = append(parts, "")
	}
	return parts
}

func parseBoardNames(response string) []string {
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(response), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "BOARD NAME") || strings.HasPrefix(line, "----------") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 1 {
			names = append(names, fields[0])
		}
	}
	return names
}

func handleControlCommand(cmd Command, boards plate.PlateGenerators, w ResponseWriter) error {
	if len(cmd) == 0 {
		return fmt.Errorf("EMPTY")
	}

	handlers := map[string]func(Command, plate.PlateGenerators, ResponseWriter) error{
		"help": handleHelp,
		"h":    handleHelp,
		"list": handleList,
		"set":  handleSet,
		"test": handleTest,
	}

	handler, ok := handlers[strings.ToLower(cmd[0])]
	if !ok {
		return fmt.Errorf("Unknown order. Use \"h\" to access the help menu")
	}

	return handler(cmd, boards, w)
}
