package runner

import (
	"fmt"
	"strings"
)

// ConsoleReporter is a simple console-based reporter
type ConsoleReporter struct {
	indent string
}

// NewConsoleReporter creates a new console reporter
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{
		indent: "",
	}
}

func (r *ConsoleReporter) OnStart(suiteCount int) {
	fmt.Printf("\nRunning %d test suite(s)...\n\n", suiteCount)
}

func (r *ConsoleReporter) OnEnd(result *TestResult) {
	separator := strings.Repeat("=", 50)
	fmt.Printf("\n%s\n", separator)
	fmt.Println("Test Results:")
	fmt.Println(separator)
	fmt.Printf("  Passed:  %d\n", result.Passed)
	fmt.Printf("  Failed:  %d\n", result.Failed)
	fmt.Printf("  Skipped: %d\n", result.Skipped)
	fmt.Printf("  Duration: %dms\n", result.Duration.Milliseconds())
	fmt.Println(separator)

	if result.Failed > 0 {
		fmt.Println("\nFailed Tests:")
		for _, suite := range result.Suites {
			for _, test := range suite.Tests {
				if test.Status == TestStatusFailed {
					if suite.Name != "" {
						fmt.Printf("\n  ✗ %s > %s\n", suite.Name, test.Name)
					} else {
						fmt.Printf("\n  ✗ %s\n", test.Name)
					}
					if test.Error != nil {
						fmt.Printf("    Error: %s\n", test.Error.Message)
						if test.Error.Stack != "" {
							lines := strings.Split(test.Error.Stack, "\n")
							for i := 1; i < len(lines) && i < 4; i++ {
								fmt.Printf("    %s\n", strings.TrimSpace(lines[i]))
							}
						}
					}
				}
			}
		}
	}
}

func (r *ConsoleReporter) OnSuiteStart(name string) {
	if name != "" {
		fmt.Printf("%s%s\n", r.indent, name)
		r.indent = "  "
	}
}

func (r *ConsoleReporter) OnSuiteEnd(_ string, _ *SuiteResult) {
	r.indent = ""
}

func (r *ConsoleReporter) OnTestStart(_ string) {
	// Nothing to output on start
}

func (r *ConsoleReporter) OnTestPass(name string, duration int64) {
	fmt.Printf("%s  ✓ %s (%dms)\n", r.indent, name, duration)
}

func (r *ConsoleReporter) OnTestFail(name string, err *TestError, duration int64) {
	fmt.Printf("%s  ✗ %s (%dms)\n", r.indent, name, duration)
	fmt.Printf("%s    → %s\n", r.indent, err.Message)
}

func (r *ConsoleReporter) OnTestSkip(name string) {
	fmt.Printf("%s  ○ %s (skipped)\n", r.indent, name)
}

func (r *ConsoleReporter) OnTestRetry(name string, attempt int) {
	fmt.Printf("%s  ↻ %s (retry %d)\n", r.indent, name, attempt)
}

// SilentReporter is a reporter that produces no output
type SilentReporter struct{}

func (r *SilentReporter) OnStart(_ int)                                    {}
func (r *SilentReporter) OnEnd(_ *TestResult)                              {}
func (r *SilentReporter) OnSuiteStart(_ string)                            {}
func (r *SilentReporter) OnSuiteEnd(_ string, _ *SuiteResult)              {}
func (r *SilentReporter) OnTestStart(_ string)                             {}
func (r *SilentReporter) OnTestPass(_ string, _ int64)                     {}
func (r *SilentReporter) OnTestFail(_ string, _ *TestError, _ int64)       {}
func (r *SilentReporter) OnTestSkip(_ string)                              {}
func (r *SilentReporter) OnTestRetry(_ string, _ int)                      {}
