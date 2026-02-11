package assertions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gollilla/best/pkg/types"
)

// CommandAssertion provides command output assertions
type CommandAssertion struct {
	output *types.CommandOutput
}

// ToSucceed asserts that the command succeeded
func (c *CommandAssertion) ToSucceed() *CommandAssertion {
	if !c.output.Success {
		panic(NewAssertionError(
			fmt.Sprintf("Expected command %q to succeed, but it failed", c.output.Command),
			"success",
			"failure",
		))
	}
	return c
}

// ToFail asserts that the command failed
func (c *CommandAssertion) ToFail() *CommandAssertion {
	if c.output.Success {
		panic(NewAssertionError(
			fmt.Sprintf("Expected command %q to fail, but it succeeded", c.output.Command),
			"failure",
			"success",
		))
	}
	return c
}

// ToContain asserts that the command output contains the expected text
func (c *CommandAssertion) ToContain(expected interface{}) *CommandAssertion {
	var matches bool

	switch e := expected.(type) {
	case string:
		matches = strings.Contains(c.output.Output, e)
	case *regexp.Regexp:
		matches = e.MatchString(c.output.Output)
	default:
		panic(NewAssertionError(
			"ToContain expects string or *regexp.Regexp",
			"string or *regexp.Regexp",
			fmt.Sprintf("%T", expected),
		))
	}

	if !matches {
		panic(NewAssertionError(
			fmt.Sprintf("Expected command output to contain %v, but output was: %q",
				expected, c.output.Output),
			expected,
			c.output.Output,
		))
	}

	return c
}

// ToHaveStatusCode asserts that the command has the expected status code
func (c *CommandAssertion) ToHaveStatusCode(code int32) *CommandAssertion {
	if c.output.StatusCode != code {
		panic(NewAssertionError(
			fmt.Sprintf("Expected status code %d, but was %d", code, c.output.StatusCode),
			code,
			c.output.StatusCode,
		))
	}
	return c
}

// And returns the assertion for chaining
func (c *CommandAssertion) And() *CommandAssertion {
	return c
}
