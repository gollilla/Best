package assertions

import (
	"context"
	"fmt"
	"time"

	"github.com/gollilla/best/pkg/events"
)

// HealthAssertion provides health-related assertions
type HealthAssertion struct {
	agent AgentInterface
}

// ToBe checks if the health is exactly the expected value
func (h *HealthAssertion) ToBe(expected float32) {
	actual := h.agent.Health()

	if actual != expected {
		panic(NewAssertionError(
			fmt.Sprintf("expected health to be %.1f", expected),
			expected,
			actual,
		))
	}
}

// ToBeAbove checks if the health is above the minimum value
func (h *HealthAssertion) ToBeAbove(min float32) {
	actual := h.agent.Health()

	if actual <= min {
		panic(NewAssertionError(
			fmt.Sprintf("expected health to be above %.1f", min),
			fmt.Sprintf("> %.1f", min),
			actual,
		))
	}
}

// ToBeBelow checks if the health is below the maximum value
func (h *HealthAssertion) ToBeBelow(max float32) {
	actual := h.agent.Health()

	if actual >= max {
		panic(NewAssertionError(
			fmt.Sprintf("expected health to be below %.1f", max),
			fmt.Sprintf("< %.1f", max),
			actual,
		))
	}
}

// ToBeFull checks if the health is at maximum (20.0)
func (h *HealthAssertion) ToBeFull() {
	const maxHealth = 20.0
	actual := h.agent.Health()

	if actual != maxHealth {
		panic(NewAssertionError(
			"expected health to be full (20.0)",
			maxHealth,
			actual,
		))
	}
}

// ToChange waits for health to change within the timeout
func (h *HealthAssertion) ToChange(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := h.agent.Emitter().WaitFor(ctx, events.EventHealthUpdate, nil)
	if err != nil {
		panic(fmt.Errorf("health change event not received within %v: %w", timeout, err))
	}
}

// ToReach waits for health to reach a specific value within the timeout
func (h *HealthAssertion) ToReach(expected float32, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := h.agent.Emitter().WaitFor(ctx, events.EventHealthUpdate, func(d events.EventData) bool {
		health, ok := d.(float32)
		if !ok {
			return false
		}
		return health == expected
	})

	if err != nil {
		panic(fmt.Errorf("health did not reach %.1f within %v: %w", expected, timeout, err))
	}

	health := data.(float32)
	if health != expected {
		panic(NewAssertionError(
			fmt.Sprintf("expected health to reach %.1f", expected),
			expected,
			health,
		))
	}
}

// ToBeAboveWithin waits for health to be above a threshold within the timeout
func (h *HealthAssertion) ToBeAboveWithin(min float32, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := h.agent.Emitter().WaitFor(ctx, events.EventHealthUpdate, func(d events.EventData) bool {
		health, ok := d.(float32)
		if !ok {
			return false
		}
		return health > min
	})

	if err != nil {
		panic(fmt.Errorf("health did not go above %.1f within %v: %w", min, timeout, err))
	}

	health := data.(float32)
	if health <= min {
		panic(NewAssertionError(
			fmt.Sprintf("expected health to be above %.1f", min),
			fmt.Sprintf("> %.1f", min),
			health,
		))
	}
}

// ToBeBelowWithin waits for health to be below a threshold within the timeout
func (h *HealthAssertion) ToBeBelowWithin(max float32, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := h.agent.Emitter().WaitFor(ctx, events.EventHealthUpdate, func(d events.EventData) bool {
		health, ok := d.(float32)
		if !ok {
			return false
		}
		return health < max
	})

	if err != nil {
		panic(fmt.Errorf("health did not go below %.1f within %v: %w", max, timeout, err))
	}

	health := data.(float32)
	if health >= max {
		panic(NewAssertionError(
			fmt.Sprintf("expected health to be below %.1f", max),
			fmt.Sprintf("< %.1f", max),
			health,
		))
	}
}
