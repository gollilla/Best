package assertions

import (
	"context"
	"fmt"
	"time"

	"github.com/gollilla/best/pkg/events"
)

// HungerAssertion provides hunger-related assertions
type HungerAssertion struct {
	agent AgentInterface
}

// ToBe checks if the hunger is exactly the expected value
func (h *HungerAssertion) ToBe(expected float32) {
	actual := h.agent.GetHunger()

	if actual != expected {
		panic(NewAssertionError(
			fmt.Sprintf("expected hunger to be %.1f", expected),
			expected,
			actual,
		))
	}
}

// ToBeAbove checks if the hunger is above the minimum value
func (h *HungerAssertion) ToBeAbove(min float32) {
	actual := h.agent.GetHunger()

	if actual <= min {
		panic(NewAssertionError(
			fmt.Sprintf("expected hunger to be above %.1f", min),
			fmt.Sprintf("> %.1f", min),
			actual,
		))
	}
}

// ToBeBelow checks if the hunger is below the maximum value
func (h *HungerAssertion) ToBeBelow(max float32) {
	actual := h.agent.GetHunger()

	if actual >= max {
		panic(NewAssertionError(
			fmt.Sprintf("expected hunger to be below %.1f", max),
			fmt.Sprintf("< %.1f", max),
			actual,
		))
	}
}

// ToBeFull checks if the hunger is at maximum (20.0)
func (h *HungerAssertion) ToBeFull() {
	const maxHunger = 20.0
	actual := h.agent.GetHunger()

	if actual != maxHunger {
		panic(NewAssertionError(
			"expected hunger to be full (20.0)",
			maxHunger,
			actual,
		))
	}
}

// ToChange waits for hunger to change within the timeout
func (h *HungerAssertion) ToChange(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := h.agent.Emitter().WaitFor(ctx, events.EventHungerUpdate, nil)
	if err != nil {
		panic(fmt.Errorf("hunger change event not received within %v: %w", timeout, err))
	}
}

// ToReach waits for hunger to reach a specific value within the timeout
func (h *HungerAssertion) ToReach(expected float32, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := h.agent.Emitter().WaitFor(ctx, events.EventHungerUpdate, func(d events.EventData) bool {
		hunger, ok := d.(float32)
		if !ok {
			return false
		}
		return hunger == expected
	})

	if err != nil {
		panic(fmt.Errorf("hunger did not reach %.1f within %v: %w", expected, timeout, err))
	}

	hunger := data.(float32)
	if hunger != expected {
		panic(NewAssertionError(
			fmt.Sprintf("expected hunger to reach %.1f", expected),
			expected,
			hunger,
		))
	}
}

// ToBeAboveWithin waits for hunger to be above a threshold within the timeout
func (h *HungerAssertion) ToBeAboveWithin(min float32, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := h.agent.Emitter().WaitFor(ctx, events.EventHungerUpdate, func(d events.EventData) bool {
		hunger, ok := d.(float32)
		if !ok {
			return false
		}
		return hunger > min
	})

	if err != nil {
		panic(fmt.Errorf("hunger did not go above %.1f within %v: %w", min, timeout, err))
	}

	hunger := data.(float32)
	if hunger <= min {
		panic(NewAssertionError(
			fmt.Sprintf("expected hunger to be above %.1f", min),
			fmt.Sprintf("> %.1f", min),
			hunger,
		))
	}
}

// ToBeBelowWithin waits for hunger to be below a threshold within the timeout
func (h *HungerAssertion) ToBeBelowWithin(max float32, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := h.agent.Emitter().WaitFor(ctx, events.EventHungerUpdate, func(d events.EventData) bool {
		hunger, ok := d.(float32)
		if !ok {
			return false
		}
		return hunger < max
	})

	if err != nil {
		panic(fmt.Errorf("hunger did not go below %.1f within %v: %w", max, timeout, err))
	}

	hunger := data.(float32)
	if hunger >= max {
		panic(NewAssertionError(
			fmt.Sprintf("expected hunger to be below %.1f", max),
			fmt.Sprintf("< %.1f", max),
			hunger,
		))
	}
}
