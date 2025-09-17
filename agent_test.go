package gopilot

import (
	"strings"
	"testing"
)

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
	if mockLLM.GetCallCount() != 2 {
		t.Fatalf("Expected 2 LLM calls, got %d", mockLLM.GetCallCount())
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
	if mockLLM.GetCallCount() != 3 {
		t.Fatalf("Expected 3 LLM calls, got %d", mockLLM.GetCallCount())
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