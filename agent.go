package gopilot

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/SadikSunbul/gopilot/clients"
)

// AgentMessage represents a single message in the conversation
type AgentMessage struct {
	Role    string      `json:"role"`    // "user", "assistant", "function"
	Content string      `json:"content"` // the message content
	Result  interface{} `json:"result,omitempty"` // function execution result
}

// AgentContext maintains conversation history and function results
type AgentContext struct {
	Messages []AgentMessage         `json:"messages"`
	Results  map[string]interface{} `json:"results"` // named function results
	MaxSteps int                    `json:"max_steps"` // maximum number of steps to prevent infinite loops
}

// NewAgentContext creates a new agent context
func NewAgentContext(maxSteps int) *AgentContext {
	if maxSteps <= 0 {
		maxSteps = 10 // default maximum steps
	}
	return &AgentContext{
		Messages: make([]AgentMessage, 0),
		Results:  make(map[string]interface{}, 0),
		MaxSteps: maxSteps,
	}
}

// AddUserMessage adds a user message to the context
func (ctx *AgentContext) AddUserMessage(content string) {
	ctx.Messages = append(ctx.Messages, AgentMessage{
		Role:    "user",
		Content: content,
	})
}

// AddAssistantMessage adds an assistant message to the context
func (ctx *AgentContext) AddAssistantMessage(content string) {
	ctx.Messages = append(ctx.Messages, AgentMessage{
		Role:    "assistant",
		Content: content,
	})
}

// AddFunctionResult adds a function execution result to the context
func (ctx *AgentContext) AddFunctionResult(functionName string, result interface{}) {
	ctx.Results[functionName] = result
	ctx.Messages = append(ctx.Messages, AgentMessage{
		Role:    "function",
		Content: fmt.Sprintf("Function %s executed", functionName),
		Result:  result,
	})
}

// GetStepCount returns the current number of steps (function calls)
func (ctx *AgentContext) GetStepCount() int {
	count := 0
	for _, msg := range ctx.Messages {
		if msg.Role == "function" {
			count++
		}
	}
	return count
}

// Agent represents a multi-step function calling agent
type Agent struct {
	gopilot *Gopilot
	context *AgentContext
}

// NewAgent creates a new multi-step agent
func NewAgent(gopilot *Gopilot, maxSteps int) *Agent {
	agent := &Agent{
		gopilot: gopilot,
		context: NewAgentContext(maxSteps),
	}
	
	// Set the multi-step system prompt
	agent.setAgentSystemPrompt()
	
	return agent
}

// setAgentSystemPrompt configures the system prompt for multi-step reasoning
func (a *Agent) setAgentSystemPrompt() {
	agentList := a.gopilot.registry.List()

	// Build function descriptions
	var functionDescriptions strings.Builder
	for _, fn := range agentList {
		functionDescriptions.WriteString(fmt.Sprintf("Function: %s\n", fn.GetName()))
		functionDescriptions.WriteString(fmt.Sprintf("Description: %s\n", fn.GetDescription()))
		functionDescriptions.WriteString("Parameters:\n")

		for name, param := range fn.GetParameters() {
			functionDescriptions.WriteString(formatParameterSchema(name, param, 1))
		}
		functionDescriptions.WriteString("\n")
	}

	// Default rules for multi-step reasoning
	rules := `
1. Analyze the user's request to determine if multiple functions are needed
2. Execute functions sequentially, using results from previous calls
3. Only provide final_answer when you have sufficient information
4. Use specific data from previous function results in subsequent calls
5. Provide clear reasoning for each action
`

	// Build final prompt
	prompt := fmt.Sprintf(agentSystemPrompt, rules, functionDescriptions.String())

	a.gopilot.llm.SetSystemPrompt(prompt)
}

// AgentResponse represents the enhanced response from the agent
type AgentResponse struct {
	Action     string                 `json:"action"`     // "function_call" or "final_answer"
	Agent      string                 `json:"agent,omitempty"`      // function name (if action is function_call)
	Parameters map[string]interface{} `json:"parameters,omitempty"` // function parameters (if action is function_call)
	Answer     string                 `json:"answer,omitempty"`     // final answer (if action is final_answer)
	Reasoning  string                 `json:"reasoning,omitempty"`  // explanation of the action
}

// ExecuteMultiStep executes a multi-step reasoning chain
func (a *Agent) ExecuteMultiStep(userInput string) (interface{}, error) {
	// Add initial user message
	a.context.AddUserMessage(userInput)

	for a.context.GetStepCount() < a.context.MaxSteps {
		// Generate next action from LLM
		prompt := a.buildContextualPrompt()
		response, err := a.gopilot.llm.Generate(prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to generate response: %w", err)
		}

		// Parse the enhanced response
		agentResp, err := a.parseAgentResponse(response)
		if err != nil {
			return nil, fmt.Errorf("failed to parse agent response: %w", err)
		}

		// Log the reasoning
		if agentResp.Reasoning != "" {
			log.Printf("Agent reasoning: %s", agentResp.Reasoning)
		}

		// Handle the action
		switch agentResp.Action {
		case "final_answer":
			a.context.AddAssistantMessage(agentResp.Answer)
			return agentResp.Answer, nil

		case "function_call":
			if agentResp.Agent == "" {
				return nil, fmt.Errorf("function name is required for function_call action")
			}

			// Execute the function
			result, err := a.gopilot.FunctionExecute(agentResp.Agent, agentResp.Parameters)
			if err != nil {
				return nil, fmt.Errorf("function execution failed: %w", err)
			}

			// Add result to context
			a.context.AddFunctionResult(agentResp.Agent, result)

		default:
			return nil, fmt.Errorf("unknown action: %s", agentResp.Action)
		}
	}

	return nil, fmt.Errorf("maximum steps (%d) reached without reaching final answer", a.context.MaxSteps)
}

// buildContextualPrompt builds a prompt that includes conversation history and previous results
func (a *Agent) buildContextualPrompt() string {
	prompt := "Based on the conversation history and previous function results, determine the next action.\n\n"
	
	// Add conversation history
	prompt += "Conversation History:\n"
	for _, msg := range a.context.Messages {
		switch msg.Role {
		case "user":
			prompt += fmt.Sprintf("User: %s\n", msg.Content)
		case "assistant":
			prompt += fmt.Sprintf("Assistant: %s\n", msg.Content)
		case "function":
			prompt += fmt.Sprintf("Function Result: %s -> %s\n", msg.Content, a.formatResult(msg.Result))
		}
	}

	// Add available results
	if len(a.context.Results) > 0 {
		prompt += "\nAvailable Results:\n"
		for name, result := range a.context.Results {
			prompt += fmt.Sprintf("- %s: %s\n", name, a.formatResult(result))
		}
	}

	prompt += "\nRespond with your next action in the specified JSON format."
	return prompt
}

// formatResult formats a function result for display in prompts
func (a *Agent) formatResult(result interface{}) string {
	if result == nil {
		return "null"
	}
	
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Sprintf("%v", result)
	}
	return string(data)
}

// parseAgentResponse parses the LLM response into an AgentResponse
func (a *Agent) parseAgentResponse(response *clients.LLMResponse) (*AgentResponse, error) {
	// Handle special case for final_answer
	if response.Agent == "final_answer" {
		answer, _ := response.Parameters["answer"].(string)
		return &AgentResponse{
			Action: "final_answer",
			Answer: answer,
		}, nil
	}
	
	// Check if this is a regular function call (old format)
	if response.Agent != "" {
		// Convert old format to new format
		agentResp := &AgentResponse{
			Action:     "function_call",
			Agent:      response.Agent,
			Parameters: response.Parameters,
		}
		return agentResp, nil
	}
	
	// For new format responses, the response should contain action directly
	// This would happen when the LLM generates the new JSON format
	var agentResp AgentResponse
	
	// Try to unmarshal the entire response as JSON to get the action field
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	
	// Try to parse as AgentResponse
	if err := json.Unmarshal(responseBytes, &agentResp); err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}
	
	return &agentResp, nil
}

// Reset resets the agent context for a new conversation
func (a *Agent) Reset() {
	a.context = NewAgentContext(a.context.MaxSteps)
}

// GetContext returns the current agent context (for debugging/inspection)
func (a *Agent) GetContext() *AgentContext {
	return a.context
}