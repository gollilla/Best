import { EventEmitter } from 'events';

export type EventMap = Record<string, unknown[]>;

export class TypedEventEmitter<T extends EventMap> {
  private emitter = new EventEmitter();

  on<K extends keyof T & string>(
    event: K,
    listener: (...args: T[K]) => void
  ): this {
    this.emitter.on(event, listener as (...args: unknown[]) => void);
    return this;
  }

  once<K extends keyof T & string>(
    event: K,
    listener: (...args: T[K]) => void
  ): this {
    this.emitter.once(event, listener as (...args: unknown[]) => void);
    return this;
  }

  emit<K extends keyof T & string>(event: K, ...args: T[K]): boolean {
    return this.emitter.emit(event, ...args);
  }

  off<K extends keyof T & string>(
    event: K,
    listener: (...args: T[K]) => void
  ): this {
    this.emitter.off(event, listener as (...args: unknown[]) => void);
    return this;
  }

  removeAllListeners<K extends keyof T & string>(event?: K): this {
    this.emitter.removeAllListeners(event);
    return this;
  }

  waitFor<K extends keyof T & string>(
    event: K,
    options?: {
      timeout?: number;
      filter?: (...args: T[K]) => boolean;
    }
  ): Promise<T[K]> {
    const { timeout, filter } = options ?? {};

    return new Promise((resolve, reject) => {
      let timer: ReturnType<typeof setTimeout> | null = null;

      const cleanup = () => {
        if (timer) clearTimeout(timer);
        this.off(event, handler);
      };

      const handler = (...args: T[K]) => {
        if (!filter || filter(...args)) {
          cleanup();
          resolve(args);
        }
      };

      if (timeout) {
        timer = setTimeout(() => {
          cleanup();
          reject(new Error(`Timeout waiting for event: ${String(event)}`));
        }, timeout);
      }

      this.on(event, handler);
    });
  }

  waitForAny<K extends keyof T & string>(
    events: K[],
    options?: { timeout?: number }
  ): Promise<{ event: K; args: T[K] }> {
    const { timeout } = options ?? {};

    return new Promise((resolve, reject) => {
      let timer: ReturnType<typeof setTimeout> | null = null;
      const handlers: Map<K, (...args: unknown[]) => void> = new Map();

      const cleanup = () => {
        if (timer) clearTimeout(timer);
        for (const [event, handler] of handlers) {
          this.off(event, handler as (...args: T[K]) => void);
        }
      };

      for (const event of events) {
        const handler = (...args: unknown[]) => {
          cleanup();
          resolve({ event, args: args as T[K] });
        };
        handlers.set(event, handler);
        this.on(event, handler as (...args: T[K]) => void);
      }

      if (timeout) {
        timer = setTimeout(() => {
          cleanup();
          reject(new Error(`Timeout waiting for events: ${events.map(String).join(', ')}`));
        }, timeout);
      }
    });
  }
}
