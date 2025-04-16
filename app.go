package gopilot

import (
	"errors"
	"github.com/SadikSunbul/Gopilot/clients"
)

type Gopilot struct {
	llm      LLMProvider
	registry *Registry
}

var systemPrompt = `You are an agent selector and parameter determiner.
Based on the user's request, select the appropriate agent from the following list and specify the required parameters.

IMPORTANT RULES:
1. Only select translate-agent if the user EXPLICITLY asks for translation
2. Do not assume translation is needed just because the text is in a different language
3. For general questions or discussions in any language, use the appropriate agent based on the intent, not the language

Available agents and their parameters:
1. weather-agent:
   - city: string (city name)

2. translate-agent:
   - text: string (text to translate)
   - from: string (source language code, e.g., "tr", "en")
   - to: string (target language code)

3. calculator-agent:
   - expression: string (mathematical expression)

If the user's request doesn't match any of these agents, use the "unsupported" agent in your response.

Provide your response ONLY in the following JSON format, without any additional text:
{
    "agent": "agent-name",
    "parameters": {
        "parameter1": "value1"
    }
}`

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

func (g *Gopilot) SetSystemPrompt() {
	// Create command here will be added more
	g.llm.SetSystemPrompt(systemPrompt)
}

func (g *Gopilot) Generate(input string) (*clients.LLMResponse, error) {
	return g.llm.Generate(input)
}
