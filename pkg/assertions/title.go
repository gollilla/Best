package assertions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// TitleAssertion provides title-related assertions
type TitleAssertion struct {
	agent AgentInterface
}

// ToReceive waits for a title to be received within the timeout
func (t *TitleAssertion) ToReceive(expected string, timeout time.Duration) {
	receiveDisplayType(t.agent, "title", expected, timeout)
}

// NotToReceive ensures no title is received within the timeout
func (t *TitleAssertion) NotToReceive(unexpected string, timeout time.Duration) {
	notReceiveDisplayType(t.agent, "title", unexpected, timeout)
}

// ToContain waits for a title containing the specified text
func (t *TitleAssertion) ToContain(text string, timeout time.Duration) {
	containDisplayType(t.agent, "title", text, timeout)
}

// SubtitleAssertion provides subtitle-related assertions
type SubtitleAssertion struct {
	agent AgentInterface
}

// ToReceive waits for a subtitle to be received within the timeout
func (s *SubtitleAssertion) ToReceive(expected string, timeout time.Duration) {
	receiveDisplayType(s.agent, "subtitle", expected, timeout)
}

// NotToReceive ensures no subtitle is received within the timeout
func (s *SubtitleAssertion) NotToReceive(unexpected string, timeout time.Duration) {
	notReceiveDisplayType(s.agent, "subtitle", unexpected, timeout)
}

// ToContain waits for a subtitle containing the specified text
func (s *SubtitleAssertion) ToContain(text string, timeout time.Duration) {
	containDisplayType(s.agent, "subtitle", text, timeout)
}

// ActionbarAssertion provides actionbar-related assertions
type ActionbarAssertion struct {
	agent AgentInterface
}

// ToReceive waits for an actionbar message to be received within the timeout
func (a *ActionbarAssertion) ToReceive(expected string, timeout time.Duration) {
	receiveDisplayType(a.agent, "actionbar", expected, timeout)
}

// NotToReceive ensures no actionbar message is received within the timeout
func (a *ActionbarAssertion) NotToReceive(unexpected string, timeout time.Duration) {
	notReceiveDisplayType(a.agent, "actionbar", unexpected, timeout)
}

// ToContain waits for an actionbar containing the specified text
func (a *ActionbarAssertion) ToContain(text string, timeout time.Duration) {
	containDisplayType(a.agent, "actionbar", text, timeout)
}

// receiveDisplayType is a helper for exact match assertions
func receiveDisplayType(agent AgentInterface, displayType, expected string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := agent.Emitter().WaitFor(ctx, events.EventTitle, func(d events.EventData) bool {
		titleDisplay, ok := d.(*types.TitleDisplay)
		if !ok {
			return false
		}
		return titleDisplay.Type == displayType && titleDisplay.Text == expected
	})

	if err != nil {
		panic(fmt.Errorf("%s not received within %v: %w", displayType, timeout, err))
	}

	titleDisplay := data.(*types.TitleDisplay)
	if titleDisplay.Text != expected {
		panic(NewAssertionError(
			fmt.Sprintf("expected %s to be %q", displayType, expected),
			expected,
			titleDisplay.Text,
		))
	}
}

// notReceiveDisplayType is a helper for ensuring messages are not received
func notReceiveDisplayType(agent AgentInterface, displayType, unexpected string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := agent.Emitter().WaitFor(ctx, events.EventTitle, func(d events.EventData) bool {
		titleDisplay, ok := d.(*types.TitleDisplay)
		if !ok {
			return false
		}
		return titleDisplay.Type == displayType && titleDisplay.Text == unexpected
	})

	if err == nil && data != nil {
		titleDisplay := data.(*types.TitleDisplay)
		panic(NewAssertionError(
			fmt.Sprintf("expected %s not to be %q", displayType, unexpected),
			fmt.Sprintf("not %q", unexpected),
			titleDisplay.Text,
		))
	}
}

// containDisplayType is a helper for partial match assertions
func containDisplayType(agent AgentInterface, displayType, text string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := agent.Emitter().WaitFor(ctx, events.EventTitle, func(d events.EventData) bool {
		titleDisplay, ok := d.(*types.TitleDisplay)
		if !ok {
			return false
		}
		return titleDisplay.Type == displayType && strings.Contains(titleDisplay.Text, text)
	})

	if err != nil {
		panic(fmt.Errorf("%s containing %q not received within %v: %w", displayType, text, timeout, err))
	}

	titleDisplay := data.(*types.TitleDisplay)
	if !strings.Contains(titleDisplay.Text, text) {
		panic(NewAssertionError(
			fmt.Sprintf("expected %s to contain %q", displayType, text),
			fmt.Sprintf("contains %q", text),
			titleDisplay.Text,
		))
	}
}
