// Package best provides a testing framework for Minecraft Bedrock Edition servers
package best

import (
	"sync"
	"time"

	"github.com/gollilla/best/pkg/agent"
	"github.com/gollilla/best/pkg/assertions"
	"github.com/gollilla/best/pkg/config"
	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/runner"
	"github.com/gollilla/best/pkg/types"
	"github.com/gollilla/best/pkg/world"
)

var (
	// globalConfig holds the automatically loaded configuration
	globalConfig     *Config
	globalConfigOnce sync.Once
)

// loadGlobalConfig loads the configuration file once (lazy loading)
func loadGlobalConfig() {
	globalConfigOnce.Do(func() {
		var err error
		globalConfig, err = LoadConfig()
		if err != nil {
			// Config file not found is not an error - use defaults
			globalConfig = DefaultConfig()
		}
	})
}

// GetConfig returns the global configuration (automatically loaded from best.config.yml)
func GetConfig() *Config {
	loadGlobalConfig()
	return globalConfig
}

// SetConfig allows users to override the global configuration
func SetConfig(cfg *Config) {
	globalConfig = cfg
}

// NewDefaultAgent creates a new agent using the global configuration
// The configuration is automatically loaded from best.config.yml
func NewDefaultAgent() *Agent {
	return NewAgentFromConfig(GetConfig())
}

// CreateAgent creates a new agent with the specified username
// It uses the global configuration for server connection settings
// and allows optional overrides via AgentOptions
//
// Example:
//
//	agent := best.CreateAgent("TestBot")  // Use config file settings
//	agent := best.CreateAgent("Player1", best.WithHost("example.com"))  // Override host
func CreateAgent(username string, options ...AgentOption) *Agent {
	cfg := GetConfig()

	// Start with configuration from file
	agentOptions := []AgentOption{
		WithHost(cfg.Server.Host),
		WithPort(uint16(cfg.Server.Port)),
		WithUsername(username), // Override username with provided value
	}

	// Add version if specified in config
	if cfg.Server.Version != "" {
		agentOptions = append(agentOptions, WithVersion(cfg.Server.Version))
	}

	// Add timeout if specified in config
	if cfg.Agent.Timeout > 0 {
		agentOptions = append(agentOptions, WithTimeout(time.Duration(cfg.Agent.Timeout)*time.Second))
	}

	// Add command prefix if specified in config
	if cfg.Agent.CommandPrefix != "" {
		agentOptions = append(agentOptions, WithCommandPrefix(cfg.Agent.CommandPrefix))
	}

	// Add command send method if specified in config
	if cfg.Agent.CommandSendMethod != "" {
		agentOptions = append(agentOptions, WithCommandSendMethod(cfg.Agent.CommandSendMethod))
	}

	// Append user-provided options (these will override config file settings)
	agentOptions = append(agentOptions, options...)

	return NewAgent(agentOptions...)
}

// Re-export main types and functions for convenience

// Agent types
type Agent = agent.Agent
type AgentOption = agent.AgentOption

var (
	NewAgent              = agent.NewAgent
	WithHost              = agent.WithHost
	WithPort              = agent.WithPort
	WithUsername          = agent.WithUsername
	WithTimeout           = agent.WithTimeout
	WithVersion           = agent.WithVersion
	WithCommandPrefix     = agent.WithCommandPrefix
	WithCommandSendMethod = agent.WithCommandSendMethod
)

// Event types
type EventName = events.EventName
type EventData = events.EventData
type Emitter = events.Emitter

const (
	// Phase 1 events
	EventJoin           = events.EventJoin
	EventSpawn          = events.EventSpawn
	EventDisconnect     = events.EventDisconnect
	EventError          = events.EventError
	EventChat           = events.EventChat
	EventPositionUpdate = events.EventPositionUpdate
	EventHealthUpdate   = events.EventHealthUpdate
	EventCommandOutput  = events.EventCommandOutput

	// Phase 2 events
	EventBlockUpdate         = events.EventBlockUpdate
	EventInventoryUpdate     = events.EventInventoryUpdate
	EventInventorySlotUpdate = events.EventInventorySlotUpdate
	EventEffectAdd           = events.EventEffectAdd
	EventEffectRemove        = events.EventEffectRemove
	EventEffectUpdate        = events.EventEffectUpdate
	EventEntityAdd           = events.EventEntityAdd
	EventEntityRemove        = events.EventEntityRemove

	// Phase 3 events (Player state)
	EventHungerUpdate     = events.EventHungerUpdate
	EventGamemodeUpdate   = events.EventGamemodeUpdate
	EventPermissionUpdate = events.EventPermissionUpdate
	EventTagUpdate        = events.EventTagUpdate

	// UI/Display events
	EventTitle       = events.EventTitle
	EventScoreUpdate = events.EventScoreUpdate
)

// Common types
type Position = types.Position
type Rotation = types.Rotation
type PlayerState = types.PlayerState
type CommandOutput = types.CommandOutput
type ChatMessage = types.ChatMessage
type ClientOptions = types.ClientOptions

// Phase 2 types
type Block = types.Block
type BlockUpdate = types.BlockUpdate
type InventoryItem = types.InventoryItem
type Effect = types.Effect
type Entity = types.Entity
type World = world.World
type BlockRegistry = world.BlockRegistry

// UI/Display types
type TitleDisplay = types.TitleDisplay
type ScoreboardEntry = types.ScoreboardEntry

// Phase 3: Assertion types
type AssertionContext = assertions.AssertionContext
type AssertionError = assertions.AssertionError
type PositionAssertion = assertions.PositionAssertion
type ChatAssertion = assertions.ChatAssertion
type CommandOutputAssertion = assertions.CommandOutputAssertion
type InventoryAssertion = assertions.InventoryAssertion
type ChatOptions = assertions.ChatOptions
type CommandOutputOptions = assertions.CommandOutputOptions

// Player state assertion types
type HealthAssertion = assertions.HealthAssertion
type HungerAssertion = assertions.HungerAssertion
type EffectAssertion = assertions.EffectAssertion
type GamemodeAssertion = assertions.GamemodeAssertion
type PermissionAssertion = assertions.PermissionAssertion
type TagAssertion = assertions.TagAssertion

// UI/Display assertion types
type TitleAssertion = assertions.TitleAssertion
type ScoreboardAssertion = assertions.ScoreboardAssertion

var (
	NewAssertionContext = assertions.NewAssertionContext
	NewAssertionError   = assertions.NewAssertionError
)

// Phase 4: Test Runner types
type TestRunner = runner.TestRunner
type TestContext = runner.TestContext
type TestFunction = runner.TestFunction
type HookFunction = runner.HookFunction
type TestCase = runner.TestCase
type TestSuite = runner.TestSuite
type TestError = runner.TestError
type TestCaseResult = runner.TestCaseResult
type SuiteResult = runner.SuiteResult
type TestResult = runner.TestResult
type TestRunnerOptions = runner.TestRunnerOptions
type Reporter = runner.Reporter
type ServerInfo = runner.ServerInfo

var (
	NewTestRunner      = runner.NewTestRunner
	NewConsoleReporter = runner.NewConsoleReporter
)

// Config types
type Config = config.Config
type ServerConfig = config.ServerConfig
type AgentConfig = config.AgentConfig

var (
	LoadConfig         = config.LoadConfig
	LoadConfigFromFile = config.LoadConfigFromFile
	DefaultConfig      = config.DefaultConfig
	SaveConfig         = config.SaveConfig
)

// NewAgentFromConfig creates a new agent from a config file
func NewAgentFromConfig(cfg *Config) *Agent {
	options := []AgentOption{
		WithHost(cfg.Server.Host),
		WithPort(uint16(cfg.Server.Port)),
		WithUsername(cfg.Agent.Username),
	}

	if cfg.Server.Version != "" {
		options = append(options, WithVersion(cfg.Server.Version))
	}

	if cfg.Agent.Timeout > 0 {
		options = append(options, WithTimeout(time.Duration(cfg.Agent.Timeout)*time.Second))
	}

	if cfg.Agent.CommandPrefix != "" {
		options = append(options, WithCommandPrefix(cfg.Agent.CommandPrefix))
	}

	return NewAgent(options...)
}
