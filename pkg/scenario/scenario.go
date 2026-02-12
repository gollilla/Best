package scenario

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gollilla/best/pkg/agent"
	"github.com/gollilla/best/pkg/config"
	"github.com/gollilla/best/pkg/scenario/llm"
	"github.com/gollilla/best/pkg/webhook"
)

// Runner is the main scenario runner
type Runner struct {
	agent    *agent.Agent
	provider llm.Provider
	executor *Executor
	options  Options
	webhook  *webhook.Client
}

// NewRunner creates a new scenario runner
func NewRunner(agent *agent.Agent, cfg *config.AIConfig, opts ...Option) (*Runner, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Expand environment variables in config
	config.ExpandEnvInConfig(&config.Config{AI: *cfg})

	provider, err := llm.NewProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM provider: %w", err)
	}

	executor := NewExecutor(agent, func(o *ExecutorOptions) {
		o.Timeout = options.Timeout
		o.StepTimeout = options.StepTimeout
		o.Verbose = options.Verbose
		o.OnStepStart = options.OnStepStart
		o.OnStepEnd = options.OnStepEnd
	})

	// Initialize webhook client if configured
	var webhookClient *webhook.Client
	if options.WebhookConfig != nil {
		webhookClient = webhook.NewClient(options.WebhookConfig)
	}

	return &Runner{
		agent:    agent,
		provider: provider,
		executor: executor,
		options:  options,
		webhook:  webhookClient,
	}, nil
}

// RunFromString executes a scenario from a string
func (r *Runner) RunFromString(ctx context.Context, scenarioText string) (*Result, error) {
	return r.run(ctx, scenarioText)
}

// RunFromFile executes a scenario from a file
func (r *Runner) RunFromFile(ctx context.Context, path string) (*Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	result, err := r.run(ctx, string(data))
	if result != nil {
		result.Scenario = path
	}
	return result, err
}

// run executes a scenario
func (r *Runner) run(ctx context.Context, scenarioText string) (*Result, error) {
	// Build scenario context for LLM
	sctx := r.executor.GetScenarioContext()

	// Convert to LLM context
	llmCtx := convertToLLMContext(sctx)

	// Parse scenario using LLM
	if r.options.Verbose {
		fmt.Println("Parsing scenario with LLM...")
	}

	parseResp, err := r.provider.ParseScenario(ctx, scenarioText, llmCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scenario: %w", err)
	}

	if parseResp.Error != "" {
		return nil, fmt.Errorf("LLM parsing error: %s", parseResp.Error)
	}

	if len(parseResp.Steps) == 0 {
		return nil, fmt.Errorf("no steps parsed from scenario")
	}

	if r.options.Verbose {
		fmt.Printf("Parsed %d steps from scenario\n", len(parseResp.Steps))
		for i, step := range parseResp.Steps {
			fmt.Printf("  %d. %s: %s\n", i+1, step.Action, step.Description)
		}
	}

	// Convert LLM steps to scenario steps
	steps := convertFromLLMSteps(parseResp.Steps)

	// Execute parsed steps
	result, err := r.executor.Execute(ctx, steps)

	// Send webhook notification if configured
	if r.webhook != nil && r.webhook.IsEnabled() && result != nil {
		webhookResult := convertToWebhookResult(result)
		if webhookErr := r.webhook.NotifyScenarioResult(ctx, webhookResult); webhookErr != nil {
			if r.options.Verbose {
				fmt.Printf("Warning: webhook notification failed: %v\n", webhookErr)
			}
		}
	}

	return result, err
}

// convertToWebhookResult converts scenario Result to webhook ScenarioResult
func convertToWebhookResult(r *Result) *webhook.ScenarioResult {
	steps := make([]webhook.StepResult, len(r.Steps))
	for i, s := range r.Steps {
		var status webhook.StepStatus
		if s.Status == StepStatusPassed {
			status = webhook.StepStatusPassed
		} else if s.Status == StepStatusFailed {
			status = webhook.StepStatusFailed
		}
		steps[i] = webhook.StepResult{
			StepNumber:  s.StepNumber,
			Description: s.Description,
			Status:      status,
			Error:       s.Error,
		}
	}

	return &webhook.ScenarioResult{
		Scenario:    r.Scenario,
		Steps:       steps,
		TotalSteps:  r.TotalSteps,
		PassedSteps: r.PassedSteps,
		FailedSteps: r.FailedSteps,
		Duration:    r.Duration,
		Success:     r.Success,
	}
}

// convertToLLMContext converts scenario context to LLM context
func convertToLLMContext(sctx *ScenarioContext) *llm.ScenarioContext {
	// Convert actions
	llmActions := make([]llm.ActionDefinition, len(sctx.AvailableActions))
	for i, a := range sctx.AvailableActions {
		params := make([]llm.ParameterDef, len(a.Parameters))
		for j, p := range a.Parameters {
			params[j] = llm.ParameterDef{
				Name:        p.Name,
				Type:        p.Type,
				Required:    p.Required,
				Description: p.Description,
				Default:     p.Default,
			}
		}
		llmActions[i] = llm.ActionDefinition{
			Name:        a.Name,
			Description: a.Description,
			Parameters:  params,
		}
	}

	// Convert assertions
	llmAssertions := make([]llm.AssertionDefinition, len(sctx.AvailableAssertions))
	for i, a := range sctx.AvailableAssertions {
		params := make([]llm.ParameterDef, len(a.Parameters))
		for j, p := range a.Parameters {
			params[j] = llm.ParameterDef{
				Name:        p.Name,
				Type:        p.Type,
				Required:    p.Required,
				Description: p.Description,
				Default:     p.Default,
			}
		}
		llmAssertions[i] = llm.AssertionDefinition{
			Name:        a.Name,
			Description: a.Description,
			Parameters:  params,
		}
	}

	return &llm.ScenarioContext{
		AvailableActions:    llmActions,
		AvailableAssertions: llmAssertions,
	}
}

// convertFromLLMSteps converts LLM steps to scenario steps
func convertFromLLMSteps(llmSteps []llm.ScenarioStep) []ScenarioStep {
	steps := make([]ScenarioStep, len(llmSteps))
	for i, s := range llmSteps {
		steps[i] = ScenarioStep{
			Action:      s.Action,
			Description: s.Description,
			Params:      s.Params,
		}
	}
	return steps
}

// RunMultipleFromFiles executes multiple scenarios from files and returns a summary
func (r *Runner) RunMultipleFromFiles(ctx context.Context, paths []string) (*Summary, error) {
	results := make([]*Result, 0, len(paths))

	for _, path := range paths {
		if r.options.Verbose {
			fmt.Printf("\n=== Running Scenario: %s ===\n", path)
		}

		result, err := r.RunFromFile(ctx, path)
		if err != nil {
			// Create a failed result for this scenario
			result = &Result{
				Scenario: path,
				Success:  false,
				Error:    err,
			}
		}
		results = append(results, result)
	}

	summary := NewSummary(results...)

	// Send webhook summary notification if configured
	if r.webhook != nil && r.webhook.IsEnabled() {
		webhookSummary := convertToWebhookSummary(summary)
		if webhookErr := r.webhook.NotifySummary(ctx, webhookSummary); webhookErr != nil {
			if r.options.Verbose {
				fmt.Printf("Warning: webhook summary notification failed: %v\n", webhookErr)
			}
		}
	}

	return summary, nil
}

// convertToWebhookSummary converts scenario Summary to webhook Summary
func convertToWebhookSummary(s *Summary) *webhook.Summary {
	results := make([]*webhook.ScenarioResult, len(s.Results))
	for i, r := range s.Results {
		results[i] = convertToWebhookResult(r)
	}

	return &webhook.Summary{
		Results:        results,
		TotalScenarios: s.TotalScenarios,
		PassedCount:    s.PassedCount,
		FailedCount:    s.FailedCount,
		TotalSteps:     s.TotalSteps,
		PassedSteps:    s.PassedSteps,
		FailedSteps:    s.FailedSteps,
		TotalDuration:  s.TotalDuration,
	}
}

// GenerateSummary generates a natural language summary using LLM
func (r *Runner) GenerateSummary(ctx context.Context, summary *Summary) (string, error) {
	input := convertToLLMSummaryInput(summary)
	return r.provider.GenerateSummary(ctx, input)
}

func convertToLLMSummaryInput(s *Summary) *llm.SummaryInput {
	scenarios := make([]llm.ScenarioResultInput, len(s.Results))
	for i, r := range s.Results {
		steps := make([]llm.StepResultInput, len(r.Steps))
		for j, step := range r.Steps {
			errStr := ""
			if step.Error != nil {
				errStr = step.Error.Error()
			}
			steps[j] = llm.StepResultInput{
				Number:      step.StepNumber,
				Description: step.Description,
				Status:      string(step.Status),
				Error:       errStr,
			}
		}
		errStr := ""
		if r.Error != nil {
			errStr = r.Error.Error()
		}
		scenarios[i] = llm.ScenarioResultInput{
			Name:        r.Scenario,
			Success:     r.Success,
			TotalSteps:  r.TotalSteps,
			PassedSteps: r.PassedSteps,
			FailedSteps: r.FailedSteps,
			Duration:    r.Duration.String(),
			Steps:       steps,
			Error:       errStr,
		}
	}
	return &llm.SummaryInput{Scenarios: scenarios}
}

// Close cleans up resources
func (r *Runner) Close() error {
	return r.provider.Close()
}

// RunFromString is a convenience function to run a scenario from a string
// It uses the global configuration for AI settings
func RunFromString(scenarioText string, agent *agent.Agent, opts ...Option) (*Result, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Add webhook config if present
	if cfg.Webhook.URL != "" {
		opts = append(opts, WithWebhook(&cfg.Webhook))
	}

	runner, err := NewRunner(agent, &cfg.AI, opts...)
	if err != nil {
		return nil, err
	}
	defer runner.Close()

	ctx := context.Background()
	return runner.RunFromString(ctx, scenarioText)
}

// RunFromFile is a convenience function to run a scenario from a file
// It uses the global configuration for AI settings
func RunFromFile(path string, agent *agent.Agent, opts ...Option) (*Result, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Add webhook config if present
	if cfg.Webhook.URL != "" {
		opts = append(opts, WithWebhook(&cfg.Webhook))
	}

	runner, err := NewRunner(agent, &cfg.AI, opts...)
	if err != nil {
		return nil, err
	}
	defer runner.Close()

	ctx := context.Background()
	return runner.RunFromFile(ctx, path)
}

// RunFromStringWithConfig runs a scenario from a string with explicit configuration
func RunFromStringWithConfig(scenarioText string, agent *agent.Agent, cfg *config.Config, opts ...Option) (*Result, error) {
	// Add webhook config if present
	if cfg.Webhook.URL != "" {
		opts = append(opts, WithWebhook(&cfg.Webhook))
	}

	runner, err := NewRunner(agent, &cfg.AI, opts...)
	if err != nil {
		return nil, err
	}
	defer runner.Close()

	ctx := context.Background()
	return runner.RunFromString(ctx, scenarioText)
}

// RunFromFileWithConfig runs a scenario from a file with explicit configuration
func RunFromFileWithConfig(path string, agent *agent.Agent, cfg *config.Config, opts ...Option) (*Result, error) {
	// Add webhook config if present
	if cfg.Webhook.URL != "" {
		opts = append(opts, WithWebhook(&cfg.Webhook))
	}

	runner, err := NewRunner(agent, &cfg.AI, opts...)
	if err != nil {
		return nil, err
	}
	defer runner.Close()

	ctx := context.Background()
	return runner.RunFromFile(ctx, path)
}

// Options configuration
type Options struct {
	Timeout       time.Duration
	StepTimeout   time.Duration
	Verbose       bool
	OnStepStart   func(stepNum int, step ScenarioStep)
	OnStepEnd     func(stepNum int, result StepResult)
	WebhookConfig *config.WebhookConfig
}

// DefaultOptions returns default options
func DefaultOptions() Options {
	return Options{
		Timeout:     5 * time.Minute,
		StepTimeout: 30 * time.Second,
		Verbose:     false,
	}
}

// Option is a function that modifies Options
type Option func(*Options)
