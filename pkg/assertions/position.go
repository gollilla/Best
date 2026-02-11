package assertions

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// PositionAssertion provides position-related assertions
type PositionAssertion struct {
	agent AgentInterface
}

// ToBe asserts that the position exactly matches the expected position
func (p *PositionAssertion) ToBe(expected types.Position) {
	actual := p.agent.Position()

	if actual.X != expected.X || actual.Y != expected.Y || actual.Z != expected.Z {
		panic(NewAssertionError(
			fmt.Sprintf("Expected position to be (%.2f, %.2f, %.2f), but was (%.2f, %.2f, %.2f)",
				expected.X, expected.Y, expected.Z, actual.X, actual.Y, actual.Z),
			expected,
			actual,
		))
	}
}

// ToBeNear asserts that the position is within tolerance of the expected position
func (p *PositionAssertion) ToBeNear(expected types.Position, tolerance float64) {
	actual := p.agent.Position()
	distance := distanceTo(actual, expected)

	if distance > tolerance {
		panic(NewAssertionError(
			fmt.Sprintf("Expected position to be within %.2f of (%.2f, %.2f, %.2f), "+
				"but was (%.2f, %.2f, %.2f) (distance: %.2f)",
				tolerance, expected.X, expected.Y, expected.Z,
				actual.X, actual.Y, actual.Z, distance),
			expected,
			actual,
		))
	}
}

// ToBeWithin asserts that the position is within the given bounds
func (p *PositionAssertion) ToBeWithin(min, max types.Position) {
	actual := p.agent.Position()

	if actual.X < min.X || actual.X > max.X ||
		actual.Y < min.Y || actual.Y > max.Y ||
		actual.Z < min.Z || actual.Z > max.Z {
		panic(NewAssertionError(
			fmt.Sprintf("Expected position to be within (%.2f, %.2f, %.2f) - (%.2f, %.2f, %.2f), "+
				"but was (%.2f, %.2f, %.2f)",
				min.X, min.Y, min.Z, max.X, max.Y, max.Z,
				actual.X, actual.Y, actual.Z),
			map[string]types.Position{"min": min, "max": max},
			actual,
		))
	}
}

// ToBeAtY asserts that the Y position is within tolerance
func (p *PositionAssertion) ToBeAtY(y float64, tolerance float64) {
	if tolerance == 0 {
		tolerance = 0.5
	}

	actual := p.agent.Position().Y
	diff := math.Abs(actual - y)

	if diff > tolerance {
		panic(NewAssertionError(
			fmt.Sprintf("Expected Y position to be %.2f (±%.2f), but was %.2f", y, tolerance, actual),
			y,
			actual,
		))
	}
}

// ToBeOnGround asserts that the player is on the ground
func (p *PositionAssertion) ToBeOnGround() {
	if !p.agent.State().IsOnGround {
		panic(NewAssertionError(
			"Expected player to be on ground",
			"on ground",
			"in air",
		))
	}
}

// ToBeInAir asserts that the player is in the air
func (p *PositionAssertion) ToBeInAir() {
	if p.agent.State().IsOnGround {
		panic(NewAssertionError(
			"Expected player to be in air",
			"in air",
			"on ground",
		))
	}
}

// ToReach waits for the player to reach the expected position within tolerance
func (p *PositionAssertion) ToReach(ctx context.Context, expected types.Position, tolerance float64) {
	// Default timeout if not set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	if tolerance == 0 {
		tolerance = 1.0
	}

	// Check immediately
	if distanceTo(p.agent.Position(), expected) <= tolerance {
		return
	}

	// Wait for position update
	filter := func(data events.EventData) bool {
		return distanceTo(p.agent.Position(), expected) <= tolerance
	}

	_, err := p.agent.Emitter().WaitFor(ctx, events.EventPositionUpdate, filter)
	if err != nil {
		panic(NewAssertionError(
			fmt.Sprintf("Timeout waiting for position to reach (%.2f, %.2f, %.2f)",
				expected.X, expected.Y, expected.Z),
			expected,
			p.agent.Position(),
		))
	}
}

// ToBeInDimension asserts that the player is in the specified dimension
func (p *PositionAssertion) ToBeInDimension(dimension string) {
	actual := p.agent.State().Dimension

	if actual != dimension {
		panic(NewAssertionError(
			fmt.Sprintf("Expected to be in dimension %q, but was %q", dimension, actual),
			dimension,
			actual,
		))
	}
}

// distanceTo calculates the Euclidean distance between two positions
func distanceTo(a, b types.Position) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}
