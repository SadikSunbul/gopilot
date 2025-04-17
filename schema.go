package gopilot

// ParameterSchema, a function parameter's schema
type ParameterSchema struct {
	Type        string                     `json:"type"`
	Description string                     `json:"description"`
	Required    bool                       `json:"required"`
	Properties  map[string]ParameterSchema `json:"properties,omitempty"`
}
