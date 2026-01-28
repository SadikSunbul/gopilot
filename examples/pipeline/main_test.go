package main

import (
	"context"
	"testing"
)

func TestPipelineLocationWeather(t *testing.T) {
	g, err := buildPipelineGopilot()
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	res, err := g.Pipeline().
		Step("get-location").
		Step("weather-agent").
		Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	out, ok := res.FinalOutput.(WeatherResponse)
	if !ok {
		t.Fatalf("unexpected output type %T", res.FinalOutput)
	}
	if out.City != "Istanbul" {
		t.Fatalf("expected Istanbul, got %q", out.City)
	}
}
