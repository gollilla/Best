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

func (t *TitleAssertion) ToReceive(expected string, timeout time.Duration) {
	receiveDisplayType(t.agent, "title", expected, timeout)
}

func (t *TitleAssertion) NotToReceive(unexpected string, timeout time.Duration) {
	notReceiveDisplayType(t.agent, "title", unexpected, timeout)
}

func (t *TitleAssertion) ToContain(text string, timeout time.Duration) {
	containDisplayType(t.agent, "title", text, timeout)
}

// SubtitleAssertion provides subtitle-related assertions
type SubtitleAssertion struct {
	agent AgentInterface
}

func (s *SubtitleAssertion) ToReceive(expected string, timeout time.Duration) {
	receiveDisplayType(s.agent, "subtitle", expected, timeout)
}

func (s *SubtitleAssertion) NotToReceive(unexpected string, timeout time.Duration) {
	notReceiveDisplayType(s.agent, "subtitle", unexpected, timeout)
}

func (s *SubtitleAssertion) ToContain(text string, timeout time.Duration) {
	containDisplayType(s.agent, "subtitle", text, timeout)
}

// ActionbarAssertion provides actionbar-related assertions
type ActionbarAssertion struct {
	agent AgentInterface
}

func (a *ActionbarAssertion) ToReceive(expected string, timeout time.Duration) {
	receiveDisplayType(a.agent, "actionbar", expected, timeout)
}

func (a *ActionbarAssertion) NotToReceive(unexpected string, timeout time.Duration) {
	notReceiveDisplayType(a.agent, "actionbar", unexpected, timeout)
}

func (a *ActionbarAssertion) ToContain(text string, timeout time.Duration) {
	containDisplayType(a.agent, "actionbar", text, timeout)
}

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

func notReceiveDisplayType(agent AgentInterface, displayType, unexpected string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ch := make(chan *types.TitleDisplay, 1)
	listenerID := agent.Emitter().On(events.EventTitle, func(d events.EventData) {
		titleDisplay, ok := d.(*types.TitleDisplay)
		if !ok {
			return
		}
		if titleDisplay.Type == displayType && titleDisplay.Text == unexpected {
			select {
			case ch <- titleDisplay:
			default:
			}
		}
	})
	defer agent.Emitter().Off(events.EventTitle, listenerID)

	select {
	case <-ctx.Done():
		return
	case titleDisplay := <-ch:
		panic(NewAssertionError(
			fmt.Sprintf("expected %s not to be %q", displayType, unexpected),
			fmt.Sprintf("not %q", unexpected),
			titleDisplay.Text,
		))
	}
}

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
