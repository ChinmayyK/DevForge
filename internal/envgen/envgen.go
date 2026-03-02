// Package envgen generates .env files from .env.template files by
// prompting the user for values.
package envgen

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chinmay/devforge/internal/logger"
	"github.com/chinmay/devforge/internal/rollback"
)

// Generator handles .env file creation from templates.
type Generator struct {
	log    *logger.Logger
	rb     *rollback.Manager
	dryRun bool
	reader *bufio.Reader
}

// NewGenerator creates a Generator with the given logger, rollback
// manager, and stdin reader for user prompts.
func NewGenerator(log *logger.Logger, rb *rollback.Manager, dryRun bool) *Generator {
	return &Generator{
		log:    log,
		rb:     rb,
		dryRun: dryRun,
		reader: bufio.NewReader(os.Stdin),
	}
}

// Generate reads the .env.template file in projectDir, prompts the
// user for each key's value, and writes a .env file. It will not
// overwrite an existing .env file.
func (g *Generator) Generate(projectDir string) error {
	templatePath := filepath.Join(projectDir, ".env.template")
	envPath := filepath.Join(projectDir, ".env")

	// Check if .env.template exists.
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		g.log.Info("no .env.template found, skipping env generation")
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to check for .env.template: %w", err)
	}

	// Never overwrite an existing .env file.
	if _, err := os.Stat(envPath); err == nil {
		g.log.Warn(".env file already exists, skipping generation to avoid overwriting")
		return nil
	}

	// Parse the template file to extract keys and optional default values.
	keys, defaults, err := g.parseTemplate(templatePath)
	if err != nil {
		return err
	}

	if g.dryRun {
		g.log.Info(fmt.Sprintf("[dry-run] would generate .env with %d key(s): %s", len(keys), strings.Join(keys, ", ")))
		return nil
	}

	g.log.Info(fmt.Sprintf("generating .env file with %d key(s)...", len(keys)))

	// Prompt user for each key.
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		defaultVal := defaults[key]
		values[key], err = g.promptForValue(key, defaultVal)
		if err != nil {
			return fmt.Errorf("failed to read value for %q: %w", key, err)
		}
	}

	// Write the .env file.
	if err := g.writeEnvFile(envPath, keys, values); err != nil {
		return err
	}

	// Register rollback: remove the generated .env file.
	g.rb.Register("remove generated .env file", func() error {
		return os.Remove(envPath)
	})

	g.log.Info(fmt.Sprintf(".env file generated at %s", envPath))
	return nil
}

// parseTemplate reads a .env.template file and extracts KEY=DEFAULT
// pairs. Lines starting with # are treated as comments. Empty lines
// are skipped.
func (g *Generator) parseTemplate(path string) ([]string, map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open .env.template: %w", err)
	}
	defer file.Close()

	var keys []string
	defaults := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		keys = append(keys, key)
		if len(parts) == 2 {
			defaults[key] = strings.TrimSpace(parts[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error reading .env.template: %w", err)
	}

	return keys, defaults, nil
}

// promptForValue asks the user to provide a value for a configuration
// key. If a default is available, it is shown and used when the user
// presses Enter without typing anything.
func (g *Generator) promptForValue(key, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Printf("  %s [%s]: ", key, defaultVal)
	} else {
		fmt.Printf("  %s: ", key)
	}

	input, err := g.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	value := strings.TrimSpace(input)
	if value == "" && defaultVal != "" {
		return defaultVal, nil
	}
	return value, nil
}

// writeEnvFile writes the key=value pairs to a .env file preserving
// the original key order from the template.
func (g *Generator) writeEnvFile(path string, keys []string, values map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, key := range keys {
		if _, err := fmt.Fprintf(writer, "%s=%s\n", key, values[key]); err != nil {
			return fmt.Errorf("failed to write key %q to .env: %w", key, err)
		}
	}
	return writer.Flush()
}
