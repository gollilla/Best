package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/gollilla/best/pkg/config"
	"github.com/liushuangls/go-anthropic/v2"
)

// AnthropicProvider implements the Provider interface using Anthropic API
type AnthropicProvider struct {
	BaseProvider
	client *anthropic.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(cfg *config.AIConfig) (*AnthropicProvider, error) {
	apiKey := os.ExpandEnv(cfg.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required (set apiKey in config or ANTHROPIC_API_KEY environment variable)")
	}

	client := anthropic.NewClient(apiKey)

	model := cfg.Model
	if model == "" {
		model = "claude-3-sonnet-20240229"
	}

	return &AnthropicProvider{
		BaseProvider: newBaseProvider(cfg),
		client:       client,
	}, nil
}

// ParseScenario implements Provider.ParseScenario
func (p *AnthropicProvider) ParseScenario(ctx context.Context, scenarioText string, sctx *ScenarioContext) (*ParseResponse, error) {
	systemPrompt, err := BuildSystemPrompt(sctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build system prompt: %w", err)
	}

	userPrompt, err := BuildUserPrompt(scenarioText)
	if err != nil {
		return nil, fmt.Errorf("failed to build user prompt: %w", err)
	}

	temperature := float32(p.temperature)
	resp, err := p.client.CreateMessages(ctx, anthropic.MessagesRequest{
		Model:  anthropic.Model(p.model),
		System: systemPrompt,
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewTextMessageContent(userPrompt),
				},
			},
		},
		Temperature: &temperature,
		MaxTokens:   p.maxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("Anthropic API error: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("no response from Anthropic")
	}

	// Extract text content from the response
	var content string
	for _, block := range resp.Content {
		if block.Type == "text" && block.Text != nil {
			content = *block.Text
			break
		}
	}

	if content == "" {
		return nil, fmt.Errorf("no text content in Anthropic response")
	}

	steps, err := ExtractJSONFromResponse(content)
	if err != nil {
		return &ParseResponse{
			Error: fmt.Sprintf("failed to parse LLM response: %v\nResponse: %s", err, content),
		}, nil
	}

	return &ParseResponse{
		Steps: steps,
	}, nil
}

// ValidateStep implements Provider.ValidateStep
func (p *AnthropicProvider) ValidateStep(ctx context.Context, step *StepResult, sctx *ScenarioContext) (*ValidationResponse, error) {
	return &ValidationResponse{
		Valid:   step.Status == "passed",
		Message: fmt.Sprintf("Step %d: %s", step.StepNumber, step.Status),
	}, nil
}

// GenerateSummary implements Provider.GenerateSummary
func (p *AnthropicProvider) GenerateSummary(ctx context.Context, results *SummaryInput) (string, error) {
	prompt := BuildSummaryPrompt(results)

	temperature := float32(p.temperature)
	resp, err := p.client.CreateMessages(ctx, anthropic.MessagesRequest{
		Model: anthropic.Model(p.model),
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewTextMessageContent(prompt),
				},
			},
		},
		Temperature: &temperature,
		MaxTokens:   p.maxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("Anthropic API error: %w", err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("no response from Anthropic")
	}

	for _, block := range resp.Content {
		if block.Type == "text" && block.Text != nil {
			return *block.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in response")
}

// Close implements Provider.Close
func (p *AnthropicProvider) Close() error {
	// Anthropic client doesn't need explicit cleanup
	return nil
}
