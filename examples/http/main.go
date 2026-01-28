package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SadikSunbul/gopilot"
	"github.com/SadikSunbul/gopilot/provider"
)

/*
Senaryo: HTTP API (LLM yerine offline plan üretimi)

Amaç:
- /run endpoint'i JSON input alır: {"input":"my location weather"}
- GoPilot, GenerateAndExecuteSmart ile:
  - tek fonksiyon yeterliyse single çalıştırır
  - değilse plan döner ve adımları uygular

Çalıştırma:
  go run ./examples/http
  curl -XPOST localhost:8080/run -d '{"input":"my location weather"}'
*/

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
				{Function: "get-location", StoreAs: "location", MapOutput: &mapOutputFalse},
				{Function: "weather-agent", Params: map[string]any{"city": "$location.city"}},
			},
		}, nil
	default:
		return &provider.Response{
			Type:       "single",
			Agent:      "unsupported",
			Parameters: map[string]any{"message": "no route for this input"},
		}, nil
	}
}

func (p *StaticProvider) Close() error { return nil }

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

func GetWeather(ctx context.Context, p WeatherParams) (WeatherResponse, error) {
	_ = ctx
	return WeatherResponse{City: p.City, Temp: 25}, nil
}

type Server struct {
	gp *gopilot.Gopilot
}

type runReq struct {
	Input string `json:"input" description:"Natural language user query" required:"true" example:"my location weather"`
}

type runResp struct {
	Result any    `json:"result"`
	Error  string `json:"error,omitempty"`
}

func NewServer() (*Server, error) {
	gp, err := gopilot.New(&StaticProvider{})
	if err != nil {
		return nil, err
	}
	if err := gp.Register(gopilot.NewFunction[LocationParams, LocationResult](
		"get-location", "returns current location", GetLocation,
	)); err != nil {
		return nil, err
	}
	if err := gp.Register(gopilot.NewFunction[WeatherParams, WeatherResponse](
		"weather-agent", "returns weather by city", GetWeather,
	)); err != nil {
		return nil, err
	}
	return &Server{gp: gp}, nil
}

func (s *Server) RunHandler(w http.ResponseWriter, r *http.Request) {
	var req runReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(runResp{Error: "invalid json"})
		return
	}

	out, err := s.gp.GenerateAndExecuteSmart(r.Context(), req.Input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(runResp{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(runResp{Result: out})
}

func main() {
	s, err := NewServer()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/run", s.RunHandler)
	_ = http.ListenAndServe(":8080", mux)
}
