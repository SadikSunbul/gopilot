package gopilot

import (
	"context"
	"errors"
	"testing"

	"github.com/SadikSunbul/gopilot/schema"
)

type testParams struct {
	Name    string `json:"name" description:"The name" required:"true"`
	Age     int    `json:"age" description:"The age"`
	Active  bool   `json:"active"`
}

type testResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func TestFunction_Basic(t *testing.T) {
	fn := NewFunction[testParams, testResponse](
		"test-function",
		"A test function",
		func(ctx context.Context, params testParams) (testResponse, error) {
			return testResponse{
				Message: "Hello, " + params.Name,
				Success: true,
			}, nil
		},
	)

	t.Run("Name", func(t *testing.T) {
		if fn.Name() != "test-function" {
			t.Errorf("expected 'test-function', got %s", fn.Name())
		}
	})

	t.Run("Description", func(t *testing.T) {
		if fn.Description() != "A test function" {
			t.Errorf("expected 'A test function', got %s", fn.Description())
		}
	})

	t.Run("Parameters", func(t *testing.T) {
		params := fn.Parameters()
		if params == nil {
			t.Fatal("parameters should not be nil")
		}
		if _, ok := params["name"]; !ok {
			t.Error("parameters should contain 'name'")
		}
	})
}

func TestFunction_Execute(t *testing.T) {
	fn := NewFunction[testParams, testResponse](
		"greet",
		"Greets a person",
		func(ctx context.Context, params testParams) (testResponse, error) {
			return testResponse{
				Message: "Hello, " + params.Name,
				Success: true,
			}, nil
		},
	)

	t.Run("with map params", func(t *testing.T) {
		result, err := fn.Execute(context.Background(), map[string]any{
			"name": "John",
			"age":  30,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp, ok := result.(testResponse)
		if !ok {
			t.Fatalf("unexpected result type: %T", result)
		}
		if resp.Message != "Hello, John" {
			t.Errorf("expected 'Hello, John', got %s", resp.Message)
		}
	})

	t.Run("with typed params", func(t *testing.T) {
		result, err := fn.ExecuteTyped(context.Background(), testParams{
			Name: "Jane",
			Age:  25,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Message != "Hello, Jane" {
			t.Errorf("expected 'Hello, Jane', got %s", result.Message)
		}
	})

	t.Run("with JSON params", func(t *testing.T) {
		result, err := fn.ExecuteWithJSON(context.Background(), `{"name": "Bob", "age": 40}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		resp, ok := result.(testResponse)
		if !ok {
			t.Fatalf("unexpected result type: %T", result)
		}
		if resp.Message != "Hello, Bob" {
			t.Errorf("expected 'Hello, Bob', got %s", resp.Message)
		}
	})
}

func TestFunction_ExecuteError(t *testing.T) {
	expectedErr := errors.New("something went wrong")
	fn := NewFunction[testParams, testResponse](
		"failing",
		"A function that fails",
		func(ctx context.Context, params testParams) (testResponse, error) {
			return testResponse{}, expectedErr
		},
	)

	_, err := fn.Execute(context.Background(), map[string]any{"name": "test"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFunction_RequiredParams(t *testing.T) {
	fn := NewFunction[testParams, testResponse](
		"test",
		"test",
		func(ctx context.Context, params testParams) (testResponse, error) {
			return testResponse{}, nil
		},
	)

	// Missing required param
	_, err := fn.Execute(context.Background(), map[string]any{
		"age": 30, // name is missing
	})
	if err == nil {
		t.Error("expected error for missing required param")
	}
}

func TestNewFunctionWithSchema(t *testing.T) {
	customSchema := map[string]schema.Parameter{
		"input": {
			Type:        "string",
			Description: "Custom input",
			Required:    true,
		},
	}

	fn := NewFunctionWithSchema[map[string]any, string](
		"custom",
		"Custom function",
		customSchema,
		func(ctx context.Context, params map[string]any) (string, error) {
			return params["input"].(string), nil
		},
	)

	if fn.Parameters()["input"].Description != "Custom input" {
		t.Error("custom schema not applied correctly")
	}
}

func TestSimpleFunction(t *testing.T) {
	fn := NewSimpleFunction(
		"simple",
		"A simple function",
		map[string]schema.Parameter{
			"value": {Type: "string", Required: true},
		},
		func(ctx context.Context, params map[string]any) (any, error) {
			return params["value"], nil
		},
	)

	t.Run("Name", func(t *testing.T) {
		if fn.Name() != "simple" {
			t.Errorf("expected 'simple', got %s", fn.Name())
		}
	})

	t.Run("Execute", func(t *testing.T) {
		result, err := fn.Execute(context.Background(), map[string]any{"value": "test"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.(string) != "test" {
			t.Errorf("expected 'test', got %v", result)
		}
	})
}

func TestUnsupportedFunction(t *testing.T) {
	fn := NewUnsupportedFunction()

	if fn.Name() != "unsupported" {
		t.Errorf("expected 'unsupported', got %s", fn.Name())
	}

	result, err := fn.Execute(context.Background(), map[string]any{
		"message": "unknown request",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, ok := result.(UnsupportedResponse)
	if !ok {
		t.Fatalf("unexpected result type: %T", result)
	}
	if resp.Success {
		t.Error("Success should be false for unsupported function")
	}
	if resp.Message != "Unsupported request: unknown request" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
}

func TestFunction_ContextCancellation(t *testing.T) {
	fn := NewFunction[testParams, testResponse](
		"slow",
		"A slow function",
		func(ctx context.Context, params testParams) (testResponse, error) {
			select {
			case <-ctx.Done():
				return testResponse{}, ctx.Err()
			default:
				return testResponse{Message: "done"}, nil
			}
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := fn.ExecuteTyped(ctx, testParams{Name: "test"})
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
