import type { Agent } from '../core/client.js';
import { AssertionError } from './index.js';

// タイミングアサーション
export class TimingAssertion {
  constructor(private player: Agent) {}

  /**
   * 指定時間内にタスクが完了することを確認
   */
  async toCompleteWithin<T>(
    task: () => Promise<T>,
    maxTime: number
  ): Promise<T> {
    const startTime = Date.now();

    try {
      const result = await Promise.race([
        task(),
        new Promise<never>((_, reject) =>
          setTimeout(
            () => reject(new Error('timeout')),
            maxTime
          )
        ),
      ]);

      return result;
    } catch (err) {
      if (err instanceof Error && err.message === 'timeout') {
        throw new AssertionError(
          `Expected task to complete within ${maxTime}ms, but it timed out`,
          `< ${maxTime}ms`,
          `> ${maxTime}ms`
        );
      }
      throw err;
    }
  }

  /**
   * 指定時間経過後も条件が満たされていることを確認
   */
  async toRemainTrueFor(
    condition: () => boolean | Promise<boolean>,
    duration: number,
    checkInterval = 100
  ): Promise<void> {
    const startTime = Date.now();

    while (Date.now() - startTime < duration) {
      const result = await condition();
      if (!result) {
        throw new AssertionError(
          `Condition became false after ${Date.now() - startTime}ms`,
          'true for full duration',
          'became false'
        );
      }
      await new Promise((r) => setTimeout(r, checkInterval));
    }
  }

  /**
   * タイムアウトすることを確認（タスクが完了しないことを期待）
   */
  async toTimeout(
    task: () => Promise<unknown>,
    timeoutMs: number
  ): Promise<void> {
    try {
      await Promise.race([
        task(),
        new Promise((resolve) => setTimeout(resolve, timeoutMs)),
      ]);

      // タスクが完了したかタイムアウトしたかを確認
      const taskCompleted = await Promise.race([
        task().then(() => true),
        new Promise<false>((resolve) => setTimeout(() => resolve(false), 0)),
      ]);

      if (taskCompleted) {
        throw new AssertionError(
          `Expected task to timeout, but it completed`,
          'timeout',
          'completed'
        );
      }
    } catch (err) {
      if (err instanceof AssertionError) throw err;
      // その他のエラーは無視（タイムアウト期待通り）
    }
  }
}

// シーケンスアサーション
export class SequenceAssertion {
  constructor(private player: Agent) {}

  /**
   * イベントが順番通りに発生することを確認
   */
  async toOccurInOrder<T extends string>(
    events: Array<{
      event: T;
      filter?: (data: unknown) => boolean;
    }>,
    options?: { timeout?: number }
  ): Promise<void> {
    const { timeout = 10000 } = options ?? {};
    let currentIndex = 0;
    const startTime = Date.now();

    return new Promise((resolve, reject) => {
      const handlers: Array<{ event: string; handler: (...args: unknown[]) => void }> = [];

      const cleanup = () => {
        for (const { event, handler } of handlers) {
          this.player.off(event, handler);
        }
      };

      const timeoutId = setTimeout(() => {
        cleanup();
        reject(
          new AssertionError(
            `Timeout waiting for event ${currentIndex + 1}: "${events[currentIndex].event}"`,
            events.map((e) => e.event),
            events.slice(0, currentIndex).map((e) => e.event)
          )
        );
      }, timeout);

      for (let i = 0; i < events.length; i++) {
        const eventDef = events[i];
        const handler = (...args: unknown[]) => {
          if (i !== currentIndex) return;

          if (eventDef.filter && !eventDef.filter(args[0])) {
            return;
          }

          currentIndex++;

          if (currentIndex >= events.length) {
            clearTimeout(timeoutId);
            cleanup();
            resolve();
          }
        };

        handlers.push({ event: eventDef.event, handler });
        this.player.on(eventDef.event, handler);
      }
    });
  }

  /**
   * 複数のイベントが指定時間内に発生することを確認（順序不問）
   */
  async toOccurAll<T extends string>(
    events: Array<{
      event: T;
      filter?: (data: unknown) => boolean;
    }>,
    options?: { timeout?: number }
  ): Promise<void> {
    const { timeout = 10000 } = options ?? {};
    const occurred = new Set<number>();

    return new Promise((resolve, reject) => {
      const handlers: Array<{ event: string; handler: (...args: unknown[]) => void }> = [];

      const cleanup = () => {
        for (const { event, handler } of handlers) {
          this.player.off(event, handler);
        }
      };

      const timeoutId = setTimeout(() => {
        cleanup();
        const missing = events
          .filter((_, i) => !occurred.has(i))
          .map((e) => e.event);
        reject(
          new AssertionError(
            `Timeout: some events did not occur. Missing: ${missing.join(', ')}`,
            events.map((e) => e.event),
            Array.from(occurred).map((i) => events[i].event)
          )
        );
      }, timeout);

      for (let i = 0; i < events.length; i++) {
        const eventDef = events[i];
        const handler = (...args: unknown[]) => {
          if (occurred.has(i)) return;

          if (eventDef.filter && !eventDef.filter(args[0])) {
            return;
          }

          occurred.add(i);

          if (occurred.size >= events.length) {
            clearTimeout(timeoutId);
            cleanup();
            resolve();
          }
        };

        handlers.push({ event: eventDef.event, handler });
        this.player.on(eventDef.event, handler);
      }
    });
  }
}

// 条件待機アサーション
export class ConditionAssertion {
  constructor(private player: Agent) {}

  /**
   * 条件が指定時間内に満たされることを確認
   */
  async toBeMetWithin(
    condition: () => boolean | Promise<boolean>,
    timeout: number,
    checkInterval = 100
  ): Promise<void> {
    const startTime = Date.now();

    while (Date.now() - startTime < timeout) {
      const result = await condition();
      if (result) {
        return;
      }
      await new Promise((r) => setTimeout(r, checkInterval));
    }

    throw new AssertionError(
      `Condition was not met within ${timeout}ms`,
      'condition met',
      'condition not met'
    );
  }

  /**
   * 条件が満たされるまで待機し、結果を返す
   */
  async toEventuallyBe<T>(
    getValue: () => T | Promise<T>,
    expected: T,
    options?: { timeout?: number; checkInterval?: number }
  ): Promise<void> {
    const { timeout = 5000, checkInterval = 100 } = options ?? {};
    const startTime = Date.now();

    while (Date.now() - startTime < timeout) {
      const actual = await getValue();
      if (actual === expected) {
        return;
      }
      await new Promise((r) => setTimeout(r, checkInterval));
    }

    const finalValue = await getValue();
    throw new AssertionError(
      `Expected value to eventually be ${expected}, but was ${finalValue}`,
      expected,
      finalValue
    );
  }

  /**
   * 値が変化することを待機
   */
  async toChange<T>(
    getValue: () => T | Promise<T>,
    options?: { timeout?: number; checkInterval?: number }
  ): Promise<{ from: T; to: T }> {
    const { timeout = 5000, checkInterval = 100 } = options ?? {};
    const initialValue = await getValue();
    const startTime = Date.now();

    while (Date.now() - startTime < timeout) {
      const currentValue = await getValue();
      if (currentValue !== initialValue) {
        return { from: initialValue, to: currentValue };
      }
      await new Promise((r) => setTimeout(r, checkInterval));
    }

    throw new AssertionError(
      `Expected value to change from ${initialValue}, but it remained the same`,
      'changed',
      'unchanged'
    );
  }

  /**
   * 値が変化しないことを確認
   */
  async notToChange<T>(
    getValue: () => T | Promise<T>,
    options?: { duration?: number; checkInterval?: number }
  ): Promise<void> {
    const { duration = 3000, checkInterval = 100 } = options ?? {};
    const initialValue = await getValue();
    const startTime = Date.now();

    while (Date.now() - startTime < duration) {
      const currentValue = await getValue();
      if (currentValue !== initialValue) {
        throw new AssertionError(
          `Expected value not to change, but it changed from ${initialValue} to ${currentValue}`,
          initialValue,
          currentValue
        );
      }
      await new Promise((r) => setTimeout(r, checkInterval));
    }
  }
}
