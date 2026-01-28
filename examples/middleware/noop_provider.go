package main

import (
	"context"

	"github.com/SadikSunbul/gopilot/provider"
)

type noopProvider struct{}

func (p *noopProvider) Generate(ctx context.Context, prompt string) (*provider.Response, error) {
	_ = ctx
	_ = prompt
	return &provider.Response{Type: "single", Agent: "unsupported", Parameters: map[string]any{"message": "noop"}}, nil
}

func (p *noopProvider) Close() error { return nil }
