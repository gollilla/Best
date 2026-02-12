package assertions

import (
	"context"
	"fmt"
	"time"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// ScoreboardAssertion provides scoreboard-related assertions
type ScoreboardAssertion struct {
	agent AgentInterface
}

// ToHaveObjective waits for a scoreboard objective to be created/displayed
func (s *ScoreboardAssertion) ToHaveObjective(objectiveName string, timeout time.Duration) {
	// First check current state
	state := s.agent.State()
	if state.Scoreboard != nil {
		// Check if objective exists in objectives map
		if _, exists := state.Scoreboard.Objectives[objectiveName]; exists {
			return // Objective exists in current state
		}
		// Also check if any entry references this objective
		for _, entry := range state.Scoreboard.Entries {
			if entry.ObjectiveName == objectiveName {
				return // Found entry for this objective
			}
		}
	}

	// If not in state, wait for event
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		// Check for display objective
		if displayInfo, ok := d.(map[string]interface{}); ok {
			if name, exists := displayInfo["objectiveName"]; exists && name == objectiveName {
				// Check if this is a display action (not a removal)
				if action, ok := displayInfo["action"].(string); ok && action == "remove" {
					return false
				}
				return true
			}
		}

		// Check for score entry (only if it's being added/modified)
		if entry, ok := d.(*types.ScoreboardEntry); ok {
			return entry.ObjectiveName == objectiveName && entry.ActionType == types.ScoreboardActionModify
		}

		return false
	})

	if err != nil {
		panic(fmt.Errorf("objective %q not found within %v: %w", objectiveName, timeout, err))
	}
}

// ToHaveScore waits for a specific score value in an objective
func (s *ScoreboardAssertion) ToHaveScore(objectiveName string, expectedScore int32, timeout time.Duration) {
	// First check current state
	state := s.agent.State()
	if state.Scoreboard != nil {
		for _, entry := range state.Scoreboard.Entries {
			if entry.ObjectiveName == objectiveName && entry.Score == expectedScore {
				return // Found matching entry in current state
			}
		}
	}

	// If not in state, wait for event
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.ObjectiveName == objectiveName && entry.Score == expectedScore
	})

	if err != nil {
		panic(fmt.Errorf("score %d in objective %q not found within %v: %w", expectedScore, objectiveName, timeout, err))
	}

	entry := data.(*types.ScoreboardEntry)
	if entry.Score != expectedScore {
		panic(NewAssertionError(
			fmt.Sprintf("expected score in objective %q to be %d", objectiveName, expectedScore),
			expectedScore,
			entry.Score,
		))
	}
}

// ToHaveScoreAbove waits for a score above the minimum value
func (s *ScoreboardAssertion) ToHaveScoreAbove(objectiveName string, minScore int32, timeout time.Duration) {
	// First check current state
	state := s.agent.State()
	if state.Scoreboard != nil {
		for _, entry := range state.Scoreboard.Entries {
			if entry.ObjectiveName == objectiveName && entry.Score > minScore {
				return // Found matching entry in current state
			}
		}
	}

	// If not in state, wait for event
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.ObjectiveName == objectiveName && entry.Score > minScore
	})

	if err != nil {
		panic(fmt.Errorf("score above %d in objective %q not found within %v: %w", minScore, objectiveName, timeout, err))
	}

	entry := data.(*types.ScoreboardEntry)
	if entry.Score <= minScore {
		panic(NewAssertionError(
			fmt.Sprintf("expected score in objective %q to be above %d", objectiveName, minScore),
			fmt.Sprintf("> %d", minScore),
			entry.Score,
		))
	}
}

// ToHaveScoreBelow waits for a score below the maximum value
func (s *ScoreboardAssertion) ToHaveScoreBelow(objectiveName string, maxScore int32, timeout time.Duration) {
	// First check current state
	state := s.agent.State()
	if state.Scoreboard != nil {
		for _, entry := range state.Scoreboard.Entries {
			if entry.ObjectiveName == objectiveName && entry.Score < maxScore {
				return // Found matching entry in current state
			}
		}
	}

	// If not in state, wait for event
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.ObjectiveName == objectiveName && entry.Score < maxScore
	})

	if err != nil {
		panic(fmt.Errorf("score below %d in objective %q not found within %v: %w", maxScore, objectiveName, timeout, err))
	}

	entry := data.(*types.ScoreboardEntry)
	if entry.Score >= maxScore {
		panic(NewAssertionError(
			fmt.Sprintf("expected score in objective %q to be below %d", objectiveName, maxScore),
			fmt.Sprintf("< %d", maxScore),
			entry.Score,
		))
	}
}

// ToHaveScoreBetween waits for a score within a range
func (s *ScoreboardAssertion) ToHaveScoreBetween(objectiveName string, minScore, maxScore int32, timeout time.Duration) {
	// First check current state
	state := s.agent.State()
	if state.Scoreboard != nil {
		for _, entry := range state.Scoreboard.Entries {
			if entry.ObjectiveName == objectiveName && entry.Score >= minScore && entry.Score <= maxScore {
				return // Found matching entry in current state
			}
		}
	}

	// If not in state, wait for event
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.ObjectiveName == objectiveName && entry.Score >= minScore && entry.Score <= maxScore
	})

	if err != nil {
		panic(fmt.Errorf("score between %d and %d in objective %q not found within %v: %w", minScore, maxScore, objectiveName, timeout, err))
	}

	entry := data.(*types.ScoreboardEntry)
	if entry.Score < minScore || entry.Score > maxScore {
		panic(NewAssertionError(
			fmt.Sprintf("expected score in objective %q to be between %d and %d", objectiveName, minScore, maxScore),
			fmt.Sprintf("%d-%d", minScore, maxScore),
			entry.Score,
		))
	}
}

// NotToHaveObjective ensures an objective does not exist or is removed
func (s *ScoreboardAssertion) NotToHaveObjective(objectiveName string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		// Check for removal
		if removeInfo, ok := d.(map[string]interface{}); ok {
			if name, exists := removeInfo["objectiveName"]; exists && name == objectiveName {
				if removed, ok := removeInfo["removed"].(bool); ok && removed {
					return true
				}
			}
		}
		return false
	})

	if err != nil {
		// Timeout means objective was not removed or never existed (which is OK)
		return
	}

	if data != nil {
		// If we got removal event, that's expected
		return
	}
}

// ToChangeScore waits for any score change in an objective
func (s *ScoreboardAssertion) ToChangeScore(objectiveName string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.ObjectiveName == objectiveName && entry.ActionType == types.ScoreboardActionModify
	})

	if err != nil {
		panic(fmt.Errorf("score change in objective %q not detected within %v: %w", objectiveName, timeout, err))
	}
}

// ToHavePlayerScore waits for a player (by EntityUniqueID) to have a specific score
func (s *ScoreboardAssertion) ToHavePlayerScore(objectiveName string, entityID int64, expectedScore int32, timeout time.Duration) {
	// First check current state
	state := s.agent.State()
	if state.Scoreboard != nil {
		for _, entry := range state.Scoreboard.Entries {
			if entry.ObjectiveName == objectiveName &&
				entry.EntityUniqueID == entityID &&
				entry.Score == expectedScore {
				return // Found matching entry in current state
			}
		}
	}

	// If not in state, wait for event
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.ObjectiveName == objectiveName &&
			entry.EntityUniqueID == entityID &&
			entry.Score == expectedScore &&
			entry.ActionType == types.ScoreboardActionModify
	})

	if err != nil {
		panic(fmt.Errorf("player %d score %d in objective %q not found within %v: %w", entityID, expectedScore, objectiveName, timeout, err))
	}

	entry := data.(*types.ScoreboardEntry)
	if entry.Score != expectedScore {
		panic(NewAssertionError(
			fmt.Sprintf("expected player %d score in objective %q to be %d", entityID, objectiveName, expectedScore),
			expectedScore,
			entry.Score,
		))
	}
}

// ToHaveFakePlayerScore waits for a fake player (by display name) to have a specific score
func (s *ScoreboardAssertion) ToHaveFakePlayerScore(objectiveName string, displayName string, expectedScore int32, timeout time.Duration) {
	// First check current state
	state := s.agent.State()
	if state.Scoreboard != nil {
		for _, entry := range state.Scoreboard.Entries {
			if entry.ObjectiveName == objectiveName &&
				entry.IdentityType == types.ScoreboardIdentityFakePlayer &&
				entry.DisplayName == displayName &&
				entry.Score == expectedScore {
				return // Found matching entry in current state
			}
		}
	}

	// If not in state, wait for event
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.ObjectiveName == objectiveName &&
			entry.IdentityType == types.ScoreboardIdentityFakePlayer &&
			entry.DisplayName == displayName &&
			entry.Score == expectedScore &&
			entry.ActionType == types.ScoreboardActionModify
	})

	if err != nil {
		panic(fmt.Errorf("fake player %q score %d in objective %q not found within %v: %w", displayName, expectedScore, objectiveName, timeout, err))
	}

	entry := data.(*types.ScoreboardEntry)
	if entry.Score != expectedScore {
		panic(NewAssertionError(
			fmt.Sprintf("expected fake player %q score in objective %q to be %d", displayName, objectiveName, expectedScore),
			expectedScore,
			entry.Score,
		))
	}
}

// ToRemoveScore waits for a score entry to be removed
func (s *ScoreboardAssertion) ToRemoveScore(objectiveName string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.ObjectiveName == objectiveName && entry.ActionType == types.ScoreboardActionRemove
	})

	if err != nil {
		panic(fmt.Errorf("score removal in objective %q not detected within %v: %w", objectiveName, timeout, err))
	}
}

// ToHaveDisplaySlot waits for an objective to be displayed in a specific slot
func (s *ScoreboardAssertion) ToHaveDisplaySlot(objectiveName string, displaySlot string, timeout time.Duration) {
	// First check current state
	state := s.agent.State()
	if state.Scoreboard != nil {
		if obj, exists := state.Scoreboard.Objectives[objectiveName]; exists {
			if obj.DisplaySlot == displaySlot {
				return // Already in the correct state
			}
		}
	}

	// If not in state, wait for event
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		displayInfo, ok := d.(map[string]interface{})
		if !ok {
			return false
		}

		name, nameOk := displayInfo["objectiveName"].(string)
		slot, slotOk := displayInfo["displaySlot"].(string)
		action, actionOk := displayInfo["action"].(string)

		return nameOk && slotOk && actionOk &&
			name == objectiveName &&
			slot == displaySlot &&
			action == "display"
	})

	if err != nil {
		panic(fmt.Errorf("objective %q not displayed in slot %q within %v: %w", objectiveName, displaySlot, timeout, err))
	}
}
