package assertions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gollilla/best/pkg/events"
)

// TagAssertion provides tag-related assertions
type TagAssertion struct {
	agent AgentInterface
}

// ToHave checks if the player has a specific tag
func (t *TagAssertion) ToHave(tag string) {
	tags := t.agent.GetTags()

	for _, existingTag := range tags {
		if existingTag == tag {
			return
		}
	}

	panic(NewAssertionError(
		fmt.Sprintf("expected player to have tag %q", tag),
		tag,
		tags,
	))
}

// NotToHave checks if the player does not have a specific tag
func (t *TagAssertion) NotToHave(tag string) {
	tags := t.agent.GetTags()

	for _, existingTag := range tags {
		if existingTag == tag {
			panic(NewAssertionError(
				fmt.Sprintf("expected player not to have tag %q", tag),
				fmt.Sprintf("not %q", tag),
				tag,
			))
		}
	}
}

// ToHaveAll checks if the player has all the specified tags
func (t *TagAssertion) ToHaveAll(expectedTags []string) {
	tags := t.agent.GetTags()
	tagMap := make(map[string]bool)
	for _, tag := range tags {
		tagMap[tag] = true
	}

	var missing []string
	for _, expectedTag := range expectedTags {
		if !tagMap[expectedTag] {
			missing = append(missing, expectedTag)
		}
	}

	if len(missing) > 0 {
		panic(NewAssertionError(
			fmt.Sprintf("expected player to have all tags %v, but missing: %v", expectedTags, missing),
			expectedTags,
			tags,
		))
	}
}

// ToHaveAny checks if the player has any of the specified tags
func (t *TagAssertion) ToHaveAny(expectedTags []string) {
	tags := t.agent.GetTags()
	tagMap := make(map[string]bool)
	for _, tag := range tags {
		tagMap[tag] = true
	}

	for _, expectedTag := range expectedTags {
		if tagMap[expectedTag] {
			return
		}
	}

	panic(NewAssertionError(
		fmt.Sprintf("expected player to have any of tags %v", expectedTags),
		expectedTags,
		tags,
	))
}

// ToHaveCount checks if the player has exactly the expected number of tags
func (t *TagAssertion) ToHaveCount(expected int) {
	tags := t.agent.GetTags()
	actual := len(tags)

	if actual != expected {
		panic(NewAssertionError(
			fmt.Sprintf("expected player to have %d tags, but found %d", expected, actual),
			expected,
			actual,
		))
	}
}

// ToHaveNone checks if the player has no tags
func (t *TagAssertion) ToHaveNone() {
	t.ToHaveCount(0)
}

// ToMatchPattern checks if the player has a tag matching the pattern (substring match)
func (t *TagAssertion) ToMatchPattern(pattern string) {
	tags := t.agent.GetTags()

	for _, tag := range tags {
		if strings.Contains(tag, pattern) {
			return
		}
	}

	panic(NewAssertionError(
		fmt.Sprintf("expected player to have a tag matching pattern %q", pattern),
		pattern,
		tags,
	))
}

// ToReceive waits for a specific tag to be added within the timeout
func (t *TagAssertion) ToReceive(tag string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := t.agent.Emitter().WaitFor(ctx, events.EventTagUpdate, func(d events.EventData) bool {
		tags, ok := d.([]string)
		if !ok {
			return false
		}

		for _, existingTag := range tags {
			if existingTag == tag {
				return true
			}
		}
		return false
	})

	if err != nil {
		panic(err)
	}

	tags := data.([]string)
	for _, existingTag := range tags {
		if existingTag == tag {
			return
		}
	}

	panic(NewAssertionError(
		fmt.Sprintf("received tag update but tag %q not found", tag),
		tag,
		tags,
	))
}

// ToLose waits for a specific tag to be removed within the timeout
func (t *TagAssertion) ToLose(tag string, timeout time.Duration) {
	// First check if the player currently has the tag
	tags := t.agent.GetTags()
	hasTag := false
	for _, existingTag := range tags {
		if existingTag == tag {
			hasTag = true
			break
		}
	}

	if !hasTag {
		// Player doesn't have the tag, so they can't lose it
		panic(NewAssertionError(
			fmt.Sprintf("expected player to lose tag %q, but they don't have it", tag),
			fmt.Sprintf("has and loses %q", tag),
			"doesn't have tag",
		))
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := t.agent.Emitter().WaitFor(ctx, events.EventTagUpdate, func(d events.EventData) bool {
		tags, ok := d.([]string)
		if !ok {
			return false
		}

		// Check if the tag is no longer in the list
		for _, existingTag := range tags {
			if existingTag == tag {
				return false // Tag still present
			}
		}
		return true // Tag removed
	})

	if err != nil {
		panic(err)
	}
}
