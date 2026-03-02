// Package template handles cloning starter template repositories from
// Git using go-git.
package template

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"

	"github.com/chinmay/devforge/internal/logger"
	"github.com/chinmay/devforge/internal/rollback"
)

// Cloner handles cloning Git repositories into local directories.
type Cloner struct {
	log    *logger.Logger
	rb     *rollback.Manager
	dryRun bool
}

// NewCloner creates a Cloner with the given logger and rollback manager.
func NewCloner(log *logger.Logger, rb *rollback.Manager, dryRun bool) *Cloner {
	return &Cloner{
		log:    log,
		rb:     rb,
		dryRun: dryRun,
	}
}

// Clone clones the repository at repoURL into the given destDir. It
// validates that the destination does not already exist and registers
// a rollback action to delete the directory if a later step fails.
func (c *Cloner) Clone(repoURL, destDir string) error {
	// Verify destination does not already exist.
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("destination directory %q already exists", destDir)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check destination directory %q: %w", destDir, err)
	}

	if c.dryRun {
		c.log.Info(fmt.Sprintf("[dry-run] would clone %s into %s", repoURL, destDir))
		return nil
	}

	c.log.Info(fmt.Sprintf("cloning template from %s into %s...", repoURL, destDir))

	_, err := git.PlainClone(destDir, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository %q: %w", repoURL, err)
	}

	// Register rollback: remove cloned directory on failure.
	c.rb.Register(fmt.Sprintf("remove cloned directory %s", destDir), func() error {
		return os.RemoveAll(destDir)
	})

	c.log.Info(fmt.Sprintf("template cloned successfully into %s", destDir))
	return nil
}
