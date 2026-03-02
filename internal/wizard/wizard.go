// Package wizard provides an interactive TUI setup experience when
// DevForge is run without a configuration file. It walks the user
// through dependency selection, template choice, and project options,
// then returns a fully populated *config.Config.
package wizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/chinmay/devforge/internal/config"
	"github.com/chinmay/devforge/internal/logger"
	"github.com/chinmay/devforge/internal/registry"
	"github.com/chinmay/devforge/internal/ux"
)

// knownDependency is a curated tool that can be selected in the wizard.
type knownDependency struct {
	Name        string
	Description string
}

// availableDeps is the curated list of common development tools.
var availableDeps = []knownDependency{
	{Name: "node", Description: "Node.js runtime"},
	{Name: "git", Description: "Version control system"},
	{Name: "docker", Description: "Container runtime"},
	{Name: "python", Description: "Python interpreter"},
	{Name: "go", Description: "Go programming language"},
	{Name: "rust", Description: "Rust toolchain (rustc)"},
	{Name: "java", Description: "Java Development Kit"},
	{Name: "ruby", Description: "Ruby interpreter"},
}

// Run launches the interactive wizard and returns a populated config.
// registryURL is the endpoint for fetching templates.
func Run(registryURL string, verbose, jsonLogs bool) (*config.Config, error) {
	// ── Banner ──────────────────────────────────────────────────
	fmt.Println()
	fmt.Printf("  %s%sDevForge Interactive Setup%s\n", ux.Blue, "", ux.Reset)
	fmt.Printf("  %s═════════════════════════%s\n", ux.Blue, ux.Reset)
	fmt.Println()
	fmt.Printf("  %sNo configuration file found. Let's set up your project interactively.%s\n\n", ux.Gray, ux.Reset)

	// ── Step 1: Select dependencies ─────────────────────────────
	depOptions := make([]huh.Option[string], len(availableDeps))
	for i, d := range availableDeps {
		depOptions[i] = huh.NewOption(
			fmt.Sprintf("%-8s — %s", d.Name, d.Description),
			d.Name,
		)
	}

	var selectedDeps []string
	depForm := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select dependencies to install").
				Description("Use space to toggle, enter to confirm").
				Options(depOptions...).
				Value(&selectedDeps),
		),
	)

	if err := depForm.Run(); err != nil {
		return nil, fmt.Errorf("dependency selection cancelled: %w", err)
	}

	if len(selectedDeps) == 0 {
		fmt.Printf("\n  %s⚠ No dependencies selected. You can add them later in devforge.yaml%s\n\n", ux.Yellow, ux.Reset)
	}

	// ── Step 2: Version for each selected dependency ────────────
	deps := make([]config.Dependency, 0, len(selectedDeps))
	for _, name := range selectedDeps {
		var version string
		versionForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("Version for %s", name)).
					Description("Leave empty for latest").
					Placeholder("latest").
					Value(&version),
			),
		)

		if err := versionForm.Run(); err != nil {
			return nil, fmt.Errorf("version input cancelled: %w", err)
		}

		version = strings.TrimSpace(version)
		if version == "" {
			version = "latest"
		}
		deps = append(deps, config.Dependency{Name: name, Version: version})
	}

	// ── Step 3: Choose template from registry ───────────────────
	templateURL, err := selectTemplate(registryURL, verbose, jsonLogs)
	if err != nil {
		return nil, err
	}

	// ── Step 4: Project options ─────────────────────────────────
	var envFile, linting, gitHooks bool

	optionsForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Generate .env file?").
				Description("Creates a starter .env in your project").
				Value(&envFile),
			huh.NewConfirm().
				Title("Setup linting?").
				Description("Configure linting tools for the project").
				Value(&linting),
			huh.NewConfirm().
				Title("Setup git hooks?").
				Description("Install pre-commit and other git hooks").
				Value(&gitHooks),
		),
	)

	if err := optionsForm.Run(); err != nil {
		return nil, fmt.Errorf("options selection cancelled: %w", err)
	}

	// ── Step 5: Confirmation ────────────────────────────────────
	fmt.Println()
	fmt.Printf("  %sSummary%s\n", ux.Blue, ux.Reset)
	fmt.Printf("  %s───────%s\n", ux.Blue, ux.Reset)
	if len(deps) > 0 {
		depNames := make([]string, len(deps))
		for i, d := range deps {
			if d.Version == "latest" {
				depNames[i] = d.Name
			} else {
				depNames[i] = fmt.Sprintf("%s@%s", d.Name, d.Version)
			}
		}
		fmt.Printf("  Dependencies: %s\n", strings.Join(depNames, ", "))
	} else {
		fmt.Printf("  Dependencies: (none)\n")
	}
	fmt.Printf("  Template:     %s\n", templateURL)
	fmt.Printf("  Env file:     %v\n", envFile)
	fmt.Printf("  Linting:      %v\n", linting)
	fmt.Printf("  Git hooks:    %v\n", gitHooks)
	fmt.Println()

	var confirmed bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Proceed with this configuration?").
				Value(&confirmed),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return nil, fmt.Errorf("confirmation cancelled: %w", err)
	}

	if !confirmed {
		return nil, fmt.Errorf("setup cancelled by user")
	}

	// ── Build config ────────────────────────────────────────────
	cfg := &config.Config{
		Dependencies: deps,
		Template:     templateURL,
		Linting:      linting,
		GitHooks:     gitHooks,
		EnvFile:      envFile,
	}

	return cfg, nil
}

// selectTemplate fetches the registry and presents a template picker.
// Falls back to manual URL input if the registry is unreachable.
func selectTemplate(registryURL string, verbose, jsonLogs bool) (string, error) {
	log, err := logger.New(verbose, jsonLogs)
	if err != nil {
		return promptManualURL()
	}
	defer log.Close()

	client := registry.NewClient(registryURL, log)
	reg, err := client.Fetch(false)
	if err != nil {
		fmt.Printf("  %s⚠ Could not fetch template registry. You can enter a URL manually.%s\n\n", ux.Yellow, ux.Reset)
		return promptManualURL()
	}

	templates := reg.ValidTemplates()
	if len(templates) == 0 {
		fmt.Printf("  %s⚠ Registry has no templates. Enter a template URL manually.%s\n\n", ux.Yellow, ux.Reset)
		return promptManualURL()
	}

	// Build options: registry templates + manual entry
	options := make([]huh.Option[string], 0, len(templates)+1)
	for _, t := range templates {
		label := fmt.Sprintf("%-20s %s", t.Name, t.Description)
		if len(label) > 60 {
			label = label[:57] + "..."
		}
		options = append(options, huh.NewOption(label, t.URL))
	}
	options = append(options, huh.NewOption("✎  Enter a custom Git URL...", "__custom__"))

	var selectedURL string
	templateForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a starter template").
				Description("Templates from the DevForge registry").
				Options(options...).
				Value(&selectedURL),
		),
	)

	if err := templateForm.Run(); err != nil {
		return "", fmt.Errorf("template selection cancelled: %w", err)
	}

	if selectedURL == "__custom__" {
		return promptManualURL()
	}

	return selectedURL, nil
}

// promptManualURL asks the user to type a Git repository URL.
func promptManualURL() (string, error) {
	var url string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Template Git URL").
				Description("e.g. https://github.com/user/template.git").
				Placeholder("https://github.com/...").
				Value(&url).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return fmt.Errorf("a template URL is required")
					}
					if !strings.HasPrefix(s, "https://") {
						return fmt.Errorf("URL must start with https://")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("URL input cancelled: %w", err)
	}

	return strings.TrimSpace(url), nil
}
