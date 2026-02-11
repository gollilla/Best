package assertions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// EffectAssertion provides effect-related assertions
type EffectAssertion struct {
	agent AgentInterface
}

// ToHave checks if the player has a specific effect
// effectID can be a full ID (e.g., "minecraft:speed") or a partial match (e.g., "speed")
func (e *EffectAssertion) ToHave(effectID string) {
	effects := e.agent.GetEffects()

	for _, effect := range effects {
		if matchesEffectID(effect.ID, effectID) {
			return
		}
	}

	panic(NewAssertionError(
		fmt.Sprintf("expected player to have effect %q", effectID),
		effectID,
		getEffectIDs(effects),
	))
}

// NotToHave checks if the player does not have a specific effect
func (e *EffectAssertion) NotToHave(effectID string) {
	effects := e.agent.GetEffects()

	for _, effect := range effects {
		if matchesEffectID(effect.ID, effectID) {
			panic(NewAssertionError(
				fmt.Sprintf("expected player not to have effect %q", effectID),
				fmt.Sprintf("not %q", effectID),
				effectID,
			))
		}
	}
}

// ToHaveLevel checks if the player has a specific effect with a specific amplifier level
func (e *EffectAssertion) ToHaveLevel(effectID string, expectedLevel int32) {
	effects := e.agent.GetEffects()

	for _, effect := range effects {
		if matchesEffectID(effect.ID, effectID) {
			if effect.Amplifier != expectedLevel {
				panic(NewAssertionError(
					fmt.Sprintf("expected effect %q to have level %d, but found %d", effectID, expectedLevel, effect.Amplifier),
					expectedLevel,
					effect.Amplifier,
				))
			}
			return
		}
	}

	panic(NewAssertionError(
		fmt.Sprintf("expected player to have effect %q with level %d, but effect not found", effectID, expectedLevel),
		effectID,
		getEffectIDs(effects),
	))
}

// ToHaveWithDuration checks if the player has a specific effect with at least a certain duration
func (e *EffectAssertion) ToHaveWithDuration(effectID string, minDuration int32) {
	effects := e.agent.GetEffects()

	for _, effect := range effects {
		if matchesEffectID(effect.ID, effectID) {
			if effect.Duration < minDuration {
				panic(NewAssertionError(
					fmt.Sprintf("expected effect %q to have at least %d ticks duration, but found %d", effectID, minDuration, effect.Duration),
					minDuration,
					effect.Duration,
				))
			}
			return
		}
	}

	panic(NewAssertionError(
		fmt.Sprintf("expected player to have effect %q, but effect not found", effectID),
		effectID,
		getEffectIDs(effects),
	))
}

// ToReceive waits for a specific effect to be received within the timeout
func (e *EffectAssertion) ToReceive(effectID string, timeout time.Duration) *types.Effect {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := e.agent.Emitter().WaitFor(ctx, events.EventEffectUpdate, func(d events.EventData) bool {
		effects, ok := d.([]types.Effect)
		if !ok {
			return false
		}

		for _, effect := range effects {
			if matchesEffectID(effect.ID, effectID) {
				return true
			}
		}
		return false
	})

	if err != nil {
		panic(err)
	}

	effects := data.([]types.Effect)
	for _, effect := range effects {
		if matchesEffectID(effect.ID, effectID) {
			return &effect
		}
	}

	panic(NewAssertionError(
		fmt.Sprintf("received effect update but effect %q not found", effectID),
		effectID,
		nil,
	))
}

// ToLose waits for a specific effect to be removed within the timeout
func (e *EffectAssertion) ToLose(effectID string, timeout time.Duration) {
	// First check if the player currently has the effect
	effects := e.agent.GetEffects()
	hasEffect := false
	for _, effect := range effects {
		if matchesEffectID(effect.ID, effectID) {
			hasEffect = true
			break
		}
	}

	if !hasEffect {
		// Player doesn't have the effect, so they can't lose it
		panic(NewAssertionError(
			fmt.Sprintf("expected player to lose effect %q, but they don't have it", effectID),
			fmt.Sprintf("has and loses %q", effectID),
			"doesn't have effect",
		))
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := e.agent.Emitter().WaitFor(ctx, events.EventEffectUpdate, func(d events.EventData) bool {
		effects, ok := d.([]types.Effect)
		if !ok {
			return false
		}

		// Check if the effect is no longer in the list
		for _, effect := range effects {
			if matchesEffectID(effect.ID, effectID) {
				return false // Effect still present
			}
		}
		return true // Effect removed
	})

	if err != nil {
		panic(err)
	}
}

// Helper functions

// matchesEffectID checks if an effect ID matches the expected pattern
// Supports full IDs (minecraft:speed) and partial matches (speed)
func matchesEffectID(actualID, expectedID string) bool {
	// Exact match
	if actualID == expectedID {
		return true
	}

	// Partial match (e.g., "speed" matches "minecraft:speed")
	if strings.Contains(actualID, expectedID) {
		return true
	}

	// Check if expected has namespace but actual doesn't
	if strings.Contains(expectedID, ":") {
		parts := strings.Split(expectedID, ":")
		if len(parts) == 2 && strings.Contains(actualID, parts[1]) {
			return true
		}
	}

	return false
}

// getEffectIDs returns a list of effect IDs the player currently has
func getEffectIDs(effects []types.Effect) []string {
	ids := make([]string, 0, len(effects))
	for _, effect := range effects {
		ids = append(ids, effect.ID)
	}
	return ids
}
