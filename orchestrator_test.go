package gopilot

import (
	"context"
	"testing"

	"github.com/SadikSunbul/gopilot/provider"
)

type planProvider struct {
	resp *provider.Response
}

func (p *planProvider) Generate(ctx context.Context, prompt string) (*provider.Response, error) {
	_ = ctx
	_ = prompt
	return p.resp, nil
}

func (p *planProvider) Close() error { return nil }

func TestGenerateAndExecuteSmartSingle(t *testing.T) {
	p := &planProvider{
		resp: &provider.Response{
			Type:       "single",
			Agent:      "a",
			Parameters: map[string]any{},
		},
	}
	g, err := New(p)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	_ = g.Register(NewSimpleFunction("a", "a", nil, func(ctx context.Context, params map[string]any) (any, error) {
		return map[string]any{"ok": true}, nil
	}))

	out, err := g.GenerateAndExecuteSmart(context.Background(), "x")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	m := out.(map[string]any)
	if m["ok"] != true {
		t.Fatalf("expected ok=true, got %#v", out)
	}
}

func TestGenerateAndExecuteSmartPlanWithStoreAsRef(t *testing.T) {
	mapOutputFalse := false
	p := &planProvider{
		resp: &provider.Response{
			Type: "plan",
			Steps: []provider.Step{
				{Function: "loc", StoreAs: "location", MapOutput: &mapOutputFalse},
				{Function: "weather", Params: map[string]any{"city": "$location.city"}},
			},
		},
	}
	g, err := New(p)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	_ = g.Register(NewSimpleFunction("loc", "loc", nil, func(ctx context.Context, params map[string]any) (any, error) {
		return map[string]any{"city": "Istanbul"}, nil
	}))

	_ = g.Register(NewSimpleFunction("weather", "weather", nil, func(ctx context.Context, params map[string]any) (any, error) {
		if params["city"] != "Istanbul" {
			t.Fatalf("expected city Istanbul, got %#v", params["city"])
		}
		return map[string]any{"temp": 25}, nil
	}))

	out, err := g.GenerateAndExecuteSmart(context.Background(), "my location weather")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	m := out.(map[string]any)
	switch v := m["temp"].(type) {
	case int:
		if v != 25 {
			t.Fatalf("expected temp 25, got %#v", out)
		}
	case float64:
		if v != 25 {
			t.Fatalf("expected temp 25, got %#v", out)
		}
	default:
		t.Fatalf("unexpected temp type %T: %#v", m["temp"], out)
	}
}
