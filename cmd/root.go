// Package cmd implements the CLI commands for DevForge using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Flag variables shared across commands.
var (
	cfgFile string
	dryRun  bool
	verbose bool
)

// rootCmd is the base command for the DevForge CLI.
var rootCmd = &cobra.Command{
	Use:   "devforge",
	Short: "DevForge — production-grade project scaffolding tool",
	Long: `DevForge is a cross-platform CLI tool that automates project setup:

  • Detects your operating system and architecture
  • Installs missing dependencies via your platform's package manager
  • Clones starter template repositories
  • Generates .env configuration files
  • Provides system health checks via the doctor command

Built for developers who value automation and consistency.`,
}

// Execute runs the root command. This is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file (default: config/default.yaml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "simulate all operations without making changes")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable debug-level logging output")
}
