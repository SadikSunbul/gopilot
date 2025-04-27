package generator

import (
	"reflect"
)

// ParameterSchema represents the schema for a function parameter
type ParameterSchema struct {
	Type        string                     `json:"type"`
	Description string                     `json:"description"`
	Required    bool                       `json:"required"`
	Properties  map[string]ParameterSchema `json:"properties,omitempty"`
}

// GenerateParameterSchema generates ParameterSchema from a struct using reflection
func GenerateParameterSchema(structType interface{}) map[string]ParameterSchema {
	params := make(map[string]ParameterSchema)
	t := reflect.TypeOf(structType)

	// If pointer, get the actual type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Return empty if not a Struct
	if t.Kind() != reflect.Struct {
		return params
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Check the JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		// Get description and required tags
		description := field.Tag.Get("description")
		required := field.Tag.Get("required") == "true"

		// Specify field type
		fieldType := "string" // default type
		switch field.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldType = "number"
		case reflect.Float32, reflect.Float64:
			fieldType = "number"
		case reflect.Bool:
			fieldType = "boolean"
		case reflect.Struct:
			// Recursive call for nested struct
			nestedParams := GenerateParameterSchema(reflect.New(field.Type).Interface())
			params[field.Name] = ParameterSchema{
				Type:        "interface",
				Description: description,
				Required:    required,
				Properties:  nestedParams,
			}
			continue
		case reflect.Slice, reflect.Array:
			fieldType = "array"
		}

		// Create ParameterSchema
		params[field.Name] = ParameterSchema{
			Type:        fieldType,
			Description: description,
			Required:    required,
		}
	}

	return params
}
