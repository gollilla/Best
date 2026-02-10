import type { TestRunner, TestFunction, HookFunction } from './runner/index.js';

let currentRunner: TestRunner | null = null;

export function setCurrentRunner(runner: TestRunner): void {
  currentRunner = runner;
}

export function getCurrentRunner(): TestRunner {
  if (!currentRunner) {
    throw new Error('No test runner available. Are you running tests with the Best CLI?');
  }
  return currentRunner;
}

// Global test functions
export function describe(name: string, fn: () => void): void {
  getCurrentRunner().describe(name, fn);
}

export function test(name: string, fn: TestFunction): void {
  getCurrentRunner().test(name, fn);
}

export function it(name: string, fn: TestFunction): void {
  getCurrentRunner().it(name, fn);
}

export function beforeAll(fn: HookFunction): void {
  getCurrentRunner().beforeAll(fn);
}

export function afterAll(fn: HookFunction): void {
  getCurrentRunner().afterAll(fn);
}

export function beforeEach(fn: HookFunction): void {
  getCurrentRunner().beforeEach(fn);
}

export function afterEach(fn: HookFunction): void {
  getCurrentRunner().afterEach(fn);
}

// Skip and only variants
export const skip = {
  test: (name: string, fn: TestFunction): void => {
    getCurrentRunner().skip.test(name, fn);
  },
  describe: (name: string, fn: () => void): void => {
    getCurrentRunner().skip.describe(name, fn);
  },
};

export const only = {
  test: (name: string, fn: TestFunction): void => {
    getCurrentRunner().only.test(name, fn);
  },
  describe: (name: string, fn: () => void): void => {
    getCurrentRunner().only.describe(name, fn);
  },
};

// Register globals on globalThis
export function registerGlobals(): void {
  const g = globalThis as Record<string, unknown>;

  g.describe = describe;
  g.test = test;
  g.it = it;
  g.beforeAll = beforeAll;
  g.afterAll = afterAll;
  g.beforeEach = beforeEach;
  g.afterEach = afterEach;
  g.skip = skip;
  g.only = only;
}

// Type declarations for global usage
declare global {
  function describe(name: string, fn: () => void): void;
  function test(name: string, fn: TestFunction): void;
  function it(name: string, fn: TestFunction): void;
  function beforeAll(fn: HookFunction): void;
  function afterAll(fn: HookFunction): void;
  function beforeEach(fn: HookFunction): void;
  function afterEach(fn: HookFunction): void;

  const skip: {
    test(name: string, fn: TestFunction): void;
    describe(name: string, fn: () => void): void;
  };

  const only: {
    test(name: string, fn: TestFunction): void;
    describe(name: string, fn: () => void): void;
  };
}
