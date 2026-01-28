package gopilot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SadikSunbul/gopilot/provider"
)

func setupTestGopilot() *Gopilot {
	g, _ := New(&mockProvider{})

	// Register test functions
	_ = g.Register(&mockFunction{
		name: "double",
		handler: func(ctx context.Context, params map[string]any) (any, error) {
			val := params["value"].(float64)
			return map[string]any{"value": val * 2}, nil
		},
	})

	_ = g.Register(&mockFunction{
		name: "add-ten",
		handler: func(ctx context.Context, params map[string]any) (any, error) {
			val := params["value"].(float64)
			return map[string]any{"value": val + 10}, nil
		},
	})

	_ = g.Register(&mockFunction{
		name: "format",
		handler: func(ctx context.Context, params map[string]any) (any, error) {
			val := params["value"].(float64)
			return map[string]any{"result": val, "formatted": true}, nil
		},
	})

	_ = g.Register(&mockFunction{
		name: "fail",
		handler: func(ctx context.Context, params map[string]any) (any, error) {
			return nil, errors.New("intentional failure")
		},
	})

	return g
}

// mockProvider implements provider.Provider for testing.
type mockProvider struct {
	systemPrompt string
}

func (m *mockProvider) Generate(ctx context.Context, prompt string) (*provider.Response, error) {
	return &provider.Response{
		Agent:      "noop",
		Parameters: map[string]any{},
		Raw:        "",
	}, nil
}

func (m *mockProvider) SetSystemPrompt(prompt string) {
	m.systemPrompt = prompt
}

func (m *mockProvider) Close() error {
	return nil
}

func TestPipelineBuilder_Basic(t *testing.T) {
	g := setupTestGopilot()

	result, err := g.Pipeline().
		Step("double").
		Step("add-ten").
		Execute(context.Background(), map[string]any{"value": 5.0})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success() {
		t.Error("pipeline should succeed")
	}

	if len(result.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(result.Steps))
	}

	// 5 * 2 = 10, 10 + 10 = 20
	finalOutput := result.FinalOutput.(map[string]any)
	if finalOutput["value"].(float64) != 20.0 {
		t.Errorf("expected value 20, got %v", finalOutput["value"])
	}
}

func TestPipelineBuilder_WithParams(t *testing.T) {
	g := setupTestGopilot()

	result, err := g.Pipeline().
		StepWithParams("double", map[string]any{"extra": "data"}).
		Execute(context.Background(), map[string]any{"value": 3.0})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success() {
		t.Error("pipeline should succeed")
	}
}

func TestPipelineBuilder_WithTransform(t *testing.T) {
	g := setupTestGopilot()

	result, err := g.Pipeline().
		StepWithTransform("double", func(output any) (map[string]any, error) {
			m := output.(map[string]any)
			// Transform the output
			return map[string]any{"value": m["value"].(float64) + 1}, nil
		}).
		Step("add-ten").
		Execute(context.Background(), map[string]any{"value": 5.0})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 * 2 = 10, transform adds 1 = 11, add-ten = 21
	finalOutput := result.FinalOutput.(map[string]any)
	if finalOutput["value"].(float64) != 21.0 {
		t.Errorf("expected value 21, got %v", finalOutput["value"])
	}
}

func TestPipelineBuilder_WithCondition(t *testing.T) {
	g := setupTestGopilot()

	// Condition that skips add-ten when value > 15
	result, err := g.Pipeline().
		Step("double").
		StepWithCondition("add-ten", func(input any) bool {
			m := input.(map[string]any)
			return m["value"].(float64) <= 15
		}).
		Execute(context.Background(), map[string]any{"value": 10.0})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 10 * 2 = 20, condition false, skip add-ten
	if !result.Steps[1].Skipped {
		t.Error("second step should be skipped")
	}
}

func TestPipelineBuilder_Error(t *testing.T) {
	g := setupTestGopilot()

	result, err := g.Pipeline().
		Step("double").
		Step("fail").
		Step("add-ten").
		Execute(context.Background(), map[string]any{"value": 5.0})

	if err == nil {
		t.Fatal("expected error")
	}

	if result.Success() {
		t.Error("pipeline should fail")
	}

	if !IsPipelineError(err) {
		t.Error("error should be PipelineError")
	}

	pErr := err.(*PipelineError)
	if pErr.Step != 1 {
		t.Errorf("expected step 1, got %d", pErr.Step)
	}
	if pErr.Function != "fail" {
		t.Errorf("expected function 'fail', got %s", pErr.Function)
	}
}

func TestPipelineBuilder_ErrorHandler(t *testing.T) {
	g := setupTestGopilot()

	errorHandled := false
	result, err := g.Pipeline().
		Step("double").
		Step("fail").
		Step("add-ten").
		OnError(func(err error, step int, fn string) error {
			errorHandled = true
			return nil // Return nil to continue
		}).
		Execute(context.Background(), map[string]any{"value": 5.0})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !errorHandled {
		t.Error("error handler should be called")
	}

	if !result.Success() {
		t.Error("pipeline should succeed when error is handled")
	}
}

func TestPipelineBuilder_Empty(t *testing.T) {
	g := setupTestGopilot()

	_, err := g.Pipeline().Execute(context.Background(), nil)

	if err != ErrPipelineEmpty {
		t.Errorf("expected ErrPipelineEmpty, got %v", err)
	}
}

func TestPipelineResult_Duration(t *testing.T) {
	g := setupTestGopilot()

	result, _ := g.Pipeline().
		Step("double").
		Execute(context.Background(), map[string]any{"value": 5.0})

	if result.Duration() < 0 {
		t.Error("duration should be non-negative")
	}
}

func TestWorkflow_Basic(t *testing.T) {
	g := setupTestGopilot()

	workflow := NewWorkflow("test-workflow").
		WithDescription("A test workflow").
		Then("double").
		Then("add-ten")

	result, err := g.ExecuteWorkflow(context.Background(), workflow, map[string]any{"value": 5.0})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Workflow != "test-workflow" {
		t.Errorf("expected workflow 'test-workflow', got %s", result.Workflow)
	}

	if !result.Success() {
		t.Error("workflow should succeed")
	}

	// 5 * 2 = 10, 10 + 10 = 20
	finalOutput := result.FinalOutput.(map[string]any)
	if finalOutput["value"].(float64) != 20.0 {
		t.Errorf("expected value 20, got %v", finalOutput["value"])
	}
}

func TestWorkflow_WithParams(t *testing.T) {
	g := setupTestGopilot()

	workflow := &Workflow{
		Name: "param-workflow",
		Steps: []WorkflowStep{
			{Function: "double", MapOutput: true},
			{Function: "add-ten", Params: map[string]any{"extra": "data"}, MapOutput: true},
		},
	}

	result, err := g.ExecuteWorkflow(context.Background(), workflow, map[string]any{"value": 3.0})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success() {
		t.Error("workflow should succeed")
	}
}

func TestWorkflow_WithOutputKey(t *testing.T) {
	g := setupTestGopilot()

	workflow := &Workflow{
		Name: "output-key-workflow",
		Steps: []WorkflowStep{
			{Function: "double", OutputKey: "doubled", MapOutput: true},
			{Function: "add-ten", OutputKey: "final", MapOutput: true},
		},
	}

	result, err := g.ExecuteWorkflow(context.Background(), workflow, map[string]any{"value": 5.0})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := result.Outputs["doubled"]; !ok {
		t.Error("outputs should contain 'doubled'")
	}
	if _, ok := result.Outputs["final"]; !ok {
		t.Error("outputs should contain 'final'")
	}
}

func TestWorkflow_ErrorStrategy_Fail(t *testing.T) {
	g := setupTestGopilot()

	workflow := &Workflow{
		Name:    "fail-workflow",
		OnError: ErrorStrategyFail,
		Steps: []WorkflowStep{
			{Function: "double", MapOutput: true},
			{Function: "fail"},
			{Function: "add-ten", MapOutput: true},
		},
	}

	result, err := g.ExecuteWorkflow(context.Background(), workflow, map[string]any{"value": 5.0})

	if err == nil {
		t.Fatal("expected error")
	}

	if result.Success() {
		t.Error("workflow should fail")
	}
}

func TestWorkflow_ErrorStrategy_Skip(t *testing.T) {
	g := setupTestGopilot()

	workflow := &Workflow{
		Name:    "skip-workflow",
		OnError: ErrorStrategySkip,
		Steps: []WorkflowStep{
			{Function: "double", MapOutput: true},
			{Function: "fail"},
			{Function: "add-ten", MapOutput: true},
		},
	}

	result, err := g.ExecuteWorkflow(context.Background(), workflow, map[string]any{"value": 5.0})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success() {
		t.Error("workflow should succeed with skip strategy")
	}

	if !result.Steps[1].Skipped {
		t.Error("failed step should be marked as skipped")
	}
}

func TestWorkflow_WithTimeout(t *testing.T) {
	g, _ := New(&mockProvider{})

	_ = g.Register(&mockFunction{
		name: "slow",
		handler: func(ctx context.Context, params map[string]any) (any, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return map[string]any{"done": true}, nil
			}
		},
	})

	workflow := NewWorkflow("timeout-workflow").
		WithTimeout(10 * time.Millisecond).
		Then("slow")

	_, err := g.ExecuteWorkflow(context.Background(), workflow, nil)

	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestWorkflow_Empty(t *testing.T) {
	g := setupTestGopilot()

	workflow := &Workflow{Name: "empty"}

	_, err := g.ExecuteWorkflow(context.Background(), workflow, nil)

	if err != ErrWorkflowEmpty {
		t.Errorf("expected ErrWorkflowEmpty, got %v", err)
	}
}

func TestWorkflow_Retries(t *testing.T) {
	g, _ := New(&mockProvider{})

	attempts := 0
	_ = g.Register(&mockFunction{
		name: "flaky",
		handler: func(ctx context.Context, params map[string]any) (any, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("temporary failure")
			}
			return map[string]any{"success": true}, nil
		},
	})

	workflow := &Workflow{
		Name: "retry-workflow",
		Steps: []WorkflowStep{
			{Function: "flaky", Retries: 3},
		},
	}

	result, err := g.ExecuteWorkflow(context.Background(), workflow, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success() {
		t.Error("workflow should succeed after retries")
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestWorkflowResult_Duration(t *testing.T) {
	g := setupTestGopilot()

	workflow := NewWorkflow("test").Then("double")
	result, _ := g.ExecuteWorkflow(context.Background(), workflow, map[string]any{"value": 5.0})

	if result.Duration() < 0 {
		t.Error("duration should be non-negative")
	}
}
