package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration for Best testing framework
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Agent   AgentConfig   `yaml:"agent"`
	AI      AIConfig      `yaml:"ai,omitempty"`
	Webhook WebhookConfig `yaml:"webhook,omitempty"`
}

// WebhookConfig contains webhook notification settings
type WebhookConfig struct {
	URL    string   `yaml:"url"`              // Webhook URL (supports ${ENV_VAR} syntax)
	Events []string `yaml:"events,omitempty"` // Events to notify: "scenario_complete", "scenario_failed", "step_failed"
}

// ServerConfig contains server connection settings
type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Version string `yaml:"version,omitempty"`
}

// AgentConfig contains agent settings
type AgentConfig struct {
	Username          string `yaml:"username"`
	Timeout           int    `yaml:"timeout,omitempty"`           // in seconds
	CommandPrefix     string `yaml:"commandPrefix,omitempty"`
	CommandSendMethod string `yaml:"commandSendMethod,omitempty"` // "text" or "request"
	CommandTimeout    int    `yaml:"commandTimeout,omitempty"`    // assertion wait timeout in seconds
}

// AIConfig contains AI/LLM settings for scenario execution
type AIConfig struct {
	Provider    string         `yaml:"provider"`              // "openai" or "anthropic"
	APIKey      string         `yaml:"apiKey"`                // API key (supports ${ENV_VAR} syntax)
	Model       string         `yaml:"model"`                 // Model name (e.g., "gpt-4", "claude-3-sonnet")
	Temperature float64        `yaml:"temperature,omitempty"` // Creativity (0.0-1.0)
	MaxTokens   int            `yaml:"maxTokens,omitempty"`   // Maximum tokens
	Timeout     int            `yaml:"timeout,omitempty"`     // API timeout in seconds
	Retries     int            `yaml:"retries,omitempty"`     // Number of retries
	Scenario    ScenarioConfig `yaml:"scenario,omitempty"`    // Scenario-specific settings
}

// ScenarioConfig contains scenario execution settings
type ScenarioConfig struct {
	Verbose     bool `yaml:"verbose,omitempty"`     // Enable verbose logging
	StepTimeout int  `yaml:"stepTimeout,omitempty"` // Step execution timeout in seconds
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:    "localhost",
			Port:    19132,
			Version: "",
		},
		Agent: AgentConfig{
			Username:          "TestBot",
			Timeout:           30,
			CommandPrefix:     "/",
			CommandSendMethod: "text",
			CommandTimeout:    5,
		},
		AI: DefaultAIConfig(),
	}
}

// DefaultAIConfig returns default AI configuration
func DefaultAIConfig() AIConfig {
	return AIConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   4096,
		Timeout:     60,
		Retries:     3,
		Scenario: ScenarioConfig{
			Verbose:     false,
			StepTimeout: 30,
		},
	}
}

// LoadConfig loads configuration from a file
// It searches for best.config.yml or best.config.yaml in the current directory and parent directories
func LoadConfig() (*Config, error) {
	// Try to find config file
	configPath, err := findConfigFile()
	if err != nil {
		// Return default config if no config file found
		return DefaultConfig(), nil
	}

	return LoadConfigFromFile(configPath)
}

// LoadConfigFromFile loads configuration from the specified file
func LoadConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// findConfigFile searches for best.config.yml or best.config.yaml
// It starts from the current directory and walks up to parent directories
func findConfigFile() (string, error) {
	filenames := []string{"best.config.yml", "best.config.yaml"}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree
	for {
		for _, filename := range filenames {
			path := filepath.Join(dir, filename)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("config file not found")
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ExpandEnvInConfig expands environment variables in the configuration
// It supports ${VAR} and $VAR syntax
func ExpandEnvInConfig(config *Config) {
	config.AI.APIKey = os.ExpandEnv(config.AI.APIKey)
	config.Webhook.URL = os.ExpandEnv(config.Webhook.URL)
}
