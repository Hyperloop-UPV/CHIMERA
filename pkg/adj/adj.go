package adj

/* Partial copy of backend adj module, adapted to be used here @JavierRibaldelRio */

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
)

// ReadHead returns the HEAD commit hash and branch of the ADJ repository at
// path. Branch is "HEAD detached" if not on a branch.
func ReadHead(path string) (hash string, branch string, err error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to open ADJ repo at %s: %w", path, err)
	}

	head, err := repo.Head()
	if err != nil {
		return "", "", fmt.Errorf("failed to read ADJ HEAD: %w", err)
	}

	branch = "HEAD detached"
	if head.Name().IsBranch() {
		branch = head.Name().Short()
	}

	return head.Hash().String(), branch, nil
}

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
		if err == nil && info.IsDir() {
			return ADJPath, nil
		}

		fmt.Printf("Local ADJ path %s unavailable (%v), trying fallback\n", ADJPath, err)

		fallback, fbErr := executableADJFallback()
		if fbErr != nil {
			return "", fmt.Errorf("ADJ path %s unavailable and fallback failed: %w", ADJPath, fbErr)
		}

		fmt.Printf("Using fallback ADJ at %s\n", fallback)
		return fallback, nil
	}

	// Clone the ADJ repository safely using atomic swap
	err := CloneADJRepo(ADJBranch)
	if err != nil {
		fallback, fbErr := executableADJFallback()
		if fbErr != nil {
			return "", fmt.Errorf("failed to get ADJ: %w (fallback unavailable: %v)", err, fbErr)
		}
		fmt.Printf("Clone failed, using fallback ADJ at %s\n", fallback)
		return fallback, nil
	}

	return destination, nil
}

// executableADJFallback returns the path to an "adj" directory located next to
// the running executable, verifying that it exists and is a directory.
func executableADJFallback() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not determine executable path: %w", err)
	}

	path := filepath.Join(filepath.Dir(exe), "adj")

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("fallback ADJ path not accessible: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("fallback ADJ path is not a directory: %s", path)
	}

	return path, nil
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

	commitHash, branch, err := ReadHead(adjDirectory)
	if err != nil {
		fmt.Printf("Warning: could not read ADJ HEAD: %v\n", err)
		commitHash = "unknown"
		branch = "unknown"
	}

	adj := ADJ{
		Info:       info,
		Boards:     boards,
		CommitHash: commitHash,
		Branch:     branch,
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
