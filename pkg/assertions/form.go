package assertions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gollilla/best/pkg/events"
	"github.com/gollilla/best/pkg/types"
)

// FormAssertion provides form-related assertions
type FormAssertion struct {
	agent AgentInterface
	form  types.Form
}

// ToReceive waits for a form to be received within the timeout
func (f *FormAssertion) ToReceive(timeout time.Duration) *FormAssertion {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := f.agent.Emitter().WaitFor(ctx, events.EventForm, nil)
	if err != nil {
		panic(NewAssertionError(
			fmt.Sprintf("Expected to receive form within %v, but timed out", timeout),
			"form received",
			"timeout",
		))
	}

	form, ok := data.(types.Form)
	if !ok {
		panic(NewAssertionError(
			"Expected form data to be types.Form",
			"types.Form",
			fmt.Sprintf("%T", data),
		))
	}

	f.form = form
	return f
}

// ToReceiveWithTitle waits for a form with the specific title within the timeout
func (f *FormAssertion) ToReceiveWithTitle(title string, timeout time.Duration) *FormAssertion {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	data, err := f.agent.Emitter().WaitFor(ctx, events.EventForm, func(d events.EventData) bool {
		form, ok := d.(types.Form)
		if !ok {
			return false
		}
		return form.GetTitle() == title
	})

	if err != nil {
		panic(NewAssertionError(
			fmt.Sprintf("Expected to receive form with title %q within %v, but timed out", title, timeout),
			fmt.Sprintf("form with title %q", title),
			"timeout",
		))
	}

	form, _ := data.(types.Form)
	f.form = form
	return f
}

// ToBeModal asserts that the form is a ModalForm
func (f *FormAssertion) ToBeModal() *FormAssertion {
	if f.form == nil {
		panic(NewAssertionError(
			"No form received yet. Call ToReceive() first",
			"form received",
			"nil",
		))
	}

	if f.form.GetType() != "modal" {
		panic(NewAssertionError(
			fmt.Sprintf("Expected form to be modal, but was %s", f.form.GetType()),
			"modal",
			f.form.GetType(),
		))
	}

	return f
}

// ToBeActionForm asserts that the form is an ActionForm
func (f *FormAssertion) ToBeActionForm() *FormAssertion {
	if f.form == nil {
		panic(NewAssertionError(
			"No form received yet. Call ToReceive() first",
			"form received",
			"nil",
		))
	}

	if f.form.GetType() != "action" {
		panic(NewAssertionError(
			fmt.Sprintf("Expected form to be action form, but was %s", f.form.GetType()),
			"action",
			f.form.GetType(),
		))
	}

	return f
}

// ToBeCustomForm asserts that the form is a CustomForm
func (f *FormAssertion) ToBeCustomForm() *FormAssertion {
	if f.form == nil {
		panic(NewAssertionError(
			"No form received yet. Call ToReceive() first",
			"form received",
			"nil",
		))
	}

	if f.form.GetType() != "form" {
		panic(NewAssertionError(
			fmt.Sprintf("Expected form to be custom form, but was %s", f.form.GetType()),
			"form",
			f.form.GetType(),
		))
	}

	return f
}

// ToHaveTitle asserts that the form has the expected title
func (f *FormAssertion) ToHaveTitle(expected string) *FormAssertion {
	if f.form == nil {
		panic(NewAssertionError(
			"No form received yet. Call ToReceive() first",
			"form received",
			"nil",
		))
	}

	actual := f.form.GetTitle()
	if actual != expected {
		panic(NewAssertionError(
			fmt.Sprintf("Expected form title to be %q, but was %q", expected, actual),
			expected,
			actual,
		))
	}

	return f
}

// ToContainTitle asserts that the form title contains the expected text
func (f *FormAssertion) ToContainTitle(expected string) *FormAssertion {
	if f.form == nil {
		panic(NewAssertionError(
			"No form received yet. Call ToReceive() first",
			"form received",
			"nil",
		))
	}

	actual := f.form.GetTitle()
	if !strings.Contains(actual, expected) {
		panic(NewAssertionError(
			fmt.Sprintf("Expected form title to contain %q, but was %q", expected, actual),
			fmt.Sprintf("title containing %q", expected),
			actual,
		))
	}

	return f
}

// ToHaveButton asserts that the action form has a button with the expected text
func (f *FormAssertion) ToHaveButton(buttonText string) *FormAssertion {
	if f.form == nil {
		panic(NewAssertionError(
			"No form received yet. Call ToReceive() first",
			"form received",
			"nil",
		))
	}

	actionForm, ok := f.form.(*types.ActionForm)
	if !ok {
		panic(NewAssertionError(
			"ToHaveButton can only be used with ActionForm",
			"ActionForm",
			f.form.GetType(),
		))
	}

	for _, btn := range actionForm.Buttons {
		if btn.Text == buttonText {
			return f
		}
	}

	panic(NewAssertionError(
		fmt.Sprintf("Expected form to have button %q, but it was not found", buttonText),
		fmt.Sprintf("button %q", buttonText),
		"not found",
	))
}

// ToHaveButtons asserts that the action form has the expected number of buttons
func (f *FormAssertion) ToHaveButtons(count int) *FormAssertion {
	if f.form == nil {
		panic(NewAssertionError(
			"No form received yet. Call ToReceive() first",
			"form received",
			"nil",
		))
	}

	actionForm, ok := f.form.(*types.ActionForm)
	if !ok {
		panic(NewAssertionError(
			"ToHaveButtons can only be used with ActionForm",
			"ActionForm",
			f.form.GetType(),
		))
	}

	actual := len(actionForm.Buttons)
	if actual != count {
		panic(NewAssertionError(
			fmt.Sprintf("Expected form to have %d buttons, but had %d", count, actual),
			count,
			actual,
		))
	}

	return f
}

// ToHaveContent asserts that the form has the expected content text
func (f *FormAssertion) ToHaveContent(expected string) *FormAssertion {
	if f.form == nil {
		panic(NewAssertionError(
			"No form received yet. Call ToReceive() first",
			"form received",
			"nil",
		))
	}

	var content string
	switch form := f.form.(type) {
	case *types.ModalForm:
		content = form.Content
	case *types.ActionForm:
		content = form.Content
	default:
		panic(NewAssertionError(
			"ToHaveContent can only be used with ModalForm or ActionForm",
			"ModalForm or ActionForm",
			f.form.GetType(),
		))
	}

	if content != expected {
		panic(NewAssertionError(
			fmt.Sprintf("Expected form content to be %q, but was %q", expected, content),
			expected,
			content,
		))
	}

	return f
}

// GetForm returns the current form being asserted
func (f *FormAssertion) GetForm() types.Form {
	return f.form
}

// And returns the assertion for chaining
func (f *FormAssertion) And() *FormAssertion {
	return f
}
