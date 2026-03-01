// Package config handles loading, parsing, and validating DevForge
// configuration files using Viper.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Dependency represents a single tool dependency that DevForge ensures
// is installed before scaffolding a project.
type Dependency struct {
	Name string `mapstructure:"name"`
}

// Config is the top-level configuration structure loaded from YAML.
type Config struct {
	// Dependencies lists the CLI tools required for the project.
	Dependencies []Dependency `mapstructure:"dependencies"`
	// Template is the Git URL of the starter template repository.
	Template string `mapstructure:"template"`
	// Linting indicates whether linting should be configured.
	Linting bool `mapstructure:"linting"`
	// GitHooks indicates whether git hooks should be set up.
	GitHooks bool `mapstructure:"gitHooks"`
	// EnvFile indicates whether a .env file should be generated.
	EnvFile bool `mapstructure:"envFile"`
}

// Load reads the configuration from the given file path. If configPath
// is empty, it falls back to looking for config/default.yaml relative
// to the working directory. It validates that all required fields are
// present and returns a typed Config.
func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("default")
		v.AddConfigPath("config")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// validate ensures all required configuration fields are present and
// contain valid values.
func validate(cfg *Config) error {
	var errs []string

	if len(cfg.Dependencies) == 0 {
		errs = append(errs, "at least one dependency must be specified")
	}
	for i, dep := range cfg.Dependencies {
		if dep.Name == "" {
			errs = append(errs, fmt.Sprintf("dependency at index %d has an empty name", i))
		}
	}

	if cfg.Template == "" {
		errs = append(errs, "template URL must be specified")
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration errors:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}
