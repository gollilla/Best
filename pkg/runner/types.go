package runner

import (
	"time"
)

// TestStatus represents the result status of a test
type TestStatus string

const (
	TestStatusPassed  TestStatus = "passed"
	TestStatusFailed  TestStatus = "failed"
	TestStatusSkipped TestStatus = "skipped"
)

// ServerInfo contains information about the test server
type ServerInfo struct {
	Host    string
	Port    int
	Version string
}

// TestContext is passed to test functions
type TestContext struct {
	timeout time.Duration
}

// Timeout sets the timeout for the current test
func (c *TestContext) Timeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout returns the current timeout setting
func (c *TestContext) GetTimeout() time.Duration {
	return c.timeout
}

// TestFunction is the signature for test functions
type TestFunction func(ctx *TestContext)

// HookFunction is the signature for hook functions (beforeAll, afterAll, etc.)
type HookFunction func(ctx *TestContext)

// TestCase represents a single test
type TestCase struct {
	Name string
	Fn   TestFunction
	Skip bool
	Only bool
}

// TestSuite represents a collection of tests
type TestSuite struct {
	Name       string
	Tests      []*TestCase
	BeforeAll  []HookFunction
	AfterAll   []HookFunction
	BeforeEach []HookFunction
	AfterEach  []HookFunction
	Skip       bool
	Only       bool
}

// TestError contains information about test errors
type TestError struct {
	Message  string
	Stack    string
	Expected interface{}
	Actual   interface{}
}

// TestCaseResult represents the result of a single test
type TestCaseResult struct {
	Name     string
	Status   TestStatus
	Duration time.Duration
	Error    *TestError
}

// SuiteResult represents the result of a test suite
type SuiteResult struct {
	Name     string
	Tests    []*TestCaseResult
	Duration time.Duration
}

// TestResult represents the overall test results
type TestResult struct {
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
	Suites   []*SuiteResult
}

// TestRunnerOptions configures the test runner
type TestRunnerOptions struct {
	Timeout        time.Duration
	Parallel       bool
	MaxConcurrency int
	Reporter       Reporter
	Bail           bool
	Retries        int
}

// DefaultOptions returns default test runner options
func DefaultOptions() TestRunnerOptions {
	return TestRunnerOptions{
		Timeout:        30 * time.Second,
		Parallel:       false,
		MaxConcurrency: 4,
		Reporter:       NewConsoleReporter(),
		Bail:           false,
		Retries:        0,
	}
}
