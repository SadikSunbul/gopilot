package gopilot

import (
	"fmt"
	"sync"

	"github.com/SadikSunbul/gopilot/pkg/generator"
)

// FunctionWrapper is an interface that all functions must implement
type FunctionWrapper interface {
	GetName() string
	GetDescription() string
	GetParameters() map[string]generator.ParameterSchema
	ExecuteWithMap(map[string]interface{}) (interface{}, error)
}

// Registry represents a thread-safe function registry
type Registry struct {
	functions map[string]FunctionWrapper
	mu        sync.RWMutex
}

// NewRegistry creates a new function registry
func NewRegistry() *Registry {
	return &Registry{
		functions: make(map[string]FunctionWrapper),
	}
}

// RegisterTyped adds a new typed function to the registry
func RegisterTyped[T any, R any](r *Registry, fn *Function[T, R]) error {
	return r.Register(fn)
}

// Register adds a new function to the registry
func (r *Registry) Register(fn FunctionWrapper) error {
	if fn == nil {
		return fmt.Errorf("function cannot be nil")
	}

	name := fn.GetName()
	if name == "" {
		return fmt.Errorf("function name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.functions[name]; exists {
		return fmt.Errorf("function %s already registered", name)
	}

	r.functions[name] = fn
	return nil
}

// Execute runs a registered function with the given parameters
func Execute[T any, R any](r *Registry, name string, params T) (R, error) {
	r.mu.RLock()
	fn, exists := r.functions[name]
	r.mu.RUnlock()

	if !exists {
		var zero R
		return zero, fmt.Errorf("function %s not found", name)
	}

	// Debug bilgisi ekle
	fmt.Printf("Function type: %T\n", fn)
	fmt.Printf("Params type: %T\n", params)

	typedFn, ok := fn.(*Function[T, R])
	if !ok {
		var zero R
		return zero, fmt.Errorf("function %s has incompatible parameter type (got: %T, want: *Function[%T, %T])", name, fn, *new(T), *new(R))
	}

	return typedFn.Execute(params)
}

// Get retrieves a function from the registry
func Get[T any, R any](r *Registry, name string) (*Function[T, R], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, exists := r.functions[name]
	if !exists {
		return nil, fmt.Errorf("function %s not found", name)
	}

	typedFn, ok := fn.(*Function[T, R])
	if !ok {
		return nil, fmt.Errorf("function %s has incompatible type", name)
	}

	return typedFn, nil
}

// Get retrieves a function from the registry
func (r *Registry) Get(name string) (FunctionWrapper, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, exists := r.functions[name]
	return fn, exists
}

// ExecuteFunction executes a registered function with the given parameters
func (r *Registry) ExecuteFunction(name string, params map[string]interface{}) (interface{}, error) {
	fn, exists := r.Get(name)
	if !exists {
		return nil, fmt.Errorf("function %s not found", name)
	}

	return fn.ExecuteWithMap(params)
}

// List returns all registered functions
func (r *Registry) List() []FunctionWrapper {
	r.mu.RLock()
	defer r.mu.RUnlock()

	functions := make([]FunctionWrapper, 0, len(r.functions))
	for _, fn := range r.functions {
		functions = append(functions, fn)
	}

	return functions
}
