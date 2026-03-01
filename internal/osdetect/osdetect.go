// Package osdetect provides operating system detection and validation
// for DevForge. Currently only macOS (darwin) is supported.
package osdetect

import (
	"fmt"
	"runtime"
)

// supportedOS lists the operating systems DevForge currently supports.
var supportedOS = map[string]string{
	"darwin": "macOS",
}

// Result holds the outcome of an OS detection check.
type Result struct {
	// OS is the raw GOOS value (e.g. "darwin", "linux").
	OS string
	// DisplayName is a human-friendly name (e.g. "macOS").
	DisplayName string
	// Arch is the CPU architecture (e.g. "arm64", "amd64").
	Arch string
	// Supported indicates whether the detected OS is supported.
	Supported bool
}

// Detect inspects the current runtime environment and returns an OS
// detection result. It returns an error only if the detected OS is
// not in the supported set.
func Detect() (Result, error) {
	goos := runtime.GOOS
	arch := runtime.GOARCH

	displayName, ok := supportedOS[goos]
	if !ok {
		return Result{
			OS:        goos,
			Arch:      arch,
			Supported: false,
		}, fmt.Errorf("unsupported operating system: %q (arch: %s). DevForge currently supports: macOS (darwin)", goos, arch)
	}

	return Result{
		OS:          goos,
		DisplayName: displayName,
		Arch:        arch,
		Supported:   true,
	}, nil
}
