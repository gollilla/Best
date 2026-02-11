package assertions

import (
	"context"
	"fmt"
	"time"

	"github.com/gollilla/best/pkg/events"
)

// PermissionAssertion provides permission-related assertions
type PermissionAssertion struct {
	agent AgentInterface
}

// Permission level constants
const (
	PermissionNormal    int32 = 0
	PermissionModerator int32 = 1
	PermissionOperator  int32 = 2 // Game master/operator
	PermissionAdmin     int32 = 3
	PermissionOwner     int32 = 4
)

// ToHaveLevel checks if the permission level is exactly the expected value
func (p *PermissionAssertion) ToHaveLevel(expected int32) {
	actual := p.agent.GetPermissionLevel()

	if actual != expected {
		panic(NewAssertionError(
			fmt.Sprintf("expected permission level to be %s (%d)", permissionName(expected), expected),
			permissionName(expected),
			permissionName(actual),
		))
	}
}

// ToBeAtLeast checks if the permission level is at least the minimum value
func (p *PermissionAssertion) ToBeAtLeast(min int32) {
	actual := p.agent.GetPermissionLevel()

	if actual < min {
		panic(NewAssertionError(
			fmt.Sprintf("expected permission level to be at least %s (%d)", permissionName(min), min),
			fmt.Sprintf(">= %s", permissionName(min)),
			permissionName(actual),
		))
	}
}

// ToBeOperator checks if the permission level is operator (2) or higher
func (p *PermissionAssertion) ToBeOperator() {
	p.ToBeAtLeast(PermissionOperator)
}

// ToBeNormal checks if the permission level is normal (0)
func (p *PermissionAssertion) ToBeNormal() {
	p.ToHaveLevel(PermissionNormal)
}

// ToBeModerator checks if the permission level is moderator (1)
func (p *PermissionAssertion) ToBeModerator() {
	p.ToHaveLevel(PermissionModerator)
}

// ToBeAdmin checks if the permission level is admin (3)
func (p *PermissionAssertion) ToBeAdmin() {
	p.ToHaveLevel(PermissionAdmin)
}

// ToBeOwner checks if the permission level is owner (4)
func (p *PermissionAssertion) ToBeOwner() {
	p.ToHaveLevel(PermissionOwner)
}

// ToChange waits for permission level to change within the timeout
func (p *PermissionAssertion) ToChange(timeout time.Duration) int32 {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := p.agent.Emitter().WaitFor(ctx, events.EventPermissionUpdate, nil)
	if err != nil {
		panic(err)
	}

	level, ok := data.(int32)
	if !ok {
		panic(fmt.Errorf("invalid permission level data type"))
	}

	return level
}

// ToChangeTo waits for permission level to change to a specific value within the timeout
func (p *PermissionAssertion) ToChangeTo(expected int32, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := p.agent.Emitter().WaitFor(ctx, events.EventPermissionUpdate, func(d events.EventData) bool {
		level, ok := d.(int32)
		if !ok {
			return false
		}
		return level == expected
	})

	if err != nil {
		panic(err)
	}

	level := data.(int32)
	if level != expected {
		panic(NewAssertionError(
			fmt.Sprintf("expected permission level to change to %s (%d)", permissionName(expected), expected),
			permissionName(expected),
			permissionName(level),
		))
	}
}

// Helper function

// permissionName returns the string name for a permission level value
func permissionName(level int32) string {
	switch level {
	case PermissionNormal:
		return "normal"
	case PermissionModerator:
		return "moderator"
	case PermissionOperator:
		return "operator"
	case PermissionAdmin:
		return "admin"
	case PermissionOwner:
		return "owner"
	default:
		return fmt.Sprintf("unknown(%d)", level)
	}
}
