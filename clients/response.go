package clients

type LLMResponse struct {
	Agent      string                 `json:"agent"`
	Parameters map[string]interface{} `json:"parameters"`
}
