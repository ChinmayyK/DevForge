// Package errors provides a structured error type for user-friendly
// DevForge CLI output, avoiding raw stack traces.
package errors

import "fmt"

// DevForgeError represents a structured, user-facing error.
type DevForgeError struct {
	Code    string
	Message string
	Hint    string
}

// Error implements the standard error interface.
func (e *DevForgeError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("[%s] %s\nHint: %s", e.Code, e.Message, e.Hint)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// New creates a new structured DevForgeError.
func New(code, message, hint string) *DevForgeError {
	return &DevForgeError{
		Code:    code,
		Message: message,
		Hint:    hint,
	}
}

// Common error codes.
const (
	CodeMissingDependency = "ERR_MISSING_DEPENDENCY"
	CodeVersionMismatch   = "ERR_VERSION_MISMATCH"
	CodeTemplateClone     = "ERR_TEMPLATE_CLONE"
	CodeExecutionFailed   = "ERR_EXECUTION_FAILED"
	CodeNetworkFailure    = "ERR_NETWORK_FAILURE"
	CodeInvalidConfig     = "ERR_INVALID_CONFIG"
	CodePathExists        = "ERR_PATH_EXISTS"
)
