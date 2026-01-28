package main

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestPlanRetries(t *testing.T) {
	var counter int32
	g, err := buildRetryPlanGopilot(&counter)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	out, err := g.GenerateAndExecuteSmart(context.Background(), "run")
	if err != nil {
		t.Fatalf("exec: %v", err)
	}

	resp, ok := out.(OKOut)
	if !ok {
		t.Fatalf("unexpected type %T", out)
	}
	if !resp.Done {
		t.Fatalf("expected done=true, got %+v", resp)
	}
	if atomic.LoadInt32(&counter) != 2 {
		t.Fatalf("expected 2 attempts, got %d", atomic.LoadInt32(&counter))
	}
}
