// Package prompt provides utilities for building system prompts.
package prompt

import (
	"fmt"
	"strings"

	"github.com/SadikSunbul/gopilot/schema"
)

// FunctionInfo contains information about a registered function.
type FunctionInfo struct {
	Name        string
	Description string
	Parameters  map[string]schema.Parameter
}

// Builder builds system prompts for the LLM.
type Builder struct {
	rules     []string
	functions []FunctionInfo
}

// NewBuilder creates a new prompt builder.
func NewBuilder() *Builder {
	return &Builder{
		rules:     defaultRules(),
		functions: make([]FunctionInfo, 0),
	}
}

func defaultRules() []string {
	return []string{
		"Analyze the user's intent carefully before selecting a function",
		"Only select a function if it clearly matches the user's request",
		"Validate all required parameters before execution",
		"If unsure about any parameter, use the \"unsupported\" function",
		"Consider the context and any previous interactions",
	}
}

// WithRules sets custom rules for the prompt.
func (b *Builder) WithRules(rules []string) *Builder {
	if len(rules) > 0 {
		b.rules = rules
	}
	return b
}

// AddFunction adds a function to the prompt.
func (b *Builder) AddFunction(info FunctionInfo) *Builder {
	b.functions = append(b.functions, info)
	return b
}

// AddFunctions adds multiple functions to the prompt.
func (b *Builder) AddFunctions(infos []FunctionInfo) *Builder {
	b.functions = append(b.functions, infos...)
	return b
}

// Build generates the system prompt.
func (b *Builder) Build() string {
	var rulesStr strings.Builder
	for i, rule := range b.rules {
		rulesStr.WriteString(fmt.Sprintf("%d. %s\n", i+1, rule))
	}

	var functionsStr strings.Builder
	for _, fn := range b.functions {
		functionsStr.WriteString(fmt.Sprintf("Function: %s\n", fn.Name))
		functionsStr.WriteString(fmt.Sprintf("Description: %s\n", fn.Description))
		functionsStr.WriteString("Parameters:\n")

		for name, param := range fn.Parameters {
			functionsStr.WriteString(formatParameter(name, param, 1))
		}
		functionsStr.WriteString("\n")
	}

	return fmt.Sprintf(systemPromptTemplate, rulesStr.String(), functionsStr.String())
}

func formatParameter(name string, param schema.Parameter, indent int) string {
	indentStr := strings.Repeat("  ", indent)
	requiredMark := ""
	if param.Required {
		requiredMark = " [required]"
	}

	description := param.Description
	if description == "" {
		description = name
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s%s: %s%s (%s)\n", indentStr, name, param.Type, requiredMark, description))

	if param.Properties != nil {
		for propName, prop := range param.Properties {
			sb.WriteString(formatParameter(propName, prop, indent+1))
		}
	}

	return sb.String()
}

const systemPromptTemplate = `You are an advanced AI Function Router and Parameter Optimizer for the GoPilot system.
Your role is to analyze user requests, determine the most appropriate function, and optimize parameter settings.

CORE RESPONSIBILITIES:
1. Function Selection: Choose the most suitable function based on user intent and context
2. Parameter Optimization: Determine and validate all required parameters
3. Error Prevention: Identify potential issues before execution
4. Context Awareness: Maintain context across multiple interactions

DECISION MAKING RULES:
%s

FUNCTION ANALYSIS:
Before selecting a function, consider:
1. Primary Intent: What is the user's main goal?
2. Context Requirements: What context is needed for successful execution?
3. Parameter Dependencies: Are there dependencies between parameters?
4. Error Scenarios: What could go wrong and how to prevent it?
5. Performance Impact: Consider the computational cost of the function

AVAILABLE FUNCTIONS AND PARAMETERS:
%s

PARAMETER VALIDATION RULES:
1. Required Parameters: Ensure all required parameters are provided
2. Type Checking: Validate parameter types match requirements
3. Value Ranges: Ensure numeric values are within acceptable ranges
4. String Formats: Validate string formats (e.g., email, URL, date)
5. Dependencies: Check for parameter interdependencies

ERROR HANDLING:
If the request cannot be mapped to any function:
1. Use the "unsupported" function
2. Provide a clear explanation of why the request cannot be fulfilled
3. Suggest alternative approaches if possible

RESPONSE FORMAT:
Respond ONLY in ONE of the following JSON formats.

1) Single-step execution (preferred when one function is enough):
{
  "type": "single",
  "agent": "agent-name",
  "parameters": { "param1": "value1" }
}

2) Multi-step plan (use when multiple registered functions must be combined):
{
  "type": "plan",
  "steps": [
    {
      "function": "function-name",
      "params": { "any": "optional static params" },
      "store_as": "optional_variable_name",
      "map_output": true,
      "retries": 0
    }
  ]
}

Rules for plans:
- The plan MUST be executable using ONLY the available registered functions.
- When you need to pass data from one step to another, either:
  - set "map_output": true (the step output becomes the next step's input params), OR
  - set "store_as": "name" and reference it in later params using "$name.field" (e.g. "$location.city").
- If you cannot build an executable plan, return a single-step response that calls "unsupported".
}`
