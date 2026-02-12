package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	bestevents "github.com/gollilla/best/pkg/events"
	bestprotocol "github.com/gollilla/best/pkg/protocol"
	"github.com/gollilla/best/pkg/state"
	"github.com/gollilla/best/pkg/types"
	"github.com/gollilla/best/pkg/world"
)

// Agent represents a virtual player that can connect to a Minecraft Bedrock server
type Agent struct {
	username    string
	options     types.ClientOptions
	client      *bestprotocol.Client
	state       *types.PlayerState
	isConnected atomic.Bool
	hasSpawned  atomic.Bool
	emitter     *bestevents.Emitter

	// Agent features
	commandPrefix     string
	commandSendMethod string        // "text" or "request"
	commandTimeout    time.Duration // assertion wait timeout

	// Player state
	inventory []types.InventoryItem
	effects   []types.Effect
	entities  map[int64]types.Entity
	scores    map[string]int32
	tags      []string
	hunger    float32
	permLevel int32

	// World management
	world *world.World

	// Internal
	pendingForms map[int32]types.Form
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewAgent creates a new agent with the given options
func NewAgent(opts ...AgentOption) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	a := &Agent{
		options:           DefaultOptions(),
		state:             state.CreateInitialState(),
		emitter:           bestevents.NewEmitter(),
		world:             world.NewWorld(),
		ctx:               ctx,
		cancel:            cancel,
		commandPrefix:     "!",
		commandSendMethod: "text",
		commandTimeout:    5 * time.Second,
		entities:          make(map[int64]types.Entity),
		scores:            make(map[string]int32),
		pendingForms:      make(map[int32]types.Form),
		tags:              make([]string, 0),
		inventory:         make([]types.InventoryItem, 0),
		effects:           make([]types.Effect, 0),
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
	a.client = bestprotocol.NewClient(a.emitter, a.state, a.username)

	// Listen for form events and store them
	a.emitter.On(bestevents.EventForm, func(data bestevents.EventData) {
		form, ok := data.(types.Form)
		if !ok {
			return
		}

		a.mu.Lock()
		a.pendingForms[form.GetID()] = form
		a.mu.Unlock()
	})

	// Listen for inventory updates and store them
	a.emitter.On(bestevents.EventInventoryUpdate, func(data bestevents.EventData) {
		items, ok := data.([]types.InventoryItem)
		if !ok {
			return
		}

		a.mu.Lock()
		a.inventory = items
		a.mu.Unlock()
	})

	// Listen for inventory slot updates
	a.emitter.On(bestevents.EventInventorySlotUpdate, func(data bestevents.EventData) {
		item, ok := data.(types.InventoryItem)
		if !ok {
			return
		}

		a.mu.Lock()
		// Update or add the item in the inventory
		found := false
		for i, existingItem := range a.inventory {
			if existingItem.Slot == item.Slot {
				a.inventory[i] = item
				found = true
				break
			}
		}
		if !found {
			a.inventory = append(a.inventory, item)
		}
		a.mu.Unlock()
	})

	return a
}

// Connect establishes connection to the Minecraft server
func (a *Agent) Connect() error {
	if a.isConnected.Load() {
		return fmt.Errorf("already connected")
	}

	// Create new context for this connection
	// This is important for reconnections after disconnect
	a.ctx, a.cancel = context.WithCancel(context.Background())

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

	// Cancel context first to stop all goroutines
	a.cancel()

	// Close the connection
	var disconnectErr error
	if err := a.client.Disconnect(); err != nil {
		disconnectErr = err
	}

	// Always reset state, even if disconnect had an error
	a.isConnected.Store(false)
	a.hasSpawned.Store(false)
	a.state = state.CreateInitialState()

	// Clear pending forms
	a.mu.Lock()
	a.pendingForms = make(map[int32]types.Form)
	a.mu.Unlock()

	// Wait for server-side session cleanup
	// This prevents "Logged in from other location" errors when reconnecting
	// with the same username shortly after disconnect
	// Increased to 3 seconds to ensure reliable cleanup
	time.Sleep(3 * time.Second)

	return disconnectErr
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

// GetScore returns the agent's current score in the specified objective
// Returns nil if the score is not found
func (a *Agent) GetScore(objectiveName string) *int32 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.state.Scoreboard == nil {
		return nil
	}

	// Get agent's EntityUniqueID
	agentEntityID := a.state.RuntimeEntityID

	// Search for the agent's score entry
	for _, entry := range a.state.Scoreboard.Entries {
		if entry.ObjectiveName == objectiveName && entry.EntityUniqueID == agentEntityID {
			score := entry.Score
			return &score
		}
	}

	return nil
}

// GetScoreByPlayer returns the score for a specific player (by display name)
// Returns nil if the score is not found
func (a *Agent) GetScoreByPlayer(objectiveName string, displayName string) *int32 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.state.Scoreboard == nil {
		return nil
	}

	// Search for the player's score entry
	for _, entry := range a.state.Scoreboard.Entries {
		if entry.ObjectiveName == objectiveName && entry.DisplayName == displayName {
			score := entry.Score
			return &score
		}
	}

	return nil
}

// GetScoreByEntityID returns the score for a specific entity ID
// Returns nil if the score is not found
func (a *Agent) GetScoreByEntityID(objectiveName string, entityID int64) *int32 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.state.Scoreboard == nil {
		return nil
	}

	// Search for the entity's score entry
	for _, entry := range a.state.Scoreboard.Entries {
		if entry.ObjectiveName == objectiveName && entry.EntityUniqueID == entityID {
			score := entry.Score
			return &score
		}
	}

	return nil
}

// GetAllScores returns all score entries for the specified objective
func (a *Agent) GetAllScores(objectiveName string) []types.ScoreboardEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.state.Scoreboard == nil {
		return []types.ScoreboardEntry{}
	}

	var scores []types.ScoreboardEntry
	for _, entry := range a.state.Scoreboard.Entries {
		if entry.ObjectiveName == objectiveName {
			scores = append(scores, *entry)
		}
	}

	return scores
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
	return a.state.PermissionLevel
}

// World returns the world manager for block and chunk access
func (a *Agent) World() *world.World {
	return a.world
}

// Emitter returns the event emitter for listening to events
func (a *Agent) Emitter() *bestevents.Emitter {
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
	_, err := a.emitter.WaitFor(ctx, bestevents.EventSpawn, nil)
	return err
}

// GetPendingForm returns a pending form by ID
func (a *Agent) GetPendingForm(id int32) (types.Form, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	form, ok := a.pendingForms[id]
	return form, ok
}

// GetLastForm returns the last received form
func (a *Agent) GetLastForm() types.Form {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Find the form with the highest ID (most recent)
	var lastForm types.Form
	var maxID int32 = -1

	for id, form := range a.pendingForms {
		if id > maxID {
			maxID = id
			lastForm = form
		}
	}

	return lastForm
}

// SubmitForm sends a form response to the server
func (a *Agent) SubmitForm(formID int32, response types.FormResponse) error {
	if !a.isConnected.Load() {
		return fmt.Errorf("not connected")
	}

	// Create form response packet
	var responseData []byte
	var err error

	// Marshal response to JSON
	responseData, err = json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal form response: %w", err)
	}

	pk := &packet.ModalFormResponse{
		FormID:       uint32(formID),
		ResponseData: protocol.Option(responseData),
	}

	// Remove from pending forms
	a.mu.Lock()
	delete(a.pendingForms, formID)
	a.mu.Unlock()

	return a.client.WritePacket(pk)
}

// ClearPendingForms clears all pending forms
func (a *Agent) ClearPendingForms() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pendingForms = make(map[int32]types.Form)
}

