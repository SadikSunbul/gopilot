package gopilot

import (
	"fmt"
	"sync"
)

// Registry, all functions are managed by this structure
type Registry struct {
	functions map[string]*Function
	mu        sync.RWMutex
}

// NewRegistry, new Registry
func NewRegistry() *Registry {
	return &Registry{
		functions: make(map[string]*Function),
	}
}

// Register, new function
func (r *Registry) register(fn *Function) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if fn == nil {
		return fmt.Errorf("function cannot be nil")
	}

	if fn.Name == "" {
		return fmt.Errorf("function name cannot be empty")
	}

	if _, exists := r.functions[fn.Name]; exists {
		return fmt.Errorf("function '%s' already registered", fn.Name)
	}

	r.functions[fn.Name] = fn
	return nil
}

// Get, name function
func (r *Registry) get(name string) (*Function, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, exists := r.functions[name]
	if !exists {
		return nil, fmt.Errorf("function '%s' not found", name)
	}

	return fn, nil
}

// List, all functions
func (r *Registry) list() []*Function {
	r.mu.RLock()
	defer r.mu.RUnlock()

	functions := make([]*Function, 0, len(r.functions))
	for _, fn := range r.functions {
		functions = append(functions, fn)
	}
	return functions
}

// Execute, function
func (r *Registry) execute(name string, params map[string]interface{}) (interface{}, error) {
	fn, err := r.get(name)
	if err != nil {
		return nil, err
	}

	// Parameter validation
	for paramName, schema := range fn.Parameters {
		if schema.Required {
			if _, exists := params[paramName]; !exists {
				return nil, fmt.Errorf("required parameter missing: %s", paramName)
			}
		}
	}

	return fn.Execute(params)
}
