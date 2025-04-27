package gopilot

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/SadikSunbul/gopilot/pkg/generator"
)

// Function represents a registered function in the system
type Function[T any, R any] struct {
	Name        string                               `json:"name"`
	Description string                               `json:"description"`
	Parameters  map[string]generator.ParameterSchema `json:"parameters"`
	Execute     func(params T) (R, error)
}

// GetName returns the function name
func (f *Function[T, R]) GetName() string {
	return f.Name
}

// GetDescription returns the function description
func (f *Function[T, R]) GetDescription() string {
	return f.Description
}

// GetParameters returns the function parameters
func (f *Function[T, R]) GetParameters() map[string]generator.ParameterSchema {
	return f.Parameters
}

// ExecuteWithMap executes the function with a parameter map
func (f *Function[T, R]) ExecuteWithMap(params map[string]interface{}) (interface{}, error) {
	var typedParams T
	mapper := &ParamsMapper{}

	if err := mapper.Map(params, &typedParams); err != nil {
		return nil, fmt.Errorf("parameter mapping failed: %w", err)
	}

	result, err := f.Execute(typedParams)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ExecuteWithJson executes the function with JSON string parameters
func (f *Function[T, R]) ExecuteWithJson(jsonParams string) (interface{}, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		return nil, fmt.Errorf("json parsing failed: %w", err)
	}

	return f.ExecuteWithMap(params)
}

// ParamsMapper helps convert generic parameters to strongly typed structs
type ParamsMapper struct{}

// Map converts generic parameters to a strongly typed struct
func (pm *ParamsMapper) Map(params map[string]interface{}, target interface{}) error {
	data, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal params to target struct: %w", err)
	}

	// Validate required fields
	v := reflect.ValueOf(target).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("required") == "true" && v.Field(i).IsZero() {
			return fmt.Errorf("required field %s is missing or empty", field.Name)
		}
	}

	return nil
}

// ExecuteWithParams is a helper function to execute a function with typed parameters
func ExecuteWithParams[T any, R any](fn func(T) (R, error), params map[string]interface{}) (interface{}, error) {
	var parameters T
	mapper := &ParamsMapper{}

	if err := mapper.Map(params, &parameters); err != nil {
		return nil, fmt.Errorf("failed to map parameters: %w", err)
	}

	result, err := fn(parameters)
	if err != nil {
		return nil, err
	}

	return result, nil
}
