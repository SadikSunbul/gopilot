package interfaces

// LLMProvider defines the basic interface that any LLM provider must implement
type LLMProvider interface {
	// Generate generates a response based on the given prompt
	Generate(prompt string) (string, error)

	// Stream generates the response in a streaming manner
	Stream(prompt string) (<-chan string, error)
}
