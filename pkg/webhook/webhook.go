// Package webhook provides webhook notification functionality for Best
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/gollilla/best/pkg/config"
)

// EventType represents a webhook event type
type EventType string

const (
	EventScenarioComplete EventType = "scenario_complete"
	EventScenarioFailed   EventType = "scenario_failed"
	EventStepFailed       EventType = "step_failed"
	EventSummary          EventType = "summary"
)

// StepStatus represents the status of a step (mirrors scenario.StepStatus)
type StepStatus string

const (
	StepStatusPassed StepStatus = "passed"
	StepStatusFailed StepStatus = "failed"
)

// ScenarioResult contains scenario execution results for webhook notifications
type ScenarioResult struct {
	Scenario    string
	Steps       []StepResult
	TotalSteps  int
	PassedSteps int
	FailedSteps int
	Duration    time.Duration
	Success     bool
}

// StepResult contains step execution result for webhook notifications
type StepResult struct {
	StepNumber  int
	Description string
	Status      StepStatus
	Error       error
}

// Summary contains summary of multiple scenario executions
type Summary struct {
	Results        []*ScenarioResult
	TotalScenarios int
	PassedCount    int
	FailedCount    int
	TotalSteps     int
	PassedSteps    int
	FailedSteps    int
	TotalDuration  time.Duration
}

// NewSummary creates a new summary from scenario results
func NewSummary(results ...*ScenarioResult) *Summary {
	s := &Summary{
		Results:        results,
		TotalScenarios: len(results),
	}

	for _, r := range results {
		if r.Success {
			s.PassedCount++
		} else {
			s.FailedCount++
		}
		s.TotalSteps += r.TotalSteps
		s.PassedSteps += r.PassedSteps
		s.FailedSteps += r.FailedSteps
		s.TotalDuration += r.Duration
	}

	return s
}

// Success returns true if all scenarios passed
func (s *Summary) Success() bool {
	return s.FailedCount == 0
}

// Client is a webhook client
type Client struct {
	config     *config.WebhookConfig
	httpClient *http.Client
}

// NewClient creates a new webhook client
func NewClient(cfg *config.WebhookConfig) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsEnabled returns true if webhook is configured
func (c *Client) IsEnabled() bool {
	return c.config != nil && c.config.URL != ""
}

// ShouldNotify returns true if the given event type should trigger a notification
func (c *Client) ShouldNotify(event EventType) bool {
	if !c.IsEnabled() {
		return false
	}
	if len(c.config.Events) == 0 {
		// Default: notify on all events
		return true
	}
	return slices.Contains(c.config.Events, string(event))
}

// NotifyScenarioResult sends a webhook notification for scenario results
func (c *Client) NotifyScenarioResult(ctx context.Context, result *ScenarioResult) error {
	if !c.IsEnabled() {
		return nil
	}

	// Determine event type
	eventType := EventScenarioComplete
	if !result.Success {
		eventType = EventScenarioFailed
	}

	if !c.ShouldNotify(eventType) {
		return nil
	}

	// Build Discord embed
	embed := c.buildResultEmbed(result)
	payload := DiscordWebhookPayload{
		Embeds: []DiscordEmbed{embed},
	}

	return c.send(ctx, payload)
}

// NotifyStepFailed sends a webhook notification for a failed step
func (c *Client) NotifyStepFailed(ctx context.Context, scenarioName string, step *StepResult) error {
	if !c.IsEnabled() || !c.ShouldNotify(EventStepFailed) {
		return nil
	}

	embed := DiscordEmbed{
		Title:       fmt.Sprintf("Step Failed: %s", scenarioName),
		Description: fmt.Sprintf("**Step %d**: %s\n**Error**: %v", step.StepNumber, step.Description, step.Error),
		Color:       ColorRed,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}

	payload := DiscordWebhookPayload{
		Embeds: []DiscordEmbed{embed},
	}

	return c.send(ctx, payload)
}

// NotifySummary sends a webhook notification with test summary
func (c *Client) NotifySummary(ctx context.Context, summary *Summary) error {
	if !c.IsEnabled() || !c.ShouldNotify(EventSummary) {
		return nil
	}

	embed := c.buildSummaryEmbed(summary)
	payload := DiscordWebhookPayload{
		Embeds: []DiscordEmbed{embed},
	}

	return c.send(ctx, payload)
}

func (c *Client) buildSummaryEmbed(summary *Summary) DiscordEmbed {
	color := ColorGreen
	status := "All Passed"
	if !summary.Success() {
		color = ColorRed
		status = "Some Failed"
	}

	description := fmt.Sprintf(
		"**Status**: %s\n**Scenarios**: %d/%d passed\n**Steps**: %d/%d passed\n**Duration**: %v",
		status,
		summary.PassedCount,
		summary.TotalScenarios,
		summary.PassedSteps,
		summary.TotalSteps,
		summary.TotalDuration.Round(time.Millisecond),
	)

	// Add scenario results
	description += "\n\n**Scenarios**:"
	for _, r := range summary.Results {
		icon := "✅"
		if !r.Success {
			icon = "❌"
		}
		description += fmt.Sprintf("\n%s %s (%d/%d steps)", icon, r.Scenario, r.PassedSteps, r.TotalSteps)
	}

	// Add failed scenario details
	var failedDetails string
	for _, r := range summary.Results {
		if !r.Success {
			for _, step := range r.Steps {
				if step.Status == StepStatusFailed {
					failedDetails += fmt.Sprintf("\n- **%s** Step %d: %s", r.Scenario, step.StepNumber, step.Description)
					if step.Error != nil {
						errStr := fmt.Sprintf("%v", step.Error)
						if len(errStr) > 50 {
							errStr = errStr[:50] + "..."
						}
						failedDetails += fmt.Sprintf(" (`%s`)", errStr)
					}
				}
			}
		}
	}
	if failedDetails != "" {
		description += "\n\n**Failed Steps**:" + failedDetails
	}

	return DiscordEmbed{
		Title:       "Test Summary",
		Description: description,
		Color:       color,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &DiscordEmbedFooter{
			Text: "Best - Minecraft Bedrock Testing",
		},
	}
}

func (c *Client) buildResultEmbed(result *ScenarioResult) DiscordEmbed {
	color := ColorGreen
	status := "Passed"
	if !result.Success {
		color = ColorRed
		status = "Failed"
	}

	description := fmt.Sprintf(
		"**Status**: %s\n**Steps**: %d/%d passed\n**Duration**: %v",
		status,
		result.PassedSteps,
		result.TotalSteps,
		result.Duration.Round(time.Millisecond),
	)

	// Add failed steps detail
	if result.FailedSteps > 0 {
		description += "\n\n**Failed Steps**:"
		for _, step := range result.Steps {
			if step.Status == StepStatusFailed {
				description += fmt.Sprintf("\n- Step %d: %s", step.StepNumber, step.Description)
				if step.Error != nil {
					description += fmt.Sprintf(" (`%v`)", step.Error)
				}
			}
		}
	}

	return DiscordEmbed{
		Title:       fmt.Sprintf("Scenario: %s", result.Scenario),
		Description: description,
		Color:       color,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Footer: &DiscordEmbedFooter{
			Text: "Best - Minecraft Bedrock Testing",
		},
	}
}

func (c *Client) send(ctx context.Context, payload DiscordWebhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
