import { Agent, createAgent } from '../core/client.js';
import { Reporter, ConsoleReporter } from './reporter.js';
import type { ClientOptions, ServerInfo } from '../types/index.js';

export interface TestRunnerOptions {
  timeout?: number;
  parallel?: boolean;
  maxConcurrency?: number;
  reporter?: Reporter;
  bail?: boolean;
  retries?: number;
}

export interface TestContext {
  player: Agent;
  server: ServerInfo;
  timeout: (ms: number) => void;
}

export type TestFunction = (ctx: TestContext) => Promise<void> | void;
export type HookFunction = (ctx: TestContext) => Promise<void> | void;

interface TestCase {
  name: string;
  fn: TestFunction;
  skip?: boolean;
  only?: boolean;
}

interface TestSuite {
  name: string;
  tests: TestCase[];
  beforeAll: HookFunction[];
  afterAll: HookFunction[];
  beforeEach: HookFunction[];
  afterEach: HookFunction[];
  skip?: boolean;
  only?: boolean;
}

export interface TestCaseResult {
  name: string;
  status: 'passed' | 'failed' | 'skipped';
  duration: number;
  error?: TestError;
}

export interface SuiteResult {
  name: string;
  tests: TestCaseResult[];
  duration: number;
}

export interface TestResult {
  passed: number;
  failed: number;
  skipped: number;
  duration: number;
  suites: SuiteResult[];
}

export interface TestError {
  message: string;
  stack?: string;
  expected?: unknown;
  actual?: unknown;
}

export class TestRunner {
  private options: Required<TestRunnerOptions>;
  private clientOptions: ClientOptions | null = null;
  private suites: TestSuite[] = [];
  private currentSuite: TestSuite | null = null;
  private globalBeforeAll: HookFunction[] = [];
  private globalAfterAll: HookFunction[] = [];
  private globalBeforeEach: HookFunction[] = [];
  private globalAfterEach: HookFunction[] = [];

  constructor(options?: TestRunnerOptions) {
    this.options = {
      timeout: options?.timeout ?? 30000,
      parallel: options?.parallel ?? false,
      maxConcurrency: options?.maxConcurrency ?? 4,
      reporter: options?.reporter ?? new ConsoleReporter(),
      bail: options?.bail ?? false,
      retries: options?.retries ?? 0,
    };
  }

  configure(options: ClientOptions): this {
    this.clientOptions = options;
    return this;
  }

  describe(name: string, fn: () => void): this {
    const suite: TestSuite = {
      name,
      tests: [],
      beforeAll: [],
      afterAll: [],
      beforeEach: [],
      afterEach: [],
    };

    const prevSuite = this.currentSuite;
    this.currentSuite = suite;
    fn();
    this.currentSuite = prevSuite;

    this.suites.push(suite);
    return this;
  }

  test(name: string, fn: TestFunction): this {
    const testCase: TestCase = { name, fn };

    if (this.currentSuite) {
      this.currentSuite.tests.push(testCase);
    } else {
      // Create implicit suite for orphan tests
      const implicitSuite: TestSuite = {
        name: '',
        tests: [testCase],
        beforeAll: [],
        afterAll: [],
        beforeEach: [],
        afterEach: [],
      };
      this.suites.push(implicitSuite);
    }

    return this;
  }

  it(name: string, fn: TestFunction): this {
    return this.test(name, fn);
  }

  beforeAll(fn: HookFunction): this {
    if (this.currentSuite) {
      this.currentSuite.beforeAll.push(fn);
    } else {
      this.globalBeforeAll.push(fn);
    }
    return this;
  }

  afterAll(fn: HookFunction): this {
    if (this.currentSuite) {
      this.currentSuite.afterAll.push(fn);
    } else {
      this.globalAfterAll.push(fn);
    }
    return this;
  }

  beforeEach(fn: HookFunction): this {
    if (this.currentSuite) {
      this.currentSuite.beforeEach.push(fn);
    } else {
      this.globalBeforeEach.push(fn);
    }
    return this;
  }

  afterEach(fn: HookFunction): this {
    if (this.currentSuite) {
      this.currentSuite.afterEach.push(fn);
    } else {
      this.globalAfterEach.push(fn);
    }
    return this;
  }

  skip = {
    test: (name: string, fn: TestFunction): this => {
      const testCase: TestCase = { name, fn, skip: true };
      if (this.currentSuite) {
        this.currentSuite.tests.push(testCase);
      }
      return this;
    },
    describe: (name: string, fn: () => void): this => {
      const suite: TestSuite = {
        name,
        tests: [],
        beforeAll: [],
        afterAll: [],
        beforeEach: [],
        afterEach: [],
        skip: true,
      };
      const prevSuite = this.currentSuite;
      this.currentSuite = suite;
      fn();
      this.currentSuite = prevSuite;
      this.suites.push(suite);
      return this;
    },
  };

  only = {
    test: (name: string, fn: TestFunction): this => {
      const testCase: TestCase = { name, fn, only: true };
      if (this.currentSuite) {
        this.currentSuite.tests.push(testCase);
      }
      return this;
    },
    describe: (name: string, fn: () => void): this => {
      const suite: TestSuite = {
        name,
        tests: [],
        beforeAll: [],
        afterAll: [],
        beforeEach: [],
        afterEach: [],
        only: true,
      };
      const prevSuite = this.currentSuite;
      this.currentSuite = suite;
      fn();
      this.currentSuite = prevSuite;
      this.suites.push(suite);
      return this;
    },
  };

  async run(): Promise<TestResult> {
    if (!this.clientOptions) {
      throw new Error('Server not configured. Call configure() first.');
    }

    const result: TestResult = {
      passed: 0,
      failed: 0,
      skipped: 0,
      duration: 0,
      suites: [],
    };

    const startTime = Date.now();
    this.options.reporter.onStart(this.suites.length);

    // Check for "only" tests
    const hasOnly =
      this.suites.some((s) => s.only) ||
      this.suites.some((s) => s.tests.some((t) => t.only));

    // Run global beforeAll
    const globalCtx = this.createContext();
    try {
      for (const hook of this.globalBeforeAll) {
        await hook(globalCtx);
      }

      for (const suite of this.suites) {
        const suiteResult = await this.runSuite(suite, hasOnly, globalCtx);
        result.suites.push(suiteResult);

        for (const test of suiteResult.tests) {
          if (test.status === 'passed') result.passed++;
          else if (test.status === 'failed') result.failed++;
          else result.skipped++;
        }

        if (this.options.bail && result.failed > 0) {
          break;
        }
      }

      // Run global afterAll
      for (const hook of this.globalAfterAll) {
        await hook(globalCtx);
      }
    } finally {
      await globalCtx.player.disconnect();
    }

    result.duration = Date.now() - startTime;
    this.options.reporter.onEnd(result);

    return result;
  }

  private createContext(): TestContext {
    // Agentは最初からexpectを持っている
    const player = createAgent(this.clientOptions!);
    let currentTimeout = this.options.timeout;

    return {
      player,
      server: {
        host: this.clientOptions!.host,
        port: this.clientOptions!.port ?? 19132,
        version: this.clientOptions!.version ?? '',
      },
      timeout: (ms: number) => {
        currentTimeout = ms;
      },
    };
  }

  private async runSuite(
    suite: TestSuite,
    hasOnly: boolean,
    globalCtx: TestContext
  ): Promise<SuiteResult> {
    const suiteResult: SuiteResult = {
      name: suite.name,
      tests: [],
      duration: 0,
    };

    const startTime = Date.now();
    this.options.reporter.onSuiteStart(suite.name);

    // Skip if needed
    if (suite.skip || (hasOnly && !suite.only && !suite.tests.some((t) => t.only))) {
      for (const test of suite.tests) {
        suiteResult.tests.push({
          name: test.name,
          status: 'skipped',
          duration: 0,
        });
        this.options.reporter.onTestSkip(test.name);
      }
      suiteResult.duration = Date.now() - startTime;
      this.options.reporter.onSuiteEnd(suite.name, suiteResult);
      return suiteResult;
    }

    // Run beforeAll hooks
    try {
      for (const hook of suite.beforeAll) {
        await hook(globalCtx);
      }
    } catch (err) {
      // If beforeAll fails, skip all tests
      for (const test of suite.tests) {
        suiteResult.tests.push({
          name: test.name,
          status: 'failed',
          duration: 0,
          error: this.toTestError(err),
        });
      }
      suiteResult.duration = Date.now() - startTime;
      return suiteResult;
    }

    // Run tests
    for (const test of suite.tests) {
      const testResult = await this.runTest(test, suite, hasOnly, globalCtx);
      suiteResult.tests.push(testResult);

      if (this.options.bail && testResult.status === 'failed') {
        break;
      }
    }

    // Run afterAll hooks
    try {
      for (const hook of suite.afterAll) {
        await hook(globalCtx);
      }
    } catch {
      // Ignore afterAll errors
    }

    suiteResult.duration = Date.now() - startTime;
    this.options.reporter.onSuiteEnd(suite.name, suiteResult);
    return suiteResult;
  }

  private async runTest(
    test: TestCase,
    suite: TestSuite,
    hasOnly: boolean,
    ctx: TestContext
  ): Promise<TestCaseResult> {
    // Skip logic
    if (test.skip || (hasOnly && !test.only && !suite.only)) {
      this.options.reporter.onTestSkip(test.name);
      return { name: test.name, status: 'skipped', duration: 0 };
    }

    this.options.reporter.onTestStart(test.name);
    const startTime = Date.now();

    let lastError: unknown;
    const maxAttempts = this.options.retries + 1;

    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      try {
        // beforeEach hooks
        for (const hook of [...this.globalBeforeEach, ...suite.beforeEach]) {
          await hook(ctx);
        }

        // Run test
        await Promise.race([
          test.fn(ctx),
          new Promise((_, reject) =>
            setTimeout(
              () => reject(new Error('Test timeout')),
              this.options.timeout
            )
          ),
        ]);

        // afterEach hooks
        for (const hook of [...suite.afterEach, ...this.globalAfterEach]) {
          await hook(ctx);
        }

        const duration = Date.now() - startTime;
        this.options.reporter.onTestPass(test.name, duration);
        return { name: test.name, status: 'passed', duration };
      } catch (err) {
        lastError = err;

        // Run afterEach even on failure
        try {
          for (const hook of [...suite.afterEach, ...this.globalAfterEach]) {
            await hook(ctx);
          }
        } catch {
          // Ignore
        }

        if (attempt < maxAttempts) {
          this.options.reporter.onTestRetry(test.name, attempt);
        }
      }
    }

    const duration = Date.now() - startTime;
    const error = this.toTestError(lastError);
    this.options.reporter.onTestFail(test.name, error, duration);
    return { name: test.name, status: 'failed', duration, error };
  }

  private toTestError(err: unknown): TestError {
    if (err instanceof Error) {
      return {
        message: err.message,
        stack: err.stack,
      };
    }
    return { message: String(err) };
  }
}

export function createTestRunner(options?: TestRunnerOptions): TestRunner {
  return new TestRunner(options);
}

export { ConsoleReporter } from './reporter.js';
export type { Reporter } from './reporter.js';
