package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/SadikSunbul/gopilot"
	"github.com/SadikSunbul/gopilot/provider/gemini"
)

/*
Senaryo (Interactive CLI):
- Terminalden doğal dilde soru sorarsın.
- Gemini sağlayıcısı, soruyu hangi agent'in çalıştıracağına ve parametrelerine karar verir.
- GoPilot bu agent'i çalıştırır ve sonucu ekrana basar.

Çalıştırma:
  export GEMINI_API_KEY="..."
  go run ./examples/cli
*/

// WeatherParams represents parameters for weather function
type WeatherParams struct {
	City string `json:"city" description:"The name of the city to get weather information for" required:"true"`
}

// WeatherResponse represents the weather information
type WeatherResponse struct {
	City      string `json:"city" description:"City name" example:"Istanbul"`
	Temp      int    `json:"temp" description:"Temperature in Celsius" example:"25"`
	Condition string `json:"condition" description:"Weather condition summary" example:"sunny"`
}

// GetWeather is the actual weather function implementation
func GetWeather(ctx context.Context, params WeatherParams) (WeatherResponse, error) {
	if params.City == "" {
		return WeatherResponse{}, fmt.Errorf("city cannot be empty")
	}
	fmt.Printf("Hava durumu sorgusu yapılıyor :%s \n", params.City)
	// Burada gerçek hava durumu API'si entegrasyonu yapılacak
	return WeatherResponse{
		City:      params.City,
		Temp:      25,
		Condition: "sunny",
	}, nil
}

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY env değişkeni gerekli")
	}

	client, err := gemini.NewClient(context.Background(), apiKey, gemini.WithModel("gemini-2.0-flash"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("failed to close gemini client: %v", err)
		}
	}()

	gp, err := gopilot.New(client, gopilot.WithStdLogger(gopilot.LogLevelInfo))
	if err != nil {
		log.Fatal("gopilot başlatılamadı:", err.Error())
	}

	if err := gp.Register(NewWeatherFunction()); err != nil {
		log.Fatal(err)
	}
	if err := gp.Register(NewTranslateFunction()); err != nil {
		log.Fatal(err)
	}

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

		response, err := gp.Generate(context.Background(), input)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\nSelected Agent: %s\n", response.Agent)
		fmt.Printf("Parameters: %+v\n\n", response.Parameters)

		// Execute the function
		result, err := gp.Execute(context.Background(), response.Agent, response.Parameters)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Result: %+v\n", result)
	}
}

// TranslationOptions represents optional translation settings
type TranslationOptions struct {
	Style string `json:"style,omitempty" description:"Translation style" enum:"formal,informal" example:"formal"`
}

// TranslationPath represents the translation path
type TranslationPath struct {
	From    string             `json:"from" description:"Source language code (ISO 639-1)" required:"true" example:"en"`
	To      string             `json:"to" description:"Target language code (ISO 639-1)" required:"true" example:"tr"`
	Options TranslationOptions `json:"options,omitempty" description:"Optional translation preferences"`
}

// TranslateParams represents all parameters for translation
type TranslateParams struct {
	Text string `json:"text" description:"The text to translate" required:"true"`
	Path struct {
		From    string `json:"from" description:"Source language code (e.g. 'tr', 'en')" required:"true"`
		To      string `json:"to" description:"Target language code (e.g. 'tr', 'en')" required:"true"`
		Options struct {
			Style string `json:"style" description:"Translation style (e.g. 'formal', 'informal')"`
		} `json:"options" description:"Additional translation options"`
	} `json:"path" description:"Translation path configuration" required:"true"`
}

// TranslateResponse represents the translation result
type TranslateResponse struct {
	Original   string `json:"original"`
	Translated string `json:"translated"`
	From       string `json:"from"`
	To         string `json:"to"`
	Style      string `json:"style,omitempty"`
}

// Translate performs the actual translation
func Translate(ctx context.Context, params TranslateParams) (TranslateResponse, error) {
	if params.Text == "" {
		return TranslateResponse{}, fmt.Errorf("text cannot be empty")
	}

	if params.Path.From == "" || params.Path.To == "" {
		return TranslateResponse{}, fmt.Errorf("from and to languages must be specified")
	}

	fmt.Println("Ceviri işlemi yapılıyor : ", params)

	// Burada gerçek çeviri API'si entegrasyonu yapılacak
	result := TranslateResponse{
		Original:   params.Text,
		Translated: "test value",
		From:       params.Path.From,
		To:         params.Path.To,
	}

	if params.Path.Options.Style != "" {
		result.Style = params.Path.Options.Style
	}

	return result, nil
}

// NewTranslateFunction creates a new translation agent
func NewTranslateFunction() *gopilot.Function[TranslateParams, TranslateResponse] {
	return gopilot.NewFunction[TranslateParams, TranslateResponse](
		"translate-agent",
		"Translates text from one language to another",
		Translate,
	)
}

// NewWeatherFunction creates a new weather agent
func NewWeatherFunction() *gopilot.Function[WeatherParams, WeatherResponse] {
	return gopilot.NewFunction[WeatherParams, WeatherResponse](
		"weather-agent",
		"Gets weather information for a specified city",
		GetWeather,
	)
}
