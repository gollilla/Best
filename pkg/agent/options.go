package agent

import (
	"time"

	"github.com/gollilla/best/pkg/types"
)

// Option is a function that configures an Agent
type AgentOption func(*Agent)

// WithHost sets the server host
func WithHost(host string) AgentOption {
	return func(a *Agent) {
		a.options.Host = host
	}
}

// WithPort sets the server port
func WithPort(port uint16) AgentOption {
	return func(a *Agent) {
		a.options.Port = port
	}
}

// WithUsername sets the player username
func WithUsername(username string) AgentOption {
	return func(a *Agent) {
		a.options.Username = username
		a.username = username
	}
}

// WithOffline sets offline mode
func WithOffline(offline bool) AgentOption {
	return func(a *Agent) {
		a.options.Offline = offline
	}
}

// WithTimeout sets the connection timeout
func WithTimeout(timeout time.Duration) AgentOption {
	return func(a *Agent) {
		a.options.Timeout = timeout
	}
}

// WithVersion sets the Minecraft version
func WithVersion(version string) AgentOption {
	return func(a *Agent) {
		a.options.Version = version
	}
}

// WithCommandPrefix sets the command prefix for agent mode
func WithCommandPrefix(prefix string) AgentOption {
	return func(a *Agent) {
		a.commandPrefix = prefix
	}
}

// DefaultOptions returns default client options
func DefaultOptions() types.ClientOptions {
	return types.ClientOptions{
		Host:     "localhost",
		Port:     19132,
		Username: "TestBot",
		Offline:  true,
		Timeout:  30 * time.Second,
		Version:  "1.21.130",
	}
}
