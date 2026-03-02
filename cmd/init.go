package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/chinmay/devforge/internal/config"
	"github.com/chinmay/devforge/internal/envgen"
	"github.com/chinmay/devforge/internal/executor"
	"github.com/chinmay/devforge/internal/installer"
	"github.com/chinmay/devforge/internal/logger"
	"github.com/chinmay/devforge/internal/osdetect"
	"github.com/chinmay/devforge/internal/rollback"
	"github.com/chinmay/devforge/internal/template"
)

var initCmd = &cobra.Command{
	Use:   "init <project-name>",
	Short: "Scaffold a new project",
	Long: `Initialize a new project by:
  1. Detecting your OS
  2. Loading configuration
  3. Installing required dependencies
  4. Cloning the starter template
  5. Generating environment configuration

If any step fails, previously completed steps are automatically rolled back.`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// ── Step 1: Detect OS ──────────────────────────────────────────
	osResult, err := osdetect.Detect()
	if err != nil {
		return fmt.Errorf("OS detection failed: %w", err)
	}
	fmt.Printf("✓ OS detected: %s (%s/%s)\n", osResult.DisplayName, osResult.OS, osResult.Arch)

	// ── Step 2: Load config ────────────────────────────────────────
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	fmt.Printf("✓ Configuration loaded (%d dependencies, template: %s)\n", len(cfg.Dependencies), cfg.Template)

	// ── Step 3: Initialize logger ──────────────────────────────────
	log, err := logger.New(verbose)
	if err != nil {
		return fmt.Errorf("logger initialization failed: %w", err)
	}
	defer log.Close()

	log.Info("DevForge init started", map[string]interface{}{
		"project": projectName,
		"dryRun":  dryRun,
	})

	// ── Step 4: Initialize rollback manager ────────────────────────
	rb := rollback.NewManager(log)

	// ── Step 5: Install dependencies ───────────────────────────────
	exec := executor.New(log, dryRun)
	inst, err := installer.New(log, exec)
	if err != nil {
		return fmt.Errorf("installer initialization failed: %w", err)
	}

	for _, dep := range cfg.Dependencies {
		installed, err := inst.IsInstalled(dep.Name)
		if err != nil {
			log.Warn(fmt.Sprintf("could not check if %q is installed: %v", dep.Name, err))
		}
		if installed {
			version, _ := inst.GetVersion(dep.Name)
			fmt.Printf("✓ %s already installed (v%s)\n", dep.Name, version)
			log.Info(fmt.Sprintf("dependency %q already installed", dep.Name))
			continue
		}
		fmt.Printf("⟳ Installing %s...\n", dep.Name)
		if err := inst.Install(dep.Name); err != nil {
			log.Error(fmt.Sprintf("failed to install %q", dep.Name))
			rbErr := rb.Execute()
			if rbErr != nil {
				log.Error(fmt.Sprintf("rollback errors: %v", rbErr))
			}
			return fmt.Errorf("dependency installation failed for %q: %w", dep.Name, err)
		}
		fmt.Printf("✓ %s installed\n", dep.Name)
	}

	// ── Step 6: Clone template ─────────────────────────────────────
	destDir, err := filepath.Abs(projectName)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	cloner := template.NewCloner(log, rb, dryRun)
	if err := cloner.Clone(cfg.Template, destDir); err != nil {
		log.Error(fmt.Sprintf("template cloning failed: %v", err))
		rbErr := rb.Execute()
		if rbErr != nil {
			log.Error(fmt.Sprintf("rollback errors: %v", rbErr))
		}
		return fmt.Errorf("template cloning failed: %w", err)
	}
	fmt.Printf("✓ Template cloned into %s\n", destDir)

	// ── Step 7: Generate env file ──────────────────────────────────
	if cfg.EnvFile {
		gen := envgen.NewGenerator(log, rb, dryRun)
		if err := gen.Generate(destDir); err != nil {
			log.Error(fmt.Sprintf("env generation failed: %v", err))
			rbErr := rb.Execute()
			if rbErr != nil {
				log.Error(fmt.Sprintf("rollback errors: %v", rbErr))
			}
			return fmt.Errorf("env file generation failed: %w", err)
		}
		fmt.Println("✓ Environment configuration complete")
	}

	// ── Step 8: Print success summary ──────────────────────────────
	printSummary(projectName, destDir, cfg, dryRun)
	log.Info("DevForge init completed successfully")

	return nil
}

// printSummary displays a clear success message with next steps.
func printSummary(name, dir string, cfg *config.Config, dryRun bool) {
	fmt.Println()
	if dryRun {
		fmt.Println("═══════════════════════════════════════════")
		fmt.Println("  DRY RUN COMPLETE — no changes were made")
		fmt.Println("═══════════════════════════════════════════")
	} else {
		fmt.Println("═══════════════════════════════════════════")
		fmt.Printf("  🚀 Project %q created successfully!\n", name)
		fmt.Println("═══════════════════════════════════════════")
	}
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Printf("    cd %s\n", dir)
	if cfg.Linting {
		fmt.Println("    # Linting is enabled")
	}
	if cfg.GitHooks {
		fmt.Println("    # Git hooks are configured")
	}
	fmt.Println()
	fmt.Println("  Run 'devforge doctor' to verify system readiness.")
	fmt.Println()
}
