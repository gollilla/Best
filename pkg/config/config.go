package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration for Best testing framework
type Config struct {
	Server ServerConfig `yaml:"server"`
	Agent  AgentConfig  `yaml:"agent"`
}

// ServerConfig contains server connection settings
type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Version string `yaml:"version,omitempty"`
}

// AgentConfig contains agent settings
type AgentConfig struct {
	Username      string `yaml:"username"`
	Offline       bool   `yaml:"offline,omitempty"`
	Timeout       int    `yaml:"timeout,omitempty"` // in seconds
	CommandPrefix string `yaml:"commandPrefix,omitempty"`
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
			Username:      "TestBot",
			Offline:       false,
			Timeout:       30,
			CommandPrefix: "/",
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
