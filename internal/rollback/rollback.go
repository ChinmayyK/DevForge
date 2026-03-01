// Package rollback provides a mechanism to undo operations when a
// multi-step process fails partway through. Actions are executed in
// reverse registration order (LIFO).
package rollback

import (
	"fmt"
	"strings"
	"sync"

	"github.com/chinmay/devforge/internal/logger"
)

// Action represents a single rollback operation with a human-readable
// description and the function to execute.
type Action struct {
	Description string
	Fn          func() error
}

// Manager collects rollback actions and executes them in reverse order
// when needed. It is safe for concurrent use.
type Manager struct {
	actions []Action
	log     *logger.Logger
	mu      sync.Mutex
}

// NewManager creates a RollbackManager with the given logger.
func NewManager(log *logger.Logger) *Manager {
	return &Manager{
		log: log,
	}
}

// Register adds a rollback action to the stack. Actions are executed
// in reverse order (last registered = first executed).
func (m *Manager) Register(description string, fn func() error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.actions = append(m.actions, Action{
		Description: description,
		Fn:          fn,
	})
	m.log.Debug(fmt.Sprintf("registered rollback action: %s", description))
}

// Execute runs all registered rollback actions in reverse order. It
// collects any errors but does not stop on individual failures, ensuring
// all rollback steps are attempted. Returns a combined error summary
// if any actions failed.
func (m *Manager) Execute() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.actions) == 0 {
		m.log.Info("no rollback actions to execute")
		return nil
	}

	m.log.Warn(fmt.Sprintf("executing %d rollback action(s)...", len(m.actions)))

	var failures []string
	for i := len(m.actions) - 1; i >= 0; i-- {
		action := m.actions[i]
		m.log.Info(fmt.Sprintf("  rollback: %s", action.Description))
		if err := action.Fn(); err != nil {
			errMsg := fmt.Sprintf("%s: %v", action.Description, err)
			failures = append(failures, errMsg)
			m.log.Error(fmt.Sprintf("  rollback failed: %s", errMsg))
		} else {
			m.log.Info(fmt.Sprintf("  rollback succeeded: %s", action.Description))
		}
	}

	// Clear actions after execution to prevent double-rollback.
	m.actions = nil

	if len(failures) > 0 {
		return fmt.Errorf("rollback completed with %d error(s):\n  - %s", len(failures), strings.Join(failures, "\n  - "))
	}

	m.log.Info("all rollback actions completed successfully")
	return nil
}

// HasActions returns true if there are registered rollback actions.
func (m *Manager) HasActions() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.actions) > 0
}
