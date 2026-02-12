package llm

import (
	"context"
	"fmt"

	"github.com/gollilla/best/pkg/config"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// ParseScenario parses a natural language scenario into executable steps
	ParseScenario(ctx context.Context, scenarioText string, sctx *ScenarioContext) (*ParseResponse, error)

	// ValidateStep validates the result of a step execution
	ValidateStep(ctx context.Context, step *StepResult, sctx *ScenarioContext) (*ValidationResponse, error)

	// GenerateSummary generates a natural language summary from test results
	GenerateSummary(ctx context.Context, results *SummaryInput) (string, error)

	// Close cleans up any resources used by the provider
	Close() error
}

// NewProvider creates a new LLM provider based on the configuration
func NewProvider(cfg *config.AIConfig) (Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("AI configuration is nil")
	}

	switch cfg.Provider {
	case "openai":
		return NewOpenAIProvider(cfg)
	case "anthropic":
		return NewAnthropicProvider(cfg)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s (supported: openai, anthropic)", cfg.Provider)
	}
}

// BaseProvider provides common functionality for LLM providers
type BaseProvider struct {
	model       string
	temperature float64
	maxTokens   int
}

// newBaseProvider creates a new base provider with common settings
func newBaseProvider(cfg *config.AIConfig) BaseProvider {
	temperature := cfg.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	return BaseProvider{
		model:       cfg.Model,
		temperature: temperature,
		maxTokens:   maxTokens,
	}
}
