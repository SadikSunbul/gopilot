// Package mapper provides utilities for mapping parameters to structs.
package mapper

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Mapper converts map[string]any to typed structs.
type Mapper struct {
	// StrictMode when true, returns an error if the map contains unknown fields.
	StrictMode bool
}

// New creates a new Mapper with default settings.
func New() *Mapper {
	return &Mapper{
		StrictMode: false,
	}
}

// NewStrict creates a new Mapper with strict mode enabled.
func NewStrict() *Mapper {
	return &Mapper{
		StrictMode: true,
	}
}

// Map converts a map to a typed struct.
func (m *Mapper) Map(params map[string]any, target any) error {
	if params == nil {
		return nil
	}

	// Convert through JSON for proper type handling
	data, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("mapper: failed to marshal params: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("mapper: failed to unmarshal to target: %w", err)
	}

	// Validate required fields
	if err := m.validateRequired(target); err != nil {
		return err
	}

	return nil
}

// MapJSON converts a JSON string to a typed struct.
func (m *Mapper) MapJSON(jsonStr string, target any) error {
	var params map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &params); err != nil {
		return fmt.Errorf("mapper: failed to parse JSON: %w", err)
	}
	return m.Map(params, target)
}

// validateRequired validates that all required fields have values.
func (m *Mapper) validateRequired(target any) error {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Check if field is required
		if field.Tag.Get("required") == "true" {
			if isZeroValue(fieldValue) {
				jsonTag := field.Tag.Get("json")
				fieldName := parseJSONFieldName(jsonTag, field.Name)
				return &RequiredFieldError{Field: fieldName}
			}
		}

		// Recursively validate nested structs
		if fieldValue.Kind() == reflect.Struct {
			if err := m.validateRequired(fieldValue.Addr().Interface()); err != nil {
				return err
			}
		}
	}

	return nil
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

func parseJSONFieldName(tag, defaultName string) string {
	if tag == "" {
		return defaultName
	}
	for i, c := range tag {
		if c == ',' {
			if i > 0 {
				return tag[:i]
			}
			return defaultName
		}
	}
	return tag
}

// RequiredFieldError is returned when a required field is missing.
type RequiredFieldError struct {
	Field string
}

func (e *RequiredFieldError) Error() string {
	return fmt.Sprintf("mapper: required field %q is missing or empty", e.Field)
}

// ToMap converts a struct to a map.
func ToMap(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("mapper: failed to marshal: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("mapper: failed to unmarshal to map: %w", err)
	}

	return result, nil
}

// MergeMaps merges multiple maps into one.
// Later maps override earlier ones for duplicate keys.
func MergeMaps(maps ...map[string]any) map[string]any {
	result := make(map[string]any)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
