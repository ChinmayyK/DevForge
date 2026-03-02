package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chinmay/devforge/internal/config"
	"github.com/chinmay/devforge/internal/osdetect"
	"github.com/chinmay/devforge/internal/ux"
)

// toolCheck holds the result of checking a single tool.
type toolCheck struct {
	Name      string
	Installed bool
	Version   string
	Error     string
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system readiness and dependency health",
	Long: `Run a series of checks against your development environment to
verify that all required tools are installed and functional.

Checks: Homebrew, Node.js, Git, Docker.`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(_ *cobra.Command, _ []string) error {
	// Load configuration to determine required dependencies.
	cfg, err := config.Load(cfgFile)
	if err != nil {
		ux.Error(fmt.Errorf("Failed to run doctor command: %v", err))
		return nil // Changed from `return` to `return nil` to match `RunE` signature
	}

	osInfo, err := osdetect.DetectFull()
	if err != nil {
		ux.Error(fmt.Errorf("OS detection failed: %v", err))
		return nil
	}
	fmt.Println()
	ux.Printf("%sDevForge Doctor — System Check%s\n", ux.Blue, ux.Reset)
	fmt.Println("--------------------------------------------------")
	ux.Printf("OS: %s/%s\n", osInfo.Name, osInfo.Arch)
	if osInfo.PackageMgr != "" {
		ux.Printf("Package Manager: %s\n", osInfo.PackageMgr)
	}
	fmt.Println()

	allReady := true

	for _, dep := range cfg.Dependencies {
		var installed bool
		var versionLabel string

		// Map config dependency names to their binary and version flags
		binary := dep.Name
		versionArg := "--version"

		// Handle known special cases
		if dep.Name == "node" {
			// node is just node
		} else if strings.ToLower(dep.Name) == "homebrew" {
			binary = "brew"
		}

		tc := checkTool(dep.Name, binary, versionArg)
		installed = tc.Installed
		versionLabel = tc.Version

		if installed {
			ux.Printf("%s%-10s%s : %s installed%s", ux.Gray, dep.Name, ux.Reset, ux.Green+ux.Check, ux.Reset)
			if versionLabel != "" {
				ux.Printf(" (%s)", versionLabel)
			}
			fmt.Println()
		} else {
			ux.Printf("%s%-10s%s : %s missing%s\n", ux.Gray, dep.Name, ux.Reset, ux.Red+ux.Cross, ux.Reset)
			allReady = false
		}
	}

	fmt.Println("--------------------------------------------------")
	if allReady {
		ux.Printf("Status: %sREADY%s\n", ux.Green, ux.Reset)
	} else {
		ux.Printf("Status: %sPARTIALLY READY%s\n", ux.Yellow, ux.Reset)
	}
	fmt.Println()

	return nil
}

// checkTool looks up a binary and tries to retrieve its version string.
func checkTool(name, binary, versionArg string) toolCheck {
	tc := toolCheck{Name: name}

	path, err := exec.LookPath(binary)
	if err != nil {
		tc.Installed = false
		tc.Error = "not found in PATH"
		return tc
	}

	out, err := exec.Command(path, versionArg).Output()
	if err != nil {
		tc.Installed = true
		tc.Version = "(installed, version unknown)"
		return tc
	}

	version := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	tc.Installed = true
	tc.Version = version
	return tc
}
