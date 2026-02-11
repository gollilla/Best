package assertions

import (
	"context"
	"fmt"
	"time"

	"github.com/gollilla/best/pkg/events"
)

// GamemodeAssertion provides gamemode-related assertions
type GamemodeAssertion struct {
	agent AgentInterface
}

// Gamemode constants
const (
	GamemodeSurvival  int32 = 0
	GamemodeCreative  int32 = 1
	GamemodeAdventure int32 = 2
	GamemodeSpectator int32 = 3
)

// ToBe checks if the gamemode is exactly the expected value
func (g *GamemodeAssertion) ToBe(expected int32) {
	actual := g.agent.Gamemode()

	if actual != expected {
		panic(NewAssertionError(
			fmt.Sprintf("expected gamemode to be %s (%d)", gamemodeName(expected), expected),
			gamemodeName(expected),
			gamemodeName(actual),
		))
	}
}

// ToBeSurvival checks if the gamemode is survival (0)
func (g *GamemodeAssertion) ToBeSurvival() {
	g.ToBe(GamemodeSurvival)
}

// ToBeCreative checks if the gamemode is creative (1)
func (g *GamemodeAssertion) ToBeCreative() {
	g.ToBe(GamemodeCreative)
}

// ToBeAdventure checks if the gamemode is adventure (2)
func (g *GamemodeAssertion) ToBeAdventure() {
	g.ToBe(GamemodeAdventure)
}

// ToBeSpectator checks if the gamemode is spectator (3)
func (g *GamemodeAssertion) ToBeSpectator() {
	g.ToBe(GamemodeSpectator)
}

// ToChange waits for gamemode to change within the timeout
func (g *GamemodeAssertion) ToChange(timeout time.Duration) int32 {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := g.agent.Emitter().WaitFor(ctx, events.EventGamemodeUpdate, nil)
	if err != nil {
		panic(err)
	}

	gamemode, ok := data.(int32)
	if !ok {
		panic(fmt.Errorf("invalid gamemode data type"))
	}

	return gamemode
}

// ToChangeTo waits for gamemode to change to a specific value within the timeout
func (g *GamemodeAssertion) ToChangeTo(expected int32, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := g.agent.Emitter().WaitFor(ctx, events.EventGamemodeUpdate, func(d events.EventData) bool {
		gamemode, ok := d.(int32)
		if !ok {
			return false
		}
		return gamemode == expected
	})

	if err != nil {
		panic(err)
	}

	gamemode := data.(int32)
	if gamemode != expected {
		panic(NewAssertionError(
			fmt.Sprintf("expected gamemode to change to %s (%d)", gamemodeName(expected), expected),
			gamemodeName(expected),
			gamemodeName(gamemode),
		))
	}
}

// Helper function

// gamemodeName returns the string name for a gamemode value
func gamemodeName(gamemode int32) string {
	switch gamemode {
	case GamemodeSurvival:
		return "survival"
	case GamemodeCreative:
		return "creative"
	case GamemodeAdventure:
		return "adventure"
	case GamemodeSpectator:
		return "spectator"
	default:
		return fmt.Sprintf("unknown(%d)", gamemode)
	}
}
