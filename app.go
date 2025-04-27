package gopilot

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/SadikSunbul/gopilot/clients"
	"github.com/SadikSunbul/gopilot/pkg/generator"
)

type Gopilot struct {
	llm      LLMProvider
	registry *Registry
}

func NewGopilot(llm LLMProvider) (*Gopilot, error) {
	if llm == nil {
		return nil, fmt.Errorf("llm is required")
	}

	return &Gopilot{
		llm:      llm,
		registry: NewRegistry(),
	}, nil
}

// FunctionRegister registers a new function
func (g *Gopilot) FunctionRegister(fn FunctionWrapper) error {
	return g.registry.Register(fn)
}

// FunctionExecute executes a registered function
func (g *Gopilot) FunctionExecute(name string, params interface{}) (interface{}, error) {
	// convert to map[string]interface{} type
	var paramMap map[string]interface{}

	switch p := params.(type) {
	case map[string]interface{}:
		paramMap = p
	default:
		// If params are not already map, let's try JSON conversion
		data, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		if err := json.Unmarshal(data, &paramMap); err != nil {
			return nil, fmt.Errorf("failed to convert params to map: %w", err)
		}
	}

	return g.registry.ExecuteFunction(name, paramMap)
}

// FunctionGet retrieves a registered function
func (g *Gopilot) FunctionGet(name string) (FunctionWrapper, error) {
	return Get[interface{}, interface{}](g.registry, name)
}

// FunctionsList returns all registered functions
func (g *Gopilot) FunctionsList() []FunctionWrapper {
	return g.registry.List()
}

// formatParameterSchema formats the parameter schema for system prompt
func formatParameterSchema(name string, param generator.ParameterSchema, indentLevel int) string {
	indent := strings.Repeat("\t", indentLevel)
	requiredMark := ""
	if param.Required {
		requiredMark = " [required]"
	}

	description := param.Description
	if description == "" {
		description = name
	}

	paramStr := fmt.Sprintf("%s%s: %s%s (%s)", indent, name, param.Type, requiredMark, description)

	if param.Type == "interface" && param.Properties != nil {
		paramStr += " {\n"
		for propName, prop := range param.Properties {
			paramStr += formatParameterSchema(propName, prop, indentLevel+1)
		}
		paramStr += fmt.Sprintf("%s}\n", indent)
	} else {
		paramStr += "\n"
	}

	return paramStr
}

// SetSystemPrompt configures the system prompt
func (g *Gopilot) SetSystemPrompt(importantRules []string, unsupportedFunction ...FunctionWrapper) {
	// Register unsupported function if not already registered
	if len(unsupportedFunction) > 0 {
		if err := g.FunctionRegister(unsupportedFunction[0]); err != nil {
			log.Fatal("unsupported function register error:", err.Error())
		}
	} else {
		if err := g.FunctionRegister(UnsupportedFunction()); err != nil {
			log.Fatal("default unsupported function register error:", err.Error())
		}
	}

	agentList := g.registry.List()

	// Build function descriptions
	var functionDescriptions strings.Builder
	for _, fn := range agentList {
		functionDescriptions.WriteString(fmt.Sprintf("Function: %s\n", fn.GetName()))
		functionDescriptions.WriteString(fmt.Sprintf("Description: %s\n", fn.GetDescription()))
		functionDescriptions.WriteString("Parameters:\n")

		for name, param := range fn.GetParameters() {
			functionDescriptions.WriteString(formatParameterSchema(name, param, 1))
		}
		functionDescriptions.WriteString("\n")
	}

	// Build rules
	var rules strings.Builder
	if importantRules == nil || len(importantRules) == 0 {
		rules.WriteString(`
1. Analyze the user's intent carefully before selecting a function
2. Only select a function if it clearly matches the user's request
3. Validate all required parameters before execution
4. If unsure about any parameter, use the "unsupported" function
5. Consider the context and any previous interactions
`)
	} else {
		for i, rule := range importantRules {
			rules.WriteString(fmt.Sprintf("%d. %s\n", i+1, rule))
		}
	}

	// Build final prompt
	prompt := fmt.Sprintf(systemPrompt, rules.String(), functionDescriptions.String())

	g.llm.SetSystemPrompt(prompt)
}

// Generate generates a response from the LLM
func (g *Gopilot) Generate(input string) (*clients.LLMResponse, error) {
	return g.llm.Generate(input)
}

// GenerateAndExecute generates a response and executes the corresponding function
func (g *Gopilot) GenerateAndExecute(input string) (interface{}, error) {
	response, err := g.Generate(input)
	if err != nil {
		return nil, err
	}
	return g.FunctionExecute(response.Agent, response.Parameters)
}

// UnsupportedParams represents parameters for unsupported function
type UnsupportedParams struct {
	Message string `json:"message" description:"Contains a simple explanation of the error." required:"true"`
}

// UnsupportedResponse represents the response from unsupported function
type UnsupportedResponse struct {
	Message string `json:"message"`
}

// UnsupportedFunction creates a new unsupported function handler
func UnsupportedFunction() *Function[UnsupportedParams, UnsupportedResponse] {
	fn := &Function[UnsupportedParams, UnsupportedResponse]{
		Name:        "unsupported",
		Description: "If the user's request doesn't match any of these agents, use the \"unsupported\" agent in your response.",
		Parameters:  generator.GenerateParameterSchema(UnsupportedParams{}),
		Execute: func(params UnsupportedParams) (UnsupportedResponse, error) {
			return UnsupportedResponse{
				Message: "you made an unsupported request: " + params.Message,
			}, nil
		},
	}
	return fn
}
