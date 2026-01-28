// Package gemini provides a Gemini LLM provider implementation.
package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/SadikSunbul/gopilot/provider"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Ensure Client implements the required interfaces.
var (
	_ provider.Provider         = (*Client)(nil)
	_ provider.PromptConfigurer = (*Client)(nil)
	_ provider.ModelConfigurer  = (*Client)(nil)
)

// ClientOption is a functional option for configuring the Gemini client.
type ClientOption func(*clientConfig)

type clientConfig struct {
	model        string
	apiEndpoint  string
	maxTokens    int32
	temperature  float32
	topP         float32
	topK         int32
	responseMIME string
}

func defaultConfig() *clientConfig {
	return &clientConfig{
		model:        "gemini-2.0-flash",
		responseMIME: "application/json",
	}
}

// WithModel sets the model to use.
func WithModel(model string) ClientOption {
	return func(c *clientConfig) {
		if model != "" {
			c.model = model
		}
	}
}

// WithAPIEndpoint sets a custom API endpoint.
func WithAPIEndpoint(endpoint string) ClientOption {
	return func(c *clientConfig) {
		c.apiEndpoint = endpoint
	}
}

// WithMaxTokens sets the maximum number of tokens to generate.
func WithMaxTokens(tokens int32) ClientOption {
	return func(c *clientConfig) {
		c.maxTokens = tokens
	}
}

// WithTemperature sets the temperature for generation.
func WithTemperature(temp float32) ClientOption {
	return func(c *clientConfig) {
		c.temperature = temp
	}
}

// WithTopP sets the top-p value for generation.
func WithTopP(topP float32) ClientOption {
	return func(c *clientConfig) {
		c.topP = topP
	}
}

// WithTopK sets the top-k value for generation.
func WithTopK(topK int32) ClientOption {
	return func(c *clientConfig) {
		c.topK = topK
	}
}

// Client is a Gemini LLM provider.
type Client struct {
	client *genai.Client
	model  *genai.GenerativeModel
	config *clientConfig
	mu     sync.RWMutex
}

// NewClient creates a new Gemini client.
func NewClient(ctx context.Context, apiKey string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("gemini: API key is required")
	}

	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	clientOpts := []option.ClientOption{option.WithAPIKey(apiKey)}
	if config.apiEndpoint != "" {
		clientOpts = append(clientOpts, option.WithEndpoint(config.apiEndpoint))
	}

	client, err := genai.NewClient(ctx, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to create client: %w", err)
	}

	model := client.GenerativeModel(config.model)
	model.ResponseMIMEType = config.responseMIME

	if config.maxTokens > 0 {
		model.MaxOutputTokens = &config.maxTokens
	}
	if config.temperature > 0 {
		model.Temperature = &config.temperature
	}
	if config.topP > 0 {
		model.TopP = &config.topP
	}
	if config.topK > 0 {
		model.TopK = &config.topK
	}

	return &Client{
		client: client,
		model:  model,
		config: config,
	}, nil
}

// SetSystemPrompt sets the system prompt for the model.
func (c *Client) SetSystemPrompt(prompt string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.model.SystemInstruction = genai.NewUserContent(genai.Text(prompt))
}

// SetModel changes the model being used.
func (c *Client) SetModel(model string) error {
	if model == "" {
		return errors.New("gemini: model name cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.config.model = model
	c.model = c.client.GenerativeModel(model)
	c.model.ResponseMIMEType = c.config.responseMIME

	if c.config.maxTokens > 0 {
		c.model.MaxOutputTokens = &c.config.maxTokens
	}
	if c.config.temperature > 0 {
		c.model.Temperature = &c.config.temperature
	}
	if c.config.topP > 0 {
		c.model.TopP = &c.config.topP
	}
	if c.config.topK > 0 {
		c.model.TopK = &c.config.topK
	}

	return nil
}

// GetModel returns the current model name.
func (c *Client) GetModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.model
}

// Generate generates a response from the model.
func (c *Client) Generate(ctx context.Context, prompt string) (*provider.Response, error) {
	c.mu.RLock()
	model := c.model
	c.mu.RUnlock()

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini: generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, errors.New("gemini: no response generated")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, errors.New("gemini: empty response content")
	}

	text, ok := candidate.Content.Parts[0].(genai.Text)
	if !ok {
		return nil, errors.New("gemini: unexpected response type")
	}

	rawResponse := string(text)

	var result provider.Response
	if err := json.Unmarshal([]byte(rawResponse), &result); err != nil {
		return nil, fmt.Errorf("gemini: failed to parse response: %w", err)
	}

	result.Raw = rawResponse
	return &result, nil
}

// Close closes the client and releases resources.
func (c *Client) Close() error {
	return c.client.Close()
}
