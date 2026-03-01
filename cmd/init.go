package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/chinmay/devforge/internal/config"
	"github.com/chinmay/devforge/internal/envgen"
	"github.com/chinmay/devforge/internal/errors"
	"github.com/chinmay/devforge/internal/executor"
	"github.com/chinmay/devforge/internal/installer"
	"github.com/chinmay/devforge/internal/logger"
	"github.com/chinmay/devforge/internal/osdetect"
	"github.com/chinmay/devforge/internal/registry"
	"github.com/chinmay/devforge/internal/rollback"
	"github.com/chinmay/devforge/internal/security"
	"github.com/chinmay/devforge/internal/semver"
	"github.com/chinmay/devforge/internal/template"
	"github.com/chinmay/devforge/internal/ux"
)

var templateName string

var initCmd = &cobra.Command{
	Use:   "init <project-name>",
	Short: "Scaffold a new project",
	Long: `Initialize a new project by:
  1. Detecting your OS and package manager
  2. Loading configuration
  3. Installing required dependencies (with version pinning)
  4. Cloning the starter template
  5. Generating environment configuration

If any step fails, previously completed steps are automatically rolled back.`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVarP(&templateName, "template", "t", "", "name of the starter template from the remote registry")
	rootCmd.AddCommand(initCmd)
}
func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Validate project name for safety.
	if err := security.ValidateName(projectName); err != nil {
		ux.Error(errors.New(errors.CodeInvalidConfig, "invalid project name", "use alphanumeric characters and dashes only"))
		return nil
	}

	destDir, err := filepath.Abs(projectName)
	if err != nil {
		ux.Error(fmt.Errorf("failed to resolve project path: %v", err))
		return nil
	}

	// ── Safety Check: Prevent accidental overwrite ─────────────────
	if _, err := os.Stat(destDir); err == nil {
		if !force {
			ux.Error(errors.New(
				errors.CodePathExists,
				fmt.Sprintf("directory %q already exists", projectName),
				"use the --force flag to overwrite",
			))
			return nil
		}
		ux.Warning("Directory %q exists. Proceeding due to --force flag.", projectName)
	}

	// ── Step 1: Detect OS ──────────────────────────────────────────
	osInfo, err := osdetect.DetectFull()
	if err != nil {
		return fmt.Errorf("OS detection failed: %w", err)
	}
	fmt.Printf("✓ OS detected: %s (%s/%s) — package manager: %s\n", osInfo.Name, osInfo.RawOS, osInfo.Arch, osInfo.PackageMgr)

	// ── Step 2: Load config ────────────────────────────────────────
	var cfg *config.Config
	var loadErr error

	if templateName != "" {
		fmt.Printf("⟳ Fetching template %q from remote registry...\n", templateName)
		client, _, err := getRegistryClient()
		if err != nil {
			return fmt.Errorf("registry client initialization failed: %w", err)
		}
		reg, err := client.Fetch(false)
		if err != nil {
			return fmt.Errorf("failed to fetch registry: %w", err)
		}

		var matched registry.Template
		found := false
		for _, t := range reg.ValidTemplates() {
			if strings.EqualFold(t.Name, templateName) {
				matched = t
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("template %q not found in registry", templateName)
		}

		// Convert GitHub URL to Raw Content URL for the devforge.yaml file
		rawURL := strings.Replace(matched.URL, "github.com", "raw.githubusercontent.com", 1)
		rawURL = strings.TrimSuffix(rawURL, ".git") + "/main/devforge.yaml"

		resp, err := http.Get(rawURL)
		if err != nil {
			return fmt.Errorf("failed to fetch remote devforge.yaml: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("template %q does not contain a valid devforge.yaml on main branch (HTTP %d)", templateName, resp.StatusCode)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read remote devforge.yaml: %w", err)
		}

		cfg, loadErr = config.LoadFromBytes(bodyBytes)
		if loadErr == nil {
			// Override the parsed template URL to ensure it matches the actual registry URL
			cfg.Template = matched.URL
		}
	} else {
		cfg, loadErr = config.Load(cfgFile)
	}

	if loadErr != nil {
		return fmt.Errorf("configuration error: %w", loadErr)
	}

	fmt.Printf("✓ Configuration loaded (%d dependencies, template: %s)\n", len(cfg.Dependencies), cfg.Template)

	// ── Step 3: Initialize logger ──────────────────────────────────
	log, err := logger.New(verbose, jsonLogs)
	if err != nil {
		return fmt.Errorf("logger initialization failed: %w", err)
	}
	defer log.Close()

	log.Info("DevForge init started", map[string]interface{}{
		"project":    projectName,
		"dryRun":     dryRun,
		"os":         osInfo.Name,
		"packageMgr": osInfo.PackageMgr,
	})

	// ── Step 4: Initialize rollback manager ────────────────────────
	rb := rollback.NewManager(log)

	// ── Step 5: Install dependencies ───────────────────────────────
	exec := executor.New(log, dryRun)
	inst, err := installer.NewFromOS(log, exec, osInfo)
	if err != nil {
		ux.Error(fmt.Errorf("installer initialization failed: %v", err))
		return nil
	}

	ux.Step("Installing dependencies")
	var wg sync.WaitGroup
	errCh := make(chan error, len(cfg.Dependencies))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Limit concurrency to 3 simultaneous installs
	sem := make(chan struct{}, 3)

	for _, dep := range cfg.Dependencies {
		wg.Add(1)
		go func(dep config.Dependency) {
			defer wg.Done()

			// Check context cancellation
			select {
			case <-ctx.Done():
				return
			case sem <- struct{}{}: // Acquire semaphore
				defer func() { <-sem }() // Release semaphore
			}

			installed, checkErr := inst.IsInstalled(dep.Name)
			if checkErr != nil {
				log.Warn(fmt.Sprintf("could not check if %q is installed: %v", dep.Name, checkErr))
			}

			if installed {
				currentVersion, _ := inst.GetVersion(dep.Name)
				ux.Success("%s already installed (v%s)", dep.Name, currentVersion)

				// Version mismatch check
				if dep.Version != "" && dep.Version != "latest" && currentVersion != "" {
					desired, parseErr := semver.Parse(dep.Version)
					current, curParseErr := semver.Parse(currentVersion)
					if parseErr == nil && curParseErr == nil && !desired.IsZero() {
						if !desired.MajorMatches(current) {
							ux.Warning("version mismatch for %q: installed=%s, wanted=%s", dep.Name, currentVersion, dep.Version)
						}
					}
				}
				return
			}

			versionLabel := dep.Version
			if versionLabel == "" || versionLabel == "latest" {
				versionLabel = "latest"
			}
			ux.Step("Installing %s (v%s)", dep.Name, versionLabel)

			if err := inst.Install(dep.Name, dep.Version); err != nil {
				errCh <- errors.New(
					errors.CodeExecutionFailed,
					fmt.Sprintf("failed to install %q", dep.Name),
					"check your network connection or package manager logs",
				)
				cancel() // Fail fast
				return
			}
			ux.Success("%s installed", dep.Name)
		}(dep)
	}

	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		ux.Error(err)
		rb.Execute()
		return nil
	}

	// ── Step 6: Clone template ─────────────────────────────────────
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

	// ── Done ───────────────────────────────────────────────────────
	fmt.Println()
	if dryRun {
		fmt.Println("═══════════════════════════════════════════")
		fmt.Println("  DRY RUN COMPLETE — no changes were made")
		fmt.Println("═══════════════════════════════════════════")
	} else {
		fmt.Println("═══════════════════════════════════════════")
		fmt.Printf("  🚀 Project %q created successfully!\n", projectName)
		fmt.Println("═══════════════════════════════════════════")
	}
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Printf("    cd %s\n", destDir)
	fmt.Println()
	fmt.Println("  Run 'devforge doctor' to verify system readiness.")
	fmt.Println()

	log.Info("DevForge init completed successfully")
	return nil
}
