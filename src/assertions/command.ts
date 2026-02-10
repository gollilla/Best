import type { CommandOutput } from '../types/index.js';
import { AssertionError } from './index.js';

export class CommandAssertion {
  constructor(private output: CommandOutput) {}

  toSucceed(): this {
    if (!this.output.success) {
      throw new AssertionError(
        `Expected command "${this.output.command}" to succeed, but it failed`,
        'success',
        'failure'
      );
    }
    return this;
  }

  toFail(): this {
    if (this.output.success) {
      throw new AssertionError(
        `Expected command "${this.output.command}" to fail, but it succeeded`,
        'failure',
        'success'
      );
    }
    return this;
  }

  toContain(expected: string | RegExp): this {
    const matches =
      typeof expected === 'string'
        ? this.output.output.includes(expected)
        : expected.test(this.output.output);

    if (!matches) {
      throw new AssertionError(
        `Expected command output to contain ${expected}, ` +
          `but output was: "${this.output.output}"`,
        expected,
        this.output.output
      );
    }
    return this;
  }

  toHaveStatusCode(code: number): this {
    if (this.output.statusCode !== code) {
      throw new AssertionError(
        `Expected status code ${code}, but was ${this.output.statusCode}`,
        code,
        this.output.statusCode
      );
    }
    return this;
  }

  and(): this {
    return this;
  }
}
