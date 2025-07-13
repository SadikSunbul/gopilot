package gopilot

import (
	"fmt"
	"strings"

	"github.com/SadikSunbul/gopilot/pkg/generator"
)

// WeatherParams represents parameters for getting weather information
type WeatherParams struct {
	City string `json:"city" description:"The name of the city to get weather information for" required:"true"`
}

// WeatherResponse represents weather information
type WeatherResponse struct {
	City        string `json:"city"`
	Temperature int    `json:"temperature"`
	Condition   string `json:"condition"`
	Humidity    int    `json:"humidity"`
	WindSpeed   int    `json:"wind_speed"`
}

// GetWeather simulates getting weather information for a city
func GetWeather(params WeatherParams) (WeatherResponse, error) {
	if params.City == "" {
		return WeatherResponse{}, fmt.Errorf("city cannot be empty")
	}

	// Simulate weather data based on city
	weatherData := map[string]WeatherResponse{
		"istanbul": {
			City:        "Istanbul",
			Temperature: 12,
			Condition:   "cloudy",
			Humidity:    75,
			WindSpeed:   15,
		},
		"london": {
			City:        "London",
			Temperature: 8,
			Condition:   "rainy",
			Humidity:    85,
			WindSpeed:   20,
		},
		"tokyo": {
			City:        "Tokyo",
			Temperature: 18,
			Condition:   "sunny",
			Humidity:    60,
			WindSpeed:   10,
		},
		"new york": {
			City:        "New York",
			Temperature: 5,
			Condition:   "snowy",
			Humidity:    70,
			WindSpeed:   25,
		},
	}

	cityKey := strings.ToLower(params.City)
	if weather, exists := weatherData[cityKey]; exists {
		return weather, nil
	}

	// Default weather for unknown cities
	return WeatherResponse{
		City:        params.City,
		Temperature: 20,
		Condition:   "partly cloudy",
		Humidity:    65,
		WindSpeed:   12,
	}, nil
}

// ClothingParams represents parameters for clothing suggestions
type ClothingParams struct {
	WeatherData WeatherResponse `json:"weather_data" description:"Weather information to base clothing suggestions on" required:"true"`
}

// ClothingResponse represents clothing suggestions
type ClothingResponse struct {
	Recommendation string   `json:"recommendation"`
	Items          []string `json:"items"`
	Reasoning      string   `json:"reasoning"`
}

// SuggestClothes suggests appropriate clothing based on weather data
func SuggestClothes(params ClothingParams) (ClothingResponse, error) {
	weather := params.WeatherData
	if weather.City == "" {
		return ClothingResponse{}, fmt.Errorf("weather data is required")
	}

	var recommendation string
	var items []string
	var reasoning string

	temp := weather.Temperature
	condition := strings.ToLower(weather.Condition)

	// Base clothing on temperature
	if temp < 0 {
		items = append(items, "heavy winter coat", "thermal underwear", "warm boots", "gloves", "hat")
		recommendation = "Dress very warmly"
		reasoning = fmt.Sprintf("Temperature is %d°C, which is freezing", temp)
	} else if temp < 10 {
		items = append(items, "warm jacket", "long pants", "closed shoes", "scarf")
		recommendation = "Dress warmly"
		reasoning = fmt.Sprintf("Temperature is %d°C, which is cold", temp)
	} else if temp < 20 {
		items = append(items, "light jacket or sweater", "long pants", "comfortable shoes")
		recommendation = "Dress in layers"
		reasoning = fmt.Sprintf("Temperature is %d°C, which is cool", temp)
	} else if temp < 30 {
		items = append(items, "light shirt", "comfortable pants", "light shoes")
		recommendation = "Dress comfortably"
		reasoning = fmt.Sprintf("Temperature is %d°C, which is pleasant", temp)
	} else {
		items = append(items, "light clothing", "shorts", "sandals", "sun hat")
		recommendation = "Dress lightly"
		reasoning = fmt.Sprintf("Temperature is %d°C, which is hot", temp)
	}

	// Adjust for weather conditions
	if strings.Contains(condition, "rain") {
		items = append(items, "waterproof jacket", "umbrella", "waterproof shoes")
		reasoning += " and it's rainy"
	} else if strings.Contains(condition, "snow") {
		items = append(items, "warm waterproof boots", "snow gloves")
		reasoning += " and it's snowy"
	} else if strings.Contains(condition, "wind") || weather.WindSpeed > 20 {
		items = append(items, "windbreaker")
		reasoning += " and it's windy"
	} else if strings.Contains(condition, "sunny") {
		items = append(items, "sunglasses", "sunscreen")
		reasoning += " and it's sunny"
	}

	return ClothingResponse{
		Recommendation: recommendation,
		Items:          items,
		Reasoning:      reasoning,
	}, nil
}

// CreateWeatherFunction creates a weather function for the registry
func CreateWeatherFunction() *Function[WeatherParams, WeatherResponse] {
	return &Function[WeatherParams, WeatherResponse]{
		Name:        "get-weather",
		Description: "Gets current weather information for a specified city",
		Parameters:  generator.GenerateParameterSchema(WeatherParams{}),
		Execute:     GetWeather,
	}
}

// CreateClothingFunction creates a clothing suggestion function for the registry
func CreateClothingFunction() *Function[ClothingParams, ClothingResponse] {
	return &Function[ClothingParams, ClothingResponse]{
		Name:        "suggest-clothes",
		Description: "Suggests appropriate clothing based on weather conditions",
		Parameters:  generator.GenerateParameterSchema(ClothingParams{}),
		Execute:     SuggestClothes,
	}
}