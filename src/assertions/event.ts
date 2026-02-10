import type { Agent } from '../core/client.js';
import type { Position, Dimension } from '../types/index.js';
import { distanceTo } from '../core/state.js';
import { AssertionError } from './index.js';

// 接続/切断アサーション
export class ConnectionAssertion {
  constructor(private player: Agent) {}

  /**
   * キックされることを待機
   */
  async toBeKicked(options?: {
    timeout?: number;
    reason?: string | RegExp;
  }): Promise<string> {
    const { timeout = 10000, reason } = options ?? {};

    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        this.player.off('disconnect', handler);
        reject(
          new AssertionError(
            'Timeout waiting for player to be kicked',
            'kicked',
            'connected'
          )
        );
      }, timeout);

      const handler = (disconnectReason: string) => {
        clearTimeout(timeoutId);
        this.player.off('disconnect', handler);

        if (reason) {
          const matches =
            typeof reason === 'string'
              ? disconnectReason.includes(reason)
              : reason.test(disconnectReason);

          if (!matches) {
            reject(
              new AssertionError(
                `Expected kick reason to match ${reason}, but was "${disconnectReason}"`,
                reason,
                disconnectReason
              )
            );
            return;
          }
        }

        resolve(disconnectReason);
      };

      this.player.on('disconnect', handler);
    });
  }

  /**
   * キックされないことを確認
   */
  async notToBeKicked(options?: { duration?: number }): Promise<void> {
    const { duration = 5000 } = options ?? {};

    return new Promise((resolve, reject) => {
      const handler = (reason: string) => {
        this.player.off('disconnect', handler);
        reject(
          new AssertionError(
            `Expected not to be kicked, but was kicked with reason: "${reason}"`,
            'connected',
            'kicked'
          )
        );
      };

      this.player.on('disconnect', handler);

      setTimeout(() => {
        this.player.off('disconnect', handler);
        resolve();
      }, duration);
    });
  }

  /**
   * BANされることを待機（キックメッセージでBAN判定）
   */
  async toBeBanned(options?: { timeout?: number }): Promise<string> {
    const { timeout = 10000 } = options ?? {};

    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        this.player.off('disconnect', handler);
        reject(
          new AssertionError(
            'Timeout waiting for player to be banned',
            'banned',
            'connected'
          )
        );
      }, timeout);

      const handler = (reason: string) => {
        clearTimeout(timeoutId);
        this.player.off('disconnect', handler);

        // BAN判定（一般的なBAN理由のパターン）
        const banPatterns = [
          /banned/i,
          /ban/i,
          /you have been banned/i,
          /permanently banned/i,
        ];

        const isBan = banPatterns.some((p) => p.test(reason));
        if (!isBan) {
          reject(
            new AssertionError(
              `Expected to be banned, but kick reason was: "${reason}"`,
              'ban message',
              reason
            )
          );
          return;
        }

        resolve(reason);
      };

      this.player.on('disconnect', handler);
    });
  }
}

// テレポートアサーション
export class TeleportAssertion {
  constructor(private player: Agent) {}

  /**
   * テレポートが発生することを待機
   */
  async toOccur(options?: {
    timeout?: number;
    minDistance?: number;
  }): Promise<Position> {
    const { timeout = 5000, minDistance = 5 } = options ?? {};
    const startPos = { ...this.player.position };

    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        this.player.off('position_update', handler);
        reject(
          new AssertionError(
            'Timeout waiting for teleport to occur',
            'teleport',
            'no teleport'
          )
        );
      }, timeout);

      const handler = (newPos: Position) => {
        const distance = distanceTo(startPos, newPos);
        if (distance >= minDistance) {
          clearTimeout(timeoutId);
          this.player.off('position_update', handler);
          resolve(newPos);
        }
      };

      this.player.on('position_update', handler);
    });
  }

  /**
   * 特定の座標へのテレポートを待機
   */
  async toDestination(
    destination: Position,
    options?: { timeout?: number; tolerance?: number }
  ): Promise<void> {
    const { timeout = 5000, tolerance = 1 } = options ?? {};

    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        this.player.off('position_update', handler);
        reject(
          new AssertionError(
            `Timeout waiting for teleport to (${destination.x}, ${destination.y}, ${destination.z})`,
            destination,
            this.player.position
          )
        );
      }, timeout);

      const handler = (newPos: Position) => {
        const distance = distanceTo(destination, newPos);
        if (distance <= tolerance) {
          clearTimeout(timeoutId);
          this.player.off('position_update', handler);
          resolve();
        }
      };

      // 既に目的地にいるか確認
      if (distanceTo(destination, this.player.position) <= tolerance) {
        resolve();
        return;
      }

      this.player.on('position_update', handler);
    });
  }

  /**
   * テレポートが発生しないことを確認
   */
  async notToOccur(options?: {
    duration?: number;
    minDistance?: number;
  }): Promise<void> {
    const { duration = 3000, minDistance = 5 } = options ?? {};
    const startPos = { ...this.player.position };

    return new Promise((resolve, reject) => {
      const handler = (newPos: Position) => {
        const distance = distanceTo(startPos, newPos);
        if (distance >= minDistance) {
          this.player.off('position_update', handler);
          reject(
            new AssertionError(
              `Expected no teleport, but teleported ${distance.toFixed(1)} blocks`,
              'no teleport',
              newPos
            )
          );
        }
      };

      this.player.on('position_update', handler);

      setTimeout(() => {
        this.player.off('position_update', handler);
        resolve();
      }, duration);
    });
  }
}

// ディメンション移動アサーション
export class DimensionAssertion {
  constructor(private player: Agent) {}

  /**
   * 現在のディメンションを確認
   */
  toBe(expected: Dimension): this {
    const actual = this.player.state.dimension;
    if (actual !== expected) {
      throw new AssertionError(
        `Expected to be in dimension "${expected}", but was in "${actual}"`,
        expected,
        actual
      );
    }
    return this;
  }

  /**
   * ディメンション変更を待機
   */
  async toChangeTo(
    expected: Dimension,
    options?: { timeout?: number }
  ): Promise<void> {
    const { timeout = 30000 } = options ?? {};

    // 既に目的のディメンションにいるか確認
    if (this.player.state.dimension === expected) {
      return;
    }

    try {
      await this.player.waitFor('dimension_change', {
        timeout,
        filter: (dim: string) => dim === expected,
      });
    } catch {
      throw new AssertionError(
        `Timeout waiting for dimension to change to "${expected}"`,
        expected,
        this.player.state.dimension
      );
    }
  }

  /**
   * オーバーワールドにいることを確認
   */
  toBeOverworld(): this {
    return this.toBe('overworld');
  }

  /**
   * ネザーにいることを確認
   */
  toBeNether(): this {
    return this.toBe('nether');
  }

  /**
   * エンドにいることを確認
   */
  toBeTheEnd(): this {
    return this.toBe('the_end');
  }
}

// 死亡アサーション
export class DeathAssertion {
  constructor(private player: Agent) {}

  /**
   * 死亡することを待機
   */
  async toOccur(options?: { timeout?: number }): Promise<void> {
    const { timeout = 10000 } = options ?? {};

    return new Promise((resolve, reject) => {
      const timeoutId = setTimeout(() => {
        this.player.off('health_update', handler);
        reject(
          new AssertionError(
            'Timeout waiting for player to die',
            0,
            this.player.health
          )
        );
      }, timeout);

      const handler = (health: number) => {
        if (health <= 0) {
          clearTimeout(timeoutId);
          this.player.off('health_update', handler);
          resolve();
        }
      };

      // 既に死んでいるか確認
      if (this.player.health <= 0) {
        clearTimeout(timeoutId);
        resolve();
        return;
      }

      this.player.on('health_update', handler);
    });
  }

  /**
   * 死亡しないことを確認
   */
  async notToOccur(options?: { duration?: number }): Promise<void> {
    const { duration = 5000 } = options ?? {};

    return new Promise((resolve, reject) => {
      const handler = (health: number) => {
        if (health <= 0) {
          this.player.off('health_update', handler);
          reject(
            new AssertionError(
              'Expected player not to die, but player died',
              'alive',
              'dead'
            )
          );
        }
      };

      this.player.on('health_update', handler);

      setTimeout(() => {
        this.player.off('health_update', handler);
        resolve();
      }, duration);
    });
  }
}

// リスポーンアサーション
export class RespawnAssertion {
  constructor(private player: Agent) {}

  /**
   * リスポーンすることを待機
   */
  async toOccur(options?: { timeout?: number }): Promise<Position> {
    const { timeout = 10000 } = options ?? {};

    try {
      const [pos] = await this.player.waitFor('respawn', { timeout });
      return pos;
    } catch {
      throw new AssertionError(
        'Timeout waiting for player to respawn',
        'respawn',
        'no respawn'
      );
    }
  }
}
