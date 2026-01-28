package gopilot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SadikSunbul/gopilot/provider"
	"github.com/SadikSunbul/gopilot/schema"
)

// testProvider is a mock provider for testing.
type testProvider struct {
	generateFunc func(ctx context.Context, prompt string) (*provider.Response, error)
	systemPrompt string
	closed       bool
}

func (p *testProvider) Generate(ctx context.Context, prompt string) (*provider.Response, error) {
	if p.generateFunc != nil {
		return p.generateFunc(ctx, prompt)
	}
	return &provider.Response{
		Agent:      "test-function",
		Parameters: map[string]any{"param": "value"},
	}, nil
}

func (p *testProvider) SetSystemPrompt(prompt string) {
	p.systemPrompt = prompt
}

func (p *testProvider) Close() error {
	p.closed = true
	return nil
}

func TestNew(t *testing.T) {
	t.Run("with valid provider", func(t *testing.T) {
		g, err := New(&testProvider{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g == nil {
			t.Fatal("gopilot should not be nil")
		}
	})

	t.Run("with nil provider", func(t *testing.T) {
		_, err := New(nil)
		if err != ErrNilProvider {
			t.Errorf("expected ErrNilProvider, got %v", err)
		}
	})
}

func TestNew_WithOptions(t *testing.T) {
	t.Run("with timeout", func(t *testing.T) {
		g, _ := New(&testProvider{}, WithTimeout(5*time.Second))
		if g.timeout != 5*time.Second {
			t.Errorf("expected timeout 5s, got %v", g.timeout)
		}
	})

	t.Run("with logger", func(t *testing.T) {
		g, _ := New(&testProvider{}, WithStdLogger(LogLevelDebug))
		if g.logger == nil {
			t.Error("logger should not be nil")
		}
	})

	t.Run("with system prompt rules", func(t *testing.T) {
		rules := []string{"Rule 1", "Rule 2"}
		g, _ := New(&testProvider{}, WithSystemPromptRules(rules))
		if len(g.systemPromptRules) != 2 {
			t.Errorf("expected 2 rules, got %d", len(g.systemPromptRules))
		}
	})
}

func TestGopilot_Register(t *testing.T) {
	g, _ := New(&testProvider{})

	fn := NewFunction[testParams, testResponse](
		"test-fn",
		"A test function",
		func(ctx context.Context, params testParams) (testResponse, error) {
			return testResponse{}, nil
		},
	)

	err := g.Register(fn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !g.Has("test-fn") {
		t.Error("function should be registered")
	}
}

func TestGopilot_Unregister(t *testing.T) {
	g, _ := New(&testProvider{})

	fn := NewFunction[testParams, testResponse](
		"test-fn",
		"A test function",
		func(ctx context.Context, params testParams) (testResponse, error) {
			return testResponse{}, nil
		},
	)

	_ = g.Register(fn)
	err := g.Unregister("test-fn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.Has("test-fn") {
		t.Error("function should be unregistered")
	}
}

func TestGopilot_Get(t *testing.T) {
	g, _ := New(&testProvider{})

	fn := NewFunction[testParams, testResponse](
		"test-fn",
		"A test function",
		func(ctx context.Context, params testParams) (testResponse, error) {
			return testResponse{}, nil
		},
	)

	_ = g.Register(fn)

	got, err := g.Get("test-fn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "test-fn" {
		t.Errorf("expected 'test-fn', got %s", got.Name())
	}
}

func TestGopilot_List(t *testing.T) {
	g, _ := New(&testProvider{})

	fn1 := NewFunction[testParams, testResponse]("fn1", "desc", nil)
	fn2 := NewFunction[testParams, testResponse]("fn2", "desc", nil)

	_ = g.Register(fn1)
	_ = g.Register(fn2)

	list := g.List()
	if len(list) != 2 {
		t.Errorf("expected 2 functions, got %d", len(list))
	}
}

func TestGopilot_ConfigurePrompt(t *testing.T) {
	p := &testProvider{}
	g, _ := New(p)

	fn := NewFunction[testParams, testResponse](
		"test-fn",
		"A test function",
		func(ctx context.Context, params testParams) (testResponse, error) {
			return testResponse{}, nil
		},
	)

	_ = g.Register(fn)
	err := g.ConfigurePrompt()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.systemPrompt == "" {
		t.Error("system prompt should be set")
	}

	if !g.promptConfigured {
		t.Error("promptConfigured should be true")
	}

	// Unsupported function should be registered
	if !g.Has("unsupported") {
		t.Error("unsupported function should be registered")
	}
}

func TestGopilot_Generate(t *testing.T) {
	p := &testProvider{
		generateFunc: func(ctx context.Context, prompt string) (*provider.Response, error) {
			return &provider.Response{
				Agent:      "weather",
				Parameters: map[string]any{"city": "Istanbul"},
			}, nil
		},
	}

	g, _ := New(p)
	_ = g.Register(NewFunction[testParams, testResponse]("weather", "desc", nil))

	resp, err := g.Generate(context.Background(), "What's the weather?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Agent != "weather" {
		t.Errorf("expected agent 'weather', got %s", resp.Agent)
	}
}

func TestGopilot_Generate_Error(t *testing.T) {
	expectedErr := errors.New("generation failed")
	p := &testProvider{
		generateFunc: func(ctx context.Context, prompt string) (*provider.Response, error) {
			return nil, expectedErr
		},
	}

	g, _ := New(p)
	g.promptConfigured = true // Skip prompt configuration

	_, err := g.Generate(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrGenerationFailed) {
		t.Errorf("expected ErrGenerationFailed, got %v", err)
	}
}

func TestGopilot_Generate_NilResponse(t *testing.T) {
	p := &testProvider{
		generateFunc: func(ctx context.Context, prompt string) (*provider.Response, error) {
			return nil, nil
		},
	}

	g, _ := New(p)
	g.promptConfigured = true

	_, err := g.Generate(context.Background(), "test")
	if err != ErrNoResponse {
		t.Errorf("expected ErrNoResponse, got %v", err)
	}
}

func TestGopilot_Execute(t *testing.T) {
	g, _ := New(&testProvider{})

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

	_ = g.Register(fn)

	result, err := g.Execute(context.Background(), "greet", map[string]any{"name": "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := result.(testResponse)
	if resp.Message != "Hello, World" {
		t.Errorf("expected 'Hello, World', got %s", resp.Message)
	}
}

func TestGopilot_ExecuteJSON(t *testing.T) {
	g, _ := New(&testProvider{})

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

	_ = g.Register(fn)

	result, err := g.ExecuteJSON(context.Background(), "greet", `{"name": "JSON"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := result.(testResponse)
	if resp.Message != "Hello, JSON" {
		t.Errorf("expected 'Hello, JSON', got %s", resp.Message)
	}
}

func TestGopilot_GenerateAndExecute(t *testing.T) {
	p := &testProvider{
		generateFunc: func(ctx context.Context, prompt string) (*provider.Response, error) {
			return &provider.Response{
				Agent:      "greet",
				Parameters: map[string]any{"name": "User"},
			}, nil
		},
	}

	g, _ := New(p)

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

	_ = g.Register(fn)

	result, err := g.GenerateAndExecute(context.Background(), "Say hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := result.(testResponse)
	if resp.Message != "Hello, User" {
		t.Errorf("expected 'Hello, User', got %s", resp.Message)
	}
}

func TestGopilot_Middleware(t *testing.T) {
	g, _ := New(&testProvider{})

	middlewareCalled := false
	middleware := func(next ExecuteFunc) ExecuteFunc {
		return func(name string, params map[string]any) (any, error) {
			middlewareCalled = true
			return next(name, params)
		}
	}

	g.middlewares = append(g.middlewares, middleware)

	fn := NewFunction[testParams, testResponse](
		"test",
		"Test",
		func(ctx context.Context, params testParams) (testResponse, error) {
			return testResponse{}, nil
		},
	)

	_ = g.Register(fn)
	_, _ = g.Execute(context.Background(), "test", map[string]any{"name": "test"})

	if !middlewareCalled {
		t.Error("middleware should be called")
	}
}

func TestGopilot_Close(t *testing.T) {
	p := &testProvider{}
	g, _ := New(p)

	err := g.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !p.closed {
		t.Error("provider should be closed")
	}
}

func TestGopilot_Provider(t *testing.T) {
	p := &testProvider{}
	g, _ := New(p)

	if g.Provider() != p {
		t.Error("Provider() should return the underlying provider")
	}
}

func TestGopilot_Registry(t *testing.T) {
	g, _ := New(&testProvider{})

	if g.Registry() == nil {
		t.Error("Registry() should not return nil")
	}
}

func TestGopilot_Pipeline(t *testing.T) {
	g, _ := New(&testProvider{})

	builder := g.Pipeline()
	if builder == nil {
		t.Error("Pipeline() should not return nil")
	}
}

func TestGopilot_Retry(t *testing.T) {
	attempts := 0
	p := &testProvider{
		generateFunc: func(ctx context.Context, prompt string) (*provider.Response, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("temporary error")
			}
			return &provider.Response{Agent: "test", Parameters: nil}, nil
		},
	}

	g, _ := New(p, WithRetry(RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}))
	g.promptConfigured = true

	resp, err := g.Generate(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Agent != "test" {
		t.Errorf("expected agent 'test', got %s", resp.Agent)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("StringParam", func(t *testing.T) {
		p := StringParam("description", true)
		if p.Type != "string" {
			t.Errorf("expected type 'string', got %s", p.Type)
		}
		if !p.Required {
			t.Error("required should be true")
		}
	})

	t.Run("IntParam", func(t *testing.T) {
		p := IntParam("description", false)
		if p.Type != "integer" {
			t.Errorf("expected type 'integer', got %s", p.Type)
		}
	})

	t.Run("NumberParam", func(t *testing.T) {
		p := NumberParam("description", false)
		if p.Type != "number" {
			t.Errorf("expected type 'number', got %s", p.Type)
		}
	})

	t.Run("BoolParam", func(t *testing.T) {
		p := BoolParam("description", false)
		if p.Type != "boolean" {
			t.Errorf("expected type 'boolean', got %s", p.Type)
		}
	})

	t.Run("ArrayParam", func(t *testing.T) {
		items := &schema.Parameter{Type: "string"}
		p := ArrayParam("description", false, items)
		if p.Type != "array" {
			t.Errorf("expected type 'array', got %s", p.Type)
		}
		if p.Items != items {
			t.Error("items not set correctly")
		}
	})

	t.Run("ObjectParam", func(t *testing.T) {
		props := map[string]schema.Parameter{
			"field": {Type: "string"},
		}
		p := ObjectParam("description", false, props)
		if p.Type != "object" {
			t.Errorf("expected type 'object', got %s", p.Type)
		}
		if len(p.Properties) != 1 {
			t.Error("properties not set correctly")
		}
	})
}
