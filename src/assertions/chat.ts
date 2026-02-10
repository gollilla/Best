import type { Agent } from '../core/client.js';
import type { ChatMessage } from '../types/index.js';
import { AssertionError } from './index.js';

export class ChatAssertion {
  constructor(private player: Agent) {}

  async toReceive(
    expected: string | RegExp,
    options?: { timeout?: number; from?: string }
  ): Promise<ChatMessage> {
    const { timeout = 5000, from } = options ?? {};

    const filter = (message: ChatMessage): boolean => {
      const matchesContent =
        typeof expected === 'string'
          ? message.message.includes(expected)
          : expected.test(message.message);

      const matchesSender = from ? message.sender === from : true;

      return matchesContent && matchesSender;
    };

    try {
      const [message] = await this.player.waitFor('chat', {
        timeout,
        filter,
      });
      return message;
    } catch {
      throw new AssertionError(
        `Timeout waiting for chat message matching ${expected}` +
          (from ? ` from ${from}` : ''),
        expected,
        undefined
      );
    }
  }

  async notToReceive(
    pattern: string | RegExp,
    options?: { duration?: number }
  ): Promise<void> {
    const { duration = 3000 } = options ?? {};

    return new Promise((resolve, reject) => {
      const handler = (message: ChatMessage) => {
        const matches =
          typeof pattern === 'string'
            ? message.message.includes(pattern)
            : pattern.test(message.message);

        if (matches) {
          this.player.off('chat', handler);
          reject(
            new AssertionError(
              `Expected not to receive chat message matching ${pattern}, ` +
                `but received: "${message.message}"`,
              undefined,
              message.message
            )
          );
        }
      };

      this.player.on('chat', handler);

      setTimeout(() => {
        this.player.off('chat', handler);
        resolve();
      }, duration);
    });
  }

  async toReceiveSystem(
    expected: string | RegExp,
    options?: { timeout?: number }
  ): Promise<ChatMessage> {
    const { timeout = 5000 } = options ?? {};

    const filter = (message: ChatMessage): boolean => {
      if (message.type !== 'system') return false;

      return typeof expected === 'string'
        ? message.message.includes(expected)
        : expected.test(message.message);
    };

    try {
      const [message] = await this.player.waitFor('chat', {
        timeout,
        filter,
      });
      return message;
    } catch {
      throw new AssertionError(
        `Timeout waiting for system message matching ${expected}`,
        expected,
        undefined
      );
    }
  }

  async toReceiveInOrder(
    expected: (string | RegExp)[],
    options?: { timeout?: number }
  ): Promise<ChatMessage[]> {
    const { timeout = 10000 } = options ?? {};
    const received: ChatMessage[] = [];
    let currentIndex = 0;

    return new Promise((resolve, reject) => {
      const startTime = Date.now();

      const handler = (message: ChatMessage) => {
        if (Date.now() - startTime > timeout) {
          cleanup();
          reject(
            new AssertionError(
              `Timeout waiting for message ${currentIndex + 1} matching ${expected[currentIndex]}`,
              expected,
              received.map((m) => m.message)
            )
          );
          return;
        }

        const pattern = expected[currentIndex];
        const matches =
          typeof pattern === 'string'
            ? message.message.includes(pattern)
            : pattern.test(message.message);

        if (matches) {
          received.push(message);
          currentIndex++;

          if (currentIndex >= expected.length) {
            cleanup();
            resolve(received);
          }
        }
      };

      const cleanup = () => {
        this.player.off('chat', handler);
      };

      this.player.on('chat', handler);

      setTimeout(() => {
        cleanup();
        if (received.length < expected.length) {
          reject(
            new AssertionError(
              `Timeout: only received ${received.length}/${expected.length} messages`,
              expected,
              received.map((m) => m.message)
            )
          );
        }
      }, timeout);
    });
  }
}
