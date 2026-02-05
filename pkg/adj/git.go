package adj

import (
	"fmt"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// CloneADJRepo clones the ADJ repository safely using atomic swap.
// If cloning fails, the previous repository is preserved.
func CloneADJRepo(branch string) error {

	fmt.Printf("Downloading ADJ branch %s\n", branch)

	parent := filepath.Dir(destination)

	// Create temporary directory in same filesystem
	tmpDir, err := os.MkdirTemp(parent, "adj_clone_*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Ensure temp dir cleanup on failure
	defer os.RemoveAll(tmpDir)

	backupDir := destination + "_backup"

	cloneOptions := &git.CloneOptions{
		URL:           repoURL,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
		Depth:         0,
		Progress:      os.Stdout,
	}

	// Step 1: Clone into temp directory
	_, err = git.PlainClone(tmpDir, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("ADJ:clone failed, fallback at %s	: %w", destination, err)
	}

	// Step 2: If destination exists, move it to backup
	if _, err := os.Stat(destination); err == nil {

		// Remove previous backup if exists
		_ = os.RemoveAll(backupDir)

		err = os.Rename(destination, backupDir)
		if err != nil {
			return fmt.Errorf("failed to backup existing repo: %w", err)
		}
	}

	// Step 3: Move temp clone to destination
	err = os.Rename(tmpDir, destination)
	if err != nil {

		// Rollback if swap fails
		_ = os.Rename(backupDir, destination)

		return fmt.Errorf("failed to move cloned repo into place: %w", err)
	}

	// Step 4: Remove backup after success
	_ = os.RemoveAll(backupDir)

	return nil
}
