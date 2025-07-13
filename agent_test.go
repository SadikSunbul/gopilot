package gopilot

import (
	"encoding/json"
	"strings"
	"testing"
	
	"github.com/SadikSunbul/gopilot/clients"
)

// MockLLMClient implements LLMProvider for testing
type MockLLMClient struct {
	responses []string
	callCount int
	systemPrompt string
}

func NewMockLLMClient(responses []string) *MockLLMClient {
	return &MockLLMClient{
		responses: responses,
		callCount: 0,
	}
}

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
	
	// For final_answer, we need to indicate this differently
	// Since the original LLMResponse doesn't support action field,
	// we'll use a special agent name
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

func (m *MockLLMClient) SetSystemPrompt(systemPrompt string) {
	m.systemPrompt = systemPrompt
}

func TestAgentSingleStep(t *testing.T) {
	// Create mock LLM with single response
	mockLLM := NewMockLLMClient([]string{
		`{"action": "function_call", "agent": "get-weather", "parameters": {"city": "Istanbul"}, "reasoning": "Need to get weather for Istanbul"}`,
		`{"action": "final_answer", "answer": "The weather in Istanbul is 12°C and cloudy with 75% humidity and 15 km/h wind.", "reasoning": "I have the weather information requested"}`,
	})
	
	// Create gopilot instance
	gopilot, err := NewGopilot(mockLLM)
	if err != nil {
		t.Fatalf("Failed to create gopilot: %v", err)
	}
	
	// Register weather function
	weatherFn := CreateWeatherFunction()
	if err := gopilot.FunctionRegister(weatherFn); err != nil {
		t.Fatalf("Failed to register weather function: %v", err)
	}
	
	// Create agent
	agent := NewAgent(gopilot, 5)
	
	// Test single step execution
	result, err := agent.ExecuteMultiStep("What's the weather in Istanbul?")
	if err != nil {
		t.Fatalf("Failed to execute multi-step: %v", err)
	}
	
	// Verify result is a string (final answer)
	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string result, got %T", result)
	}
	
	if resultStr == "" {
		t.Fatalf("Expected non-empty result")
	}
	
	// Verify that the weather function was called
	if mockLLM.callCount != 2 {
		t.Fatalf("Expected 2 LLM calls, got %d", mockLLM.callCount)
	}
}

func TestAgentMultiStep(t *testing.T) {
	// Create mock LLM with multi-step responses
	mockLLM := NewMockLLMClient([]string{
		`{"action": "function_call", "agent": "get-weather", "parameters": {"city": "Istanbul"}, "reasoning": "First need to get weather for Istanbul"}`,
		`{"action": "function_call", "agent": "suggest-clothes", "parameters": {"weather_data": {"city": "Istanbul", "temperature": 12, "condition": "cloudy", "humidity": 75, "wind_speed": 15}}, "reasoning": "Now I can suggest clothes based on the weather"}`,
		`{"action": "final_answer", "answer": "Based on today's weather in Istanbul (12°C and cloudy), I recommend dressing in layers with a light jacket or sweater, long pants, and comfortable shoes. Since it's cloudy, you might also want to bring a light jacket in case it gets cooler.", "reasoning": "I have both weather and clothing suggestions"}`,
	})
	
	// Create gopilot instance
	gopilot, err := NewGopilot(mockLLM)
	if err != nil {
		t.Fatalf("Failed to create gopilot: %v", err)
	}
	
	// Register functions
	weatherFn := CreateWeatherFunction()
	clothingFn := CreateClothingFunction()
	
	if err := gopilot.FunctionRegister(weatherFn); err != nil {
		t.Fatalf("Failed to register weather function: %v", err)
	}
	
	if err := gopilot.FunctionRegister(clothingFn); err != nil {
		t.Fatalf("Failed to register clothing function: %v", err)
	}
	
	// Create agent
	agent := NewAgent(gopilot, 5)
	
	// Test multi-step execution
	result, err := agent.ExecuteMultiStep("Based on today's weather, what should I wear in Istanbul?")
	if err != nil {
		t.Fatalf("Failed to execute multi-step: %v", err)
	}
	
	// Verify result is a string (final answer)
	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("Expected string result, got %T", result)
	}
	
	if resultStr == "" {
		t.Fatalf("Expected non-empty result")
	}
	
	// Verify that both functions were called (3 LLM calls total)
	if mockLLM.callCount != 3 {
		t.Fatalf("Expected 3 LLM calls, got %d", mockLLM.callCount)
	}
	
	// Verify context has the expected number of messages
	context := agent.GetContext()
	if len(context.Messages) < 3 { // user message + 2 function results
		t.Fatalf("Expected at least 3 messages in context, got %d", len(context.Messages))
	}
	
	// Verify that results were stored
	if len(context.Results) < 2 {
		t.Fatalf("Expected at least 2 function results, got %d", len(context.Results))
	}
}

func TestAgentMaxSteps(t *testing.T) {
	// Create mock LLM that always wants to call a function (infinite loop)
	mockLLM := NewMockLLMClient([]string{
		`{"action": "function_call", "agent": "get-weather", "parameters": {"city": "Istanbul"}, "reasoning": "Getting weather"}`,
		`{"action": "function_call", "agent": "get-weather", "parameters": {"city": "Istanbul"}, "reasoning": "Getting weather again"}`,
		`{"action": "function_call", "agent": "get-weather", "parameters": {"city": "Istanbul"}, "reasoning": "Getting weather yet again"}`,
	})
	
	// Create gopilot instance
	gopilot, err := NewGopilot(mockLLM)
	if err != nil {
		t.Fatalf("Failed to create gopilot: %v", err)
	}
	
	// Register weather function
	weatherFn := CreateWeatherFunction()
	if err := gopilot.FunctionRegister(weatherFn); err != nil {
		t.Fatalf("Failed to register weather function: %v", err)
	}
	
	// Create agent with only 2 max steps
	agent := NewAgent(gopilot, 2)
	
	// Test that max steps limit is enforced
	result, err := agent.ExecuteMultiStep("What's the weather in Istanbul?")
	if err == nil {
		t.Fatalf("Expected error due to max steps, but got result: %v", result)
	}
	
	if !strings.Contains(err.Error(), "maximum steps") {
		t.Fatalf("Expected max steps error, got: %v", err)
	}
}