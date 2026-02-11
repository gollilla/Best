package best

import (
	"github.com/gollilla/best/pkg/runner"
)

// Global test runner instance
var globalRunner *runner.TestRunner

// NewRunner creates and returns a new test runner
// Users should create and manage their own agents in test hooks
func NewRunner(options *runner.TestRunnerOptions) *runner.TestRunner {
	globalRunner = runner.NewTestRunner(options)
	return globalRunner
}

// Describe defines a test suite using the global runner
func Describe(name string, fn func()) *runner.TestRunner {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.Describe(name, fn)
}

// Test defines a test case using the global runner
func Test(name string, fn runner.TestFunction) *runner.TestRunner {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.Test(name, fn)
}

// It is an alias for Test using the global runner
func It(name string, fn runner.TestFunction) *runner.TestRunner {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.It(name, fn)
}

// BeforeAll registers a hook to run before all tests using the global runner
func BeforeAll(fn runner.HookFunction) *runner.TestRunner {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.BeforeAll(fn)
}

// AfterAll registers a hook to run after all tests using the global runner
func AfterAll(fn runner.HookFunction) *runner.TestRunner {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.AfterAll(fn)
}

// BeforeEach registers a hook to run before each test using the global runner
func BeforeEach(fn runner.HookFunction) *runner.TestRunner {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.BeforeEach(fn)
}

// AfterEach registers a hook to run after each test using the global runner
func AfterEach(fn runner.HookFunction) *runner.TestRunner {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.AfterEach(fn)
}

// SkipTest defines a test case that should be skipped using the global runner
func SkipTest(name string, fn runner.TestFunction) *runner.TestRunner {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.SkipTest(name, fn)
}

// SkipDescribe defines a test suite that should be skipped using the global runner
func SkipDescribe(name string, fn func()) *runner.TestRunner {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.SkipDescribe(name, fn)
}

// OnlyTest defines a test case that should be run exclusively using the global runner
func OnlyTest(name string, fn runner.TestFunction) *runner.TestRunner {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.OnlyTest(name, fn)
}

// OnlyDescribe defines a test suite that should be run exclusively using the global runner
func OnlyDescribe(name string, fn func()) *runner.TestRunner {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.OnlyDescribe(name, fn)
}

// Run executes all registered test suites using the global runner
func Run() (*runner.TestResult, error) {
	if globalRunner == nil {
		panic("test runner not configured. Call NewRunner() first")
	}
	return globalRunner.Run()
}
