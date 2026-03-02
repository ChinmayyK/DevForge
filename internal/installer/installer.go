// Package installer defines the interface for platform-specific package
// installers and provides a factory to create the appropriate one.
package installer

import (
	"fmt"
	"runtime"

	"github.com/chinmay/devforge/internal/executor"
	"github.com/chinmay/devforge/internal/logger"
)

// Installer is the interface that platform-specific package managers
// must implement.
type Installer interface {
	// IsInstalled checks whether a given dependency is already present
	// on the system.
	IsInstalled(name string) (bool, error)

	// Install installs the given dependency.
	Install(name string) error

	// GetVersion returns the installed version string for a dependency,
	// or an error if the dependency is not installed.
	GetVersion(name string) (string, error)
}

// New returns the appropriate Installer for the current OS. Currently
// only macOS (Homebrew) is supported.
func New(log *logger.Logger, exec *executor.Executor) (Installer, error) {
	switch runtime.GOOS {
	case "darwin":
		return NewBrewInstaller(log, exec)
	default:
		return nil, fmt.Errorf("no installer available for OS %q", runtime.GOOS)
	}
}
