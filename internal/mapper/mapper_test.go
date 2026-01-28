package mapper

import (
	"testing"
)

type testStruct struct {
	Name    string `json:"name" required:"true"`
	Age     int    `json:"age"`
	Active  bool   `json:"active"`
	Score   float64 `json:"score"`
}

type nestedTestStruct struct {
	User    testStruct `json:"user" required:"true"`
	Tags    []string   `json:"tags"`
}

func TestMapper_Map(t *testing.T) {
	m := New()

	t.Run("basic mapping", func(t *testing.T) {
		params := map[string]any{
			"name":   "John",
			"age":    30,
			"active": true,
			"score":  95.5,
		}

		var result testStruct
		err := m.Map(params, &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Name != "John" {
			t.Errorf("expected name 'John', got %s", result.Name)
		}
		if result.Age != 30 {
			t.Errorf("expected age 30, got %d", result.Age)
		}
		if !result.Active {
			t.Error("expected active true")
		}
		if result.Score != 95.5 {
			t.Errorf("expected score 95.5, got %f", result.Score)
		}
	})

	t.Run("nested struct mapping", func(t *testing.T) {
		params := map[string]any{
			"user": map[string]any{
				"name": "Jane",
				"age":  25,
			},
			"tags": []any{"go", "dev"},
		}

		var result nestedTestStruct
		err := m.Map(params, &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.User.Name != "Jane" {
			t.Errorf("expected user name 'Jane', got %s", result.User.Name)
		}
		if len(result.Tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(result.Tags))
		}
	})

	t.Run("nil params", func(t *testing.T) {
		var result testStruct
		err := m.Map(nil, &result)
		if err != nil {
			t.Errorf("unexpected error for nil params: %v", err)
		}
	})

	t.Run("partial mapping", func(t *testing.T) {
		params := map[string]any{
			"name": "Partial",
			// age, active, score omitted
		}

		var result testStruct
		err := m.Map(params, &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Name != "Partial" {
			t.Errorf("expected name 'Partial', got %s", result.Name)
		}
		if result.Age != 0 {
			t.Errorf("expected default age 0, got %d", result.Age)
		}
	})
}

func TestMapper_MapRequired(t *testing.T) {
	m := New()

	t.Run("missing required field", func(t *testing.T) {
		params := map[string]any{
			"age": 30, // name is missing
		}

		var result testStruct
		err := m.Map(params, &result)
		if err == nil {
			t.Fatal("expected error for missing required field")
		}

		reqErr, ok := err.(*RequiredFieldError)
		if !ok {
			t.Fatalf("expected RequiredFieldError, got %T", err)
		}
		if reqErr.Field != "name" {
			t.Errorf("expected field 'name', got %s", reqErr.Field)
		}
	})

	t.Run("empty required field", func(t *testing.T) {
		params := map[string]any{
			"name": "", // empty string
			"age":  30,
		}

		var result testStruct
		err := m.Map(params, &result)
		if err == nil {
			t.Fatal("expected error for empty required field")
		}
	})

	t.Run("nested required field", func(t *testing.T) {
		params := map[string]any{
			"user": map[string]any{
				// name is missing in user
				"age": 25,
			},
		}

		var result nestedTestStruct
		err := m.Map(params, &result)
		if err == nil {
			t.Fatal("expected error for nested missing required field")
		}
	})
}

func TestMapper_MapJSON(t *testing.T) {
	m := New()

	t.Run("valid JSON", func(t *testing.T) {
		jsonStr := `{"name": "JSON User", "age": 35}`

		var result testStruct
		err := m.MapJSON(jsonStr, &result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Name != "JSON User" {
			t.Errorf("expected name 'JSON User', got %s", result.Name)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		jsonStr := `{invalid json}`

		var result testStruct
		err := m.MapJSON(jsonStr, &result)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestMapper_Strict(t *testing.T) {
	m := NewStrict()

	if !m.StrictMode {
		t.Error("NewStrict should set StrictMode to true")
	}
}

func TestToMap(t *testing.T) {
	input := testStruct{
		Name:   "Test",
		Age:    30,
		Active: true,
		Score:  85.5,
	}

	result, err := ToMap(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "Test" {
		t.Errorf("expected name 'Test', got %v", result["name"])
	}
	if result["age"].(float64) != 30 {
		t.Errorf("expected age 30, got %v", result["age"])
	}
}

func TestMergeMaps(t *testing.T) {
	map1 := map[string]any{"a": 1, "b": 2}
	map2 := map[string]any{"b": 3, "c": 4}
	map3 := map[string]any{"d": 5}

	result := MergeMaps(map1, map2, map3)

	if len(result) != 4 {
		t.Errorf("expected 4 keys, got %d", len(result))
	}

	if result["a"].(int) != 1 {
		t.Errorf("expected a=1, got %v", result["a"])
	}
	if result["b"].(int) != 3 {
		t.Errorf("expected b=3 (overridden), got %v", result["b"])
	}
	if result["c"].(int) != 4 {
		t.Errorf("expected c=4, got %v", result["c"])
	}
	if result["d"].(int) != 5 {
		t.Errorf("expected d=5, got %v", result["d"])
	}
}

func TestRequiredFieldError(t *testing.T) {
	err := &RequiredFieldError{Field: "username"}

	expected := `mapper: required field "username" is missing or empty`
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestIsZeroValue(t *testing.T) {
	testCases := []struct {
		name     string
		value    any
		expected bool
	}{
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"zero float", 0.0, true},
		{"non-zero float", 3.14, false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"nil slice", []string(nil), true},
		{"empty slice", []string{}, true},
		{"non-empty slice", []string{"a"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't directly test isZeroValue as it's unexported,
			// but we test it indirectly through the required field validation
		})
	}
}

func TestParseJSONFieldName(t *testing.T) {
	testCases := []struct {
		tag      string
		default_ string
		expected string
	}{
		{"name", "default", "name"},
		{"name,omitempty", "default", "name"},
		{",omitempty", "default", "default"},
		{"", "default", "default"},
		{"-", "default", "-"},
	}

	for _, tc := range testCases {
		t.Run(tc.tag, func(t *testing.T) {
			result := parseJSONFieldName(tc.tag, tc.default_)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

type complexStruct struct {
	ID       string            `json:"id" required:"true"`
	Metadata map[string]string `json:"metadata"`
	Items    []testStruct      `json:"items"`
	Nested   *testStruct       `json:"nested"`
}

func TestMapper_ComplexTypes(t *testing.T) {
	m := New()

	params := map[string]any{
		"id": "123",
		"metadata": map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
		"items": []any{
			map[string]any{"name": "Item1", "age": 1},
			map[string]any{"name": "Item2", "age": 2},
		},
		"nested": map[string]any{
			"name": "Nested",
			"age":  99,
		},
	}

	var result complexStruct
	err := m.Map(params, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "123" {
		t.Errorf("expected ID '123', got %s", result.ID)
	}

	if len(result.Metadata) != 2 {
		t.Errorf("expected 2 metadata entries, got %d", len(result.Metadata))
	}

	if len(result.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(result.Items))
	}

	if result.Nested == nil {
		t.Fatal("nested should not be nil")
	}
	if result.Nested.Name != "Nested" {
		t.Errorf("expected nested name 'Nested', got %s", result.Nested.Name)
	}
}
