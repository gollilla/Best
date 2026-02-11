package agent

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/protocol"
	"github.com/gollilla/best/pkg/state"
	"github.com/gollilla/best/pkg/types"
	"github.com/gollilla/best/pkg/world"
)

// Agent represents a virtual player that can connect to a Minecraft Bedrock server
type Agent struct {
	username       string
	options        types.ClientOptions
	client         *protocol.Client
	state          *types.PlayerState
	isConnected    atomic.Bool
	hasSpawned     atomic.Bool
	emitter        *events.Emitter

	// Agent features
	commandPrefix  string

	// Player state
	inventory      []types.InventoryItem
	effects        []types.Effect
	entities       map[int64]types.Entity
	scores         map[string]int32
	tags           []string
	hunger         float32
	permLevel      int32

	// World management
	world          *world.World

	// Internal
	pendingForms   map[int32]types.Form
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewAgent creates a new agent with the given options
func NewAgent(opts ...AgentOption) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	a := &Agent{
		options:       DefaultOptions(),
		state:         state.CreateInitialState(),
		emitter:       events.NewEmitter(),
		world:         world.NewWorld(),
		ctx:           ctx,
		cancel:        cancel,
		commandPrefix: "!",
		entities:      make(map[int64]types.Entity),
		scores:        make(map[string]int32),
		pendingForms:  make(map[int32]types.Form),
		tags:          make([]string, 0),
		inventory:     make([]types.InventoryItem, 0),
		effects:       make([]types.Effect, 0),
	}

	// Apply options
	for _, opt := range opts {
		opt(a)
	}

	// Set username from options if not already set
	if a.username == "" {
		a.username = a.options.Username
	}

	// Create protocol client
	a.client = protocol.NewClient(a.emitter, a.state)

	return a
}

// Connect establishes connection to the Minecraft server
func (a *Agent) Connect() error {
	if a.isConnected.Load() {
		return fmt.Errorf("already connected")
	}

	if err := a.client.Connect(a.options); err != nil {
		return err
	}

	a.isConnected.Store(true)

	// Perform spawn sequence after connection is established
	if err := a.client.DoSpawn(); err != nil {
		return err
	}

	a.hasSpawned.Store(true)
	return nil
}

// Disconnect closes the connection
func (a *Agent) Disconnect() error {
	if !a.isConnected.Load() {
		return nil
	}

	a.cancel()

	if err := a.client.Disconnect(); err != nil {
		return err
	}

	a.isConnected.Store(false)
	a.state = state.CreateInitialState()
	return nil
}

// Username returns the agent's username
func (a *Agent) Username() string {
	return a.username
}

// IsConnected returns whether the agent is currently connected
func (a *Agent) IsConnected() bool {
	return a.isConnected.Load()
}

// Position returns the current position
func (a *Agent) Position() types.Position {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state.Position
}

// Health returns the current health
func (a *Agent) Health() float32 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state.Health
}

// Gamemode returns the current gamemode
func (a *Agent) Gamemode() int32 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state.Gamemode
}

// State returns a copy of the current player state
func (a *Agent) State() types.PlayerState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return *a.state
}

// GetInventory returns a copy of the inventory
func (a *Agent) GetInventory() []types.InventoryItem {
	a.mu.RLock()
	defer a.mu.RUnlock()
	inv := make([]types.InventoryItem, len(a.inventory))
	copy(inv, a.inventory)
	return inv
}

// GetEffects returns a copy of active effects
func (a *Agent) GetEffects() []types.Effect {
	a.mu.RLock()
	defer a.mu.RUnlock()
	effects := make([]types.Effect, len(a.effects))
	copy(effects, a.effects)
	return effects
}

// GetEntities returns a copy of nearby entities
func (a *Agent) GetEntities() []types.Entity {
	a.mu.RLock()
	defer a.mu.RUnlock()
	entities := make([]types.Entity, 0, len(a.entities))
	for _, entity := range a.entities {
		entities = append(entities, entity)
	}
	return entities
}

// GetScore returns a score value for an objective
func (a *Agent) GetScore(objective string) (int32, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	score, ok := a.scores[objective]
	return score, ok
}

// GetTags returns a copy of player tags
func (a *Agent) GetTags() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	tags := make([]string, len(a.tags))
	copy(tags, a.tags)
	return tags
}

// GetHunger returns the current hunger level
func (a *Agent) GetHunger() float32 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.hunger
}

// GetPermissionLevel returns the current permission level
func (a *Agent) GetPermissionLevel() int32 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.permLevel
}

// World returns the world manager for block and chunk access
func (a *Agent) World() *world.World {
	return a.world
}

// Emitter returns the event emitter for listening to events
func (a *Agent) Emitter() *events.Emitter {
	return a.emitter
}

// Context returns the agent's context
func (a *Agent) Context() context.Context {
	return a.ctx
}

// WaitForSpawn waits for the spawn event
func (a *Agent) WaitForSpawn(ctx context.Context) error {
	// If already spawned, return immediately
	if a.hasSpawned.Load() {
		return nil
	}

	// Otherwise wait for the spawn event
	_, err := a.emitter.WaitFor(ctx, events.EventSpawn, nil)
	return err
}
