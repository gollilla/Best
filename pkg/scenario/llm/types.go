// Package llm provides LLM provider implementations for scenario parsing
package llm

// ScenarioStep represents a single step in a scenario (used by LLM)
type ScenarioStep struct {
	Action      string                 `json:"action"`
	Description string                 `json:"description,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
}

// ActionDefinition defines an action that can be executed
type ActionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  []ParameterDef `json:"parameters"`
}

// AssertionDefinition defines an assertion that can be checked
type AssertionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  []ParameterDef `json:"parameters"`
}

// ParameterDef defines a parameter for an action or assertion
type ParameterDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
}

// StepResult represents the result of executing a step
type StepResult struct {
	StepNumber  int    `json:"stepNumber"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Status      string `json:"status"`
}

// ScenarioContext contains context information for scenario execution
type ScenarioContext struct {
	AvailableActions    []ActionDefinition     `json:"availableActions"`
	AvailableAssertions []AssertionDefinition  `json:"availableAssertions"`
	AgentState          map[string]interface{} `json:"agentState,omitempty"`
	History             []StepResult           `json:"history,omitempty"`
}

// ParseResponse represents the response from parsing a scenario
type ParseResponse struct {
	Steps []ScenarioStep `json:"steps"`
	Error string         `json:"error,omitempty"`
}

// ValidationResponse represents the response from validating a step
type ValidationResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Message represents a chat message for LLM conversation
type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// CompletionRequest represents a generic completion request
type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// CompletionResponse represents a generic completion response
type CompletionResponse struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// SummaryInput contains test results for summary generation
type SummaryInput struct {
	Scenarios []ScenarioResultInput `json:"scenarios"`
}

// ScenarioResultInput contains a single scenario result
type ScenarioResultInput struct {
	Name        string            `json:"name"`
	Success     bool              `json:"success"`
	TotalSteps  int               `json:"totalSteps"`
	PassedSteps int               `json:"passedSteps"`
	FailedSteps int               `json:"failedSteps"`
	Duration    string            `json:"duration"`
	Steps       []StepResultInput `json:"steps,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// StepResultInput contains a single step result
type StepResultInput struct {
	Number      int    `json:"number"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}
