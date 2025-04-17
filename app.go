package gopilot

import (
	"errors"
	"fmt"
	"github.com/SadikSunbul/gopilot/clients"
	"log"
	"strings"
)

type Gopilot struct {
	llm      LLMProvider
	registry *Registry
}

func NewGopilot(llm LLMProvider) (*Gopilot, error) {
	if llm == nil {
		return nil, errors.New("llm is required")
	}

	return &Gopilot{
		llm:      llm,
		registry: NewRegistry(),
	}, nil
}

func (g *Gopilot) FunctionRegister(fn *Function) error {
	return g.registry.register(fn)
}

func (g *Gopilot) FunctionExecute(name string, params map[string]interface{}) (interface{}, error) {
	return g.registry.execute(name, params)
}

func (g *Gopilot) FunctionGet(name string) (*Function, error) {
	return g.registry.get(name)
}

func (g *Gopilot) FunctionsList() []*Function {
	return g.registry.list()
}

// recursively formats the parameter schema, along with the required information
func formatParameterSchema(name string, param ParameterSchema, indentLevel int) string {
	indent := strings.Repeat("\t", indentLevel)
	requiredMark := ""
	if param.Required {
		requiredMark = " [required]"
	}
	paramStr := fmt.Sprintf("%s%s: %s%s (%s)", indent, name, param.Type, requiredMark, param.Description)

	if param.Type == "interface" && len(param.Properties) > 0 {
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

func (g *Gopilot) SetSystemPrompt(importantRules []string, unsupportedFunction *func() *Function) {
	if unsupportedFunction == nil {
		err := g.registry.register(UnsupportedFunction())
		if err != nil {
			log.Fatal("unsupported function register error:", err.Error())
		}
	}
	agentList := g.registry.list()

	agentParameter := ""
	importantRulesParameter := ""

	if importantRules == nil || len(importantRules) == 0 {
		importantRulesParameter = `
1. Only select the *translate-agent* if the user **explicitly** asks for translation.
2. Do not assume translation is needed just because the text is in a different language.
3. For general questions or discussions in any language, choose the appropriate agent based on the *intent*, not the language.
`
	} else {
		for i, v := range importantRules {
			importantRulesParameter += fmt.Sprintf("%d. %s \n", i, v)
		}
	}

	for index, value := range agentList {
		agentParameter += fmt.Sprintf("%d. %s (%s):\n", index, value.Name, value.Description)
		if len(value.Parameters) > 0 {
			for name, p := range value.Parameters {
				agentParameter += fmt.Sprintf("\t- %s", formatParameterSchema(name, p, 2))
			}
		}
	}

	g.llm.SetSystemPrompt(fmt.Sprintf(systemPrompt, importantRulesParameter, agentParameter))
	fmt.Println(fmt.Sprintf(systemPrompt, importantRulesParameter, agentParameter))
}

func (g *Gopilot) Generate(input string) (*clients.LLMResponse, error) {
	return g.llm.Generate(input)
}

func (g *Gopilot) GenerateAndExecute(input string) (interface{}, error) {
	response, err := g.Generate(input)
	if err != nil {
		return nil, err
	}
	return g.FunctionExecute(response.Agent, response.Parameters)
}

func UnsupportedFunction() *Function {
	return &Function{
		Name:        "unsupported",
		Description: "If the user's request doesn't match any of these agents, use the \"unsupported\" agent in your response.",
		Parameters: map[string]ParameterSchema{
			"message": {
				Type:        "string",
				Description: "Contains a simple explanation of the error.",
				Required:    true,
			},
		},
		Execute: func(params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"message": "you made an unsupported request",
			}, nil
		},
	}
}
