package assertions

// AssertionContext provides assertion methods for an agent
type AssertionContext struct {
	agent AgentInterface

	// Basic assertions
	positionAssertion       *PositionAssertion
	chatAssertion           *ChatAssertion
	commandOutputAssertion  *CommandOutputAssertion
	inventoryAssertion      *InventoryAssertion
	formAssertion           *FormAssertion

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
	ctx.commandOutputAssertion = &CommandOutputAssertion{agent: a}
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

// CommandOutput returns CommandOutput assertions
// Use this to wait for CommandOutputPacket responses from the server
func (c *AssertionContext) CommandOutput() *CommandOutputAssertion {
	return c.commandOutputAssertion
}

// Inventory returns inventory assertions
func (c *AssertionContext) Inventory() *InventoryAssertion {
	return c.inventoryAssertion
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
