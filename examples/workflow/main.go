package main

import (
	"context"
	"fmt"

	"github.com/SadikSunbul/gopilot"
)

/*
Senaryo: "Konum -> Hava Durumu" (Declarative Workflow)

Amaç:
- Workflow ile adımları deklaratif şekilde tanımla.
- İlk adım konumu üretir: {"city":"Istanbul"}
- MapOutput=true olduğu için bu çıktı sonraki adımın input parametrelerine dönüşür.
- İkinci adım weather-agent çalışır.

Çalıştırma:
  go run ./examples/workflow
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
	City string `json:"city" description:"City name"`
	Temp int    `json:"temp" description:"Temperature in Celsius" example:"25"`
}

func GetWeather(ctx context.Context, params WeatherParams) (WeatherResponse, error) {
	_ = ctx
	return WeatherResponse{City: params.City, Temp: 25}, nil
}

func buildWorkflowGopilot() (*gopilot.Gopilot, error) {
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
	g, err := buildWorkflowGopilot()
	if err != nil {
		panic(err)
	}

	workflow := gopilot.NewWorkflow("location-weather").
		WithDescription("Get location then get weather").
		WithErrorStrategy(gopilot.ErrorStrategyFail).
		AddStep(gopilot.WorkflowStep{
			Name:      "location",
			Function:  "get-location",
			MapOutput: true,
			OutputKey: "location",
		}).
		AddStep(gopilot.WorkflowStep{
			Name:      "weather",
			Function:  "weather-agent",
			MapOutput: true,
			OutputKey: "weather",
		})

	res, err := g.ExecuteWorkflow(context.Background(), workflow, map[string]any{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Workflow outputs: %+v\n", res.Outputs)
	fmt.Printf("Final output: %+v\n", res.FinalOutput)
}
