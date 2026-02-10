import type { Agent } from '../core/client.js';
import type { Position } from '../types/index.js';
import { distanceTo } from '../core/state.js';
import { AssertionError } from './index.js';

export class PositionAssertion {
  constructor(private player: Agent) {}

  toBe(expected: Position): void {
    const actual = this.player.position;
    if (
      actual.x !== expected.x ||
      actual.y !== expected.y ||
      actual.z !== expected.z
    ) {
      throw new AssertionError(
        `Expected position to be (${expected.x}, ${expected.y}, ${expected.z}), ` +
          `but was (${actual.x}, ${actual.y}, ${actual.z})`,
        expected,
        actual
      );
    }
  }

  toBeNear(expected: Position, tolerance: number): void {
    const actual = this.player.position;
    const distance = distanceTo(actual, expected);

    if (distance > tolerance) {
      throw new AssertionError(
        `Expected position to be within ${tolerance} of ` +
          `(${expected.x}, ${expected.y}, ${expected.z}), ` +
          `but was (${actual.x}, ${actual.y}, ${actual.z}) ` +
          `(distance: ${distance.toFixed(2)})`,
        expected,
        actual
      );
    }
  }

  toBeWithin(min: Position, max: Position): void {
    const actual = this.player.position;

    if (
      actual.x < min.x ||
      actual.x > max.x ||
      actual.y < min.y ||
      actual.y > max.y ||
      actual.z < min.z ||
      actual.z > max.z
    ) {
      throw new AssertionError(
        `Expected position to be within ` +
          `(${min.x}, ${min.y}, ${min.z}) - (${max.x}, ${max.y}, ${max.z}), ` +
          `but was (${actual.x}, ${actual.y}, ${actual.z})`,
        { min, max },
        actual
      );
    }
  }

  toBeAtY(y: number, tolerance = 0.5): void {
    const actual = this.player.position.y;
    const diff = Math.abs(actual - y);

    if (diff > tolerance) {
      throw new AssertionError(
        `Expected Y position to be ${y} (±${tolerance}), but was ${actual}`,
        y,
        actual
      );
    }
  }

  toBeOnGround(): void {
    if (!this.player.state.isOnGround) {
      throw new AssertionError(
        'Expected player to be on ground',
        'on ground',
        'in air'
      );
    }
  }

  toBeInAir(): void {
    if (this.player.state.isOnGround) {
      throw new AssertionError(
        'Expected player to be in air',
        'in air',
        'on ground'
      );
    }
  }

  async toReach(
    expected: Position,
    options?: { timeout?: number; tolerance?: number }
  ): Promise<void> {
    const { timeout = 10000, tolerance = 1 } = options ?? {};

    return new Promise((resolve, reject) => {
      const startTime = Date.now();

      const check = () => {
        const distance = distanceTo(this.player.position, expected);
        if (distance <= tolerance) {
          this.player.off('position_update', check);
          resolve();
          return;
        }

        if (Date.now() - startTime > timeout) {
          this.player.off('position_update', check);
          reject(
            new AssertionError(
              `Timeout waiting for position to reach ` +
                `(${expected.x}, ${expected.y}, ${expected.z})`,
              expected,
              this.player.position
            )
          );
          return;
        }
      };

      // Check immediately
      check();

      // Listen for updates
      this.player.on('position_update', check);
    });
  }

  toBeInDimension(dimension: 'overworld' | 'nether' | 'the_end'): void {
    const actual = this.player.state.dimension;
    if (actual !== dimension) {
      throw new AssertionError(
        `Expected to be in dimension "${dimension}", but was "${actual}"`,
        dimension,
        actual
      );
    }
  }
}
