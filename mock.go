package gopilot

import (
	"encoding/json"

	"github.com/SadikSunbul/gopilot/clients"
)

// MockLLMClient implements LLMProvider for testing multi-step scenarios
type MockLLMClient struct {
	responses    []string
	callCount    int
	systemPrompt string
}

// NewMockLLMClient creates a new mock LLM client with predefined responses
func NewMockLLMClient(responses []string) *MockLLMClient {
	return &MockLLMClient{
		responses: responses,
		callCount: 0,
	}
}

// SetResponses updates the mock responses (useful for testing)
func (m *MockLLMClient) SetResponses(responses []string) {
	m.responses = responses
	m.callCount = 0
}

// Generate simulates LLM response generation
func (m *MockLLMClient) Generate(prompt string) (*clients.LLMResponse, error) {
	if m.callCount >= len(m.responses) {
		// Return unsupported if we've exhausted responses
		return &clients.LLMResponse{
			Agent: "unsupported",
			Parameters: map[string]interface{}{
				"message": "No more responses available",
			},
		}, nil
	}

	response := m.responses[m.callCount]
	m.callCount++

	// Parse the mock response as JSON to extract the intended structure
	var agentResp map[string]interface{}
	if err := json.Unmarshal([]byte(response), &agentResp); err != nil {
		return nil, err
	}

	// Check if this is a function_call action
	if action, ok := agentResp["action"].(string); ok && action == "function_call" {
		// Extract agent and parameters for function calls
		agent, _ := agentResp["agent"].(string)
		parameters, _ := agentResp["parameters"].(map[string]interface{})

		return &clients.LLMResponse{
			Agent:      agent,
			Parameters: parameters,
		}, nil
	}

	// For final_answer, we use a special agent name
	if action, ok := agentResp["action"].(string); ok && action == "final_answer" {
		answer, _ := agentResp["answer"].(string)
		return &clients.LLMResponse{
			Agent: "final_answer",
			Parameters: map[string]interface{}{
				"answer": answer,
			},
		}, nil
	}

	// Fallback: try to parse as standard LLMResponse
	var result clients.LLMResponse
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SetSystemPrompt sets the system prompt (for compatibility)
func (m *MockLLMClient) SetSystemPrompt(systemPrompt string) {
	m.systemPrompt = systemPrompt
}

// GetCallCount returns the number of times Generate has been called
func (m *MockLLMClient) GetCallCount() int {
	return m.callCount
}

// GetSystemPrompt returns the current system prompt
func (m *MockLLMClient) GetSystemPrompt() string {
	return m.systemPrompt
}