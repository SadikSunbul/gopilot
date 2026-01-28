package gopilot

import (
	"context"
	"sync"
)

// Registry is a thread-safe function registry.
type Registry struct {
	functions map[string]FunctionWrapper
	mu        sync.RWMutex
}

// NewRegistry creates a new function registry.
func NewRegistry() *Registry {
	return &Registry{
		functions: make(map[string]FunctionWrapper),
	}
}

// Register adds a function to the registry.
func (r *Registry) Register(fn FunctionWrapper) error {
	if fn == nil {
		return ErrNilFunction
	}

	name := fn.Name()
	if name == "" {
		return ErrEmptyFunctionName
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.functions[name]; exists {
		return ErrFunctionExists
	}

	r.functions[name] = fn
	return nil
}

// Unregister removes a function from the registry.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.functions[name]; !exists {
		return ErrFunctionNotFound
	}

	delete(r.functions, name)
	return nil
}

// Get retrieves a function from the registry.
func (r *Registry) Get(name string) (FunctionWrapper, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, exists := r.functions[name]
	if !exists {
		return nil, ErrFunctionNotFound
	}

	return fn, nil
}

// Has checks if a function exists in the registry.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.functions[name]
	return exists
}

// Execute executes a registered function with the given parameters.
func (r *Registry) Execute(ctx context.Context, name string, params map[string]any) (any, error) {
	fn, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	result, err := fn.Execute(ctx, params)
	if err != nil {
		return nil, NewExecutionError(name, err)
	}

	return result, nil
}

// List returns all registered functions.
func (r *Registry) List() []FunctionWrapper {
	r.mu.RLock()
	defer r.mu.RUnlock()

	functions := make([]FunctionWrapper, 0, len(r.functions))
	for _, fn := range r.functions {
		functions = append(functions, fn)
	}

	return functions
}

// Names returns the names of all registered functions.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.functions))
	for name := range r.functions {
		names = append(names, name)
	}

	return names
}

// Count returns the number of registered functions.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.functions)
}

// Clear removes all functions from the registry.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.functions = make(map[string]FunctionWrapper)
}

// RegisterTyped is a helper function to register a typed function.
func RegisterTyped[T any, R any](r *Registry, fn *Function[T, R]) error {
	return r.Register(fn)
}

// GetTyped retrieves a typed function from the registry.
func GetTyped[T any, R any](r *Registry, name string) (*Function[T, R], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, exists := r.functions[name]
	if !exists {
		return nil, ErrFunctionNotFound
	}

	typedFn, ok := fn.(*Function[T, R])
	if !ok {
		return nil, NewParameterError(name, "incompatible function type")
	}

	return typedFn, nil
}

// ExecuteTyped executes a typed function with typed parameters.
func ExecuteTyped[T any, R any](r *Registry, ctx context.Context, name string, params T) (R, error) {
	var zero R

	fn, err := GetTyped[T, R](r, name)
	if err != nil {
		return zero, err
	}

	return fn.ExecuteTyped(ctx, params)
}
