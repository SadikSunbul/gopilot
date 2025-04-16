package clients

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	client       *genai.Client
	model        *genai.GenerativeModel
	systemPrompt string
}

func NewGeminiClient(apiKey string, genaiModel string) (*GeminiClient, error) {
	if apiKey == "" {
		return nil, errors.New("API key required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	if genaiModel == "" {
		genaiModel = "gemini-2.0-flash"
	}

	model := client.GenerativeModel(genaiModel)

	// Configure the model
	model.ResponseMIMEType = "application/json"

	return &GeminiClient{
		client:       client,
		model:        model,
		systemPrompt: "",
	}, nil
}

func (g *GeminiClient) SetSystemPrompt(systemPrompt string) {
	g.model.SystemInstruction = genai.NewUserContent(genai.Text(systemPrompt))
}

func (g *GeminiClient) Generate(prompt string) (*LLMResponse, error) {
	ctx := context.Background()

	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		return nil, errors.New("response not generated")
	}

	response := string(resp.Candidates[0].Content.Parts[0].(genai.Text))

	var result LLMResponse
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (g *GeminiClient) Close() error {
	return g.client.Close()
}
