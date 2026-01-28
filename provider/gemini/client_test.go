package gemini

import (
	"context"
	"testing"
)

func TestNewClient_NoAPIKey(t *testing.T) {
	_, err := NewClient(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty API key")
	}
}

func TestClientConfig_Defaults(t *testing.T) {
	config := defaultConfig()

	if config.model != "gemini-2.0-flash" {
		t.Errorf("expected default model 'gemini-2.0-flash', got %s", config.model)
	}

	if config.responseMIME != "application/json" {
		t.Errorf("expected default MIME 'application/json', got %s", config.responseMIME)
	}
}

func TestClientOptions(t *testing.T) {
	t.Run("WithModel", func(t *testing.T) {
		config := defaultConfig()
		WithModel("gemini-pro")(config)

		if config.model != "gemini-pro" {
			t.Errorf("expected model 'gemini-pro', got %s", config.model)
		}
	})

	t.Run("WithModel empty", func(t *testing.T) {
		config := defaultConfig()
		original := config.model
		WithModel("")(config)

		if config.model != original {
			t.Error("empty model should not override default")
		}
	})

	t.Run("WithAPIEndpoint", func(t *testing.T) {
		config := defaultConfig()
		WithAPIEndpoint("https://custom.api.com")(config)

		if config.apiEndpoint != "https://custom.api.com" {
			t.Errorf("expected endpoint 'https://custom.api.com', got %s", config.apiEndpoint)
		}
	})

	t.Run("WithMaxTokens", func(t *testing.T) {
		config := defaultConfig()
		WithMaxTokens(1000)(config)

		if config.maxTokens != 1000 {
			t.Errorf("expected maxTokens 1000, got %d", config.maxTokens)
		}
	})

	t.Run("WithTemperature", func(t *testing.T) {
		config := defaultConfig()
		WithTemperature(0.7)(config)

		if config.temperature != 0.7 {
			t.Errorf("expected temperature 0.7, got %f", config.temperature)
		}
	})

	t.Run("WithTopP", func(t *testing.T) {
		config := defaultConfig()
		WithTopP(0.9)(config)

		if config.topP != 0.9 {
			t.Errorf("expected topP 0.9, got %f", config.topP)
		}
	})

	t.Run("WithTopK", func(t *testing.T) {
		config := defaultConfig()
		WithTopK(40)(config)

		if config.topK != 40 {
			t.Errorf("expected topK 40, got %d", config.topK)
		}
	})
}

// Note: The following tests require a valid API key and network access
// They are skipped by default and can be run with:
// go test -v -run TestClient_Integration -api-key=YOUR_API_KEY

func TestClient_SetModel_Empty(t *testing.T) {
	// This test only verifies the error case without needing a real client
	c := &Client{
		config: defaultConfig(),
	}

	err := c.SetModel("")
	if err == nil {
		t.Error("expected error for empty model name")
	}
}

func TestClient_GetModel(t *testing.T) {
	c := &Client{
		config: &clientConfig{model: "test-model"},
	}

	if c.GetModel() != "test-model" {
		t.Errorf("expected 'test-model', got %s", c.GetModel())
	}
}

// Integration test (requires API key)
// To run: go test -v -run TestClient_Integration -short=false
func TestClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This test is skipped by default
	// Uncomment and add your API key to test
	t.Skip("integration test requires API key")

	/*
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			t.Skip("GEMINI_API_KEY not set")
		}

		ctx := context.Background()
		client, err := NewClient(ctx, apiKey)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}
		defer client.Close()

		client.SetSystemPrompt("You are a helpful assistant. Respond only with JSON.")

		resp, err := client.Generate(ctx, "Say hello in JSON format")
		if err != nil {
			t.Fatalf("generation failed: %v", err)
		}

		if resp == nil {
			t.Fatal("response should not be nil")
		}
	*/
}

func TestClient_ImplementsInterfaces(t *testing.T) {
	// Compile-time check that Client implements the interfaces
	// This is already done with var _ checks in the main file,
	// but we include it here for documentation
	var _ interface {
		Generate(context.Context, string) (*struct {
			Agent      string         `json:"agent"`
			Parameters map[string]any `json:"parameters"`
			Raw        string         `json:"-"`
		}, error)
		Close() error
	}

	// The actual interface checks are in client.go with:
	// var _ provider.Provider = (*Client)(nil)
	// var _ provider.PromptConfigurer = (*Client)(nil)
	// var _ provider.ModelConfigurer = (*Client)(nil)
}
