package gopilot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SadikSunbul/gopilot/provider"
)

// OrchestratorConfig controls multi-step execution limits.
type OrchestratorConfig struct {
	// MaxSteps is a safety limit to avoid runaway plans.
	MaxSteps int
}

func defaultOrchestratorConfig() OrchestratorConfig {
	return OrchestratorConfig{MaxSteps: 20}
}

// GenerateAndExecuteSmart executes either a single function (agent+parameters) or a multi-step plan.
// - If the provider returns a plan (Response.Steps), the plan is executed step-by-step.
// - Otherwise it falls back to single-step execution (Response.Agent + Response.Parameters).
func (g *Gopilot) GenerateAndExecuteSmart(ctx context.Context, input string) (any, error) {
	resp, err := g.Generate(ctx, input)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, ErrNoResponse
	}

	if isPlanResponse(resp) {
		return g.executePlan(ctx, resp, defaultOrchestratorConfig())
	}
	return g.executeSingle(ctx, resp)
}

func (g *Gopilot) executePlan(ctx context.Context, resp *provider.Response, cfg OrchestratorConfig) (any, error) {
	if err := validatePlan(resp, &cfg); err != nil {
		return nil, err
	}

	// currentParams starts from top-level parameters (if any); this allows a plan to seed context.
	currentParams := cloneMap(resp.Parameters)

	vars := map[string]any{}
	var lastOutput any

	for i, step := range resp.Steps {
		if err := validateStep(i, step); err != nil {
			return nil, err
		}

		execParams, err := buildStepParams(currentParams, step.Params, vars)
		if err != nil {
			return nil, NewPipelineError(i, step.Function, err)
		}

		// Execute with retries
		out, execErr := g.executeWithRetries(ctx, step.Function, execParams, step.Retries)
		if execErr != nil {
			return nil, NewPipelineError(i, step.Function, execErr)
		}

		lastOutput = out
		storeOutput(vars, step.StoreAs, out)

		if shouldMapOutput(step.MapOutput) {
			nextParams, err := toParams(out)
			if err != nil {
				return nil, NewPipelineError(i, step.Function, fmt.Errorf("output conversion failed: %w", err))
			}
			currentParams = nextParams
		}
	}

	return lastOutput, nil
}

func (g *Gopilot) executeSingle(ctx context.Context, resp *provider.Response) (any, error) {
	return g.Execute(ctx, resp.Agent, resp.Parameters)
}

func isPlanResponse(resp *provider.Response) bool {
	return len(resp.Steps) > 0 || strings.EqualFold(resp.Type, "plan")
}

func validatePlan(resp *provider.Response, cfg *OrchestratorConfig) error {
	if resp == nil {
		return ErrNoResponse
	}
	if len(resp.Steps) == 0 {
		return fmt.Errorf("plan response has no steps")
	}
	if cfg.MaxSteps <= 0 {
		cfg.MaxSteps = defaultOrchestratorConfig().MaxSteps
	}
	if len(resp.Steps) > cfg.MaxSteps {
		return fmt.Errorf("plan has too many steps: %d (max %d)", len(resp.Steps), cfg.MaxSteps)
	}
	return nil
}

func validateStep(i int, step provider.Step) error {
	if strings.TrimSpace(step.Function) == "" {
		return fmt.Errorf("plan step %d has empty function", i)
	}
	return nil
}

func buildStepParams(current, stepParams map[string]any, vars map[string]any) (map[string]any, error) {
	execParams := mergeMaps(current, stepParams)
	resolved := resolveRefs(execParams, vars)
	m, ok := resolved.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("resolved params are not an object")
	}
	return m, nil
}

func (g *Gopilot) executeWithRetries(ctx context.Context, fn string, params map[string]any, retries int) (any, error) {
	attempts := retries + 1
	if attempts < 1 {
		attempts = 1
	}
	var out any
	var err error
	for attempt := 0; attempt < attempts; attempt++ {
		out, err = g.Execute(ctx, fn, params)
		if err == nil {
			return out, nil
		}
	}
	return nil, err
}

func storeOutput(vars map[string]any, storeAs string, out any) {
	vars["prev"] = out
	if storeAs != "" {
		vars[storeAs] = out
	}
}

func shouldMapOutput(mapOutput *bool) bool {
	// Default map_output is true when omitted.
	if mapOutput == nil {
		return true
	}
	return *mapOutput
}

func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any)
	for k, v := range m {
		out[k] = v
	}
	return out
}

func resolveRefs(v any, vars map[string]any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = resolveRefs(val, vars)
		}
		return out
	case []any:
		out := make([]any, 0, len(t))
		for _, val := range t {
			out = append(out, resolveRefs(val, vars))
		}
		return out
	case string:
		// Only treat "$x.y" as a reference if the whole value is a ref token.
		if strings.HasPrefix(t, "$") && len(t) > 1 {
			if resolved, ok := lookupVarPath(vars, strings.TrimPrefix(t, "$")); ok {
				return resolved
			}
		}
		return t
	default:
		return v
	}
}

func lookupVarPath(vars map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil, false
	}

	cur, ok := vars[parts[0]]
	if !ok {
		return nil, false
	}

	for _, p := range parts[1:] {
		m, ok := anyToMap(cur)
		if !ok {
			return nil, false
		}
		cur, ok = m[p]
		if !ok {
			return nil, false
		}
	}

	return cur, true
}

func anyToMap(v any) (map[string]any, bool) {
	if v == nil {
		return nil, false
	}
	if m, ok := v.(map[string]any); ok {
		return m, true
	}
	// best-effort: marshal/unmarshal into map
	b, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, false
	}
	return m, true
}
