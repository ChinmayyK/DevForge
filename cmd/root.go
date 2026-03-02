// Package cmd implements the CLI commands for DevForge using Cobra.
package cmd

import (
	"os"

	"github.com/chinmay/devforge/internal/ux"
	"github.com/spf13/cobra"
)

// Version is set via ldflags at build time.
var Version = "dev"

// Flag variables shared across commands.
var (
	cfgFile  string
	dryRun   bool
	verbose  bool
	jsonLogs bool
	force    bool
)

// rootCmd is the base command for the DevForge CLI.
var rootCmd = &cobra.Command{
	Use:   "devforge <command>",
	Short: "DevForge — Development Environment Automation CLI",
	Long: `DevForge — Development Environment Automation CLI

A powerful, standalone CLI tool that hardens and automates project scaffolding:
  ✔ Detects OS and architecture automatically
  ✔ Installs required toolchains safely
  ✔ Uses smart dependency resolution and version pinning
  ✔ Clones organization templates securely
  ✔ Provides professional configuration management

Built for elite developers who value speed, safety, and consistency.`,
}

// Execute runs the root command. This is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ux.Error(err)
		os.Exit(1)
	}
}

// SetVersion allows main.go to inject the build-time version.
func SetVersion(v string) {
	Version = v
	rootCmd.Version = v
}

func init() {
	// Custom premium help template.
	rootCmd.SetHelpTemplate(`{{.Long}}

Usage:
  {{.UseLine}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
`)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file (default: config/default.yaml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "simulate operations without making changes")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&jsonLogs, "json-logs", false, "structured JSON output")
	rootCmd.PersistentFlags().BoolVar(&force, "force", false, "force overwrite of existing directories/files")
}
