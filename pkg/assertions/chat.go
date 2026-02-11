package assertions

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// ChatAssertion provides chat-related assertions
type ChatAssertion struct {
	agent AgentInterface
}

// ToReceive waits for a chat message matching the expected pattern
func (c *ChatAssertion) ToReceive(ctx context.Context, expected interface{}, options *ChatOptions) *types.ChatMessage {
	// Default timeout
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	if options == nil {
		options = &ChatOptions{}
	}

	// Create filter
	filter := func(data events.EventData) bool {
		msg, ok := data.(*types.ChatMessage)
		if !ok {
			return false
		}

		// Check sender filter
		if options.From != "" && msg.Sender != options.From {
			return false
		}

		// Check message content
		return matchesPattern(msg.Message, expected)
	}

	// Wait for matching message
	data, err := c.agent.Emitter().WaitFor(ctx, events.EventChat, filter)
	if err != nil {
		fromStr := ""
		if options.From != "" {
			fromStr = fmt.Sprintf(" from %s", options.From)
		}
		panic(NewAssertionError(
			fmt.Sprintf("Timeout waiting for chat message matching %v%s", expected, fromStr),
			expected,
			nil,
		))
	}

	msg := data.(*types.ChatMessage)
	return msg
}

// NotToReceive asserts that no matching message is received within duration
func (c *ChatAssertion) NotToReceive(ctx context.Context, pattern interface{}, duration time.Duration) {
	if duration == 0 {
		duration = 3 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	// Listen for messages
	eventCh := make(chan events.EventData, 10)
	c.agent.Emitter().On(events.EventChat, func(data events.EventData) {
		select {
		case eventCh <- data:
		default:
		}
	})

	// Wait for duration or matching message
	for {
		select {
		case <-ctx.Done():
			// Duration elapsed without matching message - success
			return

		case data := <-eventCh:
			msg, ok := data.(*types.ChatMessage)
			if !ok {
				continue
			}

			if matchesPattern(msg.Message, pattern) {
				panic(NewAssertionError(
					fmt.Sprintf("Expected not to receive chat message matching %v, but received: %q",
						pattern, msg.Message),
					nil,
					msg.Message,
				))
			}
		}
	}
}

// ToReceiveSystem waits for a system message matching the expected pattern
func (c *ChatAssertion) ToReceiveSystem(ctx context.Context, expected interface{}) *types.ChatMessage {
	// Default timeout
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	filter := func(data events.EventData) bool {
		msg, ok := data.(*types.ChatMessage)
		if !ok {
			return false
		}

		// Check if system message
		if msg.Type != "system" {
			return false
		}

		return matchesPattern(msg.Message, expected)
	}

	data, err := c.agent.Emitter().WaitFor(ctx, events.EventChat, filter)
	if err != nil {
		panic(NewAssertionError(
			fmt.Sprintf("Timeout waiting for system message matching %v", expected),
			expected,
			nil,
		))
	}

	msg := data.(*types.ChatMessage)
	return msg
}

// ToReceiveInOrder waits for messages in the specified order
func (c *ChatAssertion) ToReceiveInOrder(ctx context.Context, expected []interface{}) []*types.ChatMessage {
	// Default timeout
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	received := make([]*types.ChatMessage, 0, len(expected))
	currentIndex := 0

	// Listen for messages
	eventCh := make(chan events.EventData, 10)
	c.agent.Emitter().On(events.EventChat, func(data events.EventData) {
		select {
		case eventCh <- data:
		default:
		}
	})

	for currentIndex < len(expected) {
		select {
		case <-ctx.Done():
			// Timeout
			panic(NewAssertionError(
				fmt.Sprintf("Timeout: only received %d/%d messages", len(received), len(expected)),
				expected,
				messagesContent(received),
			))

		case data := <-eventCh:
			msg, ok := data.(*types.ChatMessage)
			if !ok {
				continue
			}

			pattern := expected[currentIndex]
			if matchesPattern(msg.Message, pattern) {
				received = append(received, msg)
				currentIndex++
			}
		}
	}

	return received
}

// ChatOptions provides options for chat assertions
type ChatOptions struct {
	From string // Filter by sender
}

// matchesPattern checks if text matches the pattern (string or regexp)
func matchesPattern(text string, pattern interface{}) bool {
	switch p := pattern.(type) {
	case string:
		return strings.Contains(text, p)
	case *regexp.Regexp:
		return p.MatchString(text)
	default:
		return false
	}
}

// messagesContent extracts message content from a slice of chat messages
func messagesContent(messages []*types.ChatMessage) []string {
	content := make([]string, len(messages))
	for i, msg := range messages {
		content[i] = msg.Message
	}
	return content
}
