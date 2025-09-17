# GoPilot: AI-Powered Function Router for Go

<p align="center">
  <img src="gopilot.jpeg" alt="Gopilot Logo" width="200"/>
</p>

[![Development Status](https://img.shields.io/badge/Status-In%20Development-yellow)]()
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.23-blue)]()
[![go-pilot](https://img.shields.io/badge/go--pilot-Visit%20Site-blue)](https://go-pilot.vercel.app/)
[![go-pilot](https://img.shields.io/badge/go--pilot-pkg%20go%20dev-blue)](https://pkg.go.dev/github.com/SadikSunbul/gopilot#FunctionWrapper)

> ⚠️ **Note**: This project is under active development and currently supports only the Gemini LLM.

## Overview

GoPilot is an intelligent automation library that enables natural language interaction with your Go functions. It automatically routes user queries to appropriate functions, manages parameter mapping, and manages execution flow - all through simple speech inputs.

### Key Features

- 🤖 **Natural Language Processing**: Process user queries in natural language
- 🎯 **Automatic Function Routing**: Map queries to the most appropriate function
- 🔄 **Type-Safe Parameter Mapping**: Convert dynamic inputs to strongly-typed parameters
- 🛡️ **Validation Built-in**: Automatic validation of required parameters
- 🔌 **Easy Integration**: Simple API for registering and executing functions
- 🎨 **Flexible Response Handling**: Support for various response types and formats

## Installation

```bash
go get github.com/SadikSunbul/gopilot
```

## Quick Start

Here's a simple example that demonstrates how to use GoPilot for **single-step** function calling:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/SadikSunbul/gopilot"
    "github.com/SadikSunbul/gopilot/clients"
    "github.com/SadikSunbul/gopilot/pkg/generator"
)

// Define your function parameters
type WeatherParams struct {
    City string `json:"city" description:"The name of the city to get weather information for" required:"true"`
}

// Define your function response
type WeatherResponse struct {
    City      string `json:"city"`
    Temp      int    `json:"temp"`
    Condition string `json:"condition"`
}

// Implement your function logic
func GetWeather(params WeatherParams) (WeatherResponse, error) {
    if params.City == "" {
        return WeatherResponse{}, fmt.Errorf("city cannot be empty")
    }
    
    // Your weather API integration here
    return WeatherResponse{
        City:      params.City,
        Temp:      25,
        Condition: "sunny",
    }, nil
}

func main() {
    // Initialize Gemini client
    client, err := clients.NewGeminiClient(context.Background(), "your-api-key", "gemini-2.0-flash")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Create GoPilot instance
    gp, err := gopilot.NewGopilot(client)
    if err != nil {
        log.Fatal("failed to initialize gopilot:", err)
    }

    // Register your function
    weatherFn := &gopilot.Function[WeatherParams, WeatherResponse]{
        Name:        "weather-agent",
        Description: "Gets weather information for a specified city",
        Parameters:  generator.GenerateParameterSchema(WeatherParams{}),
        Execute:     GetWeather,
    }
    
    if err := gp.FunctionRegister(weatherFn); err != nil {
        log.Fatal(err)
    }

    // Set system prompt (required)
    gp.SetSystemPrompt(nil)

    // Process user query
    input := "What's the weather like in Istanbul?"
    
    // Option 1: Generate and Execute separately
    response, err := gp.Generate(input)
    if err != nil {
        log.Fatal(err)
    }
    
    result, err := gp.FunctionExecute(response.Agent, response.Parameters)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %+v\n", result)

    // Option 2: Generate and Execute in one step
    result, err = gp.GenerateAndExecute(input)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %+v\n", result)
}
```

## Advanced Usage

## Advanced Usage

### Multi-Step Function Calling Agent

GoPilot now supports **multi-step reasoning** where the LLM can chain multiple function calls together to solve complex problems. This is perfect for scenarios where you need to gather information from one function and use it in another.

#### Example: Weather + Clothing Recommendation

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
    // Initialize Gemini client
    client, err := clients.NewGeminiClient(context.Background(), "your-api-key", "gemini-2.0-flash")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Create GoPilot instance
    gp, err := gopilot.NewGopilot(client)
    if err != nil {
        log.Fatal(err)
    }

    // Register multi-step functions
    weatherFn := gopilot.CreateWeatherFunction()
    clothingFn := gopilot.CreateClothingFunction()
    
    gp.FunctionRegister(weatherFn)
    gp.FunctionRegister(clothingFn)

    // Create multi-step agent
    agent := gopilot.NewAgent(gp, 5) // max 5 steps

    // Multi-step execution
    result, err := agent.ExecuteMultiStep("Based on today's weather, what should I wear in Istanbul?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Result: %v\n", result)
}
```

#### How Multi-Step Works

1. **User Query**: "Based on today's weather, what should I wear in Istanbul?"
2. **Step 1**: Agent calls `get-weather` function with `{"city": "Istanbul"}`
3. **Step 2**: Agent calls `suggest-clothes` function with weather data from Step 1
4. **Final Answer**: Agent provides comprehensive clothing recommendation

#### Agent Features

- **Context Management**: Maintains conversation history and function results
- **Result Chaining**: Results from previous functions are available to subsequent calls
- **Safety Limits**: Maximum step count prevents infinite loops
- **Dynamic Reasoning**: LLM decides when to stop and provide final answer

### Complex Parameter Types

GoPilot supports nested and complex parameter types:

```go
type TranslateParams struct {
    Text string `json:"text" description:"The text to translate" required:"true"`
    Path struct {
        From    string `json:"from" description:"Source language code" required:"true"`
        To      string `json:"to" description:"Target language code" required:"true"`
        Options struct {
            Style string `json:"style" description:"Translation style"`
        } `json:"options" description:"Additional options"`
    } `json:"path" description:"Translation configuration" required:"true"`
}

type TranslateResponse struct {
    Original   string `json:"original"`
    Translated string `json:"translated"`
    From       string `json:"from"`
    To         string `json:"to"`
    Style      string `json:"style,omitempty"`
}

func Translate(params TranslateParams) (TranslateResponse, error) {
    // Your translation logic here
}

// Register the translation function
translateFn := &gopilot.Function[TranslateParams, TranslateResponse]{
    Name:        "translate-agent",
    Description: "Translates text between languages",
    Parameters:  generator.GenerateParameterSchema(TranslateParams{}),
    Execute:     Translate,
}
```

### Interactive CLI Example

Here's how to create an interactive CLI application:

```go
reader := bufio.NewReader(os.Stdin)
fmt.Println("Welcome! Type 'exit' to exit.")

for {
    fmt.Print("\nQuestion: ")
    input, err := reader.ReadString('\n')
    if err != nil {
        log.Fatal(err)
    }

    input = strings.TrimSpace(input)
    if input == "exit" {
        break
    }

    result, err := gp.GenerateAndExecute(input)
    if err != nil {
        log.Printf("Error: %v\n", err)
        continue
    }

    fmt.Printf("Result: %+v\n", result)
}
```

## Best Practices

1. **Parameter Validation**
   - Always validate required parameters
   - Use descriptive error messages
   - Add proper documentation using struct tags

2. **Error Handling**
   - Return specific error types
   - Handle both function-specific and system errors
   - Provide context in error messages

3. **Function Registration**
   - Use descriptive function names
   - Provide clear function descriptions
   - Document parameter requirements

4. **Type Safety**
   - Use strongly-typed parameters and responses
   - Leverage Go's type system for validation
   - Avoid interface{} when possible

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

## Security

For security concerns, please review our [Security Policy](SECURITY.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📚 [Documentation](https://go-pilot.vercel.app/)
- 🐛 [Issue Tracker](https://github.com/SadikSunbul/gopilot/issues)
- 💬 [Discussions](https://github.com/SadikSunbul/gopilot/discussions)

## Acknowledgments

- Thanks to all contributors who have helped shape GoPilot
- Special thanks to the Go community for their support and feedback


