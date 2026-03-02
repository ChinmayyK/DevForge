// DevForge is a production-grade CLI tool for automated project scaffolding.
// It detects your OS, installs dependencies, clones starter templates, and
// generates environment configuration — all with rollback support.
package main

import "github.com/chinmay/devforge/cmd"

func main() {
	cmd.Execute()
}
