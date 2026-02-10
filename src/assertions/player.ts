import type { Agent } from '../core/client.js';
import type { Effect, GameMode, PermissionLevel } from '../types/index.js';
import { GameModeMap } from '../types/index.js';
import { AssertionError } from './index.js';

// 体力アサーション
export class HealthAssertion {
  constructor(private player: Agent) {}

  toBe(expected: number, tolerance = 0.5): this {
    const actual = this.player.health;
    if (Math.abs(actual - expected) > tolerance) {
      throw new AssertionError(
        `Expected health to be ${expected} (±${tolerance}), but was ${actual}`,
        expected,
        actual
      );
    }
    return this;
  }

  toBeAbove(min: number): this {
    const actual = this.player.health;
    if (actual <= min) {
      throw new AssertionError(
        `Expected health to be above ${min}, but was ${actual}`,
        `> ${min}`,
        actual
      );
    }
    return this;
  }

  toBeBelow(max: number): this {
    const actual = this.player.health;
    if (actual >= max) {
      throw new AssertionError(
        `Expected health to be below ${max}, but was ${actual}`,
        `< ${max}`,
        actual
      );
    }
    return this;
  }

  toBeFull(): this {
    const actual = this.player.health;
    if (actual < 20) {
      throw new AssertionError(
        `Expected health to be full (20), but was ${actual}`,
        20,
        actual
      );
    }
    return this;
  }

  toBeDead(): this {
    const actual = this.player.health;
    if (actual > 0) {
      throw new AssertionError(
        `Expected player to be dead, but health was ${actual}`,
        0,
        actual
      );
    }
    return this;
  }

  async toReach(target: number, options?: { timeout?: number }): Promise<void> {
    const { timeout = 10000 } = options ?? {};

    return new Promise((resolve, reject) => {
      const startTime = Date.now();

      const check = (health: number) => {
        if (health >= target) {
          this.player.off('health_update', check);
          resolve();
          return;
        }

        if (Date.now() - startTime > timeout) {
          this.player.off('health_update', check);
          reject(
            new AssertionError(
              `Timeout waiting for health to reach ${target}`,
              target,
              this.player.health
            )
          );
        }
      };

      if (this.player.health >= target) {
        resolve();
        return;
      }

      this.player.on('health_update', check);

      setTimeout(() => {
        this.player.off('health_update', check);
        if (this.player.health < target) {
          reject(
            new AssertionError(
              `Timeout waiting for health to reach ${target}`,
              target,
              this.player.health
            )
          );
        }
      }, timeout);
    });
  }
}

// 満腹度アサーション
export class HungerAssertion {
  constructor(private player: Agent) {}

  toBe(expected: number, tolerance = 0.5): this {
    const actual = this.player.getHunger();
    if (Math.abs(actual - expected) > tolerance) {
      throw new AssertionError(
        `Expected hunger to be ${expected} (±${tolerance}), but was ${actual}`,
        expected,
        actual
      );
    }
    return this;
  }

  toBeAbove(min: number): this {
    const actual = this.player.getHunger();
    if (actual <= min) {
      throw new AssertionError(
        `Expected hunger to be above ${min}, but was ${actual}`,
        `> ${min}`,
        actual
      );
    }
    return this;
  }

  toBeBelow(max: number): this {
    const actual = this.player.getHunger();
    if (actual >= max) {
      throw new AssertionError(
        `Expected hunger to be below ${max}, but was ${actual}`,
        `< ${max}`,
        actual
      );
    }
    return this;
  }

  toBeFull(): this {
    const actual = this.player.getHunger();
    if (actual < 20) {
      throw new AssertionError(
        `Expected hunger to be full (20), but was ${actual}`,
        20,
        actual
      );
    }
    return this;
  }
}

// エフェクトアサーション
export class EffectAssertion {
  constructor(private player: Agent) {}

  toHave(effectId: string): this {
    const effects = this.player.getEffects();
    const normalizedId = effectId.startsWith('minecraft:') ? effectId : `minecraft:${effectId}`;

    const found = effects.find((e) => e.id === normalizedId);
    if (!found) {
      throw new AssertionError(
        `Expected player to have effect "${normalizedId}"`,
        normalizedId,
        effects.map((e) => e.id)
      );
    }
    return this;
  }

  notToHave(effectId: string): this {
    const effects = this.player.getEffects();
    const normalizedId = effectId.startsWith('minecraft:') ? effectId : `minecraft:${effectId}`;

    const found = effects.find((e) => e.id === normalizedId);
    if (found) {
      throw new AssertionError(
        `Expected player not to have effect "${normalizedId}"`,
        undefined,
        normalizedId
      );
    }
    return this;
  }

  toHaveLevel(effectId: string, level: number): this {
    const effects = this.player.getEffects();
    const normalizedId = effectId.startsWith('minecraft:') ? effectId : `minecraft:${effectId}`;

    const found = effects.find((e) => e.id === normalizedId);
    if (!found) {
      throw new AssertionError(
        `Expected player to have effect "${normalizedId}"`,
        normalizedId,
        effects.map((e) => e.id)
      );
    }

    // amplifier は 0-indexed (amplifier 0 = レベル 1)
    const actualLevel = found.amplifier + 1;
    if (actualLevel !== level) {
      throw new AssertionError(
        `Expected effect "${normalizedId}" to be level ${level}, but was level ${actualLevel}`,
        level,
        actualLevel
      );
    }
    return this;
  }

  toHaveMinDuration(effectId: string, minDurationTicks: number): this {
    const effects = this.player.getEffects();
    const normalizedId = effectId.startsWith('minecraft:') ? effectId : `minecraft:${effectId}`;

    const found = effects.find((e) => e.id === normalizedId);
    if (!found) {
      throw new AssertionError(
        `Expected player to have effect "${normalizedId}"`,
        normalizedId,
        undefined
      );
    }

    if (found.duration < minDurationTicks) {
      throw new AssertionError(
        `Expected effect "${normalizedId}" to have at least ${minDurationTicks} ticks remaining, but had ${found.duration}`,
        minDurationTicks,
        found.duration
      );
    }
    return this;
  }

  async toReceive(
    effectId: string,
    options?: { timeout?: number }
  ): Promise<Effect> {
    const { timeout = 5000 } = options ?? {};
    const normalizedId = effectId.startsWith('minecraft:') ? effectId : `minecraft:${effectId}`;

    try {
      const [effect] = await this.player.waitFor('effect_add', {
        timeout,
        filter: (e: Effect) => e.id === normalizedId,
      });
      return effect;
    } catch {
      throw new AssertionError(
        `Timeout waiting for effect "${normalizedId}"`,
        normalizedId,
        undefined
      );
    }
  }
}

// ゲームモードアサーション
export class GamemodeAssertion {
  constructor(private player: Agent) {}

  toBe(expected: GameMode | number): this {
    const actualNum = this.player.gamemode;
    const actual = GameModeMap[actualNum] ?? 'unknown';

    const expectedMode = typeof expected === 'number' ? GameModeMap[expected] : expected;

    if (actual !== expectedMode) {
      throw new AssertionError(
        `Expected gamemode to be "${expectedMode}", but was "${actual}"`,
        expectedMode,
        actual
      );
    }
    return this;
  }

  toBeSurvival(): this {
    return this.toBe('survival');
  }

  toBeCreative(): this {
    return this.toBe('creative');
  }

  toBeAdventure(): this {
    return this.toBe('adventure');
  }

  toBeSpectator(): this {
    return this.toBe('spectator');
  }

  async toChangeTo(expected: GameMode, options?: { timeout?: number }): Promise<void> {
    const { timeout = 5000 } = options ?? {};

    const expectedNum = Object.entries(GameModeMap).find(([_, v]) => v === expected)?.[0];
    if (expectedNum === undefined) {
      throw new Error(`Invalid gamemode: ${expected}`);
    }

    try {
      await this.player.waitFor('gamemode_update', {
        timeout,
        filter: (gm: number) => gm === Number(expectedNum),
      });
    } catch {
      throw new AssertionError(
        `Timeout waiting for gamemode to change to "${expected}"`,
        expected,
        GameModeMap[this.player.gamemode]
      );
    }
  }
}

// 権限レベルアサーション
export class PermissionAssertion {
  constructor(private player: Agent) {}

  toBeOperator(): this {
    const level = this.player.getPermissionLevel();
    if (level < 2) {
      throw new AssertionError(
        'Expected player to be operator',
        'operator',
        this.getLevelName(level)
      );
    }
    return this;
  }

  notToBeOperator(): this {
    const level = this.player.getPermissionLevel();
    if (level >= 2) {
      throw new AssertionError(
        'Expected player not to be operator',
        'non-operator',
        'operator'
      );
    }
    return this;
  }

  toHaveLevel(expected: number): this {
    const actual = this.player.getPermissionLevel();
    if (actual !== expected) {
      throw new AssertionError(
        `Expected permission level ${expected}, but was ${actual}`,
        expected,
        actual
      );
    }
    return this;
  }

  toHaveMinLevel(min: number): this {
    const actual = this.player.getPermissionLevel();
    if (actual < min) {
      throw new AssertionError(
        `Expected permission level at least ${min}, but was ${actual}`,
        `>= ${min}`,
        actual
      );
    }
    return this;
  }

  private getLevelName(level: number): string {
    switch (level) {
      case 0:
        return 'visitor';
      case 1:
        return 'member';
      case 2:
        return 'operator';
      case 3:
        return 'custom';
      default:
        return `level ${level}`;
    }
  }
}

// タグアサーション
export class TagAssertion {
  constructor(private player: Agent) {}

  toHave(tag: string): this {
    const tags = this.player.getTags();
    if (!tags.includes(tag)) {
      throw new AssertionError(
        `Expected player to have tag "${tag}"`,
        tag,
        tags
      );
    }
    return this;
  }

  notToHave(tag: string): this {
    const tags = this.player.getTags();
    if (tags.includes(tag)) {
      throw new AssertionError(
        `Expected player not to have tag "${tag}"`,
        undefined,
        tag
      );
    }
    return this;
  }

  toHaveAny(...expectedTags: string[]): this {
    const tags = this.player.getTags();
    const hasAny = expectedTags.some((t) => tags.includes(t));

    if (!hasAny) {
      throw new AssertionError(
        `Expected player to have at least one of tags: ${expectedTags.join(', ')}`,
        expectedTags,
        tags
      );
    }
    return this;
  }

  toHaveAll(...expectedTags: string[]): this {
    const tags = this.player.getTags();
    const missing = expectedTags.filter((t) => !tags.includes(t));

    if (missing.length > 0) {
      throw new AssertionError(
        `Expected player to have all tags: ${expectedTags.join(', ')}. Missing: ${missing.join(', ')}`,
        expectedTags,
        tags
      );
    }
    return this;
  }
}
