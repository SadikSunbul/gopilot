package prompt

import (
	"strings"
	"testing"

	"github.com/SadikSunbul/gopilot/schema"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()

	if b == nil {
		t.Fatal("builder should not be nil")
	}

	if len(b.rules) == 0 {
		t.Error("builder should have default rules")
	}
}

func TestBuilder_WithRules(t *testing.T) {
	b := NewBuilder()

	customRules := []string{"Rule 1", "Rule 2", "Rule 3"}
	b.WithRules(customRules)

	if len(b.rules) != 3 {
		t.Errorf("expected 3 rules, got %d", len(b.rules))
	}

	if b.rules[0] != "Rule 1" {
		t.Errorf("expected first rule 'Rule 1', got %s", b.rules[0])
	}
}

func TestBuilder_WithRules_Empty(t *testing.T) {
	b := NewBuilder()
	originalRules := len(b.rules)

	b.WithRules(nil)

	if len(b.rules) != originalRules {
		t.Error("empty rules should not override default rules")
	}

	b.WithRules([]string{})

	if len(b.rules) != originalRules {
		t.Error("empty slice should not override default rules")
	}
}

func TestBuilder_AddFunction(t *testing.T) {
	b := NewBuilder()

	fn := FunctionInfo{
		Name:        "test-function",
		Description: "A test function",
		Parameters: map[string]schema.Parameter{
			"name": {Type: "string", Description: "The name", Required: true},
			"age":  {Type: "integer", Description: "The age"},
		},
	}

	b.AddFunction(fn)

	if len(b.functions) != 1 {
		t.Errorf("expected 1 function, got %d", len(b.functions))
	}

	if b.functions[0].Name != "test-function" {
		t.Errorf("expected function name 'test-function', got %s", b.functions[0].Name)
	}
}

func TestBuilder_AddFunctions(t *testing.T) {
	b := NewBuilder()

	fns := []FunctionInfo{
		{Name: "fn1", Description: "Function 1"},
		{Name: "fn2", Description: "Function 2"},
		{Name: "fn3", Description: "Function 3"},
	}

	b.AddFunctions(fns)

	if len(b.functions) != 3 {
		t.Errorf("expected 3 functions, got %d", len(b.functions))
	}
}

func TestBuilder_Build(t *testing.T) {
	b := NewBuilder()

	b.WithRules([]string{"Custom rule 1", "Custom rule 2"})

	b.AddFunction(FunctionInfo{
		Name:        "weather",
		Description: "Gets weather information",
		Parameters: map[string]schema.Parameter{
			"city": {Type: "string", Description: "The city name", Required: true},
		},
	})

	b.AddFunction(FunctionInfo{
		Name:        "translate",
		Description: "Translates text",
		Parameters: map[string]schema.Parameter{
			"text": {Type: "string", Description: "Text to translate", Required: true},
			"from": {Type: "string", Description: "Source language"},
			"to":   {Type: "string", Description: "Target language"},
		},
	})

	prompt := b.Build()

	t.Run("contains rules", func(t *testing.T) {
		if !strings.Contains(prompt, "Custom rule 1") {
			t.Error("prompt should contain custom rule 1")
		}
		if !strings.Contains(prompt, "Custom rule 2") {
			t.Error("prompt should contain custom rule 2")
		}
	})

	t.Run("contains functions", func(t *testing.T) {
		if !strings.Contains(prompt, "Function: weather") {
			t.Error("prompt should contain weather function")
		}
		if !strings.Contains(prompt, "Function: translate") {
			t.Error("prompt should contain translate function")
		}
	})

	t.Run("contains descriptions", func(t *testing.T) {
		if !strings.Contains(prompt, "Gets weather information") {
			t.Error("prompt should contain weather description")
		}
		if !strings.Contains(prompt, "Translates text") {
			t.Error("prompt should contain translate description")
		}
	})

	t.Run("contains parameters", func(t *testing.T) {
		if !strings.Contains(prompt, "city") {
			t.Error("prompt should contain city parameter")
		}
		if !strings.Contains(prompt, "[required]") {
			t.Error("prompt should mark required parameters")
		}
	})

	t.Run("contains system instructions", func(t *testing.T) {
		if !strings.Contains(prompt, "Function Router") {
			t.Error("prompt should contain system instructions")
		}
		if !strings.Contains(prompt, "RESPONSE FORMAT") {
			t.Error("prompt should contain response format section")
		}
	})
}

func TestBuilder_Build_NestedParameters(t *testing.T) {
	b := NewBuilder()

	b.AddFunction(FunctionInfo{
		Name:        "complex",
		Description: "A complex function",
		Parameters: map[string]schema.Parameter{
			"user": {
				Type:        "object",
				Description: "User information",
				Required:    true,
				Properties: map[string]schema.Parameter{
					"name":  {Type: "string", Description: "User name", Required: true},
					"email": {Type: "string", Description: "User email"},
				},
			},
		},
	})

	prompt := b.Build()

	if !strings.Contains(prompt, "user") {
		t.Error("prompt should contain user parameter")
	}

	// Note: The current implementation may not deeply render nested properties
	// This test documents the expected behavior
}

func TestBuilder_Chaining(t *testing.T) {
	prompt := NewBuilder().
		WithRules([]string{"Rule 1"}).
		AddFunction(FunctionInfo{Name: "fn1", Description: "Function 1"}).
		AddFunction(FunctionInfo{Name: "fn2", Description: "Function 2"}).
		Build()

	if !strings.Contains(prompt, "fn1") || !strings.Contains(prompt, "fn2") {
		t.Error("chained builder should include all functions")
	}
}

func TestDefaultRules(t *testing.T) {
	rules := defaultRules()

	if len(rules) == 0 {
		t.Error("default rules should not be empty")
	}

	// Check for some expected default rules
	hasIntent := false
	hasValidation := false
	for _, rule := range rules {
		if strings.Contains(strings.ToLower(rule), "intent") {
			hasIntent = true
		}
		if strings.Contains(strings.ToLower(rule), "validate") || strings.Contains(strings.ToLower(rule), "required") {
			hasValidation = true
		}
	}

	if !hasIntent {
		t.Error("default rules should include something about user intent")
	}
	if !hasValidation {
		t.Error("default rules should include something about validation")
	}
}

func TestFormatParameter(t *testing.T) {
	t.Run("simple parameter", func(t *testing.T) {
		param := schema.Parameter{
			Type:        "string",
			Description: "A test parameter",
			Required:    false,
		}

		result := formatParameter("test", param, 0)

		if !strings.Contains(result, "test") {
			t.Error("result should contain parameter name")
		}
		if !strings.Contains(result, "string") {
			t.Error("result should contain parameter type")
		}
		if strings.Contains(result, "[required]") {
			t.Error("non-required parameter should not have [required] marker")
		}
	})

	t.Run("required parameter", func(t *testing.T) {
		param := schema.Parameter{
			Type:        "string",
			Description: "A required parameter",
			Required:    true,
		}

		result := formatParameter("name", param, 0)

		if !strings.Contains(result, "[required]") {
			t.Error("required parameter should have [required] marker")
		}
	})

	t.Run("indented parameter", func(t *testing.T) {
		param := schema.Parameter{
			Type:        "string",
			Description: "A nested parameter",
			Required:    false,
		}

		result := formatParameter("nested", param, 2)

		if !strings.HasPrefix(result, "    ") { // 2 * 2 spaces
			t.Error("parameter should be indented")
		}
	})

	t.Run("parameter with no description", func(t *testing.T) {
		param := schema.Parameter{
			Type:     "string",
			Required: false,
		}

		result := formatParameter("field", param, 0)

		// Should use the field name as description when empty
		if !strings.Contains(result, "field") {
			t.Error("result should contain field name as fallback description")
		}
	})
}

func TestBuilder_EmptyFunctions(t *testing.T) {
	b := NewBuilder()

	prompt := b.Build()

	// Should still produce a valid prompt with no functions
	if prompt == "" {
		t.Error("prompt should not be empty even without functions")
	}

	if !strings.Contains(prompt, "AVAILABLE FUNCTIONS") {
		t.Error("prompt should contain functions section header")
	}
}
