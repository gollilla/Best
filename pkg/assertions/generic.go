package assertions

import (
	"fmt"
	"reflect"
	"strings"
)

// === Generic value assertions ===
// These assertions don't require an agent context

// IsTrue asserts that a condition is true
func IsTrue(condition bool, message string) {
	if !condition {
		panic(NewAssertionError(
			message,
			"true",
			"false",
		))
	}
}

// IsFalse asserts that a condition is false
func IsFalse(condition bool, message string) {
	if condition {
		panic(NewAssertionError(
			message,
			"false",
			"true",
		))
	}
}

// Equal asserts that two values are equal
func Equal(actual, expected interface{}, message string) {
	if !reflect.DeepEqual(actual, expected) {
		panic(NewAssertionError(
			message,
			fmt.Sprintf("%v", expected),
			fmt.Sprintf("%v", actual),
		))
	}
}

// NotEqual asserts that two values are not equal
func NotEqual(actual, expected interface{}, message string) {
	if reflect.DeepEqual(actual, expected) {
		panic(NewAssertionError(
			message,
			fmt.Sprintf("not %v", expected),
			fmt.Sprintf("%v", actual),
		))
	}
}

// IsNil asserts that a value is nil
func IsNil(value interface{}, message string) {
	if value != nil && !reflect.ValueOf(value).IsNil() {
		panic(NewAssertionError(
			message,
			"nil",
			fmt.Sprintf("%v", value),
		))
	}
}

// NotNil asserts that a value is not nil
func NotNil(value interface{}, message string) {
	if value == nil || reflect.ValueOf(value).IsNil() {
		panic(NewAssertionError(
			message,
			"not nil",
			"nil",
		))
	}
}

// === Numeric comparisons ===

// GreaterThan asserts that a numeric value is greater than another
func GreaterThan(actual, threshold float64, message string) {
	if actual <= threshold {
		panic(NewAssertionError(
			message,
			fmt.Sprintf("> %v", threshold),
			fmt.Sprintf("%v", actual),
		))
	}
}

// GreaterThanOrEqual asserts that a numeric value is greater than or equal to another
func GreaterThanOrEqual(actual, threshold float64, message string) {
	if actual < threshold {
		panic(NewAssertionError(
			message,
			fmt.Sprintf(">= %v", threshold),
			fmt.Sprintf("%v", actual),
		))
	}
}

// LessThan asserts that a numeric value is less than another
func LessThan(actual, threshold float64, message string) {
	if actual >= threshold {
		panic(NewAssertionError(
			message,
			fmt.Sprintf("< %v", threshold),
			fmt.Sprintf("%v", actual),
		))
	}
}

// LessThanOrEqual asserts that a numeric value is less than or equal to another
func LessThanOrEqual(actual, threshold float64, message string) {
	if actual > threshold {
		panic(NewAssertionError(
			message,
			fmt.Sprintf("<= %v", threshold),
			fmt.Sprintf("%v", actual),
		))
	}
}

// InRange asserts that a numeric value is within a range (inclusive)
func InRange(actual, min, max float64, message string) {
	if actual < min || actual > max {
		panic(NewAssertionError(
			message,
			fmt.Sprintf("between %v and %v", min, max),
			fmt.Sprintf("%v", actual),
		))
	}
}

// === String assertions ===

// Contains asserts that a string contains a substring
func Contains(str, substr string, message string) {
	if !strings.Contains(str, substr) {
		panic(NewAssertionError(
			message,
			fmt.Sprintf("contains '%s'", substr),
			fmt.Sprintf("'%s'", str),
		))
	}
}

// NotContains asserts that a string does not contain a substring
func NotContains(str, substr string, message string) {
	if strings.Contains(str, substr) {
		panic(NewAssertionError(
			message,
			fmt.Sprintf("does not contain '%s'", substr),
			fmt.Sprintf("'%s'", str),
		))
	}
}

// HasPrefix asserts that a string has a specific prefix
func HasPrefix(str, prefix string, message string) {
	if !strings.HasPrefix(str, prefix) {
		panic(NewAssertionError(
			message,
			fmt.Sprintf("starts with '%s'", prefix),
			fmt.Sprintf("'%s'", str),
		))
	}
}

// HasSuffix asserts that a string has a specific suffix
func HasSuffix(str, suffix string, message string) {
	if !strings.HasSuffix(str, suffix) {
		panic(NewAssertionError(
			message,
			fmt.Sprintf("ends with '%s'", suffix),
			fmt.Sprintf("'%s'", str),
		))
	}
}

// IsEmpty asserts that a string is empty
func IsEmpty(str string, message string) {
	if str != "" {
		panic(NewAssertionError(
			message,
			"empty string",
			fmt.Sprintf("'%s'", str),
		))
	}
}

// NotEmpty asserts that a string is not empty
func NotEmpty(str string, message string) {
	if str == "" {
		panic(NewAssertionError(
			message,
			"non-empty string",
			"empty string",
		))
	}
}

// === Collection assertions ===

// LengthEqual asserts that a slice/array/map has a specific length
func LengthEqual(collection interface{}, expectedLen int, message string) {
	v := reflect.ValueOf(collection)
	actualLen := v.Len()

	if actualLen != expectedLen {
		panic(NewAssertionError(
			message,
			fmt.Sprintf("length %d", expectedLen),
			fmt.Sprintf("length %d", actualLen),
		))
	}
}

// IsEmptyCollection asserts that a collection is empty
func IsEmptyCollection(collection interface{}, message string) {
	v := reflect.ValueOf(collection)
	if v.Len() != 0 {
		panic(NewAssertionError(
			message,
			"empty collection",
			fmt.Sprintf("length %d", v.Len()),
		))
	}
}

// NotEmptyCollection asserts that a collection is not empty
func NotEmptyCollection(collection interface{}, message string) {
	v := reflect.ValueOf(collection)
	if v.Len() == 0 {
		panic(NewAssertionError(
			message,
			"non-empty collection",
			"empty collection",
		))
	}
}

// ContainsElement asserts that a slice contains a specific element
func ContainsElement(slice interface{}, element interface{}, message string) {
	v := reflect.ValueOf(slice)

	for i := 0; i < v.Len(); i++ {
		if reflect.DeepEqual(v.Index(i).Interface(), element) {
			return
		}
	}

	panic(NewAssertionError(
		message,
		fmt.Sprintf("contains %v", element),
		fmt.Sprintf("does not contain %v", element),
	))
}
