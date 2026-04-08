package adj

/* Partial copy of backend adj module, adapted to be used here @JavierRibaldelRio */

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// repoURL is the URL of the ADJ repository to clone.
	repoURL = "https://github.com/Hyperloop-UPV/adj.git"
	// destination is the local folder where the ADJ repository will be cloned.
	destination = "./adj"
)

// GetADJ retrieves the ADJ repository, either by cloning it from GitHub or using a local path if provided.
func GetADJ(ADJBranch string, ADJPath string) (string, error) {

	path, err := GetPath(ADJBranch, ADJPath)
	if err != nil {
		return "", fmt.Errorf("failed to get ADJ: %w", err)
	}

	return path, nil
}

// GetPath retrieves the path to the ADJ repository, either by cloning it from GitHub or using a local path if provided.
func GetPath(ADJBranch string, ADJPath string) (string, error) {

	// If ADJPath is provided, use the local repository instead of cloning
	if ADJPath != "" {
		fmt.Printf("Using local ADJ repository at %s\n", ADJPath)

		// Check if the provided ADJPath exists and is a directory
		info, err := os.Stat(ADJPath)

		// Handle errors when accessing the ADJPath
		if err != nil {

			if os.IsNotExist(err) {
				return "", fmt.Errorf("ADJ path does not exist: %s", ADJPath)
			}

			if os.IsPermission(err) {
				return "", fmt.Errorf("permission denied accessing ADJ path: %s", ADJPath)
			}

			return "", fmt.Errorf("error accessing ADJ path: %w", err)
		}

		if !info.IsDir() {
			return "", fmt.Errorf("ADJ path is not a directory: %s", ADJPath)
		}

		// If ADJPath is empty, clone the repository

		return ADJPath, nil
	}

	// Clone the ADJ repository safely using atomic swap
	err := CloneADJRepo(ADJBranch)
	if err != nil {
		return "", fmt.Errorf("failed to get ADJ: %w", err)
	}

	return destination, nil
}

// downloadADJ retrieves the ADJ repository, either by cloning it from GitHub or using a local path if provided, and reads the general info and boards list.
func downloadADJ(AdjBranch string, AdjPath string) (string, json.RawMessage, json.RawMessage, error) {

	// Get the ADJ repository path (either by cloning or using local path) and get the path
	destination, err := GetPath(AdjBranch, AdjPath)

	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to get ADJ: %w", err)
	}

	// Read the general info and boards list from the ADJ repository

	info, err := os.ReadFile(filepath.Join(destination, "general_info.json"))
	if err != nil {
		return "", nil, nil, err
	}

	boardsList, err := os.ReadFile(filepath.Join(destination, "boards.json"))
	if err != nil {
		return "", nil, nil, err
	}

	return destination, info, boardsList, nil
}

// NewADJ creates a new ADJ instance by downloading the ADJ repository, reading the general info and boards list, and parsing the data into the ADJ struct.
func NewADJ(AdjBranch string, AdjPath string) (ADJ, error) {

	// Download the ADJ repository, read the general info and boards list, and get the path to the ADJ repository
	adjDirectory, infoRaw, boardsRaw, err := downloadADJ(AdjBranch, AdjPath)
	if err != nil {
		return ADJ{}, err
	}

	var infoJSON InfoJSON
	if err := json.Unmarshal(infoRaw, &infoJSON); err != nil {
		println("Info JSON unmarshal error")
		return ADJ{}, err
	}

	var info = Info{
		Ports:      infoJSON.Ports,
		MessageIds: infoJSON.MessageIds,
		Units:      make(map[string]string),
	}
	for key, value := range infoJSON.Units {
		info.Units[key] = value
	}

	var boardsList map[string]string

	if err := json.Unmarshal(boardsRaw, &boardsList); err != nil {
		return ADJ{}, err
	}

	boards, err := getBoards(adjDirectory, boardsList)
	if err != nil {
		return ADJ{}, err
	}

	info.BoardIds, err = getBoardIds(adjDirectory, boardsList)
	if err != nil {
		return ADJ{}, err
	}

	info.Addresses, err = getAddresses(boards)
	if err != nil {
		return ADJ{}, err
	}
	for target, address := range infoJSON.Addresses {
		info.Addresses[target] = address
	}

	adj := ADJ{
		Info:   info,
		Boards: boards,
	}

	// Check that ADJ has backend address and UDP port defined.
	if adj.Info.Addresses["backend"] == "" {
		return ADJ{}, fmt.Errorf("ADJ is missing backend address")
	}

	if adj.Info.Ports["UDP"] == 0 {
		return ADJ{}, fmt.Errorf("ADJ is missing UDP port")
	}

	return adj, nil
}
