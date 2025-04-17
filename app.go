package gopilot

import (
	"errors"
	"fmt"
	"github.com/SadikSunbul/gopilot/clients"
	"log"
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

// if unsupportedFunctionActive is nil default gopilot.UnsupportedFunction()
func (g *Gopilot) SetSystemPrompt(importantRules []string, unsupportedFunction *func() *Function) {
	if unsupportedFunction == nil {
		err := g.registry.register(UnsupportedFunction())
		if err != nil {
			log.Fatal("unsported function is register it error:", err.Error())
		}
	}
	agentlist := g.registry.list()

	agentparameter := ""
	importantRulesParameter := ""

	if importantRules == nil || len(importantRules) == 0 {
		importantRulesParameter = "\n1. Only select translate-agent if the user EXPLICITLY asks for translation\n2. Do not assume translation is needed just because the text is in a different language\n3. For general questions or discussions in any language, use the appropriate agent based on the intent, not the language"
	} else {
		for i, v := range importantRules {
			importantRulesParameter += fmt.Sprintf("%d. %s \n", i, v)
		}
	}

	for index, value := range agentlist {
		agentparameter += fmt.Sprintf("%d. %s (%s):\n", index, value.Name, value.Description)
		if len(value.Parameters) > 0 {
			for name, p := range value.Parameters {
				parmeter := fmt.Sprintf("%s: %s (%s)", name, p.Type, p.Description)
				agentparameter += fmt.Sprintf("\t - %s \n", parmeter)
			}
		}

	}

	// Create command here will be added more
	g.llm.SetSystemPrompt(fmt.Sprintf(systemPrompt, importantRulesParameter, agentparameter))

	fmt.Println(fmt.Sprintf(systemPrompt, importantRulesParameter, agentparameter))
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
		Parameters:  map[string]ParameterSchema{},
		Execute: func(params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"message": "you made an unsupported request",
			}, nil
		},
	}
}
