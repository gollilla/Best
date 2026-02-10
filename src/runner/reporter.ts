import type { TestResult, SuiteResult, TestError } from './index.js';

export interface Reporter {
  onStart(suiteCount: number): void;
  onEnd(result: TestResult): void;
  onSuiteStart(name: string): void;
  onSuiteEnd(name: string, result: SuiteResult): void;
  onTestStart(name: string): void;
  onTestPass(name: string, duration: number): void;
  onTestFail(name: string, error: TestError, duration: number): void;
  onTestSkip(name: string): void;
  onTestRetry(name: string, attempt: number): void;
}

export class ConsoleReporter implements Reporter {
  private indent = '';

  onStart(suiteCount: number): void {
    console.log(`\nRunning ${suiteCount} test suite(s)...\n`);
  }

  onEnd(result: TestResult): void {
    console.log('\n' + '='.repeat(50));
    console.log('Test Results:');
    console.log('='.repeat(50));
    console.log(`  Passed:  ${result.passed}`);
    console.log(`  Failed:  ${result.failed}`);
    console.log(`  Skipped: ${result.skipped}`);
    console.log(`  Duration: ${result.duration}ms`);
    console.log('='.repeat(50));

    if (result.failed > 0) {
      console.log('\nFailed Tests:');
      for (const suite of result.suites) {
        for (const test of suite.tests) {
          if (test.status === 'failed') {
            console.log(`\n  ✗ ${suite.name ? suite.name + ' > ' : ''}${test.name}`);
            if (test.error) {
              console.log(`    Error: ${test.error.message}`);
              if (test.error.stack) {
                const stackLines = test.error.stack.split('\n').slice(1, 4);
                stackLines.forEach((line) => console.log(`    ${line.trim()}`));
              }
            }
          }
        }
      }
    }
  }

  onSuiteStart(name: string): void {
    if (name) {
      console.log(`${this.indent}${name}`);
      this.indent = '  ';
    }
  }

  onSuiteEnd(_name: string, _result: SuiteResult): void {
    this.indent = '';
  }

  onTestStart(_name: string): void {
    // Nothing to output on start
  }

  onTestPass(name: string, duration: number): void {
    console.log(`${this.indent}  ✓ ${name} (${duration}ms)`);
  }

  onTestFail(name: string, error: TestError, duration: number): void {
    console.log(`${this.indent}  ✗ ${name} (${duration}ms)`);
    console.log(`${this.indent}    → ${error.message}`);
  }

  onTestSkip(name: string): void {
    console.log(`${this.indent}  ○ ${name} (skipped)`);
  }

  onTestRetry(name: string, attempt: number): void {
    console.log(`${this.indent}  ↻ ${name} (retry ${attempt})`);
  }
}

export class SilentReporter implements Reporter {
  onStart(): void {}
  onEnd(): void {}
  onSuiteStart(): void {}
  onSuiteEnd(): void {}
  onTestStart(): void {}
  onTestPass(): void {}
  onTestFail(): void {}
  onTestSkip(): void {}
  onTestRetry(): void {}
}
