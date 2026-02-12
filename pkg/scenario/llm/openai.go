package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/gollilla/best/pkg/config"
	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface using OpenAI API
type OpenAIProvider struct {
	BaseProvider
	client *openai.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(cfg *config.AIConfig) (*OpenAIProvider, error) {
	apiKey := os.ExpandEnv(cfg.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required (set apiKey in config or OPENAI_API_KEY environment variable)")
	}

	client := openai.NewClient(apiKey)

	model := cfg.Model
	if model == "" {
		model = "gpt-4"
	}

	return &OpenAIProvider{
		BaseProvider: newBaseProvider(cfg),
		client:       client,
	}, nil
}

// ParseScenario implements Provider.ParseScenario
func (p *OpenAIProvider) ParseScenario(ctx context.Context, scenarioText string, sctx *ScenarioContext) (*ParseResponse, error) {
	systemPrompt, err := BuildSystemPrompt(sctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build system prompt: %w", err)
	}

	userPrompt, err := BuildUserPrompt(scenarioText)
	if err != nil {
		return nil, fmt.Errorf("failed to build user prompt: %w", err)
	}

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: float32(p.temperature),
		MaxTokens:   p.maxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content := resp.Choices[0].Message.Content

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
func (p *OpenAIProvider) ValidateStep(ctx context.Context, step *StepResult, sctx *ScenarioContext) (*ValidationResponse, error) {
	return &ValidationResponse{
		Valid:   step.Status == "passed",
		Message: fmt.Sprintf("Step %d: %s", step.StepNumber, step.Status),
	}, nil
}

// GenerateSummary implements Provider.GenerateSummary
func (p *OpenAIProvider) GenerateSummary(ctx context.Context, results *SummaryInput) (string, error) {
	prompt := BuildSummaryPrompt(results)

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: float32(p.temperature),
		MaxTokens:   p.maxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// Close implements Provider.Close
func (p *OpenAIProvider) Close() error {
	// OpenAI client doesn't need explicit cleanup
	return nil
}
