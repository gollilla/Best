// Package scenario provides natural language scenario execution for Minecraft Bedrock server testing
package scenario

import (
	"time"
)

// StepStatus represents the status of a scenario step
type StepStatus string

const (
	StepStatusPending  StepStatus = "pending"
	StepStatusRunning  StepStatus = "running"
	StepStatusPassed   StepStatus = "passed"
	StepStatusFailed   StepStatus = "failed"
	StepStatusSkipped  StepStatus = "skipped"
)

// ScenarioStep represents a single step in a scenario
type ScenarioStep struct {
	Action      string                 `json:"action"`
	Description string                 `json:"description,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty"`
}

// StepResult represents the result of executing a scenario step
type StepResult struct {
	StepNumber  int           `json:"stepNumber"`
	Description string        `json:"description"`
	Action      string        `json:"action"`
	Status      StepStatus    `json:"status"`
	Duration    time.Duration `json:"duration"`
	Error       error         `json:"error,omitempty"`
}

// Result represents the result of executing a scenario
type Result struct {
	Scenario    string        `json:"scenario"`
	Steps       []StepResult  `json:"steps"`
	TotalSteps  int           `json:"totalSteps"`
	PassedSteps int           `json:"passedSteps"`
	FailedSteps int           `json:"failedSteps"`
	Duration    time.Duration `json:"duration"`
	Success     bool          `json:"success"`
	Error       error         `json:"error,omitempty"`
}

// ActionDefinition defines an action that can be executed by the scenario engine
// This is re-exported from the actions package
type ActionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  []ParameterDef `json:"parameters"`
}

// AssertionDefinition defines an assertion that can be checked by the scenario engine
// This is re-exported from the actions package
type AssertionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  []ParameterDef `json:"parameters"`
}

// ParameterDef defines a parameter for an action or assertion
// This is re-exported from the actions package
type ParameterDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "number", "boolean", "duration"
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
}

// ScenarioContext contains context information for scenario execution
type ScenarioContext struct {
	AvailableActions    []ActionDefinition     `json:"availableActions"`
	AvailableAssertions []AssertionDefinition  `json:"availableAssertions"`
	AgentState          map[string]interface{} `json:"agentState,omitempty"`
	History             []StepResult           `json:"history,omitempty"`
}
