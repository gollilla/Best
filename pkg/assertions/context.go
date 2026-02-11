package assertions

import (
	"github.com/gollilla/best/pkg/types"
)

// AssertionContext provides assertion methods for an agent
type AssertionContext struct {
	agent AgentInterface

	// Basic assertions
	positionAssertion  *PositionAssertion
	chatAssertion      *ChatAssertion
	inventoryAssertion *InventoryAssertion
	formAssertion      *FormAssertion

	// Player state assertions
	healthAssertion     *HealthAssertion
	hungerAssertion     *HungerAssertion
	effectAssertion     *EffectAssertion
	gamemodeAssertion   *GamemodeAssertion
	permissionAssertion *PermissionAssertion
	tagAssertion        *TagAssertion

	// UI/Display assertions
	titleAssertion      *TitleAssertion
	subtitleAssertion   *SubtitleAssertion
	actionbarAssertion  *ActionbarAssertion
	scoreboardAssertion *ScoreboardAssertion
}

// NewAssertionContext creates a new assertion context for an agent
func NewAssertionContext(a AgentInterface) *AssertionContext {
	ctx := &AssertionContext{
		agent: a,
	}

	// Initialize assertions
	ctx.positionAssertion = &PositionAssertion{agent: a}
	ctx.chatAssertion = &ChatAssertion{agent: a}
	ctx.inventoryAssertion = &InventoryAssertion{agent: a}
	ctx.formAssertion = &FormAssertion{agent: a}

	// Initialize player state assertions
	ctx.healthAssertion = &HealthAssertion{agent: a}
	ctx.hungerAssertion = &HungerAssertion{agent: a}
	ctx.effectAssertion = &EffectAssertion{agent: a}
	ctx.gamemodeAssertion = &GamemodeAssertion{agent: a}
	ctx.permissionAssertion = &PermissionAssertion{agent: a}
	ctx.tagAssertion = &TagAssertion{agent: a}

	// Initialize UI/Display assertions
	ctx.titleAssertion = &TitleAssertion{agent: a}
	ctx.subtitleAssertion = &SubtitleAssertion{agent: a}
	ctx.actionbarAssertion = &ActionbarAssertion{agent: a}
	ctx.scoreboardAssertion = &ScoreboardAssertion{agent: a}

	return ctx
}

// === Connection assertions ===

// ToBeConnected asserts that the agent is connected
func (c *AssertionContext) ToBeConnected() error {
	if !c.agent.IsConnected() {
		return NewAssertionError(
			"Expected player to be connected",
			"connected",
			"disconnected",
		)
	}
	return nil
}

// ToBeDisconnected asserts that the agent is disconnected
func (c *AssertionContext) ToBeDisconnected() error {
	if c.agent.IsConnected() {
		return NewAssertionError(
			"Expected player to be disconnected",
			"disconnected",
			"connected",
		)
	}
	return nil
}

// === Getter methods for specific assertion types ===

// Position returns position assertions
func (c *AssertionContext) Position() *PositionAssertion {
	return c.positionAssertion
}

// Chat returns chat assertions
func (c *AssertionContext) Chat() *ChatAssertion {
	return c.chatAssertion
}

// Inventory returns inventory assertions
func (c *AssertionContext) Inventory() *InventoryAssertion {
	return c.inventoryAssertion
}

// Command returns command assertions for the given output or executes a command
// Accepts either:
//   - *types.CommandOutput: for manual command execution
//   - string: executes the command and returns assertions
func (c *AssertionContext) Command(cmdOrOutput interface{}) *CommandAssertion {
	var output *types.CommandOutput

	switch v := cmdOrOutput.(type) {
	case *types.CommandOutput:
		// Use existing output
		output = v
	case string:
		// Execute command and get output
		var err error
		output, err = c.agent.Command(v)
		if err != nil {
			panic(NewAssertionError(
				"Command execution failed: "+err.Error(),
				"successful execution",
				"error: "+err.Error(),
			))
		}
	default:
		panic(NewAssertionError(
			"Command expects *types.CommandOutput or string",
			"*types.CommandOutput or string",
			"unknown type",
		))
	}

	return &CommandAssertion{output: output}
}

// Form returns form assertions
func (c *AssertionContext) Form() *FormAssertion {
	return c.formAssertion
}

// === Player state assertion getters ===

// Health returns health assertions
func (c *AssertionContext) Health() *HealthAssertion {
	return c.healthAssertion
}

// Hunger returns hunger assertions
func (c *AssertionContext) Hunger() *HungerAssertion {
	return c.hungerAssertion
}

// Effect returns effect assertions
func (c *AssertionContext) Effect() *EffectAssertion {
	return c.effectAssertion
}

// Gamemode returns gamemode assertions
func (c *AssertionContext) Gamemode() *GamemodeAssertion {
	return c.gamemodeAssertion
}

// Permission returns permission assertions
func (c *AssertionContext) Permission() *PermissionAssertion {
	return c.permissionAssertion
}

// Tag returns tag assertions
func (c *AssertionContext) Tag() *TagAssertion {
	return c.tagAssertion
}

// === UI/Display assertion getters ===

// Title returns title assertions
func (c *AssertionContext) Title() *TitleAssertion {
	return c.titleAssertion
}

// Subtitle returns subtitle assertions
func (c *AssertionContext) Subtitle() *SubtitleAssertion {
	return c.subtitleAssertion
}

// Actionbar returns actionbar assertions
func (c *AssertionContext) Actionbar() *ActionbarAssertion {
	return c.actionbarAssertion
}

// Scoreboard returns scoreboard assertions
func (c *AssertionContext) Scoreboard() *ScoreboardAssertion {
	return c.scoreboardAssertion
}

// === Generic assertions ===
// Generic assertion methods are defined in generic.go
// They can be called directly on AssertionContext:
// - IsTrue, IsFalse
// - Equal, NotEqual
// - IsNil, NotNil
// - GreaterThan, LessThan, InRange
// - Contains, HasPrefix, HasSuffix
// - LengthEqual, ContainsElement
