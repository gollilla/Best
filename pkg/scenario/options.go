package scenario

import (
	"time"

	"github.com/gollilla/best/pkg/config"
)

// WithTimeout sets the overall timeout for scenario execution
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithStepTimeout sets the timeout for each individual step
func WithStepTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.StepTimeout = timeout
	}
}

// WithVerbose enables verbose logging
func WithVerbose(verbose bool) Option {
	return func(o *Options) {
		o.Verbose = verbose
	}
}

// WithOnStepStart sets a callback for step start events
func WithOnStepStart(fn func(stepNum int, step ScenarioStep)) Option {
	return func(o *Options) {
		o.OnStepStart = fn
	}
}

// WithOnStepEnd sets a callback for step end events
func WithOnStepEnd(fn func(stepNum int, result StepResult)) Option {
	return func(o *Options) {
		o.OnStepEnd = fn
	}
}

// WithWebhook sets the webhook configuration for notifications
func WithWebhook(cfg *config.WebhookConfig) Option {
	return func(o *Options) {
		o.WebhookConfig = cfg
	}
}
