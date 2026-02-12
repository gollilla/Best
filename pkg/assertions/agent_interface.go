package assertions

import (
	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// AgentInterface defines the methods needed by assertions
// This interface prevents circular imports between agent and assertions packages
type AgentInterface interface {
	// Connection
	IsConnected() bool

	// State accessors
	Position() types.Position
	State() types.PlayerState
	Health() float32
	Gamemode() int32

	// Collections
	GetInventory() []types.InventoryItem
	GetEffects() []types.Effect
	GetEntities() []types.Entity
	GetTags() []string
	GetHunger() float32
	GetPermissionLevel() int32

	// Scoreboard
	GetScore(objectiveName string) *int32
	GetScoreByPlayer(objectiveName string, displayName string) *int32
	GetScoreByEntityID(objectiveName string, entityID int64) *int32
	GetAllScores(objectiveName string) []types.ScoreboardEntry

	// Form handling
	GetPendingForm(id int32) (types.Form, bool)
	GetLastForm() types.Form
	SubmitForm(formID int32, response types.FormResponse) error
	ClearPendingForms()

	// Event system
	Emitter() *events.Emitter
}
