package scenario

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gollilla/best/pkg/agent"
	"github.com/gollilla/best/pkg/scenario/actions"
)

// Executor executes scenario steps
type Executor struct {
	agent    *agent.Agent
	registry *actions.Registry
	options  ExecutorOptions
}

// ExecutorOptions contains options for the executor
type ExecutorOptions struct {
	Timeout     time.Duration
	StepTimeout time.Duration
	Verbose     bool
	OnStepStart func(stepNum int, step ScenarioStep)
	OnStepEnd   func(stepNum int, result StepResult)
}

// DefaultExecutorOptions returns default executor options
func DefaultExecutorOptions() ExecutorOptions {
	return ExecutorOptions{
		Timeout:     5 * time.Minute,
		StepTimeout: 30 * time.Second,
		Verbose:     false,
	}
}

// NewExecutor creates a new scenario executor
func NewExecutor(agent *agent.Agent, opts ...func(*ExecutorOptions)) *Executor {
	options := DefaultExecutorOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Executor{
		agent:    agent,
		registry: actions.NewRegistry(),
		options:  options,
	}
}

// Execute executes a list of scenario steps
func (e *Executor) Execute(ctx context.Context, steps []ScenarioStep) (*Result, error) {
	result := &Result{
		Steps:      make([]StepResult, 0, len(steps)),
		TotalSteps: len(steps),
	}

	startTime := time.Now()

	// Create timeout context for the entire execution
	execCtx, cancel := context.WithTimeout(ctx, e.options.Timeout)
	defer cancel()

	for i, step := range steps {
		stepNum := i + 1

		// Notify step start
		if e.options.OnStepStart != nil {
			e.options.OnStepStart(stepNum, step)
		}

		stepResult := e.executeStep(execCtx, stepNum, step)
		result.Steps = append(result.Steps, stepResult)

		// Notify step end
		if e.options.OnStepEnd != nil {
			e.options.OnStepEnd(stepNum, stepResult)
		}

		if stepResult.Status == StepStatusFailed {
			result.FailedSteps++
			result.Success = false
			result.Error = stepResult.Error
			break
		}

		result.PassedSteps++

		// Check if context was cancelled
		if execCtx.Err() != nil {
			result.Error = execCtx.Err()
			break
		}
	}

	result.Duration = time.Since(startTime)
	result.Success = result.FailedSteps == 0 && result.Error == nil

	return result, nil
}

// executeStep executes a single scenario step
func (e *Executor) executeStep(ctx context.Context, stepNum int, step ScenarioStep) StepResult {
	startTime := time.Now()

	result := StepResult{
		StepNumber:  stepNum,
		Description: step.Description,
		Action:      step.Action,
		Status:      StepStatusRunning,
	}

	// Create timeout context for this step
	stepCtx, cancel := context.WithTimeout(ctx, e.options.StepTimeout)
	defer cancel()

	// Execute the action or assertion
	var err error
	if e.isAssertion(step.Action) {
		err = e.executeAssertion(stepCtx, step)
	} else {
		err = e.executeAction(stepCtx, step)
	}

	result.Duration = time.Since(startTime)

	if err != nil {
		result.Status = StepStatusFailed
		result.Error = err
	} else {
		result.Status = StepStatusPassed
	}

	return result
}

// isAssertion checks if the action name is an assertion
func (e *Executor) isAssertion(name string) bool {
	return strings.HasPrefix(name, "assert_") || e.registry.IsAssertion(name)
}

// executeAction executes an action
func (e *Executor) executeAction(ctx context.Context, step ScenarioStep) error {
	// Recover from panics (assertions might panic)
	var panicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					panicErr = err
				} else {
					panicErr = fmt.Errorf("%v", r)
				}
			}
		}()

		panicErr = e.registry.ExecuteAction(ctx, e.agent, step.Action, step.Params)
	}()

	return panicErr
}

// executeAssertion executes an assertion
func (e *Executor) executeAssertion(ctx context.Context, step ScenarioStep) error {
	// Recover from panics (assertions might panic)
	var panicErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					panicErr = err
				} else {
					panicErr = fmt.Errorf("%v", r)
				}
			}
		}()

		panicErr = e.registry.ExecuteAssertion(ctx, e.agent, step.Action, step.Params)
	}()

	return panicErr
}

// GetRegistry returns the action/assertion registry
func (e *Executor) GetRegistry() *actions.Registry {
	return e.registry
}

// GetScenarioContext builds a ScenarioContext for LLM parsing
func (e *Executor) GetScenarioContext() *ScenarioContext {
	// Convert actions.ActionDefinition to scenario.ActionDefinition
	actionDefs := e.registry.GetActionDefinitions()
	scenarioActions := make([]ActionDefinition, len(actionDefs))
	for i, def := range actionDefs {
		params := make([]ParameterDef, len(def.Parameters))
		for j, p := range def.Parameters {
			params[j] = ParameterDef{
				Name:        p.Name,
				Type:        p.Type,
				Required:    p.Required,
				Description: p.Description,
				Default:     p.Default,
			}
		}
		scenarioActions[i] = ActionDefinition{
			Name:        def.Name,
			Description: def.Description,
			Parameters:  params,
		}
	}

	// Convert actions.AssertionDefinition to scenario.AssertionDefinition
	assertionDefs := e.registry.GetAssertionDefinitions()
	scenarioAssertions := make([]AssertionDefinition, len(assertionDefs))
	for i, def := range assertionDefs {
		params := make([]ParameterDef, len(def.Parameters))
		for j, p := range def.Parameters {
			params[j] = ParameterDef{
				Name:        p.Name,
				Type:        p.Type,
				Required:    p.Required,
				Description: p.Description,
				Default:     p.Default,
			}
		}
		scenarioAssertions[i] = AssertionDefinition{
			Name:        def.Name,
			Description: def.Description,
			Parameters:  params,
		}
	}

	return &ScenarioContext{
		AvailableActions:    scenarioActions,
		AvailableAssertions: scenarioAssertions,
	}
}
