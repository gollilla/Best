package agent

import (
	"github.com/gollilla/best/pkg/assertions"
)

// Expect returns an assertion context for this agent
func (a *Agent) Expect() *assertions.AssertionContext {
	return assertions.NewAssertionContext(a)
}
