package main

import (
	"context"
	"fmt"

	"github.com/SadikSunbul/gopilot"
)

/*
Senaryo: "Konum -> Hava Durumu" (Pipeline)

Amaç:
- Bir kullanıcının konumunu (get-location) bul.
- Bu çıktıyı bir sonraki adıma parametre olarak aktar.
- Hava durumunu (weather-agent) çalıştır.

Bu örnek LLM kullanmaz; doğrudan Pipeline ile fonksiyonları "zincirler".

Çalıştırma:
  go run ./examples/pipeline
*/

type LocationParams struct{}

type LocationResult struct {
	City string `json:"city" description:"City name resolved from the user's current location" example:"Istanbul"`
}

func GetLocation(ctx context.Context, _ LocationParams) (LocationResult, error) {
	_ = ctx
	return LocationResult{City: "Istanbul"}, nil
}

type WeatherParams struct {
	City string `json:"city" description:"City name to fetch weather for" required:"true" example:"Istanbul"`
}

type WeatherResponse struct {
	City      string `json:"city" description:"City name"`
	Temp      int    `json:"temp" description:"Temperature in Celsius" example:"25"`
	Condition string `json:"condition" description:"Weather condition summary" example:"sunny"`
}

func GetWeather(ctx context.Context, params WeatherParams) (WeatherResponse, error) {
	_ = ctx
	return WeatherResponse{
		City:      params.City,
		Temp:      25,
		Condition: "sunny",
	}, nil
}

func buildPipelineGopilot() (*gopilot.Gopilot, error) {
	// Provider gerekmiyor çünkü Generate kullanmıyoruz; plan/pipeline direkt Execute çağırıyor.
	// Yine de New için bir provider şart, o yüzden offline minimal provider kullanıyoruz.
	g, err := gopilot.New(&noopProvider{})
	if err != nil {
		return nil, err
	}

	if err := g.Register(gopilot.NewFunction[LocationParams, LocationResult](
		"get-location",
		"Returns current location",
		GetLocation,
	)); err != nil {
		return nil, err
	}

	if err := g.Register(gopilot.NewFunction[WeatherParams, WeatherResponse](
		"weather-agent",
		"Gets weather info for a city",
		GetWeather,
	)); err != nil {
		return nil, err
	}

	return g, nil
}

func main() {
	g, err := buildPipelineGopilot()
	if err != nil {
		panic(err)
	}

	result, err := g.Pipeline().
		Step("get-location").
		Step("weather-agent").
		Execute(context.Background(), map[string]any{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Pipeline final output: %+v\n", result.FinalOutput)
}
