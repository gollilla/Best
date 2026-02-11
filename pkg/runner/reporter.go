package runner

// Reporter is the interface for test result reporting
type Reporter interface {
	OnStart(suiteCount int)
	OnEnd(result *TestResult)
	OnSuiteStart(name string)
	OnSuiteEnd(name string, result *SuiteResult)
	OnTestStart(name string)
	OnTestPass(name string, duration int64)
	OnTestFail(name string, err *TestError, duration int64)
	OnTestSkip(name string)
	OnTestRetry(name string, attempt int)
}
