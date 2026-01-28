package gopilot

import (
	"context"
	"sync"
	"testing"

	"github.com/SadikSunbul/gopilot/schema"
)

// mockFunction is a test helper function wrapper.
type mockFunction struct {
	name        string
	description string
	params      map[string]schema.Parameter
	handler     func(ctx context.Context, params map[string]any) (any, error)
}

func (m *mockFunction) Name() string                         { return m.name }
func (m *mockFunction) Description() string                  { return m.description }
func (m *mockFunction) Parameters() map[string]schema.Parameter { return m.params }
func (m *mockFunction) Execute(ctx context.Context, params map[string]any) (any, error) {
	return m.handler(ctx, params)
}

func newMockFunction(name string) *mockFunction {
	return &mockFunction{
		name:        name,
		description: "test function",
		params:      make(map[string]schema.Parameter),
		handler: func(ctx context.Context, params map[string]any) (any, error) {
			return "result", nil
		},
	}
}

func TestRegistry_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		r := NewRegistry()
		fn := newMockFunction("test")

		err := r.Register(fn)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !r.Has("test") {
			t.Error("function should be registered")
		}
	})

	t.Run("nil function", func(t *testing.T) {
		r := NewRegistry()

		err := r.Register(nil)
		if err != ErrNilFunction {
			t.Errorf("expected ErrNilFunction, got %v", err)
		}
	})

	t.Run("empty function name", func(t *testing.T) {
		r := NewRegistry()
		fn := newMockFunction("")

		err := r.Register(fn)
		if err != ErrEmptyFunctionName {
			t.Errorf("expected ErrEmptyFunctionName, got %v", err)
		}
	})

	t.Run("duplicate registration", func(t *testing.T) {
		r := NewRegistry()
		fn := newMockFunction("test")

		_ = r.Register(fn)
		err := r.Register(fn)
		if err != ErrFunctionExists {
			t.Errorf("expected ErrFunctionExists, got %v", err)
		}
	})
}

func TestRegistry_Unregister(t *testing.T) {
	t.Run("successful unregistration", func(t *testing.T) {
		r := NewRegistry()
		fn := newMockFunction("test")
		_ = r.Register(fn)

		err := r.Unregister("test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if r.Has("test") {
			t.Error("function should be unregistered")
		}
	})

	t.Run("unregister non-existent function", func(t *testing.T) {
		r := NewRegistry()

		err := r.Unregister("non-existent")
		if err != ErrFunctionNotFound {
			t.Errorf("expected ErrFunctionNotFound, got %v", err)
		}
	})
}

func TestRegistry_Get(t *testing.T) {
	t.Run("get existing function", func(t *testing.T) {
		r := NewRegistry()
		fn := newMockFunction("test")
		_ = r.Register(fn)

		got, err := r.Get("test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if got.Name() != "test" {
			t.Errorf("expected name 'test', got %s", got.Name())
		}
	})

	t.Run("get non-existent function", func(t *testing.T) {
		r := NewRegistry()

		_, err := r.Get("non-existent")
		if err != ErrFunctionNotFound {
			t.Errorf("expected ErrFunctionNotFound, got %v", err)
		}
	})
}

func TestRegistry_Has(t *testing.T) {
	r := NewRegistry()
	fn := newMockFunction("test")
	_ = r.Register(fn)

	if !r.Has("test") {
		t.Error("Has should return true for registered function")
	}

	if r.Has("non-existent") {
		t.Error("Has should return false for non-existent function")
	}
}

func TestRegistry_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		r := NewRegistry()
		fn := &mockFunction{
			name: "add",
			handler: func(ctx context.Context, params map[string]any) (any, error) {
				a := params["a"].(float64)
				b := params["b"].(float64)
				return a + b, nil
			},
		}
		_ = r.Register(fn)

		result, err := r.Execute(context.Background(), "add", map[string]any{"a": 1.0, "b": 2.0})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.(float64) != 3.0 {
			t.Errorf("expected 3.0, got %v", result)
		}
	})

	t.Run("execute non-existent function", func(t *testing.T) {
		r := NewRegistry()

		_, err := r.Execute(context.Background(), "non-existent", nil)
		if err != ErrFunctionNotFound {
			t.Errorf("expected ErrFunctionNotFound, got %v", err)
		}
	})
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(newMockFunction("fn1"))
	_ = r.Register(newMockFunction("fn2"))
	_ = r.Register(newMockFunction("fn3"))

	list := r.List()
	if len(list) != 3 {
		t.Errorf("expected 3 functions, got %d", len(list))
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(newMockFunction("fn1"))
	_ = r.Register(newMockFunction("fn2"))

	names := r.Names()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestRegistry_Count(t *testing.T) {
	r := NewRegistry()

	if r.Count() != 0 {
		t.Error("empty registry should have count 0")
	}

	_ = r.Register(newMockFunction("fn1"))
	_ = r.Register(newMockFunction("fn2"))

	if r.Count() != 2 {
		t.Errorf("expected count 2, got %d", r.Count())
	}
}

func TestRegistry_Clear(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(newMockFunction("fn1"))
	_ = r.Register(newMockFunction("fn2"))

	r.Clear()

	if r.Count() != 0 {
		t.Error("registry should be empty after Clear")
	}
}

func TestRegistry_Concurrency(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fn := newMockFunction("fn" + string(rune('0'+i%10)))
			_ = r.Register(fn) // Ignore errors as some may be duplicates
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.List()
			_ = r.Names()
			_ = r.Count()
		}()
	}

	wg.Wait()
}
