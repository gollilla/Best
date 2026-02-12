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
	StatusCode *int32 // Filter by status code (nil means any)
}

// ToReceive waits for a CommandOutput matching the expected pattern
// Usage: agent.Expect().CommandOutput().ToReceive("text", 3*time.Second, nil)
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

		// Check status code filter
		if options.StatusCode != nil && output.StatusCode != *options.StatusCode {
			return false
		}

		// Check output content
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

	output := data.(*types.CommandOutput)
	return output
}

// ToReceiveAny waits for any CommandOutput to be received
// Usage: agent.Expect().CommandOutput().ToReceiveAny(3*time.Second)
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

	output := data.(*types.CommandOutput)
	return output
}

// ToReceiveWithContext waits for a CommandOutput with a custom context
func (c *CommandOutputAssertion) ToReceiveWithContext(ctx context.Context, expected interface{}, options *CommandOutputOptions) *types.CommandOutput {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

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

	output := data.(*types.CommandOutput)
	return output
}

// NotToReceive asserts that no matching CommandOutput is received within duration
func (c *CommandOutputAssertion) NotToReceive(ctx context.Context, pattern interface{}, duration time.Duration) {
	if duration == 0 {
		duration = 3 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	eventCh := make(chan events.EventData, 10)
	c.agent.Emitter().On(events.EventCommandOutput, func(data events.EventData) {
		select {
		case eventCh <- data:
		default:
		}
	})

	for {
		select {
		case <-ctx.Done():
			// Duration elapsed without matching output - success
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

// ToContain waits for a CommandOutput containing the expected text
// Usage: agent.Expect().CommandOutput().ToContain("success", 3*time.Second)
func (c *CommandOutputAssertion) ToContain(expected string, timeout time.Duration) *types.CommandOutput {
	return c.ToReceive(expected, timeout, nil)
}

// ToMatch waits for a CommandOutput matching the given regex pattern
// Usage: agent.Expect().CommandOutput().ToMatch(regexp.MustCompile(`\d+ players`), 3*time.Second)
func (c *CommandOutputAssertion) ToMatch(pattern *regexp.Regexp, timeout time.Duration) *types.CommandOutput {
	return c.ToReceive(pattern, timeout, nil)
}

// ToReceiveWithStatusCode waits for a CommandOutput with a specific status code
func (c *CommandOutputAssertion) ToReceiveWithStatusCode(statusCode int32, timeout time.Duration) *types.CommandOutput {
	options := &CommandOutputOptions{StatusCode: &statusCode}
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

	_ = options // Used for error message context
	output := data.(*types.CommandOutput)
	return output
}

// ToReceiveSuccess waits for a successful CommandOutput (StatusCode == 0)
func (c *CommandOutputAssertion) ToReceiveSuccess(timeout time.Duration) *types.CommandOutput {
	return c.ToReceiveWithStatusCode(0, timeout)
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
	c.agent.Emitter().On(events.EventCommandOutput, func(data events.EventData) {
		select {
		case eventCh <- data:
		default:
		}
	})

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

			pattern := expected[currentIndex]
			if matchesCommandOutputPattern(output.Output, pattern) {
				received = append(received, output)
				currentIndex++
			}
		}
	}

	return received
}

// matchesCommandOutputPattern checks if output matches the pattern (string or regexp)
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

// commandOutputsContent extracts output content from a slice of CommandOutputs
func commandOutputsContent(outputs []*types.CommandOutput) []string {
	content := make([]string, len(outputs))
	for i, output := range outputs {
		content[i] = output.Output
	}
	return content
}
