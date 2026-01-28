package schema

import (
	"testing"
)

type simpleStruct struct {
	Name   string  `json:"name" description:"The name" required:"true"`
	Age    int     `json:"age" description:"The age"`
	Active bool    `json:"active"`
	Score  float64 `json:"score"`
}

type nestedStruct struct {
	User simpleStruct `json:"user" description:"User info"`
	Tags []string     `json:"tags" description:"Tags"`
}

type complexStruct struct {
	ID      string        `json:"id" required:"true"`
	Data    nestedStruct  `json:"data"`
	Options *simpleStruct `json:"options"`
}

func TestGenerateSchema_Simple(t *testing.T) {
	schema := GenerateSchema(simpleStruct{})

	t.Run("field count", func(t *testing.T) {
		if len(schema) != 4 {
			t.Errorf("expected 4 fields, got %d", len(schema))
		}
	})

	t.Run("string field", func(t *testing.T) {
		param, ok := schema["name"]
		if !ok {
			t.Fatal("name field should exist")
		}
		if param.Type != "string" {
			t.Errorf("expected type 'string', got %s", param.Type)
		}
		if param.Description != "The name" {
			t.Errorf("expected description 'The name', got %s", param.Description)
		}
		if !param.Required {
			t.Error("name should be required")
		}
	})

	t.Run("int field", func(t *testing.T) {
		param, ok := schema["age"]
		if !ok {
			t.Fatal("age field should exist")
		}
		if param.Type != "integer" {
			t.Errorf("expected type 'integer', got %s", param.Type)
		}
	})

	t.Run("bool field", func(t *testing.T) {
		param, ok := schema["active"]
		if !ok {
			t.Fatal("active field should exist")
		}
		if param.Type != "boolean" {
			t.Errorf("expected type 'boolean', got %s", param.Type)
		}
	})

	t.Run("float field", func(t *testing.T) {
		param, ok := schema["score"]
		if !ok {
			t.Fatal("score field should exist")
		}
		if param.Type != "number" {
			t.Errorf("expected type 'number', got %s", param.Type)
		}
	})
}

func TestGenerateSchema_Nested(t *testing.T) {
	schema := GenerateSchema(nestedStruct{})

	t.Run("nested struct", func(t *testing.T) {
		param, ok := schema["user"]
		if !ok {
			t.Fatal("user field should exist")
		}
		if param.Type != "object" {
			t.Errorf("expected type 'object', got %s", param.Type)
		}
		if param.Properties == nil {
			t.Fatal("nested properties should not be nil")
		}
		if len(param.Properties) != 4 {
			t.Errorf("expected 4 nested properties, got %d", len(param.Properties))
		}
	})

	t.Run("array field", func(t *testing.T) {
		param, ok := schema["tags"]
		if !ok {
			t.Fatal("tags field should exist")
		}
		if param.Type != "array" {
			t.Errorf("expected type 'array', got %s", param.Type)
		}
		if param.Items == nil {
			t.Fatal("array items should not be nil")
		}
		if param.Items.Type != "string" {
			t.Errorf("expected items type 'string', got %s", param.Items.Type)
		}
	})
}

func TestGenerateSchema_Complex(t *testing.T) {
	schema := GenerateSchema(complexStruct{})

	t.Run("field count", func(t *testing.T) {
		if len(schema) != 3 {
			t.Errorf("expected 3 fields, got %d", len(schema))
		}
	})

	t.Run("pointer field", func(t *testing.T) {
		param, ok := schema["options"]
		if !ok {
			t.Fatal("options field should exist")
		}
		// Pointer fields should resolve to their underlying type
		if param.Type != "object" {
			t.Errorf("expected type 'object', got %s", param.Type)
		}
	})
}

func TestGenerateSchema_Pointer(t *testing.T) {
	schema := GenerateSchema(&simpleStruct{})

	if len(schema) != 4 {
		t.Errorf("expected 4 fields from pointer, got %d", len(schema))
	}
}

func TestGenerateSchema_NonStruct(t *testing.T) {
	schema := GenerateSchema("not a struct")

	if len(schema) != 0 {
		t.Error("non-struct should return empty schema")
	}
}

func TestGenerateSchema_WithTags(t *testing.T) {
	type taggedStruct struct {
		Field1 string `json:"field_1" enum:"a,b,c" default:"a" example:"b"`
		Field2 string `json:"field_2" pattern:"^[a-z]+$"`
		Field3 string `json:"-"` // Should be ignored
	}

	schema := GenerateSchema(taggedStruct{})

	t.Run("field count (ignoring json:-)", func(t *testing.T) {
		if len(schema) != 2 {
			t.Errorf("expected 2 fields, got %d", len(schema))
		}
	})

	t.Run("enum tag", func(t *testing.T) {
		param := schema["field_1"]
		if len(param.Enum) != 3 {
			t.Errorf("expected 3 enum values, got %d", len(param.Enum))
		}
	})

	t.Run("default tag", func(t *testing.T) {
		param := schema["field_1"]
		if param.Default != "a" {
			t.Errorf("expected default 'a', got %v", param.Default)
		}
	})

	t.Run("example tag", func(t *testing.T) {
		param := schema["field_1"]
		if param.Example != "b" {
			t.Errorf("expected example 'b', got %v", param.Example)
		}
	})

	t.Run("pattern tag", func(t *testing.T) {
		param := schema["field_2"]
		if param.Pattern != "^[a-z]+$" {
			t.Errorf("expected pattern '^[a-z]+$', got %s", param.Pattern)
		}
	})
}

func TestGenerateSchema_UnexportedFields(t *testing.T) {
	type structWithUnexported struct {
		Public  string `json:"public"`
		private string // unexported (intentionally ignored)
	}

	// Validate that unexported fields are ignored by GenerateSchema.
	// Keep the field referenced to avoid linters that flag unused fields.
	_ = structWithUnexported{}.private

	schema := GenerateSchema(structWithUnexported{})

	if len(schema) != 1 {
		t.Errorf("expected 1 field (unexported should be ignored), got %d", len(schema))
	}

	if _, ok := schema["public"]; !ok {
		t.Error("public field should exist")
	}
}

func TestMergeSchemas(t *testing.T) {
	schema1 := map[string]Parameter{
		"a": {Type: "string"},
		"b": {Type: "integer"},
	}

	schema2 := map[string]Parameter{
		"b": {Type: "number"}, // Override
		"c": {Type: "boolean"},
	}

	merged := MergeSchemas(schema1, schema2)

	if len(merged) != 3 {
		t.Errorf("expected 3 fields, got %d", len(merged))
	}

	if merged["b"].Type != "number" {
		t.Error("later schema should override earlier")
	}
}

func TestValidateParams(t *testing.T) {
	schema := map[string]Parameter{
		"name":     {Type: "string", Required: true},
		"age":      {Type: "integer", Required: false},
		"required": {Type: "string", Required: true},
	}

	t.Run("valid params", func(t *testing.T) {
		params := map[string]any{
			"name":     "John",
			"required": "value",
		}

		err := ValidateParams(params, schema)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing required", func(t *testing.T) {
		params := map[string]any{
			"name": "John",
			// missing "required"
		}

		err := ValidateParams(params, schema)
		if err == nil {
			t.Error("expected error for missing required field")
		}
	})

	t.Run("empty required", func(t *testing.T) {
		params := map[string]any{
			"name":     "John",
			"required": "", // empty
		}

		err := ValidateParams(params, schema)
		if err == nil {
			t.Error("expected error for empty required field")
		}
	})

	t.Run("nil required", func(t *testing.T) {
		params := map[string]any{
			"name":     "John",
			"required": nil,
		}

		err := ValidateParams(params, schema)
		if err == nil {
			t.Error("expected error for nil required field")
		}
	})
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:  "name",
		Reason: "is required",
	}

	expected := "schema: validation failed for field name: is required"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestArrayOfStructs(t *testing.T) {
	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type container struct {
		Items []item `json:"items"`
	}

	schema := GenerateSchema(container{})

	items, ok := schema["items"]
	if !ok {
		t.Fatal("items field should exist")
	}

	if items.Type != "array" {
		t.Errorf("expected type 'array', got %s", items.Type)
	}

	if items.Items == nil {
		t.Fatal("items should have Items defined")
	}

	if items.Items.Type != "object" {
		t.Errorf("expected items type 'object', got %s", items.Items.Type)
	}

	if items.Items.Properties == nil {
		t.Fatal("array items should have properties for struct elements")
	}

	if len(items.Items.Properties) != 2 {
		t.Errorf("expected 2 properties in array items, got %d", len(items.Items.Properties))
	}
}

func TestMapField(t *testing.T) {
	type withMap struct {
		Data map[string]any `json:"data"`
	}

	schema := GenerateSchema(withMap{})

	data, ok := schema["data"]
	if !ok {
		t.Fatal("data field should exist")
	}

	if data.Type != "object" {
		t.Errorf("expected type 'object' for map, got %s", data.Type)
	}
}

func TestIntegerTypes(t *testing.T) {
	type intTypes struct {
		Int8   int8   `json:"int8"`
		Int16  int16  `json:"int16"`
		Int32  int32  `json:"int32"`
		Int64  int64  `json:"int64"`
		Uint8  uint8  `json:"uint8"`
		Uint16 uint16 `json:"uint16"`
		Uint32 uint32 `json:"uint32"`
		Uint64 uint64 `json:"uint64"`
	}

	schema := GenerateSchema(intTypes{})

	for name, param := range schema {
		if param.Type != "integer" {
			t.Errorf("field %s: expected type 'integer', got %s", name, param.Type)
		}
	}
}
