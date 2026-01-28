// Package gopilot provides an AI-powered function router for Go applications.
// It enables natural language interaction with Go functions through LLM providers.
package gopilot

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SadikSunbul/gopilot/internal/prompt"
	"github.com/SadikSunbul/gopilot/provider"
	"github.com/SadikSunbul/gopilot/schema"
)

// Gopilot is the main entry point for the gopilot library.
type Gopilot struct {
	provider           provider.Provider
	registry           *Registry
	logger             Logger
	timeout            time.Duration
	systemPromptRules  []string
	middlewares        []Middleware
	unsupportedHandler FunctionWrapper
	retryConfig        *RetryConfig
	promptConfigured   bool
}

// New creates a new Gopilot instance with the given provider and options.
func New(p provider.Provider, opts ...Option) (*Gopilot, error) {
	if p == nil {
		return nil, ErrNilProvider
	}

	g := &Gopilot{
		provider: p,
		registry: NewRegistry(),
		logger:   newDefaultLogger(),
		timeout:  30 * time.Second,
	}

	// Apply options
	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

// Register registers a function with the gopilot instance.
func (g *Gopilot) Register(fn FunctionWrapper) error {
	if err := g.registry.Register(fn); err != nil {
		return err
	}
	g.logger.Debug("registered function: %s", fn.Name())
	return nil
}

// RegisterFunc is a helper to register a typed function.
func RegisterFunc[T any, R any](g *Gopilot, fn *Function[T, R]) error {
	return g.Register(fn)
}

// Unregister removes a function from the registry.
func (g *Gopilot) Unregister(name string) error {
	return g.registry.Unregister(name)
}

// Get retrieves a registered function by name.
func (g *Gopilot) Get(name string) (FunctionWrapper, error) {
	return g.registry.Get(name)
}

// Has checks if a function is registered.
func (g *Gopilot) Has(name string) bool {
	return g.registry.Has(name)
}

// List returns all registered functions.
func (g *Gopilot) List() []FunctionWrapper {
	return g.registry.List()
}

// ConfigurePrompt configures the system prompt with the registered functions.
// This must be called before Generate or GenerateAndExecute.
func (g *Gopilot) ConfigurePrompt() error {
	// Register unsupported handler if not set
	if g.unsupportedHandler == nil {
		g.unsupportedHandler = NewUnsupportedFunction()
	}

	if err := g.registry.Register(g.unsupportedHandler); err != nil {
		// Ignore if already registered
		if err != ErrFunctionExists {
			return fmt.Errorf("failed to register unsupported handler: %w", err)
		}
	}

	// Build the system prompt
	builder := prompt.NewBuilder()
	if len(g.systemPromptRules) > 0 {
		builder.WithRules(g.systemPromptRules)
	}

	for _, fn := range g.registry.List() {
		builder.AddFunction(prompt.FunctionInfo{
			Name:        fn.Name(),
			Description: fn.Description(),
			Parameters:  fn.Parameters(),
		})
	}

	systemPrompt := builder.Build()

	// Configure the provider if it supports it
	if configurer, ok := g.provider.(provider.PromptConfigurer); ok {
		configurer.SetSystemPrompt(systemPrompt)
	}

	g.promptConfigured = true
	g.logger.Info("system prompt configured with %d functions", g.registry.Count())
	return nil
}

// Generate sends a prompt to the LLM and returns the response.
func (g *Gopilot) Generate(ctx context.Context, input string) (*provider.Response, error) {
	if !g.promptConfigured {
		if err := g.ConfigurePrompt(); err != nil {
			return nil, err
		}
	}

	// Apply timeout if set
	if g.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.timeout)
		defer cancel()
	}

	g.logger.Debug("generating response for input: %s", input)

	var resp *provider.Response
	var err error

	if g.retryConfig != nil {
		resp, err = g.generateWithRetry(ctx, input)
	} else {
		resp, err = g.provider.Generate(ctx, input)
	}

	if err != nil {
		g.logger.Error("generation failed: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrGenerationFailed, err)
	}

	if resp == nil {
		return nil, ErrNoResponse
	}

	g.logger.Debug("generated response: agent=%s", resp.Agent)
	return resp, nil
}

func (g *Gopilot) generateWithRetry(ctx context.Context, input string) (*provider.Response, error) {
	var lastErr error
	delay := g.retryConfig.InitialDelay

	for attempt := 0; attempt < g.retryConfig.MaxAttempts; attempt++ {
		resp, err := g.provider.Generate(ctx, input)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		g.logger.Warn("generation attempt %d failed: %v", attempt+1, err)

		if attempt < g.retryConfig.MaxAttempts-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			delay = time.Duration(float64(delay) * g.retryConfig.Multiplier)
			if delay > g.retryConfig.MaxDelay {
				delay = g.retryConfig.MaxDelay
			}
		}
	}

	return nil, lastErr
}

// Execute executes a registered function with the given parameters.
func (g *Gopilot) Execute(ctx context.Context, name string, params map[string]any) (any, error) {
	// Apply middleware chain
	exec := func(name string, params map[string]any) (any, error) {
		return g.registry.Execute(ctx, name, params)
	}

	for i := len(g.middlewares) - 1; i >= 0; i-- {
		exec = g.middlewares[i](exec)
	}

	g.logger.Debug("executing function: %s", name)
	result, err := exec(name, params)
	if err != nil {
		g.logger.Error("execution failed for %s: %v", name, err)
		return nil, err
	}

	return result, nil
}

// GenerateAndExecute generates a response and executes the corresponding function.
func (g *Gopilot) GenerateAndExecute(ctx context.Context, input string) (any, error) {
	resp, err := g.Generate(ctx, input)
	if err != nil {
		return nil, err
	}

	return g.Execute(ctx, resp.Agent, resp.Parameters)
}

// ExecuteJSON executes a function with JSON parameters.
func (g *Gopilot) ExecuteJSON(ctx context.Context, name string, jsonParams string) (any, error) {
	var params map[string]any
	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		return nil, fmt.Errorf("failed to parse JSON parameters: %w", err)
	}
	return g.Execute(ctx, name, params)
}

// Pipeline creates a new pipeline builder.
func (g *Gopilot) Pipeline() *PipelineBuilder {
	return NewPipelineBuilder(g)
}

// ExecuteWorkflow executes a workflow definition.
func (g *Gopilot) ExecuteWorkflow(ctx context.Context, workflow *Workflow, input any) (*WorkflowResult, error) {
	return workflow.Execute(ctx, g, input)
}

// Provider returns the underlying LLM provider.
func (g *Gopilot) Provider() provider.Provider {
	return g.provider
}

// Registry returns the function registry.
func (g *Gopilot) Registry() *Registry {
	return g.registry
}

// Close closes the underlying provider.
func (g *Gopilot) Close() error {
	return g.provider.Close()
}

// Helper types for common parameter schemas

// StringParam creates a string parameter schema.
func StringParam(description string, required bool) schema.Parameter {
	return schema.Parameter{
		Type:        "string",
		Description: description,
		Required:    required,
	}
}

// IntParam creates an integer parameter schema.
func IntParam(description string, required bool) schema.Parameter {
	return schema.Parameter{
		Type:        "integer",
		Description: description,
		Required:    required,
	}
}

// NumberParam creates a number parameter schema.
func NumberParam(description string, required bool) schema.Parameter {
	return schema.Parameter{
		Type:        "number",
		Description: description,
		Required:    required,
	}
}

// BoolParam creates a boolean parameter schema.
func BoolParam(description string, required bool) schema.Parameter {
	return schema.Parameter{
		Type:        "boolean",
		Description: description,
		Required:    required,
	}
}

// ArrayParam creates an array parameter schema.
func ArrayParam(description string, required bool, items *schema.Parameter) schema.Parameter {
	return schema.Parameter{
		Type:        "array",
		Description: description,
		Required:    required,
		Items:       items,
	}
}

// ObjectParam creates an object parameter schema.
func ObjectParam(description string, required bool, properties map[string]schema.Parameter) schema.Parameter {
	return schema.Parameter{
		Type:        "object",
		Description: description,
		Required:    required,
		Properties:  properties,
	}
}
