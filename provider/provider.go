// Package provider defines the interface for LLM providers.
package provider

import (
	"context"
)

// Step represents a single step in a multi-step execution plan.
// Providers can return either a single (agent+parameters) response or a plan consisting of steps.
type Step struct {
	// Function is the name of the registered function to execute.
	Function string `json:"function"`

	// Params are optional static parameters for this step.
	Params map[string]any `json:"params,omitempty"`

	// StoreAs, if set, stores this step's output under the given variable name.
	// It can later be referenced in params as "$<store_as>.<field>".
	StoreAs string `json:"store_as,omitempty"`

	// MapOutput controls whether the output of this step becomes the input params for the next step.
	// If omitted (nil), the default is treated as true by the orchestrator.
	MapOutput *bool `json:"map_output,omitempty"`

	// Retries is the number of retries for this step (0 means no retries).
	Retries int `json:"retries,omitempty"`
}

// Response represents the response from an LLM provider.
type Response struct {
	// Type indicates how to interpret the response.
	// Supported values:
	// - "single": use Agent + Parameters
	// - "plan": use Steps
	// If empty, implementations should fall back to Agent + Parameters for backward compatibility.
	Type string `json:"type,omitempty"`

	// Agent is the name of the function to execute.
	Agent string `json:"agent"`

	// Parameters are the parameters for the function.
	Parameters map[string]any `json:"parameters"`

	// Steps is an optional multi-step plan produced by the provider.
	Steps []Step `json:"steps,omitempty"`

	// Explanation is optional human-readable planning context.
	Explanation string `json:"explanation,omitempty"`

	// Raw is the raw response from the provider (optional).
	Raw string `json:"-"`
}

// Provider defines the interface that all LLM providers must implement.
type Provider interface {
	// Generate generates a response based on the given prompt.
	// The context should be used for cancellation and timeouts.
	Generate(ctx context.Context, prompt string) (*Response, error)

	// Close releases any resources held by the provider.
	Close() error
}

// PromptConfigurer is an optional interface for providers that support
// configuring the system prompt.
type PromptConfigurer interface {
	// SetSystemPrompt sets the system prompt for the provider.
	SetSystemPrompt(prompt string)
}

// ModelConfigurer is an optional interface for providers that support
// runtime model configuration.
type ModelConfigurer interface {
	// SetModel changes the model being used.
	SetModel(model string) error

	// GetModel returns the current model name.
	GetModel() string
}

// StreamingProvider is an optional interface for providers that support
// streaming responses.
type StreamingProvider interface {
	Provider

	// GenerateStream generates a streaming response.
	GenerateStream(ctx context.Context, prompt string) (<-chan StreamChunk, error)
}

// StreamChunk represents a chunk in a streaming response.
type StreamChunk struct {
	// Content is the content of this chunk.
	Content string

	// Done indicates if this is the final chunk.
	Done bool

	// Error is set if an error occurred.
	Error error
}

// HealthChecker is an optional interface for providers that support
// health checks.
type HealthChecker interface {
	// HealthCheck performs a health check on the provider.
	HealthCheck(ctx context.Context) error
}
