// Package actions provides action and assertion registry for scenario execution
package actions

import (
	"context"
	"fmt"
	"sync"

	"github.com/gollilla/best/pkg/agent"
	"github.com/gollilla/best/pkg/types"
)

// ActionDefinition defines an action that can be executed by the scenario engine
type ActionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  []ParameterDef `json:"parameters"`
}

// AssertionDefinition defines an assertion that can be checked by the scenario engine
type AssertionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  []ParameterDef `json:"parameters"`
}

// ParameterDef defines a parameter for an action or assertion
type ParameterDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "number", "boolean", "duration"
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
}

// ActionFunc is a function that executes an action
type ActionFunc func(ctx context.Context, agent *agent.Agent, params map[string]interface{}) error

// AssertionFunc is a function that executes an assertion
type AssertionFunc func(ctx context.Context, agent *agent.Agent, params map[string]interface{}) error

// ActionEntry represents a registered action with its definition and executor
type ActionEntry struct {
	Definition ActionDefinition
	Execute    ActionFunc
}

// AssertionEntry represents a registered assertion with its definition and executor
type AssertionEntry struct {
	Definition AssertionDefinition
	Assert     AssertionFunc
}

// Registry holds all registered actions and assertions
type Registry struct {
	mu          sync.RWMutex
	actions     map[string]ActionEntry
	assertions  map[string]AssertionEntry
	// Scenario context state
	lastPosition *types.Position
}

// NewRegistry creates a new action/assertion registry with builtin actions
func NewRegistry() *Registry {
	r := &Registry{
		actions:    make(map[string]ActionEntry),
		assertions: make(map[string]AssertionEntry),
	}

	// Register builtin actions and assertions
	registerBuiltinActions(r)
	registerBuiltinAssertions(r)

	return r
}

// RegisterAction registers a new action
func (r *Registry) RegisterAction(name string, def ActionDefinition, fn ActionFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	def.Name = name
	r.actions[name] = ActionEntry{
		Definition: def,
		Execute:    fn,
	}
}

// RegisterAssertion registers a new assertion
func (r *Registry) RegisterAssertion(name string, def AssertionDefinition, fn AssertionFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	def.Name = name
	r.assertions[name] = AssertionEntry{
		Definition: def,
		Assert:     fn,
	}
}

// GetAction returns an action by name
func (r *Registry) GetAction(name string) (ActionEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.actions[name]
	return entry, ok
}

// GetAssertion returns an assertion by name
func (r *Registry) GetAssertion(name string) (AssertionEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.assertions[name]
	return entry, ok
}

// ExecuteAction executes an action by name
func (r *Registry) ExecuteAction(ctx context.Context, agent *agent.Agent, name string, params map[string]interface{}) error {
	entry, ok := r.GetAction(name)
	if !ok {
		return fmt.Errorf("action not found: %s", name)
	}

	return entry.Execute(ctx, agent, params)
}

// ExecuteAssertion executes an assertion by name
func (r *Registry) ExecuteAssertion(ctx context.Context, agent *agent.Agent, name string, params map[string]interface{}) error {
	entry, ok := r.GetAssertion(name)
	if !ok {
		return fmt.Errorf("assertion not found: %s", name)
	}

	return entry.Assert(ctx, agent, params)
}

// GetActionDefinitions returns all action definitions
func (r *Registry) GetActionDefinitions() []ActionDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]ActionDefinition, 0, len(r.actions))
	for _, entry := range r.actions {
		defs = append(defs, entry.Definition)
	}
	return defs
}

// GetAssertionDefinitions returns all assertion definitions
func (r *Registry) GetAssertionDefinitions() []AssertionDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]AssertionDefinition, 0, len(r.assertions))
	for _, entry := range r.assertions {
		defs = append(defs, entry.Definition)
	}
	return defs
}

// IsAction checks if a name is a registered action
func (r *Registry) IsAction(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.actions[name]
	return ok
}

// IsAssertion checks if a name is a registered assertion
func (r *Registry) IsAssertion(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.assertions[name]
	return ok
}

// SetLastPosition stores the last known position for relative movement assertions
func (r *Registry) SetLastPosition(pos types.Position) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastPosition = &pos
}

// GetLastPosition returns the last known position
func (r *Registry) GetLastPosition() *types.Position {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastPosition
}

// ClearContext clears the scenario context state
func (r *Registry) ClearContext() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastPosition = nil
}
