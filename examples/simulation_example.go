package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/SadikSunbul/gopilot"
)

// SimulationExample demonstrates step-by-step multi-step function calling
func main() {
	fmt.Println("🚀 GoPilot Multi-Step Agent - Detailed Simulation")
	fmt.Println("=" + fmt.Sprintf("%50s", "="))
	fmt.Println()

	// 1. SETUP: Create mock responses that simulate intelligent reasoning
	fmt.Println("📋 SIMULATION SETUP")
	fmt.Println("User Request: 'Based on today's weather, what should I wear in Istanbul?'")
	fmt.Println()

	// These responses simulate what a real LLM would generate
	mockResponses := []string{
		// Step 1: Agent decides to get weather first
		`{
			"action": "function_call",
			"agent": "get-weather", 
			"parameters": {"city": "Istanbul"},
			"reasoning": "I need to get the current weather in Istanbul before I can recommend appropriate clothing"
		}`,
		// Step 2: Agent uses weather data to suggest clothes
		`{
			"action": "function_call",
			"agent": "suggest-clothes",
			"parameters": {
				"weather_data": {
					"city": "Istanbul",
					"temperature": 12,
					"condition": "cloudy", 
					"humidity": 75,
					"wind_speed": 15
				}
			},
			"reasoning": "Now that I have the weather data (12°C, cloudy, 75% humidity), I can suggest appropriate clothing"
		}`,
		// Step 3: Agent provides final comprehensive answer
		`{
			"action": "final_answer",
			"answer": "Based on today's weather in Istanbul (12°C and cloudy with 75% humidity and 15 km/h wind), I recommend dressing in layers: wear a light jacket or sweater, long pants, and comfortable shoes. Since it's cloudy and humid, you might want to bring a light umbrella as there's a chance of rain. The wind speed of 15 km/h means you should wear something that won't be too loose.",
			"reasoning": "I now have complete weather information and clothing suggestions, so I can provide a comprehensive final answer"
		}`,
	}

	// 2. INITIALIZE: Set up the multi-step agent
	fmt.Println("🔧 AGENT INITIALIZATION")
	mockLLM := gopilot.NewMockLLMClient(mockResponses)
	
	gp, err := gopilot.NewGopilot(mockLLM)
	if err != nil {
		log.Fatal("Failed to initialize gopilot:", err)
	}

	// Register the demo functions
	weatherFn := gopilot.CreateWeatherFunction()
	clothingFn := gopilot.CreateClothingFunction()
	
	gp.FunctionRegister(weatherFn)
	gp.FunctionRegister(clothingFn)

	// Create multi-step agent with max 5 steps
	agent := gopilot.NewAgent(gp, 5)
	fmt.Println("✅ Agent created with weather and clothing functions registered")
	fmt.Println("✅ Maximum steps set to 5 to prevent infinite loops")
	fmt.Println()

	// 3. EXECUTION: Run the multi-step reasoning
	fmt.Println("🧠 MULTI-STEP REASONING EXECUTION")
	fmt.Println("─────────────────────────────────────")

	userInput := "Based on today's weather, what should I wear in Istanbul?"
	fmt.Printf("👤 User: %s\n\n", userInput)

	// Execute and track each step
	fmt.Println("🤖 Agent begins multi-step reasoning...\n")

	result, err := agent.ExecuteMultiStep(userInput)
	if err != nil {
		log.Fatal("Error:", err)
	}

	fmt.Printf("🎯 FINAL RESULT: %s\n\n", result)

	// 4. ANALYSIS: Show the detailed step-by-step breakdown
	fmt.Println("📊 DETAILED STEP ANALYSIS")
	fmt.Println("─────────────────────────")

	context := agent.GetContext()
	stepNum := 1

	fmt.Printf("Initial Context:\n")
	fmt.Printf("  • Max Steps: %d\n", context.MaxSteps)
	fmt.Printf("  • Messages: %d\n", len(context.Messages))
	fmt.Printf("  • Function Results: %d\n\n", len(context.Results))

	// Show each step in detail
	for _, msg := range context.Messages {
		switch msg.Role {
		case "user":
			fmt.Printf("📝 Initial Request: %s\n\n", msg.Content)
		case "function":
			fmt.Printf("⚙️  STEP %d: %s\n", stepNum, msg.Content)
			if msg.Result != nil {
				resultBytes, _ := json.MarshalIndent(msg.Result, "   ", "  ")
				fmt.Printf("   📤 Result: %s\n\n", string(resultBytes))
			}
			stepNum++
		case "assistant":
			fmt.Printf("🤖 Final Answer Generated: %s\n\n", msg.Content)
		}
	}

	// 5. FUNCTION RESULTS: Show all stored results
	fmt.Println("💾 STORED FUNCTION RESULTS")
	fmt.Println("──────────────────────────")
	for funcName, result := range context.Results {
		fmt.Printf("🔍 Function: %s\n", funcName)
		resultBytes, _ := json.MarshalIndent(result, "   ", "  ")
		fmt.Printf("   Result: %s\n\n", string(resultBytes))
	}

	// 6. SIMULATION SUMMARY
	fmt.Println("📈 SIMULATION SUMMARY")
	fmt.Println("────────────────────")
	fmt.Printf("✅ Total LLM Calls: %d\n", mockLLM.GetCallCount())
	fmt.Printf("✅ Functions Executed: %d\n", len(context.Results))
	fmt.Printf("✅ Steps Completed: %d/%d\n", context.GetStepCount(), context.MaxSteps)
	fmt.Printf("✅ Final Answer Provided: Yes\n")
	fmt.Println()

	// 7. ARCHITECTURE EXPLANATION
	fmt.Println("🏗️  ARCHITECTURE OVERVIEW")
	fmt.Println("─────────────────────────")
	fmt.Println("1. 🧠 Agent: Orchestrates multi-step reasoning")
	fmt.Println("2. 📚 AgentContext: Maintains conversation history & results")
	fmt.Println("3. 🎯 Function Registry: Available functions (weather, clothing)")
	fmt.Println("4. 🤖 LLM: Makes intelligent decisions about next actions")
	fmt.Println("5. 🔄 Execution Loop: Continues until final_answer or max steps")
	fmt.Println()

	// 8. EXAMPLE OF CHAINING
	fmt.Println("🔗 FUNCTION CHAINING EXAMPLE")
	fmt.Println("───────────────────────────")
	fmt.Println("Weather Function Result → Used as input to → Clothing Function")
	fmt.Println()
	fmt.Println("Step 1: get-weather('Istanbul') →")
	fmt.Println("  {city: 'Istanbul', temperature: 12, condition: 'cloudy', humidity: 75}")
	fmt.Println()
	fmt.Println("Step 2: suggest-clothes(weather_data_from_step_1) →")
	fmt.Println("  {recommendation: 'Dress in layers', items: ['light jacket', 'long pants']}")
	fmt.Println()
	fmt.Println("Step 3: final_answer(combining_all_data) →")
	fmt.Println("  'Based on 12°C cloudy weather in Istanbul, wear layers with light jacket...'")
}