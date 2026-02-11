package runner

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gollilla/best/pkg/assertions"
)

// TestRunner manages and executes test suites
type TestRunner struct {
	options          TestRunnerOptions
	suites           []*TestSuite
	currentSuite     *TestSuite
	globalBeforeAll  []HookFunction
	globalAfterAll   []HookFunction
	globalBeforeEach []HookFunction
	globalAfterEach  []HookFunction
}

// NewTestRunner creates a new test runner
func NewTestRunner(options *TestRunnerOptions) *TestRunner {
	opts := DefaultOptions()
	if options != nil {
		if options.Timeout > 0 {
			opts.Timeout = options.Timeout
		}
		if options.Reporter != nil {
			opts.Reporter = options.Reporter
		}
		opts.Parallel = options.Parallel
		if options.MaxConcurrency > 0 {
			opts.MaxConcurrency = options.MaxConcurrency
		}
		opts.Bail = options.Bail
		if options.Retries > 0 {
			opts.Retries = options.Retries
		}
	}

	return &TestRunner{
		options:          opts,
		suites:           make([]*TestSuite, 0),
		globalBeforeAll:  make([]HookFunction, 0),
		globalAfterAll:   make([]HookFunction, 0),
		globalBeforeEach: make([]HookFunction, 0),
		globalAfterEach:  make([]HookFunction, 0),
	}
}

// Describe defines a test suite
func (r *TestRunner) Describe(name string, fn func()) *TestRunner {
	suite := &TestSuite{
		Name:       name,
		Tests:      make([]*TestCase, 0),
		BeforeAll:  make([]HookFunction, 0),
		AfterAll:   make([]HookFunction, 0),
		BeforeEach: make([]HookFunction, 0),
		AfterEach:  make([]HookFunction, 0),
	}

	prevSuite := r.currentSuite
	r.currentSuite = suite
	fn()
	r.currentSuite = prevSuite

	r.suites = append(r.suites, suite)
	return r
}

// Test defines a test case
func (r *TestRunner) Test(name string, fn TestFunction) *TestRunner {
	testCase := &TestCase{
		Name: name,
		Fn:   fn,
	}

	if r.currentSuite != nil {
		r.currentSuite.Tests = append(r.currentSuite.Tests, testCase)
	} else {
		// Create implicit suite for orphan tests
		implicitSuite := &TestSuite{
			Name:       "",
			Tests:      []*TestCase{testCase},
			BeforeAll:  make([]HookFunction, 0),
			AfterAll:   make([]HookFunction, 0),
			BeforeEach: make([]HookFunction, 0),
			AfterEach:  make([]HookFunction, 0),
		}
		r.suites = append(r.suites, implicitSuite)
	}

	return r
}

// It is an alias for Test
func (r *TestRunner) It(name string, fn TestFunction) *TestRunner {
	return r.Test(name, fn)
}

// BeforeAll registers a hook to run before all tests
func (r *TestRunner) BeforeAll(fn HookFunction) *TestRunner {
	if r.currentSuite != nil {
		r.currentSuite.BeforeAll = append(r.currentSuite.BeforeAll, fn)
	} else {
		r.globalBeforeAll = append(r.globalBeforeAll, fn)
	}
	return r
}

// AfterAll registers a hook to run after all tests
func (r *TestRunner) AfterAll(fn HookFunction) *TestRunner {
	if r.currentSuite != nil {
		r.currentSuite.AfterAll = append(r.currentSuite.AfterAll, fn)
	} else {
		r.globalAfterAll = append(r.globalAfterAll, fn)
	}
	return r
}

// BeforeEach registers a hook to run before each test
func (r *TestRunner) BeforeEach(fn HookFunction) *TestRunner {
	if r.currentSuite != nil {
		r.currentSuite.BeforeEach = append(r.currentSuite.BeforeEach, fn)
	} else {
		r.globalBeforeEach = append(r.globalBeforeEach, fn)
	}
	return r
}

// AfterEach registers a hook to run after each test
func (r *TestRunner) AfterEach(fn HookFunction) *TestRunner {
	if r.currentSuite != nil {
		r.currentSuite.AfterEach = append(r.currentSuite.AfterEach, fn)
	} else {
		r.globalAfterEach = append(r.globalAfterEach, fn)
	}
	return r
}

// SkipTest defines a test case that should be skipped
func (r *TestRunner) SkipTest(name string, fn TestFunction) *TestRunner {
	testCase := &TestCase{
		Name: name,
		Fn:   fn,
		Skip: true,
	}

	if r.currentSuite != nil {
		r.currentSuite.Tests = append(r.currentSuite.Tests, testCase)
	}
	return r
}

// SkipDescribe defines a test suite that should be skipped
func (r *TestRunner) SkipDescribe(name string, fn func()) *TestRunner {
	suite := &TestSuite{
		Name:       name,
		Tests:      make([]*TestCase, 0),
		BeforeAll:  make([]HookFunction, 0),
		AfterAll:   make([]HookFunction, 0),
		BeforeEach: make([]HookFunction, 0),
		AfterEach:  make([]HookFunction, 0),
		Skip:       true,
	}

	prevSuite := r.currentSuite
	r.currentSuite = suite
	fn()
	r.currentSuite = prevSuite

	r.suites = append(r.suites, suite)
	return r
}

// OnlyTest defines a test case that should be run exclusively
func (r *TestRunner) OnlyTest(name string, fn TestFunction) *TestRunner {
	testCase := &TestCase{
		Name: name,
		Fn:   fn,
		Only: true,
	}

	if r.currentSuite != nil {
		r.currentSuite.Tests = append(r.currentSuite.Tests, testCase)
	}
	return r
}

// OnlyDescribe defines a test suite that should be run exclusively
func (r *TestRunner) OnlyDescribe(name string, fn func()) *TestRunner {
	suite := &TestSuite{
		Name:       name,
		Tests:      make([]*TestCase, 0),
		BeforeAll:  make([]HookFunction, 0),
		AfterAll:   make([]HookFunction, 0),
		BeforeEach: make([]HookFunction, 0),
		AfterEach:  make([]HookFunction, 0),
		Only:       true,
	}

	prevSuite := r.currentSuite
	r.currentSuite = suite
	fn()
	r.currentSuite = prevSuite

	r.suites = append(r.suites, suite)
	return r
}

// Run executes all registered test suites
func (r *TestRunner) Run() (*TestResult, error) {
	result := &TestResult{
		Passed:   0,
		Failed:   0,
		Skipped:  0,
		Duration: 0,
		Suites:   make([]*SuiteResult, 0),
	}

	startTime := time.Now()
	r.options.Reporter.OnStart(len(r.suites))

	// Check for "only" tests
	hasOnly := r.hasOnlyTests()

	// Create global context
	globalCtx := r.createContext()

	// Run global beforeAll hooks
	if err := r.runHooks(r.globalBeforeAll, globalCtx); err != nil {
		return nil, fmt.Errorf("global beforeAll hook failed: %w", err)
	}

	// Run test suites
	for _, suite := range r.suites {
		suiteResult := r.runSuite(suite, hasOnly, globalCtx)
		result.Suites = append(result.Suites, suiteResult)

		for _, test := range suiteResult.Tests {
			switch test.Status {
			case TestStatusPassed:
				result.Passed++
			case TestStatusFailed:
				result.Failed++
			case TestStatusSkipped:
				result.Skipped++
			}
		}

		if r.options.Bail && result.Failed > 0 {
			break
		}
	}

	// Run global afterAll hooks (ignore errors)
	_ = r.runHooks(r.globalAfterAll, globalCtx)

	result.Duration = time.Since(startTime)
	r.options.Reporter.OnEnd(result)

	return result, nil
}

func (r *TestRunner) createContext() *TestContext {
	return &TestContext{
		timeout: r.options.Timeout,
	}
}

func (r *TestRunner) hasOnlyTests() bool {
	for _, suite := range r.suites {
		if suite.Only {
			return true
		}
		for _, test := range suite.Tests {
			if test.Only {
				return true
			}
		}
	}
	return false
}

func (r *TestRunner) runSuite(suite *TestSuite, hasOnly bool, globalCtx *TestContext) *SuiteResult {
	suiteResult := &SuiteResult{
		Name:     suite.Name,
		Tests:    make([]*TestCaseResult, 0),
		Duration: 0,
	}

	startTime := time.Now()
	r.options.Reporter.OnSuiteStart(suite.Name)

	// Skip if needed
	if suite.Skip || (hasOnly && !suite.Only && !r.hasSuiteOnlyTest(suite)) {
		for _, test := range suite.Tests {
			suiteResult.Tests = append(suiteResult.Tests, &TestCaseResult{
				Name:     test.Name,
				Status:   TestStatusSkipped,
				Duration: 0,
			})
			r.options.Reporter.OnTestSkip(test.Name)
		}
		suiteResult.Duration = time.Since(startTime)
		r.options.Reporter.OnSuiteEnd(suite.Name, suiteResult)
		return suiteResult
	}

	// Run beforeAll hooks
	if err := r.runHooks(suite.BeforeAll, globalCtx); err != nil {
		// If beforeAll fails, mark all tests as failed
		testErr := r.toTestError(err)
		for _, test := range suite.Tests {
			suiteResult.Tests = append(suiteResult.Tests, &TestCaseResult{
				Name:     test.Name,
				Status:   TestStatusFailed,
				Duration: 0,
				Error:    testErr,
			})
		}
		suiteResult.Duration = time.Since(startTime)
		return suiteResult
	}

	// Run tests
	for _, test := range suite.Tests {
		testResult := r.runTest(test, suite, hasOnly, globalCtx)
		suiteResult.Tests = append(suiteResult.Tests, testResult)

		if r.options.Bail && testResult.Status == TestStatusFailed {
			break
		}
	}

	// Run afterAll hooks (ignore errors)
	_ = r.runHooks(suite.AfterAll, globalCtx)

	suiteResult.Duration = time.Since(startTime)
	r.options.Reporter.OnSuiteEnd(suite.Name, suiteResult)
	return suiteResult
}

func (r *TestRunner) hasSuiteOnlyTest(suite *TestSuite) bool {
	for _, test := range suite.Tests {
		if test.Only {
			return true
		}
	}
	return false
}

func (r *TestRunner) runTest(test *TestCase, suite *TestSuite, hasOnly bool, ctx *TestContext) *TestCaseResult {
	// Skip logic
	if test.Skip || (hasOnly && !test.Only && !suite.Only) {
		r.options.Reporter.OnTestSkip(test.Name)
		return &TestCaseResult{
			Name:     test.Name,
			Status:   TestStatusSkipped,
			Duration: 0,
		}
	}

	r.options.Reporter.OnTestStart(test.Name)
	startTime := time.Now()

	var lastError interface{}
	maxAttempts := r.options.Retries + 1

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := r.executeTest(test, suite, ctx)

		if err == nil {
			duration := time.Since(startTime)
			r.options.Reporter.OnTestPass(test.Name, duration.Milliseconds())
			return &TestCaseResult{
				Name:     test.Name,
				Status:   TestStatusPassed,
				Duration: duration,
			}
		}

		lastError = err

		if attempt < maxAttempts {
			r.options.Reporter.OnTestRetry(test.Name, attempt)
		}
	}

	duration := time.Since(startTime)
	testErr := r.toTestError(lastError)
	r.options.Reporter.OnTestFail(test.Name, testErr, duration.Milliseconds())
	return &TestCaseResult{
		Name:     test.Name,
		Status:   TestStatusFailed,
		Duration: duration,
		Error:    testErr,
	}
}

func (r *TestRunner) executeTest(test *TestCase, suite *TestSuite, ctx *TestContext) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%v", recovered)
		}
	}()

	// Run beforeEach hooks
	allBeforeEach := append(r.globalBeforeEach, suite.BeforeEach...)
	if err := r.runHooks(allBeforeEach, ctx); err != nil {
		return err
	}

	// Run test with timeout
	done := make(chan struct{})
	var testErr error

	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				// Check if it's an AssertionError (clean error message without stack)
				if _, ok := recovered.(*assertions.AssertionError); ok {
					testErr = fmt.Errorf("%v", recovered)
				} else if err, ok := recovered.(error); ok {
					// Regular error, include stack trace
					testErr = fmt.Errorf("%v\nStack: %s", err, string(debug.Stack()))
				} else {
					// Unknown panic type, include stack trace
					testErr = fmt.Errorf("%v\nStack: %s", recovered, string(debug.Stack()))
				}
			}
			close(done)
		}()

		test.Fn(ctx)
	}()

	select {
	case <-done:
		if testErr != nil {
			return testErr
		}
	case <-time.After(ctx.timeout):
		return fmt.Errorf("test timeout after %v", ctx.timeout)
	}

	// Run afterEach hooks (ignore errors in afterEach)
	allAfterEach := append(suite.AfterEach, r.globalAfterEach...)
	_ = r.runHooks(allAfterEach, ctx)

	return nil
}

func (r *TestRunner) runHooks(hooks []HookFunction, ctx *TestContext) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%v", recovered)
		}
	}()

	for _, hook := range hooks {
		hook(ctx)
	}
	return nil
}

func (r *TestRunner) toTestError(err interface{}) *TestError {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case error:
		return &TestError{
			Message: e.Error(),
			Stack:   string(debug.Stack()),
		}
	default:
		return &TestError{
			Message: fmt.Sprintf("%v", e),
			Stack:   string(debug.Stack()),
		}
	}
}
