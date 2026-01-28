package main

import (
	"context"
	"testing"
)

func TestMiddlewareInjectsRequestID(t *testing.T) {
	g, err := buildMiddlewareGopilot()
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	out, err := g.Execute(context.Background(), "echo", map[string]any{"message": "hi"})
	if err != nil {
		t.Fatalf("exec: %v", err)
	}

	resp, ok := out.(EchoResp)
	if !ok {
		t.Fatalf("unexpected type %T", out)
	}
	if resp.RequestID != "req-123" {
		t.Fatalf("expected req-123, got %q", resp.RequestID)
	}
}
