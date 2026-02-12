package types

import (
	"time"
)

// Position represents a 3D position
type Position struct {
	X float64
	Y float64
	Z float64
}

// Rotation represents yaw and pitch
type Rotation struct {
	Yaw   float32
	Pitch float32
}

// PlayerState represents the complete state of a player
type PlayerState struct {
	RuntimeEntityID int64
	Position        Position
	Rotation        Rotation
	Health          float32
	Gamemode        int32
	Dimension       string
	IsOnGround      bool
	PermissionLevel int32
}

// CommandOutput represents the result of a command execution (CommandOutputPacket)
type CommandOutput struct {
	Command    string
	Success    bool
	Output     string
	StatusCode int32
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Type      string
	Sender    string
	Message   string
	Timestamp int64
	XUID      string
}

// Form represents a form (modal, action, or custom)
type Form interface {
	GetID() int32
	GetType() string
	GetTitle() string
}

// ModalForm represents a yes/no dialog
type ModalForm struct {
	ID      int32
	Title   string
	Content string
	Button1 string
	Button2 string
}

func (f *ModalForm) GetID() int32      { return f.ID }
func (f *ModalForm) GetType() string   { return "modal" }
func (f *ModalForm) GetTitle() string  { return f.Title }

// ActionForm represents a button list
type ActionForm struct {
	ID      int32
	Title   string
	Content string
	Buttons []ActionButton
}

func (f *ActionForm) GetID() int32     { return f.ID }
func (f *ActionForm) GetType() string  { return "action" }
func (f *ActionForm) GetTitle() string { return f.Title }

// ActionButton represents a button in an action form
type ActionButton struct {
	Text  string
	Image *ButtonImage
}

// ButtonImage represents an image on a button
type ButtonImage struct {
	Type string // "path" or "url"
	Data string
}

// CustomForm represents a form with input elements
type CustomForm struct {
	ID      int32
	Title   string
	Content []FormElement
}

func (f *CustomForm) GetID() int32     { return f.ID }
func (f *CustomForm) GetType() string  { return "form" }
func (f *CustomForm) GetTitle() string { return f.Title }

// FormElement represents an element in a custom form
type FormElement interface {
	GetType() string
}

// InventoryItem represents an item in inventory
type InventoryItem struct {
	ID           string
	Count        int32
	Slot         int32
	Damage       *int32
	Enchantments []Enchantment
}

// Enchantment represents an enchantment on an item
type Enchantment struct {
	ID    string
	Level int32
}

// Effect represents a status effect
type Effect struct {
	ID        string
	Amplifier int32
	Duration  int32
	Visible   bool
}

// Entity represents an entity in the world
type Entity struct {
	RuntimeID int64
	Type      string
	Position  Position
	NameTag   *string
}

// Block represents a block in the world
type Block struct {
	Name      string
	Position  Position
	RuntimeID int32
}

// BlockUpdate represents a block update event
type BlockUpdate struct {
	Position  Position
	RuntimeID int32
}

// BlockBreakData represents block breaking information
type BlockBreakData struct {
	Position  Position
	Completed bool
	Progress  float64
}

// ScoreboardEntry represents a scoreboard entry
type ScoreboardEntry struct {
	Objective   string
	Score       int32
	DisplayName *string
}

// TitleDisplay represents a title/subtitle/actionbar display
type TitleDisplay struct {
	Type    string // "title", "subtitle", "actionbar"
	Text    string
	FadeIn  int32
	Stay    int32
	FadeOut int32
}

// SoundPlay represents a sound being played
type SoundPlay struct {
	Name     string
	Position Position
	Volume   float32
	Pitch    float32
}

// ParticleSpawn represents a particle effect
type ParticleSpawn struct {
	Name     string
	Position Position
}

// ServerInfo represents server information
type ServerInfo struct {
	Host    string
	Port    uint16
	Version string
}

// ClientOptions represents connection options
type ClientOptions struct {
	Host     string
	Port     uint16
	Username string
	Timeout  time.Duration
	Version  string
}

// FormResponse can be null, bool (modal), int (action), or []interface{} (custom)
type FormResponse interface{}
