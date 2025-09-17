# Multi-Step Function Calling Agent - Detailed Implementation Guide

## Overview of Changes

I've implemented a comprehensive multi-step function calling agent that enables the LLM to reason through complex problems by chaining multiple function calls together. Here's what was added:

## 🏗️ Key Components Added

### 1. `agent.go` - Core Multi-Step Agent
```go
// AgentContext maintains conversation history and function results
type AgentContext struct {
    Messages []AgentMessage         `json:"messages"`
    Results  map[string]interface{} `json:"results"` // stored function results
    MaxSteps int                    `json:"max_steps"` // prevents infinite loops
}

// Agent orchestrates multi-step function calling
type Agent struct {
    gopilot *Gopilot
    context *AgentContext
}
```

### 2. `examples.go` - Demo Functions
```go
// WeatherResponse - Gets weather data for cities
type WeatherResponse struct {
    City        string `json:"city"`
    Temperature int    `json:"temperature"`
    Condition   string `json:"condition"`
    Humidity    int    `json:"humidity"`
    WindSpeed   int    `json:"wind_speed"`
}

// ClothingResponse - Suggests clothes based on weather
type ClothingResponse struct {
    Recommendation string   `json:"recommendation"`
    Items          []string `json:"items"`
    Reasoning      string   `json:"reasoning"`
}
```

### 3. `mock.go` - Testing Infrastructure
```go
// MockLLMClient for deterministic testing
type MockLLMClient struct {
    responses    []string
    callCount    int
    systemPrompt string
}
```

## 🚀 Usage Examples

### Single-Step vs Multi-Step Comparison

**Before (Single-Step):**
```go
// Old way - direct function execution
gp := gopilot.NewGopilot(client)
gp.FunctionRegister(weatherFn)
result, err := gp.FunctionExecute("get-weather", map[string]interface{}{
    "city": "Istanbul",
})
```

**After (Multi-Step Agent):**
```go
// New way - intelligent multi-step reasoning
agent := gopilot.NewAgent(gp, 5) // max 5 steps
result, err := agent.ExecuteMultiStep("Based on today's weather, what should I wear in Istanbul?")

// Agent automatically:
// Step 1: Calls get-weather("Istanbul") 
// Step 2: Uses weather data to call suggest-clothes(weather_data)
// Step 3: Provides final comprehensive answer
```

## 🔄 Multi-Step Execution Flow

Here's a step-by-step simulation of how the agent works:

### Step-by-Step Process Simulation

```go
package main

import (
    "fmt"
    "github.com/SadikSunbul/gopilot"
)

func demonstrateMultiStep() {
    // 1. User asks complex question
    userInput := "Based on today's weather, what should I wear in Istanbul?"
    
    // 2. Agent breaks it down into steps
    fmt.Println("🧠 Agent reasoning:")
    fmt.Println("   'This requires weather data first, then clothing suggestions'")
    
    // 3. Step 1: Get weather
    fmt.Println("⚙️  Step 1: Calling get-weather")
    weatherResult := map[string]interface{}{
        "city":        "Istanbul",
        "temperature": 12,
        "condition":   "cloudy",
        "humidity":    75,
        "wind_speed":  15,
    }
    fmt.Printf("   📤 Weather result: %+v\n", weatherResult)
    
    // 4. Step 2: Use weather data for clothing suggestions
    fmt.Println("⚙️  Step 2: Calling suggest-clothes with weather data")
    clothingResult := map[string]interface{}{
        "recommendation": "Dress in layers",
        "items":         []string{"light jacket", "long pants", "comfortable shoes"},
        "reasoning":     "Temperature is 12°C, which is cool",
    }
    fmt.Printf("   📤 Clothing result: %+v\n", clothingResult)
    
    // 5. Step 3: Generate final answer
    fmt.Println("🎯 Final answer generated combining all data")
    finalAnswer := "Based on today's weather in Istanbul (12°C and cloudy), I recommend dressing in layers with a light jacket, long pants, and comfortable shoes."
    fmt.Printf("   🤖 Answer: %s\n", finalAnswer)
}
```

## 📋 Complete Working Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/SadikSunbul/gopilot"
    "github.com/SadikSunbul/gopilot/clients"
)

func main() {
    // Initialize with real Gemini API or mock for testing
    var client gopilot.LLMProvider
    
    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey != "" {
        // Use real Gemini
        geminiClient, err := clients.NewGeminiClient(context.Background(), apiKey, "gemini-2.0-flash")
        if err != nil {
            log.Fatal(err)
        }
        client = geminiClient
    } else {
        // Use mock for testing
        mockResponses := []string{
            `{"action": "function_call", "agent": "get-weather", "parameters": {"city": "Istanbul"}}`,
            `{"action": "function_call", "agent": "suggest-clothes", "parameters": {"weather_data": {"city": "Istanbul", "temperature": 12, "condition": "cloudy", "humidity": 75}}}`,
            `{"action": "final_answer", "answer": "Based on 12°C cloudy weather, wear layers with light jacket..."}`,
        }
        client = gopilot.NewMockLLMClient(mockResponses)
    }
    
    // Create GoPilot instance
    gp, err := gopilot.NewGopilot(client)
    if err != nil {
        log.Fatal(err)
    }
    
    // Register functions
    weatherFn := gopilot.CreateWeatherFunction()
    clothingFn := gopilot.CreateClothingFunction()
    
    gp.FunctionRegister(weatherFn)
    gp.FunctionRegister(clothingFn)
    
    // Create multi-step agent
    agent := gopilot.NewAgent(gp, 5) // max 5 steps
    
    // Execute complex reasoning
    result, err := agent.ExecuteMultiStep("Based on today's weather, what should I wear in Istanbul?")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Final Answer: %s\n", result)
    
    // Inspect the reasoning process
    context := agent.GetContext()
    fmt.Printf("\nSteps taken: %d\n", context.GetStepCount())
    fmt.Printf("Functions called: %d\n", len(context.Results))
    
    // Show each function result
    for funcName, result := range context.Results {
        fmt.Printf("- %s: %+v\n", funcName, result)
    }
}
```

## 🧪 Testing Examples

```go
func TestMultiStepAgent(t *testing.T) {
    // Create deterministic mock responses
    mockLLM := gopilot.NewMockLLMClient([]string{
        `{"action": "function_call", "agent": "get-weather", "parameters": {"city": "Istanbul"}}`,
        `{"action": "function_call", "agent": "suggest-clothes", "parameters": {"weather_data": {...}}}`,
        `{"action": "final_answer", "answer": "Comprehensive clothing recommendation"}`,
    })
    
    gp, _ := gopilot.NewGopilot(mockLLM)
    gp.FunctionRegister(gopilot.CreateWeatherFunction())
    gp.FunctionRegister(gopilot.CreateClothingFunction())
    
    agent := gopilot.NewAgent(gp, 5)
    
    result, err := agent.ExecuteMultiStep("What should I wear in Istanbul?")
    
    assert.NoError(t, err)
    assert.Contains(t, result.(string), "recommendation")
    assert.Equal(t, 3, mockLLM.GetCallCount()) // 3 LLM calls made
}
```

## 🛡️ Safety Features

### Maximum Steps Prevention
```go
// Prevents infinite loops
agent := gopilot.NewAgent(gp, 3) // Will stop after 3 function calls

// If agent tries to exceed limit:
result, err := agent.ExecuteMultiStep("complex request")
// Returns error: "maximum steps (3) reached without reaching final answer"
```

### Context Management
```go
// Access conversation history and results
context := agent.GetContext()
fmt.Printf("Messages: %d\n", len(context.Messages))
fmt.Printf("Function results: %d\n", len(context.Results))
fmt.Printf("Steps used: %d/%d\n", context.GetStepCount(), context.MaxSteps)

// Reset for new conversation
agent.Reset()
```

## 🎯 Key Benefits

1. **Intelligent Chaining**: Functions automatically use results from previous calls
2. **Natural Language**: Complex requests in plain English
3. **Type Safety**: All function parameters are strongly typed
4. **Testing Support**: Mock client for deterministic testing
5. **Safety Controls**: Maximum step limits prevent infinite loops
6. **Backward Compatible**: Existing single-step usage still works

## 🔍 Before vs After Comparison

| Feature | Before (Single-Step) | After (Multi-Step Agent) |
|---------|---------------------|--------------------------|
| **Complexity** | One function per request | Multiple chained functions |
| **Context** | No memory between calls | Maintains conversation history |
| **Intelligence** | Manual parameter mapping | Automatic reasoning and chaining |
| **Use Cases** | Simple, direct tasks | Complex multi-step scenarios |
| **Safety** | Basic error handling | Step limits, comprehensive validation |

This implementation provides a solid foundation for complex AI-powered automation while maintaining the simplicity and reliability of the existing single-step functionality.