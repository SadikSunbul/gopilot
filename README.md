
<p align="center">
  <img src="gopilot.jpeg" alt="Gopilot Logo" width="200"/>
</p>

[![Development Status](https://img.shields.io/badge/Status-In%20Development-yellow)]()
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.23-blue)]()
[![go-pilot](https://img.shields.io/badge/go--pilot-Visit%20Site-blue)](https://go-pilot.vercel.app/)


> ⚠️ **Note**: This project is under active development and currently supports only the Gemini LLM.

## What is GoPilot?

GoPilot is an intelligent automation platform that enables users to perform complex tasks across systems, APIs, or agents using simple natural language inputs. Powered by advanced language models, GoPilot interprets user commands, selects the appropriate functions or agents, configures parameters, and executes tasks seamlessly—saving time and effort.

For details, go to our [website](https://go-pilot.vercel.app/)

## How It Works

Imagine an application with a complex settings menu. Instead of navigating multiple layers to adjust a setting, a user can simply say, "Set the app's font size to 15." If a GoPilot module is configured for this task, it analyzes the input, identifies the relevant agent or API, sets the necessary parameters, and executes the command as if the user had manually made the change. The result is a streamlined, intuitive experience that feels like having a personal assistant.

## Key Features

- **Natural Language Processing**: Understands and processes conversational or imperfect user inputs.
- **Agent Selection**: Automatically selects the most suitable function, API, or agent for the task.
- **Parameter Automation**: Populates required parameters without user intervention.
- **Versatile Integration**: Integrates with existing systems, APIs, or custom agents for diverse applications.

## Installation

```cli
  go get github.com/SadikSunbul/gopilot
```

## Example Usage
For details, go to our [website](https://go-pilot.vercel.app/)

Below is an example of how to use GoPilot to register and execute a function, such as fetching weather information for a city:

```go

func main() {
    // Initialize Gemini client
    client, err := clients.NewGeminiClient(apiKey, "gemini-2.0-flash")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Initialize GoPilot
    gp, err := gopilot.NewGopilot(client)
    if err != nil {
        log.Fatal("gopilot is not run:", err.Error())
    }

    // Register a weather agent
    if err := gp.FunctionRegister(
        func() *gopilot.Function {
            return &gopilot.Function{
                Name:        "weather-agent",
                Description: "Gets weather information for a specified city",
                Parameters: map[string]gopilot.ParameterSchema{
                    "city": {
                        Type:        "string",
                        Description: "The name of the city to get weather information for",
                        Required:    true,
                    },
                },
                Execute: func(params map[string]interface{}) (interface{}, error) {
                    city, ok := params["city"].(string)
                    if !ok {
                        return nil, errors.New("city parameter must be a string")
                    }
                    // Placeholder for real weather API integration
                    return map[string]interface{}{
                        "city":      city,
                        "temp":      25,
                        "condition": "sunny",
                    }, nil
                },
            }
        }()); err != nil {
        log.Fatal(err)
    }

    // Set system prompt (not optional)
    gp.SetSystemPrompt()

    // Generate and execute a command
    input := "Get the weather for Istanbul"
    response, err := gp.Generate(input)
    if err != nil {
        log.Fatal(err)
    }
    result, err := gp.FunctionExecute(response.Agent, response.Parameters)
    if err != nil {
        log.Fatal(err)
    }

    // Alternatively, use the combined method
    // result, err := gp.GenerateAndExecute(input)
}
```

## Installation

1. Ensure you have Go `>=1.23` installed.
2. Clone the repository:

   ```bash
   git clone https://github.com/SadikSunbul/gopilot.git
   ```
3. Install dependencies:

   ```bash
   cd gopilot
   go mod tidy
   ```
4. Set up your Gemini API key as an environment variable:

   ```bash
   export GEMINI_API_KEY=your-api-key
   ```

## Usage

1. Import the GoPilot package into your Go project.
2. Initialize a client for your chosen LLM (currently only Gemini is supported).
3. Create and register functions or agents with defined parameters and execution logic.
4. Use `Generate` and `FunctionExecute` or the combined `GenerateAndExecute` to process user inputs.

## Contributing

We welcome contributions to GoPilot! To get started:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/your-feature`).
3. Make your changes and commit (`git commit -m "Add your feature"`).
4. Push to your branch (`git push origin feature/your-feature`).
5. Open a Pull Request.

Please read our CONTRIBUTING.md for more details.

## Roadmap

- Support for additional LLMs (e.g., OpenAI, Anthropic).
- Enhanced error handling and logging.
- Pre-built agents for common tasks (e.g., file management, notifications).
- Web-based interface for easier interaction.
- Comprehensive test suite.

## License

This project is licensed under the MIT License. See LICENSE for details.

## Contact

For questions, suggestions, or issues, please:

- Open an issue on GitHub.
- Reach out to the maintainer: [Sadik Sunbul](https://github.com/SadikSunbul)


