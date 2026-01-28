package main

import (
	"context"
	"testing"

	"github.com/SadikSunbul/gopilot"
)

func TestWorkflowLocationWeather(t *testing.T) {
	g, err := buildWorkflowGopilot()
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	workflow := gopilot.NewWorkflow("location-weather").
		AddStep(gopilot.WorkflowStep{Function: "get-location", MapOutput: true, OutputKey: "location"}).
		AddStep(gopilot.WorkflowStep{Function: "weather-agent", MapOutput: true, OutputKey: "weather"})

	res, err := g.ExecuteWorkflow(context.Background(), workflow, map[string]any{})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if res.Outputs["location"] == nil {
		t.Fatalf("expected location output")
	}
	if res.Outputs["weather"] == nil {
		t.Fatalf("expected weather output")
	}
}
