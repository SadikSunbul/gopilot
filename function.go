package gopilot

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SadikSunbul/gopilot/internal/mapper"
	"github.com/SadikSunbul/gopilot/schema"
)

// FunctionWrapper is the interface that all registered functions must implement.
type FunctionWrapper interface {
	// Name returns the function name.
	Name() string

	// Description returns the function description.
	Description() string

	// Parameters returns the function's parameter schema.
	Parameters() map[string]schema.Parameter

	// Execute executes the function with the given parameters.
	Execute(ctx context.Context, params map[string]any) (any, error)
}

// Function is a generic function wrapper that provides type-safe execution.
type Function[T any, R any] struct {
	// FnName is the name of the function.
	FnName string `json:"name"`

	// FnDescription is the description of the function.
	FnDescription string `json:"description"`

	// FnParameters is the parameter schema for the function.
	FnParameters map[string]schema.Parameter `json:"parameters"`

	// Handler is the function to execute.
	Handler func(ctx context.Context, params T) (R, error)
}

// Ensure Function implements FunctionWrapper.
var _ FunctionWrapper = (*Function[any, any])(nil)

// Name returns the function name.
func (f *Function[T, R]) Name() string {
	return f.FnName
}

// Description returns the function description.
func (f *Function[T, R]) Description() string {
	return f.FnDescription
}

// Parameters returns the function's parameter schema.
func (f *Function[T, R]) Parameters() map[string]schema.Parameter {
	return f.FnParameters
}

// Execute executes the function with the given parameters.
func (f *Function[T, R]) Execute(ctx context.Context, params map[string]any) (any, error) {
	var typedParams T
	m := mapper.New()

	if err := m.Map(params, &typedParams); err != nil {
		return nil, fmt.Errorf("parameter mapping failed: %w", err)
	}

	return f.Handler(ctx, typedParams)
}

// ExecuteTyped executes the function and returns a typed result.
func (f *Function[T, R]) ExecuteTyped(ctx context.Context, params T) (R, error) {
	return f.Handler(ctx, params)
}

// ExecuteWithJSON executes the function with JSON string parameters.
func (f *Function[T, R]) ExecuteWithJSON(ctx context.Context, jsonParams string) (any, error) {
	var params map[string]any
	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %w", err)
	}
	return f.Execute(ctx, params)
}

// NewFunction creates a new function with the given configuration.
func NewFunction[T any, R any](name, description string, handler func(ctx context.Context, params T) (R, error)) *Function[T, R] {
	var params T
	return &Function[T, R]{
		FnName:        name,
		FnDescription: description,
		FnParameters:  schema.GenerateSchema(params),
		Handler:       handler,
	}
}

// NewFunctionWithSchema creates a new function with a custom parameter schema.
func NewFunctionWithSchema[T any, R any](
	name, description string,
	params map[string]schema.Parameter,
	handler func(ctx context.Context, params T) (R, error),
) *Function[T, R] {
	return &Function[T, R]{
		FnName:        name,
		FnDescription: description,
		FnParameters:  params,
		Handler:       handler,
	}
}

// SimpleFunction is a non-generic function wrapper for simpler use cases.
type SimpleFunction struct {
	FnName        string
	FnDescription string
	FnParameters  map[string]schema.Parameter
	Handler       func(ctx context.Context, params map[string]any) (any, error)
}

// Ensure SimpleFunction implements FunctionWrapper.
var _ FunctionWrapper = (*SimpleFunction)(nil)

// Name returns the function name.
func (f *SimpleFunction) Name() string {
	return f.FnName
}

// Description returns the function description.
func (f *SimpleFunction) Description() string {
	return f.FnDescription
}

// Parameters returns the function's parameter schema.
func (f *SimpleFunction) Parameters() map[string]schema.Parameter {
	return f.FnParameters
}

// Execute executes the function with the given parameters.
func (f *SimpleFunction) Execute(ctx context.Context, params map[string]any) (any, error) {
	return f.Handler(ctx, params)
}

// NewSimpleFunction creates a new simple function.
func NewSimpleFunction(
	name, description string,
	params map[string]schema.Parameter,
	handler func(ctx context.Context, params map[string]any) (any, error),
) *SimpleFunction {
	return &SimpleFunction{
		FnName:        name,
		FnDescription: description,
		FnParameters:  params,
		Handler:       handler,
	}
}

// UnsupportedParams represents parameters for the unsupported function.
type UnsupportedParams struct {
	Message string `json:"message" description:"Contains a simple explanation of the error." required:"true"`
}

// UnsupportedResponse represents the response from the unsupported function.
type UnsupportedResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// NewUnsupportedFunction creates the default unsupported function handler.
func NewUnsupportedFunction() *Function[UnsupportedParams, UnsupportedResponse] {
	return NewFunction[UnsupportedParams, UnsupportedResponse](
		"unsupported",
		"If the user's request doesn't match any available function, use this function to explain why.",
		func(ctx context.Context, params UnsupportedParams) (UnsupportedResponse, error) {
			return UnsupportedResponse{
				Message: "Unsupported request: " + params.Message,
				Success: false,
			}, nil
		},
	)
}
