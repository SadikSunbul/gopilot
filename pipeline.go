package gopilot

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// PipelineBuilder provides a fluent API for building and executing pipelines.
type PipelineBuilder struct {
	gopilot *Gopilot
	steps   []PipelineStep
	onError func(err error, step int, fn string) error
}

// PipelineStep represents a single step in a pipeline.
type PipelineStep struct {
	// Function is the name of the function to execute.
	Function string

	// Params are optional static parameters for the step.
	// These are merged with the output from the previous step.
	Params map[string]any

	// Transform is an optional function to transform the output before passing to the next step.
	Transform func(output any) (map[string]any, error)

	// Condition is an optional function that determines if this step should be executed.
	Condition func(input any) bool
}

// NewPipelineBuilder creates a new pipeline builder.
func NewPipelineBuilder(g *Gopilot) *PipelineBuilder {
	return &PipelineBuilder{
		gopilot: g,
		steps:   make([]PipelineStep, 0),
	}
}

// Step adds a step to the pipeline.
func (p *PipelineBuilder) Step(function string) *PipelineBuilder {
	p.steps = append(p.steps, PipelineStep{
		Function: function,
	})
	return p
}

// StepWithParams adds a step with static parameters.
func (p *PipelineBuilder) StepWithParams(function string, params map[string]any) *PipelineBuilder {
	p.steps = append(p.steps, PipelineStep{
		Function: function,
		Params:   params,
	})
	return p
}

// StepWithTransform adds a step with a transform function.
func (p *PipelineBuilder) StepWithTransform(function string, transform func(output any) (map[string]any, error)) *PipelineBuilder {
	p.steps = append(p.steps, PipelineStep{
		Function:  function,
		Transform: transform,
	})
	return p
}

// StepWithCondition adds a conditional step.
func (p *PipelineBuilder) StepWithCondition(function string, condition func(input any) bool) *PipelineBuilder {
	p.steps = append(p.steps, PipelineStep{
		Function:  function,
		Condition: condition,
	})
	return p
}

// AddStep adds a complete step configuration.
func (p *PipelineBuilder) AddStep(step PipelineStep) *PipelineBuilder {
	p.steps = append(p.steps, step)
	return p
}

// OnError sets an error handler for the pipeline.
func (p *PipelineBuilder) OnError(handler func(err error, step int, fn string) error) *PipelineBuilder {
	p.onError = handler
	return p
}

// Execute executes the pipeline with the given input.
func (p *PipelineBuilder) Execute(ctx context.Context, input any) (*PipelineResult, error) {
	if len(p.steps) == 0 {
		return nil, ErrPipelineEmpty
	}

	result := &PipelineResult{
		StartTime: time.Now(),
		Steps:     make([]StepResult, 0, len(p.steps)),
	}

	// Convert input to params
	currentParams, err := toParams(input)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input: %w", err)
	}

	for i, step := range p.steps {
		stepResult := StepResult{
			Index:     i,
			Function:  step.Function,
			StartTime: time.Now(),
		}

		// Check condition
		if step.Condition != nil && !step.Condition(currentParams) {
			stepResult.Skipped = true
			stepResult.EndTime = time.Now()
			result.Steps = append(result.Steps, stepResult)
			continue
		}

		// Merge static params with current params
		execParams := currentParams
		if step.Params != nil {
			execParams = mergeMaps(currentParams, step.Params)
		}

		stepResult.Input = execParams

		// Execute the function
		output, err := p.gopilot.Execute(ctx, step.Function, execParams)
		stepResult.EndTime = time.Now()

		if err != nil {
			stepResult.Error = err

			if p.onError != nil {
				if handledErr := p.onError(err, i, step.Function); handledErr != nil {
					result.Steps = append(result.Steps, stepResult)
					result.EndTime = time.Now()
					result.Error = NewPipelineError(i, step.Function, handledErr)
					return result, result.Error
				}
				// Error was handled, continue
				result.Steps = append(result.Steps, stepResult)
				continue
			}

			result.Steps = append(result.Steps, stepResult)
			result.EndTime = time.Now()
			result.Error = NewPipelineError(i, step.Function, err)
			return result, result.Error
		}

		stepResult.Output = output

		// Transform output for next step
		if step.Transform != nil {
			currentParams, err = step.Transform(output)
			if err != nil {
				stepResult.Error = err
				result.Steps = append(result.Steps, stepResult)
				result.EndTime = time.Now()
				result.Error = NewPipelineError(i, step.Function, fmt.Errorf("transform failed: %w", err))
				return result, result.Error
			}
		} else {
			currentParams, err = toParams(output)
			if err != nil {
				stepResult.Error = err
				result.Steps = append(result.Steps, stepResult)
				result.EndTime = time.Now()
				result.Error = NewPipelineError(i, step.Function, fmt.Errorf("output conversion failed: %w", err))
				return result, result.Error
			}
		}

		result.Steps = append(result.Steps, stepResult)
	}

	result.EndTime = time.Now()
	result.Output = p.steps[len(p.steps)-1].Function

	// Get the last successful output
	for i := len(result.Steps) - 1; i >= 0; i-- {
		if !result.Steps[i].Skipped && result.Steps[i].Error == nil {
			result.FinalOutput = result.Steps[i].Output
			break
		}
	}

	return result, nil
}

// PipelineResult contains the result of a pipeline execution.
type PipelineResult struct {
	// Steps contains the result of each step.
	Steps []StepResult

	// FinalOutput is the output of the last successful step.
	FinalOutput any

	// Output is the name of the last function executed.
	Output string

	// Error is set if the pipeline failed.
	Error error

	// StartTime is when the pipeline started.
	StartTime time.Time

	// EndTime is when the pipeline ended.
	EndTime time.Time
}

// Duration returns the total duration of the pipeline.
func (r *PipelineResult) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

// Success returns true if the pipeline completed without errors.
func (r *PipelineResult) Success() bool {
	return r.Error == nil
}

// StepResult contains the result of a single step.
type StepResult struct {
	// Index is the step index.
	Index int

	// Function is the function name.
	Function string

	// Input is the input parameters.
	Input map[string]any

	// Output is the step output.
	Output any

	// Error is set if the step failed.
	Error error

	// Skipped is true if the step was skipped due to condition.
	Skipped bool

	// StartTime is when the step started.
	StartTime time.Time

	// EndTime is when the step ended.
	EndTime time.Time
}

// Duration returns the duration of the step.
func (r *StepResult) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

// Workflow represents a declarative workflow definition.
type Workflow struct {
	// Name is the workflow name.
	Name string `json:"name"`

	// Description is the workflow description.
	Description string `json:"description,omitempty"`

	// Steps are the workflow steps.
	Steps []WorkflowStep `json:"steps"`

	// OnError defines the error handling strategy.
	OnError WorkflowErrorStrategy `json:"on_error,omitempty"`

	// Timeout is the maximum duration for the workflow.
	Timeout time.Duration `json:"timeout,omitempty"`
}

// WorkflowStep represents a single step in a workflow.
type WorkflowStep struct {
	// Name is an optional name for the step.
	Name string `json:"name,omitempty"`

	// Function is the function to execute.
	Function string `json:"function"`

	// Params are static parameters for the step.
	Params map[string]any `json:"params,omitempty"`

	// MapOutput determines how to map output to the next step.
	// If true, the output is converted to params for the next step.
	MapOutput bool `json:"map_output,omitempty"`

	// OutputKey if set, the output is stored under this key.
	OutputKey string `json:"output_key,omitempty"`

	// Condition is an optional condition expression.
	Condition string `json:"condition,omitempty"`

	// Retries is the number of retries for this step.
	Retries int `json:"retries,omitempty"`
}

// WorkflowErrorStrategy defines how errors are handled.
type WorkflowErrorStrategy string

const (
	// ErrorStrategyFail fails the workflow on error.
	ErrorStrategyFail WorkflowErrorStrategy = "fail"

	// ErrorStrategySkip skips the failed step and continues.
	ErrorStrategySkip WorkflowErrorStrategy = "skip"

	// ErrorStrategyRetry retries the failed step.
	ErrorStrategyRetry WorkflowErrorStrategy = "retry"
)

// WorkflowResult contains the result of a workflow execution.
type WorkflowResult struct {
	// Workflow is the workflow that was executed.
	Workflow string

	// Steps contains the result of each step.
	Steps []StepResult

	// Outputs contains named outputs from steps with OutputKey set.
	Outputs map[string]any

	// FinalOutput is the output of the last successful step.
	FinalOutput any

	// Error is set if the workflow failed.
	Error error

	// StartTime is when the workflow started.
	StartTime time.Time

	// EndTime is when the workflow ended.
	EndTime time.Time
}

// Duration returns the total duration of the workflow.
func (r *WorkflowResult) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

// Success returns true if the workflow completed without errors.
func (r *WorkflowResult) Success() bool {
	return r.Error == nil
}

// Execute executes the workflow.
func (w *Workflow) Execute(ctx context.Context, g *Gopilot, input any) (*WorkflowResult, error) {
	if len(w.Steps) == 0 {
		return nil, ErrWorkflowEmpty
	}

	// Apply timeout if set
	if w.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, w.Timeout)
		defer cancel()
	}

	result := &WorkflowResult{
		Workflow:  w.Name,
		StartTime: time.Now(),
		Steps:     make([]StepResult, 0, len(w.Steps)),
		Outputs:   make(map[string]any),
	}

	currentParams, err := toParams(input)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input: %w", err)
	}

	for i, step := range w.Steps {
		stepResult := StepResult{
			Index:     i,
			Function:  step.Function,
			StartTime: time.Now(),
		}

		// Merge static params
		execParams := currentParams
		if step.Params != nil {
			execParams = mergeMaps(currentParams, step.Params)
		}

		stepResult.Input = execParams

		// Execute with retries
		var output any
		var execErr error
		attempts := step.Retries + 1

		for attempt := 0; attempt < attempts; attempt++ {
			output, execErr = g.Execute(ctx, step.Function, execParams)
			if execErr == nil {
				break
			}

			if attempt < attempts-1 {
				g.logger.Warn("workflow step %d (%s) failed, retrying: %v", i, step.Function, execErr)
			}
		}

		stepResult.EndTime = time.Now()

		if execErr != nil {
			stepResult.Error = execErr

			switch w.OnError {
			case ErrorStrategySkip:
				stepResult.Skipped = true
				result.Steps = append(result.Steps, stepResult)
				continue
			case ErrorStrategyFail, "":
				result.Steps = append(result.Steps, stepResult)
				result.EndTime = time.Now()
				result.Error = NewPipelineError(i, step.Function, execErr)
				return result, result.Error
			}
		}

		stepResult.Output = output

		// Store named output
		if step.OutputKey != "" {
			result.Outputs[step.OutputKey] = output
		}

		// Map output for next step
		if step.MapOutput {
			currentParams, err = toParams(output)
			if err != nil {
				stepResult.Error = err
				result.Steps = append(result.Steps, stepResult)
				result.EndTime = time.Now()
				result.Error = NewPipelineError(i, step.Function, fmt.Errorf("output conversion failed: %w", err))
				return result, result.Error
			}
		}

		result.Steps = append(result.Steps, stepResult)
	}

	result.EndTime = time.Now()

	// Get last successful output
	for i := len(result.Steps) - 1; i >= 0; i-- {
		if !result.Steps[i].Skipped && result.Steps[i].Error == nil {
			result.FinalOutput = result.Steps[i].Output
			break
		}
	}

	return result, nil
}

// NewWorkflow creates a new workflow with the given name.
func NewWorkflow(name string) *Workflow {
	return &Workflow{
		Name:  name,
		Steps: make([]WorkflowStep, 0),
	}
}

// WithDescription sets the workflow description.
func (w *Workflow) WithDescription(desc string) *Workflow {
	w.Description = desc
	return w
}

// WithTimeout sets the workflow timeout.
func (w *Workflow) WithTimeout(d time.Duration) *Workflow {
	w.Timeout = d
	return w
}

// WithErrorStrategy sets the error handling strategy.
func (w *Workflow) WithErrorStrategy(strategy WorkflowErrorStrategy) *Workflow {
	w.OnError = strategy
	return w
}

// AddStep adds a step to the workflow.
func (w *Workflow) AddStep(step WorkflowStep) *Workflow {
	w.Steps = append(w.Steps, step)
	return w
}

// Then adds a simple step to the workflow.
func (w *Workflow) Then(function string) *Workflow {
	return w.AddStep(WorkflowStep{
		Function:  function,
		MapOutput: true,
	})
}

// ThenWithParams adds a step with parameters.
func (w *Workflow) ThenWithParams(function string, params map[string]any) *Workflow {
	return w.AddStep(WorkflowStep{
		Function:  function,
		Params:    params,
		MapOutput: true,
	})
}

// Helper functions

func toParams(v any) (map[string]any, error) {
	if v == nil {
		return make(map[string]any), nil
	}

	// Already a map
	if m, ok := v.(map[string]any); ok {
		return m, nil
	}

	// Convert through JSON
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func mergeMaps(maps ...map[string]any) map[string]any {
	result := make(map[string]any)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
