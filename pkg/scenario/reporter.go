package scenario

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Reporter reports scenario execution results
type Reporter interface {
	ReportResult(result *Result)
	ReportSummary(summary *Summary)
}

// Summary contains summary of multiple scenario executions
type Summary struct {
	Results        []*Result
	TotalScenarios int
	PassedCount    int
	FailedCount    int
	TotalSteps     int
	PassedSteps    int
	FailedSteps    int
	TotalDuration  time.Duration
}

// NewSummary creates a new summary from scenario results
func NewSummary(results ...*Result) *Summary {
	s := &Summary{
		Results:        results,
		TotalScenarios: len(results),
	}

	for _, r := range results {
		if r.Success {
			s.PassedCount++
		} else {
			s.FailedCount++
		}
		s.TotalSteps += r.TotalSteps
		s.PassedSteps += r.PassedSteps
		s.FailedSteps += r.FailedSteps
		s.TotalDuration += r.Duration
	}

	return s
}

// Success returns true if all scenarios passed
func (s *Summary) Success() bool {
	return s.FailedCount == 0
}

// ConsoleReporter reports results to the console
type ConsoleReporter struct {
	writer io.Writer
}

// NewConsoleReporter creates a new console reporter
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{
		writer: os.Stdout,
	}
}

// NewConsoleReporterWithWriter creates a new console reporter with a custom writer
func NewConsoleReporterWithWriter(w io.Writer) *ConsoleReporter {
	return &ConsoleReporter{
		writer: w,
	}
}

// ReportResult reports the scenario result to the console
func (r *ConsoleReporter) ReportResult(result *Result) {
	fmt.Fprintln(r.writer)
	fmt.Fprintln(r.writer, strings.Repeat("=", 60))

	if result.Scenario != "" {
		fmt.Fprintf(r.writer, "Scenario: %s\n", result.Scenario)
	}
	fmt.Fprintln(r.writer, strings.Repeat("=", 60))

	for _, step := range result.Steps {
		statusIcon := r.getStatusIcon(step.Status)
		fmt.Fprintf(r.writer, "%s Step %d: %s\n", statusIcon, step.StepNumber, step.Action)

		if step.Description != "" {
			fmt.Fprintf(r.writer, "   Description: %s\n", step.Description)
		}

		fmt.Fprintf(r.writer, "   Duration: %v\n", step.Duration)

		if step.Error != nil {
			fmt.Fprintf(r.writer, "   Error: %v\n", step.Error)
		}
	}

	fmt.Fprintln(r.writer, strings.Repeat("-", 60))
	fmt.Fprintf(r.writer, "Total: %d steps | Passed: %d | Failed: %d\n",
		result.TotalSteps, result.PassedSteps, result.FailedSteps)
	fmt.Fprintf(r.writer, "Duration: %v\n", result.Duration)

	if result.Success {
		fmt.Fprintln(r.writer, "Result: PASSED")
	} else {
		fmt.Fprintln(r.writer, "Result: FAILED")
		if result.Error != nil {
			fmt.Fprintf(r.writer, "Error: %v\n", result.Error)
		}
	}

	fmt.Fprintln(r.writer, strings.Repeat("=", 60))
}

// ReportSummary reports the summary of multiple scenario results
func (r *ConsoleReporter) ReportSummary(summary *Summary) {
	fmt.Fprintln(r.writer)
	fmt.Fprintln(r.writer, strings.Repeat("=", 60))
	fmt.Fprintln(r.writer, "TEST SUMMARY")
	fmt.Fprintln(r.writer, strings.Repeat("=", 60))

	// List all scenarios
	for _, result := range summary.Results {
		icon := "[PASS]"
		if !result.Success {
			icon = "[FAIL]"
		}
		fmt.Fprintf(r.writer, "%s %s (%d/%d steps, %v)\n",
			icon, result.Scenario, result.PassedSteps, result.TotalSteps, result.Duration.Round(time.Millisecond))
	}

	fmt.Fprintln(r.writer, strings.Repeat("-", 60))

	// Summary stats
	fmt.Fprintf(r.writer, "Scenarios: %d passed, %d failed, %d total\n",
		summary.PassedCount, summary.FailedCount, summary.TotalScenarios)
	fmt.Fprintf(r.writer, "Steps:     %d passed, %d failed, %d total\n",
		summary.PassedSteps, summary.FailedSteps, summary.TotalSteps)
	fmt.Fprintf(r.writer, "Duration:  %v\n", summary.TotalDuration.Round(time.Millisecond))

	fmt.Fprintln(r.writer)

	if summary.Success() {
		fmt.Fprintln(r.writer, "Result: ALL PASSED")
	} else {
		fmt.Fprintln(r.writer, "Result: SOME FAILED")

		// Show failed details
		fmt.Fprintln(r.writer)
		fmt.Fprintln(r.writer, "Failed scenarios:")
		for _, result := range summary.Results {
			if !result.Success {
				fmt.Fprintf(r.writer, "  - %s\n", result.Scenario)
				for _, step := range result.Steps {
					if step.Status == StepStatusFailed {
						fmt.Fprintf(r.writer, "      Step %d: %s\n", step.StepNumber, step.Description)
						if step.Error != nil {
							fmt.Fprintf(r.writer, "      Error: %v\n", step.Error)
						}
					}
				}
			}
		}
	}

	fmt.Fprintln(r.writer, strings.Repeat("=", 60))
}

// getStatusIcon returns an icon for the given status
func (r *ConsoleReporter) getStatusIcon(status StepStatus) string {
	switch status {
	case StepStatusPassed:
		return "[PASS]"
	case StepStatusFailed:
		return "[FAIL]"
	case StepStatusSkipped:
		return "[SKIP]"
	case StepStatusRunning:
		return "[RUN ]"
	default:
		return "[    ]"
	}
}

// StepReporter creates callbacks for real-time step reporting
func StepReporter() (func(int, ScenarioStep), func(int, StepResult)) {
	onStart := func(stepNum int, step ScenarioStep) {
		fmt.Printf("  [RUN ] Step %d: %s", stepNum, step.Action)
		if step.Description != "" {
			fmt.Printf(" - %s", step.Description)
		}
		fmt.Println()
	}

	onEnd := func(stepNum int, result StepResult) {
		icon := "[PASS]"
		if result.Status == StepStatusFailed {
			icon = "[FAIL]"
		}
		fmt.Printf("  %s Step %d completed in %v\n", icon, stepNum, result.Duration)
		if result.Error != nil {
			fmt.Printf("        Error: %v\n", result.Error)
		}
	}

	return onStart, onEnd
}
