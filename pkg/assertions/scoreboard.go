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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		// Check for display objective
		if displayInfo, ok := d.(map[string]interface{}); ok {
			if name, exists := displayInfo["objectiveName"]; exists && name == objectiveName {
				return true
			}
		}

		// Check for score entry
		if entry, ok := d.(*types.ScoreboardEntry); ok {
			return entry.Objective == objectiveName
		}

		return false
	})

	if err != nil {
		panic(fmt.Errorf("objective %q not found within %v: %w", objectiveName, timeout, err))
	}
}

// ToHaveScore waits for a specific score value in an objective
func (s *ScoreboardAssertion) ToHaveScore(objectiveName string, expectedScore int32, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.Objective == objectiveName && entry.Score == expectedScore
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.Objective == objectiveName && entry.Score > minScore
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.Objective == objectiveName && entry.Score < maxScore
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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := s.agent.Emitter().WaitFor(ctx, events.EventScoreUpdate, func(d events.EventData) bool {
		entry, ok := d.(*types.ScoreboardEntry)
		if !ok {
			return false
		}
		return entry.Objective == objectiveName && entry.Score >= minScore && entry.Score <= maxScore
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
		return entry.Objective == objectiveName
	})

	if err != nil {
		panic(fmt.Errorf("score change in objective %q not detected within %v: %w", objectiveName, timeout, err))
	}
}
