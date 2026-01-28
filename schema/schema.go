// Package schema provides utilities for generating parameter schemas from Go structs.
package schema

import (
	"reflect"
	"strings"
)

// Parameter represents the schema for a function parameter.
type Parameter struct {
	// Type is the parameter type (string, number, boolean, array, object).
	Type string `json:"type"`

	// Description describes the parameter.
	Description string `json:"description,omitempty"`

	// Required indicates if the parameter is required.
	Required bool `json:"required,omitempty"`

	// Properties contains nested properties for object types.
	Properties map[string]Parameter `json:"properties,omitempty"`

	// Items describes the schema for array elements.
	Items *Parameter `json:"items,omitempty"`

	// Enum lists allowed values for enum types.
	Enum []any `json:"enum,omitempty"`

	// Default is the default value for the parameter.
	Default any `json:"default,omitempty"`

	// Example provides an example value.
	Example any `json:"example,omitempty"`

	// Minimum is the minimum value for numeric types.
	Minimum *float64 `json:"minimum,omitempty"`

	// Maximum is the maximum value for numeric types.
	Maximum *float64 `json:"maximum,omitempty"`

	// MinLength is the minimum length for string types.
	MinLength *int `json:"minLength,omitempty"`

	// MaxLength is the maximum length for string types.
	MaxLength *int `json:"maxLength,omitempty"`

	// Pattern is a regex pattern for string validation.
	Pattern string `json:"pattern,omitempty"`
}

// GenerateSchema generates a parameter schema from a struct.
// The struct should use the following tags:
//   - json: field name in JSON
//   - description: field description
//   - required: "true" if required
//   - enum: comma-separated list of allowed values
//   - default: default value
//   - example: example value
//   - min: minimum value
//   - max: maximum value
//   - minLength: minimum string length
//   - maxLength: maximum string length
//   - pattern: regex pattern
func GenerateSchema(v any) map[string]Parameter {
	t := reflect.TypeOf(v)
	if t == nil {
		return make(map[string]Parameter)
	}
	return generateSchemaFromType(t)
}

func generateSchemaFromType(t reflect.Type) map[string]Parameter {
	params := make(map[string]Parameter)

	if t == nil {
		return params
	}

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Return empty if not a struct
	if t.Kind() != reflect.Struct {
		return params
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag for field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		// Parse JSON tag to get field name
		fieldName := parseJSONTag(jsonTag, field.Name)

		// Generate parameter schema
		param := generateParameterFromField(field)
		params[fieldName] = param
	}

	return params
}

func parseJSONTag(tag, defaultName string) string {
	if tag == "" {
		return defaultName
	}

	parts := strings.Split(tag, ",")
	if parts[0] == "" {
		return defaultName
	}
	return parts[0]
}

func generateParameterFromField(field reflect.StructField) Parameter {
	param := Parameter{
		Description: field.Tag.Get("description"),
		Required:    field.Tag.Get("required") == "true",
	}

	// Determine type
	param.Type = getTypeString(field.Type)

	// Handle nested structs
	if field.Type.Kind() == reflect.Struct {
		param.Properties = generateSchemaFromType(field.Type)
	}

	// Handle slices and arrays
	if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array {
		elemType := field.Type.Elem()
		itemParam := Parameter{
			Type: getTypeString(elemType),
		}
		if elemType.Kind() == reflect.Struct {
			itemParam.Properties = generateSchemaFromType(elemType)
		}
		param.Items = &itemParam
	}

	// Handle pointer types
	if field.Type.Kind() == reflect.Ptr {
		elemType := field.Type.Elem()
		param.Type = getTypeString(elemType)
		if elemType.Kind() == reflect.Struct {
			param.Properties = generateSchemaFromType(elemType)
		}
	}

	// Parse enum tag
	if enumTag := field.Tag.Get("enum"); enumTag != "" {
		parts := strings.Split(enumTag, ",")
		for _, part := range parts {
			param.Enum = append(param.Enum, strings.TrimSpace(part))
		}
	}

	// Parse default tag
	if defaultTag := field.Tag.Get("default"); defaultTag != "" {
		param.Default = defaultTag
	}

	// Parse example tag
	if exampleTag := field.Tag.Get("example"); exampleTag != "" {
		param.Example = exampleTag
	}

	// Parse pattern tag
	if patternTag := field.Tag.Get("pattern"); patternTag != "" {
		param.Pattern = patternTag
	}

	return param
}

func getTypeString(t reflect.Type) string {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Struct:
		return "object"
	case reflect.Map:
		return "object"
	case reflect.Interface:
		return "any"
	default:
		return "string"
	}
}

// MergeSchemas merges multiple parameter schemas into one.
func MergeSchemas(schemas ...map[string]Parameter) map[string]Parameter {
	result := make(map[string]Parameter)
	for _, schema := range schemas {
		for k, v := range schema {
			result[k] = v
		}
	}
	return result
}

// ValidateParams validates parameters against a schema.
func ValidateParams(params map[string]any, schema map[string]Parameter) error {
	for name, paramSchema := range schema {
		if paramSchema.Required {
			val, exists := params[name]
			if !exists {
				return &ValidationError{
					Field:  name,
					Reason: "required field is missing",
				}
			}
			if val == nil || (reflect.ValueOf(val).Kind() == reflect.String && val.(string) == "") {
				return &ValidationError{
					Field:  name,
					Reason: "required field is empty",
				}
			}
		}
	}
	return nil
}

// ValidationError represents a schema validation error.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return "schema: validation failed for field " + e.Field + ": " + e.Reason
}
