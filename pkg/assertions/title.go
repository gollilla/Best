package assertions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// TitleAssertion provides title/subtitle/actionbar-related assertions
type TitleAssertion struct {
	agent AgentInterface
}

// ToReceive waits for a title to be received within the timeout
func (t *TitleAssertion) ToReceive(expected string, timeout time.Duration) {
	t.toReceiveType("title", expected, timeout)
}

// ToReceiveTitle waits for a title to be received (alias for ToReceive)
func (t *TitleAssertion) ToReceiveTitle(expected string, timeout time.Duration) {
	t.ToReceive(expected, timeout)
}

// ToReceiveSubtitle waits for a subtitle to be received within the timeout
func (t *TitleAssertion) ToReceiveSubtitle(expected string, timeout time.Duration) {
	t.toReceiveType("subtitle", expected, timeout)
}

// ToReceiveActionbar waits for an actionbar message to be received within the timeout
func (t *TitleAssertion) ToReceiveActionbar(expected string, timeout time.Duration) {
	t.toReceiveType("actionbar", expected, timeout)
}

// NotToReceive ensures no title is received within the timeout
func (t *TitleAssertion) NotToReceive(unexpected string, timeout time.Duration) {
	t.notToReceiveType("title", unexpected, timeout)
}

// NotToReceiveSubtitle ensures no subtitle is received within the timeout
func (t *TitleAssertion) NotToReceiveSubtitle(unexpected string, timeout time.Duration) {
	t.notToReceiveType("subtitle", unexpected, timeout)
}

// NotToReceiveActionbar ensures no actionbar message is received within the timeout
func (t *TitleAssertion) NotToReceiveActionbar(unexpected string, timeout time.Duration) {
	t.notToReceiveType("actionbar", unexpected, timeout)
}

// ToContain waits for a title containing the specified text
func (t *TitleAssertion) ToContain(text string, timeout time.Duration) {
	t.toContainType("title", text, timeout)
}

// ToContainSubtitle waits for a subtitle containing the specified text
func (t *TitleAssertion) ToContainSubtitle(text string, timeout time.Duration) {
	t.toContainType("subtitle", text, timeout)
}

// ToContainActionbar waits for an actionbar containing the specified text
func (t *TitleAssertion) ToContainActionbar(text string, timeout time.Duration) {
	t.toContainType("actionbar", text, timeout)
}

// toReceiveType is a helper for exact match assertions
func (t *TitleAssertion) toReceiveType(titleType, expected string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := t.agent.Emitter().WaitFor(ctx, events.EventTitle, func(d events.EventData) bool {
		titleDisplay, ok := d.(*types.TitleDisplay)
		if !ok {
			return false
		}
		return titleDisplay.Type == titleType && titleDisplay.Text == expected
	})

	if err != nil {
		panic(fmt.Errorf("%s not received within %v: %w", titleType, timeout, err))
	}

	titleDisplay := data.(*types.TitleDisplay)
	if titleDisplay.Text != expected {
		panic(NewAssertionError(
			fmt.Sprintf("expected %s to be %q", titleType, expected),
			expected,
			titleDisplay.Text,
		))
	}
}

// notToReceiveType is a helper for ensuring messages are not received
func (t *TitleAssertion) notToReceiveType(titleType, unexpected string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := t.agent.Emitter().WaitFor(ctx, events.EventTitle, func(d events.EventData) bool {
		titleDisplay, ok := d.(*types.TitleDisplay)
		if !ok {
			return false
		}
		return titleDisplay.Type == titleType && titleDisplay.Text == unexpected
	})

	if err == nil && data != nil {
		titleDisplay := data.(*types.TitleDisplay)
		panic(NewAssertionError(
			fmt.Sprintf("expected %s not to be %q", titleType, unexpected),
			fmt.Sprintf("not %q", unexpected),
			titleDisplay.Text,
		))
	}
}

// toContainType is a helper for partial match assertions
func (t *TitleAssertion) toContainType(titleType, text string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := t.agent.Emitter().WaitFor(ctx, events.EventTitle, func(d events.EventData) bool {
		titleDisplay, ok := d.(*types.TitleDisplay)
		if !ok {
			return false
		}
		return titleDisplay.Type == titleType && strings.Contains(titleDisplay.Text, text)
	})

	if err != nil {
		panic(fmt.Errorf("%s containing %q not received within %v: %w", titleType, text, timeout, err))
	}

	titleDisplay := data.(*types.TitleDisplay)
	if !strings.Contains(titleDisplay.Text, text) {
		panic(NewAssertionError(
			fmt.Sprintf("expected %s to contain %q", titleType, text),
			fmt.Sprintf("contains %q", text),
			titleDisplay.Text,
		))
	}
}
