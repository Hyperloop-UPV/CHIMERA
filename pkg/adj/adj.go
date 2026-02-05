package adj

import (
	"fmt"
	"os"
)

const (
	// repoURL is the URL of the ADJ repository to clone.
	repoURL = "https://github.com/Hyperloop-UPV/adj.git"
	// destination is the local folder where the ADJ repository will be cloned.
	destination = "./adj"
)

func GetADJ(ADJBranch string, ADJPath string) (string, error) {

	path, err := GetPath(ADJBranch, ADJPath)
	if err != nil {
		return "", fmt.Errorf("failed to get ADJ: %w", err)
	}

	return path, nil
}

func GetPath(ADJBranch string, ADJPath string) (string, error) {

	// If ADJPath is provided, use the local repository instead of cloning
	if ADJPath != "" {
		fmt.Printf("Using local ADJ repository at %s\n", ADJPath)

		info, err := os.Stat(ADJPath)

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
