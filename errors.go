// Package gopilot provides an AI-powered function router for Go applications.
package gopilot

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions.
var (
	// ErrNilProvider is returned when a nil provider is passed to New.
	ErrNilProvider = errors.New("gopilot: provider cannot be nil")

	// ErrNilFunction is returned when attempting to register a nil function.
	ErrNilFunction = errors.New("gopilot: function cannot be nil")

	// ErrEmptyFunctionName is returned when a function has an empty name.
	ErrEmptyFunctionName = errors.New("gopilot: function name cannot be empty")

	// ErrFunctionNotFound is returned when a requested function is not registered.
	ErrFunctionNotFound = errors.New("gopilot: function not found")

	// ErrFunctionExists is returned when attempting to register a function with a name that already exists.
	ErrFunctionExists = errors.New("gopilot: function already registered")

	// ErrInvalidParameters is returned when function parameters are invalid.
	ErrInvalidParameters = errors.New("gopilot: invalid parameters")

	// ErrMissingRequiredParam is returned when a required parameter is missing.
	ErrMissingRequiredParam = errors.New("gopilot: required parameter is missing")

	// ErrPipelineEmpty is returned when attempting to execute an empty pipeline.
	ErrPipelineEmpty = errors.New("gopilot: pipeline has no steps")

	// ErrWorkflowEmpty is returned when attempting to execute an empty workflow.
	ErrWorkflowEmpty = errors.New("gopilot: workflow has no steps")

	// ErrGenerationFailed is returned when LLM generation fails.
	ErrGenerationFailed = errors.New("gopilot: generation failed")

	// ErrNoResponse is returned when LLM returns no response.
	ErrNoResponse = errors.New("gopilot: no response from provider")

	// ErrInvalidResponse is returned when LLM response cannot be parsed.
	ErrInvalidResponse = errors.New("gopilot: invalid response format")
)

// ExecutionError represents an error that occurred during function execution.
type ExecutionError struct {
	// Function is the name of the function that failed.
	Function string

	// Cause is the underlying error.
	Cause error
}

// Error implements the error interface.
func (e *ExecutionError) Error() string {
	return fmt.Sprintf("gopilot: execution failed for function %q: %v", e.Function, e.Cause)
}

// Unwrap returns the underlying error.
func (e *ExecutionError) Unwrap() error {
	return e.Cause
}

// NewExecutionError creates a new ExecutionError.
func NewExecutionError(function string, cause error) *ExecutionError {
	return &ExecutionError{
		Function: function,
		Cause:    cause,
	}
}

// PipelineError represents an error that occurred during pipeline execution.
type PipelineError struct {
	// Step is the index of the step that failed (0-based).
	Step int

	// Function is the name of the function that failed.
	Function string

	// Cause is the underlying error.
	Cause error
}

// Error implements the error interface.
func (e *PipelineError) Error() string {
	return fmt.Sprintf("gopilot: pipeline failed at step %d (%s): %v", e.Step, e.Function, e.Cause)
}

// Unwrap returns the underlying error.
func (e *PipelineError) Unwrap() error {
	return e.Cause
}

// NewPipelineError creates a new PipelineError.
func NewPipelineError(step int, function string, cause error) *PipelineError {
	return &PipelineError{
		Step:     step,
		Function: function,
		Cause:    cause,
	}
}

// ParameterError represents an error related to function parameters.
type ParameterError struct {
	// Parameter is the name of the problematic parameter.
	Parameter string

	// Reason describes why the parameter is invalid.
	Reason string
}

// Error implements the error interface.
func (e *ParameterError) Error() string {
	return fmt.Sprintf("gopilot: parameter %q error: %s", e.Parameter, e.Reason)
}

// NewParameterError creates a new ParameterError.
func NewParameterError(parameter, reason string) *ParameterError {
	return &ParameterError{
		Parameter: parameter,
		Reason:    reason,
	}
}

// IsExecutionError checks if the error is an ExecutionError.
func IsExecutionError(err error) bool {
	var execErr *ExecutionError
	return errors.As(err, &execErr)
}

// IsPipelineError checks if the error is a PipelineError.
func IsPipelineError(err error) bool {
	var pipeErr *PipelineError
	return errors.As(err, &pipeErr)
}

// IsParameterError checks if the error is a ParameterError.
func IsParameterError(err error) bool {
	var paramErr *ParameterError
	return errors.As(err, &paramErr)
}
