package gopilot

import "github.com/nolva/gopilot/clients"

// LLMProvider defines the basic interface that any LLM provider must implement
type LLMProvider interface {
	// Generate generates a response based on the given prompt
	Generate(prompt string) (*clients.LLMResponse, error)

	SetSystemPrompt(systemPrompt string)
}
