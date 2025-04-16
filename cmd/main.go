package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	gopilot "github.com/SadikSunbul/Gopilot"
	"github.com/SadikSunbul/Gopilot/clients"
)

func main() {
	apiKey := "your-api-key-here"

	client, err := clients.NewGeminiClient(apiKey, "gemini-2.0-flash")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	gp, err := gopilot.NewGopilot(client)
	if err != nil {
		log.Fatal("gopilot is not run:", err.Error())
	}

	if err := gp.FunctionRegister(NewWeatherFunction()); err != nil {
		log.Fatal(err)
	}
	if err := gp.FunctionRegister(NewTranslateFunction()); err != nil {
		log.Fatal(err)
	}
	if err := gp.FunctionRegister(NewCalculatorFunction()); err != nil {
		log.Fatal(err)
	}

	gp.SetSystemPrompt()

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
			fmt.Println("Bye..")
			break
		}

		response, err := gp.Generate(input)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\nSelected Agent: %s\n", response.Agent)
		fmt.Printf("Parameters: %+v\n\n", response.Parameters)

		// Execute the function
		result, err := gp.FunctionExecute(response.Agent, response.Parameters)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Result: %+v\n", result)
	}
}

// NewCalculatorFunction, calculator agent for new function
func NewCalculatorFunction() *gopilot.Function {
	return &gopilot.Function{
		Name:        "calculator-agent",
		Description: "Calculates mathematical expressions",
		Parameters: map[string]gopilot.ParameterSchema{
			"expression": {
				Type:        "string",
				Description: "The mathematical expression to calculate (e.g. '2 + 2', 'sin(30)')",
				Required:    true,
			},
		},
		Execute: func(params map[string]interface{}) (interface{}, error) {
			expression, ok := params["expression"].(string)
			if !ok {
				return nil, errors.New("expression parameter must be a string")
			}

			// Here we will integrate the real mathematical expression evaluation
			// For now, we are returning an example response
			return map[string]interface{}{
				"expression": expression,
				"result":     4.0,
			}, nil
		},
	}
}

// NewTranslateFunction, translation agent for new function
func NewTranslateFunction() *gopilot.Function {
	return &gopilot.Function{
		Name:        "translate-agent",
		Description: "Translates text from one language to another",
		Parameters: map[string]gopilot.ParameterSchema{
			"text": {
				Type:        "string",
				Description: "The text to translate",
				Required:    true,
			},
			"from": {
				Type:        "string",
				Description: "Source language code (e.g. 'tr', 'en')",
				Required:    true,
			},
			"to": {
				Type:        "string",
				Description: "Target language code (e.g. 'tr', 'en')",
				Required:    true,
			},
		},
		Execute: func(params map[string]interface{}) (interface{}, error) {
			text, ok1 := params["text"].(string)
			from, ok2 := params["from"].(string)
			to, ok3 := params["to"].(string)

			if !ok1 || !ok2 || !ok3 {
				return nil, errors.New("all parameters must be strings")
			}

			// Here we will integrate the real translation API
			// For now, we are returning an example response
			return map[string]interface{}{
				"original":   text,
				"translated": "Translated text example",
				"from":       from,
				"to":         to,
			}, nil
		},
	}
}

// NewWeatherFunction, weather agent for new function
func NewWeatherFunction() *gopilot.Function {
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

			// Burada gerçek hava durumu API'si entegrasyonu yapılacak
			// Şimdilik örnek bir yanıt dönüyoruz
			return map[string]interface{}{
				"city":      city,
				"temp":      25,
				"condition": "sunny",
			}, nil
		},
	}
}
