package gopilot

// This represents an agent function
type Function struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Parameters  map[string]ParameterSchema `json:"parameters"`
	Execute     func(params map[string]interface{}) (interface{}, error)
}
