package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/SadikSunbul/gopilot"
	"github.com/SadikSunbul/gopilot/clients"
)

func main() {
	fmt.Println("🤖 GoPilot Multi-Step Agent Demo")
	fmt.Println("=================================")
	fmt.Println()

	// Check for API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  GEMINI_API_KEY environment variable not set.")
		fmt.Println("   This demo will run with a mock client for testing purposes.")
		fmt.Println()
		runMockDemo()
		return
	}

	// Initialize real Gemini client
	client, err := clients.NewGeminiClient(context.Background(), apiKey, "gemini-2.0-flash")
	if err != nil {
		log.Fatal("Failed to initialize Gemini client:", err)
	}
	defer client.Close()

	// Create GoPilot instance
	gp, err := gopilot.NewGopilot(client)
	if err != nil {
		log.Fatal("Failed to initialize gopilot:", err)
	}

	// Register example functions
	weatherFn := gopilot.CreateWeatherFunction()
	clothingFn := gopilot.CreateClothingFunction()

	if err := gp.FunctionRegister(weatherFn); err != nil {
		log.Fatal("Failed to register weather function:", err)
	}

	if err := gp.FunctionRegister(clothingFn); err != nil {
		log.Fatal("Failed to register clothing function:", err)
	}

	// Create multi-step agent
	agent := gopilot.NewAgent(gp, 5)

	fmt.Println("🌟 Multi-step agent ready! Try these examples:")
	fmt.Println("   • 'What should I wear in Istanbul today?'")
	fmt.Println("   • 'Based on the weather in London, what clothes should I pack?'")
	fmt.Println("   • 'What's the weather like in Tokyo?'")
	fmt.Println("   • Type 'exit' to quit")
	fmt.Println()

	// Interactive loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("👤 You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading input:", err)
		}

		input = strings.TrimSpace(input)
		if strings.ToLower(input) == "exit" {
			fmt.Println("👋 Goodbye!")
			break
		}

		if input == "" {
			continue
		}

		fmt.Println("🤔 Thinking...")

		// Execute multi-step reasoning
		result, err := agent.ExecuteMultiStep(input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("🤖 Agent: %v\n\n", result)

		// Reset agent for next conversation
		agent.Reset()
	}
}

func runMockDemo() {
	fmt.Println("Running with mock LLM client...")
	fmt.Println()

	// Create mock responses that simulate the weather + clothing chain
	mockResponses := []string{
		`{"agent": "get-weather", "parameters": {"city": "Istanbul"}}`,
		`{"agent": "final_answer", "parameters": {"answer": "Based on the weather in Istanbul today (12°C and cloudy with 75% humidity), I recommend dressing in layers. Wear a light jacket or sweater, long pants, and comfortable shoes. Since it's cloudy and humid, you might want to bring a light umbrella just in case of rain."}}`,
	}

	mockLLM := gopilot.NewMockLLMClient(mockResponses)

	// Create GoPilot instance with mock
	gp, err := gopilot.NewGopilot(mockLLM)
	if err != nil {
		log.Fatal("Failed to initialize gopilot:", err)
	}

	// Register example functions
	weatherFn := gopilot.CreateWeatherFunction()
	clothingFn := gopilot.CreateClothingFunction()

	if err := gp.FunctionRegister(weatherFn); err != nil {
		log.Fatal("Failed to register weather function:", err)
	}

	if err := gp.FunctionRegister(clothingFn); err != nil {
		log.Fatal("Failed to register clothing function:", err)
	}

	// Create multi-step agent
	agent := gopilot.NewAgent(gp, 5)

	fmt.Println("🌟 Demo: Multi-step reasoning for clothing recommendation")
	fmt.Println()

	userInput := "Based on today's weather, what should I wear in Istanbul?"
	fmt.Printf("👤 User: %s\n", userInput)
	fmt.Println("🤔 Agent is thinking through multiple steps...")
	fmt.Println()

	// Execute multi-step reasoning
	result, err := agent.ExecuteMultiStep(userInput)
	if err != nil {
		log.Fatal("Error:", err)
	}

	fmt.Printf("🤖 Final Answer: %v\n", result)
	fmt.Println()

	// Show the step-by-step process
	context := agent.GetContext()
	fmt.Println("📝 Step-by-step process:")
	stepNum := 1
	for _, msg := range context.Messages {
		if msg.Role == "function" {
			fmt.Printf("   Step %d: %s\n", stepNum, msg.Content)
			stepNum++
		}
	}
	fmt.Printf("   Step %d: Provided final answer\n", stepNum)
}