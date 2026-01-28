package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/SadikSunbul/gopilot"
	"github.com/SadikSunbul/gopilot/provider"
)

/*
Senaryo (Offline Demo):
- Bu örnek, LLM çağrısı yapmadan (API key olmadan) GoPilot kullanımını göstermek için yazıldı.
- Gerçekte LLM sağlayıcısı (örn. Gemini) kullanıcı cümlesinden "hangi agent çalışsın" ve "parametreler ne olsun"
  bilgisini üretir.
- Burada aynı şeyi basit bir "StaticProvider" ile simüle ediyoruz:
  - Eğer input içinde "weather" geçerse weather-agent seçilir.
  - Eğer input içinde "translate" geçerse translate-agent seçilir.
  - Eğer input "my location weather" gibi bir şeyse, tek adım yetmediği için bir PLAN döndürür:
    1) get-location -> city üretir (store_as=location)
    2) weather-agent -> city="$location.city" ile çağrılır
  - Aksi halde unsupported seçilir.

Çalıştırma:
  go run ./examples/offline
*/

type LocationParams struct {
	// No params for offline demo.
}

type LocationResult struct {
	City string `json:"city" description:"City name resolved from the user's current location" example:"Istanbul"`
}

const (
	locationFnName = "get-location"
	weatherFnName  = "weather-agent"
)

func GetLocation(ctx context.Context, params LocationParams) (LocationResult, error) {
	_ = ctx
	_ = params
	// Offline demo: konumu sabit döndürüyoruz.
	return LocationResult{City: "Istanbul"}, nil
}

type WeatherParams struct {
	City string `json:"city" description:"City name" required:"true"`
}

type WeatherResponse struct {
	City      string `json:"city" description:"City name"`
	Temp      int    `json:"temp" description:"Temperature in Celsius" example:"25"`
	Condition string `json:"condition" description:"Weather condition summary" example:"sunny"`
}

func GetWeather(ctx context.Context, params WeatherParams) (WeatherResponse, error) {
	if strings.TrimSpace(params.City) == "" {
		return WeatherResponse{}, fmt.Errorf("city cannot be empty")
	}

	return WeatherResponse{
		City:      params.City,
		Temp:      25,
		Condition: "sunny",
	}, nil
}

type TranslateParams struct {
	Text string `json:"text" description:"Text to translate" required:"true" example:"hello world"`
	Path struct {
		From string `json:"from" description:"Source language code (ISO 639-1)" required:"true" example:"en"`
		To   string `json:"to" description:"Target language code (ISO 639-1)" required:"true" example:"tr"`
	} `json:"path" description:"Translation path configuration" required:"true"`
}

type TranslateResponse struct {
	Original   string `json:"original" description:"Original text" example:"hello world"`
	Translated string `json:"translated" description:"Translated text" example:"merhaba dünya"`
	From       string `json:"from" description:"Source language code" example:"en"`
	To         string `json:"to" description:"Target language code" example:"tr"`
}

func Translate(ctx context.Context, params TranslateParams) (TranslateResponse, error) {
	if strings.TrimSpace(params.Text) == "" {
		return TranslateResponse{}, fmt.Errorf("text cannot be empty")
	}
	if strings.TrimSpace(params.Path.From) == "" || strings.TrimSpace(params.Path.To) == "" {
		return TranslateResponse{}, fmt.Errorf("from/to must be specified")
	}

	// Offline demo: gerçek çeviri yapmıyoruz.
	return TranslateResponse{
		Original:   params.Text,
		Translated: "[demo-translation] " + params.Text,
		From:       params.Path.From,
		To:         params.Path.To,
	}, nil
}

// StaticProvider: LLM yerine deterministic seçim/parametre üretimi.
type StaticProvider struct{}

func (p *StaticProvider) Generate(ctx context.Context, prompt string) (*provider.Response, error) {
	_ = ctx
	text := strings.ToLower(prompt)

	switch {
	case strings.Contains(text, "my location") && strings.Contains(text, "weather"):
		mapOutputFalse := false
		return &provider.Response{
			Type: "plan",
			Steps: []provider.Step{
				{
					Function:  locationFnName,
					Params:    map[string]any{},
					StoreAs:   "location",
					MapOutput: &mapOutputFalse,
				},
				{
					Function: weatherFnName,
					Params: map[string]any{
						"city": "$location.city",
					},
				},
			},
		}, nil
	case strings.Contains(text, "weather"):
		return &provider.Response{
			Agent: "weather-agent",
			Parameters: map[string]any{
				"city": "Istanbul",
			},
		}, nil
	case strings.Contains(text, "translate"):
		return &provider.Response{
			Agent: "translate-agent",
			Parameters: map[string]any{
				"text": "hello world",
				"path": map[string]any{
					"from": "en",
					"to":   "tr",
				},
			},
		}, nil
	default:
		return &provider.Response{
			Agent: "unsupported",
			Parameters: map[string]any{
				"message": "No matching agent for this prompt (offline demo).",
			},
		}, nil
	}
}

func (p *StaticProvider) Close() error { return nil }

func main() {
	gp, err := gopilot.New(&StaticProvider{})
	if err != nil {
		panic(err)
	}

	// Agent kayıtları
	if err := gp.Register(gopilot.NewFunction[LocationParams, LocationResult](
		locationFnName,
		"Returns the user's current location (offline demo returns a fixed city).",
		GetLocation,
	)); err != nil {
		panic(err)
	}

	if err := gp.Register(gopilot.NewFunction[WeatherParams, WeatherResponse](
		weatherFnName,
		"Gets weather information for a specified city",
		GetWeather,
	)); err != nil {
		panic(err)
	}

	if err := gp.Register(gopilot.NewFunction[TranslateParams, TranslateResponse](
		"translate-agent",
		"Translates text between languages",
		Translate,
	)); err != nil {
		panic(err)
	}

	// Offline demo input'ları
	for _, input := range []string{
		"weather please",
		"my location weather",
		"translate this",
		"some other request",
	} {
		out, err := gp.GenerateAndExecuteSmart(context.Background(), input)
		if err != nil {
			fmt.Printf("input=%q error=%v\n", input, err)
			continue
		}
		fmt.Printf("input=%q result=%+v\n", input, out)
	}
}
