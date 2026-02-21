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

// CommandOutputAssertion provides CommandOutput-related assertions
type CommandOutputAssertion struct {
	agent AgentInterface
}

// CommandOutputOptions provides options for CommandOutput assertions
type CommandOutputOptions struct {
	StatusCode *int32
}

// ToReceive waits for a CommandOutput matching the expected pattern
func (c *CommandOutputAssertion) ToReceive(expected interface{}, timeout time.Duration, options *CommandOutputOptions) *types.CommandOutput {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if options == nil {
		options = &CommandOutputOptions{}
	}

	filter := func(data events.EventData) bool {
		output, ok := data.(*types.CommandOutput)
		if !ok {
			return false
		}
		if options.StatusCode != nil && output.StatusCode != *options.StatusCode {
			return false
		}
		return matchesCommandOutputPattern(output.Output, expected)
	}

	data, err := c.agent.Emitter().WaitFor(ctx, events.EventCommandOutput, filter)
	if err != nil {
		panic(NewAssertionError(
			fmt.Sprintf("Timeout waiting for CommandOutput matching %v", expected),
			expected,
			nil,
		))
	}

	return data.(*types.CommandOutput)
}

// ToReceiveAny waits for any CommandOutput to be received
func (c *CommandOutputAssertion) ToReceiveAny(timeout time.Duration) *types.CommandOutput {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := c.agent.Emitter().WaitFor(ctx, events.EventCommandOutput, nil)
	if err != nil {
		panic(NewAssertionError(
			"Timeout waiting for any CommandOutput",
			"any CommandOutput",
			nil,
		))
	}

	return data.(*types.CommandOutput)
}

// ToContain waits for a CommandOutput containing the expected text
func (c *CommandOutputAssertion) ToContain(expected string, timeout time.Duration) *types.CommandOutput {
	return c.ToReceive(expected, timeout, nil)
}

// ToMatch waits for a CommandOutput matching the given regex pattern
func (c *CommandOutputAssertion) ToMatch(pattern *regexp.Regexp, timeout time.Duration) *types.CommandOutput {
	return c.ToReceive(pattern, timeout, nil)
}

// ToReceiveWithStatusCode waits for a CommandOutput with a specific status code
func (c *CommandOutputAssertion) ToReceiveWithStatusCode(statusCode int32, timeout time.Duration) *types.CommandOutput {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	filter := func(data events.EventData) bool {
		output, ok := data.(*types.CommandOutput)
		if !ok {
			return false
		}
		return output.StatusCode == statusCode
	}

	data, err := c.agent.Emitter().WaitFor(ctx, events.EventCommandOutput, filter)
	if err != nil {
		panic(NewAssertionError(
			fmt.Sprintf("Timeout waiting for CommandOutput with status code %d", statusCode),
			statusCode,
			nil,
		))
	}

	return data.(*types.CommandOutput)
}

// ToReceiveSuccess waits for a successful CommandOutput (StatusCode == 0)
func (c *CommandOutputAssertion) ToReceiveSuccess(timeout time.Duration) *types.CommandOutput {
	return c.ToReceiveWithStatusCode(0, timeout)
}

// NotToReceive asserts that no matching CommandOutput is received within duration
func (c *CommandOutputAssertion) NotToReceive(ctx context.Context, pattern interface{}, duration time.Duration) {
	if duration == 0 {
		duration = 3 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	eventCh := make(chan events.EventData, 10)
	listenerID := c.agent.Emitter().On(events.EventCommandOutput, func(data events.EventData) {
		select {
		case eventCh <- data:
		default:
		}
	})
	defer c.agent.Emitter().Off(events.EventCommandOutput, listenerID)

	for {
		select {
		case <-ctx.Done():
			return

		case data := <-eventCh:
			output, ok := data.(*types.CommandOutput)
			if !ok {
				continue
			}
			if matchesCommandOutputPattern(output.Output, pattern) {
				panic(NewAssertionError(
					fmt.Sprintf("Expected not to receive CommandOutput matching %v, but received: %q",
						pattern, output.Output),
					nil,
					output.Output,
				))
			}
		}
	}
}

// ToReceiveInOrder waits for CommandOutputs in the specified order
func (c *CommandOutputAssertion) ToReceiveInOrder(ctx context.Context, expected []interface{}) []*types.CommandOutput {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	received := make([]*types.CommandOutput, 0, len(expected))
	currentIndex := 0

	eventCh := make(chan events.EventData, 10)
	listenerID := c.agent.Emitter().On(events.EventCommandOutput, func(data events.EventData) {
		select {
		case eventCh <- data:
		default:
		}
	})
	defer c.agent.Emitter().Off(events.EventCommandOutput, listenerID)

	for currentIndex < len(expected) {
		select {
		case <-ctx.Done():
			panic(NewAssertionError(
				fmt.Sprintf("Timeout: only received %d/%d CommandOutputs", len(received), len(expected)),
				expected,
				commandOutputsContent(received),
			))

		case data := <-eventCh:
			output, ok := data.(*types.CommandOutput)
			if !ok {
				continue
			}
			if matchesCommandOutputPattern(output.Output, expected[currentIndex]) {
				received = append(received, output)
				currentIndex++
			}
		}
	}

	return received
}

func matchesCommandOutputPattern(text string, pattern interface{}) bool {
	switch p := pattern.(type) {
	case string:
		return strings.Contains(text, p)
	case *regexp.Regexp:
		return p.MatchString(text)
	default:
		return false
	}
}

func commandOutputsContent(outputs []*types.CommandOutput) []string {
	content := make([]string, len(outputs))
	for i, output := range outputs {
		content[i] = output.Output
	}
	return content
}
