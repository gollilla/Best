import type { Agent } from '../core/client.js';
import type { Position, Entity, ScoreboardEntry } from '../types/index.js';
import { AssertionError } from './index.js';

// ブロックアサーション
export class BlockAssertion {
  constructor(private player: Agent) {}

  /**
   * 指定座標のブロックが指定タイプか確認
   */
  toBeAt(position: Position, blockId: string): this {
    const x = Math.floor(position.x);
    const y = Math.floor(position.y);
    const z = Math.floor(position.z);

    const block = this.player.getBlock(x, y, z);
    const normalizedId = blockId.startsWith('minecraft:') ? blockId : `minecraft:${blockId}`;

    if (!block || block.name !== normalizedId) {
      throw new AssertionError(
        `Expected block at (${x}, ${y}, ${z}) to be "${normalizedId}", but was "${block?.name ?? 'unknown'}"`,
        normalizedId,
        block?.name
      );
    }
    return this;
  }

  /**
   * 指定座標が空気ブロックか確認
   */
  toBeAirAt(position: Position): this {
    return this.toBeAt(position, 'minecraft:air');
  }

  /**
   * 指定座標が空気でないか確認
   */
  notToBeAirAt(position: Position): this {
    const x = Math.floor(position.x);
    const y = Math.floor(position.y);
    const z = Math.floor(position.z);

    const block = this.player.getBlock(x, y, z);

    if (!block || block.name === 'minecraft:air') {
      throw new AssertionError(
        `Expected block at (${x}, ${y}, ${z}) not to be air`,
        'non-air block',
        'minecraft:air'
      );
    }
    return this;
  }

  /**
   * 指定座標が通行可能か確認
   */
  toBePassableAt(position: Position): this {
    const x = Math.floor(position.x);
    const y = Math.floor(position.y);
    const z = Math.floor(position.z);

    if (!this.player.isPassable(x, y, z)) {
      throw new AssertionError(
        `Expected block at (${x}, ${y}, ${z}) to be passable`,
        'passable',
        'solid'
      );
    }
    return this;
  }

  /**
   * 指定座標が固体ブロックか確認
   */
  toBeSolidAt(position: Position): this {
    const x = Math.floor(position.x);
    const y = Math.floor(position.y);
    const z = Math.floor(position.z);

    if (!this.player.isSolid(x, y, z)) {
      throw new AssertionError(
        `Expected block at (${x}, ${y}, ${z}) to be solid`,
        'solid',
        'passable'
      );
    }
    return this;
  }

  /**
   * ブロック変更を待機
   */
  async toChangeTo(
    position: Position,
    blockId: string,
    options?: { timeout?: number }
  ): Promise<void> {
    const { timeout = 5000 } = options ?? {};
    const normalizedId = blockId.startsWith('minecraft:') ? blockId : `minecraft:${blockId}`;
    const x = Math.floor(position.x);
    const y = Math.floor(position.y);
    const z = Math.floor(position.z);

    try {
      await this.player.waitFor('block_update', {
        timeout,
        filter: (update: { position: Position; runtimeId: number }) =>
          Math.floor(update.position.x) === x &&
          Math.floor(update.position.y) === y &&
          Math.floor(update.position.z) === z,
      });

      // 変更後のブロックを確認
      const block = this.player.getBlock(x, y, z);
      if (!block || block.name !== normalizedId) {
        throw new AssertionError(
          `Block changed but not to "${normalizedId}", was "${block?.name}"`,
          normalizedId,
          block?.name
        );
      }
    } catch {
      throw new AssertionError(
        `Timeout waiting for block at (${x}, ${y}, ${z}) to change to "${normalizedId}"`,
        normalizedId,
        this.player.getBlockName(x, y, z)
      );
    }
  }
}

// エンティティアサーション
export class EntityAssertion {
  constructor(private player: Agent) {}

  /**
   * 指定タイプのエンティティが存在するか確認
   */
  toExist(entityType: string): this {
    const entities = this.player.getEntities();
    const normalizedType = entityType.startsWith('minecraft:') ? entityType : `minecraft:${entityType}`;

    const found = entities.find((e) => e.type === normalizedType);
    if (!found) {
      throw new AssertionError(
        `Expected entity "${normalizedType}" to exist`,
        normalizedType,
        entities.map((e) => e.type)
      );
    }
    return this;
  }

  /**
   * 指定タイプのエンティティが存在しないか確認
   */
  notToExist(entityType: string): this {
    const entities = this.player.getEntities();
    const normalizedType = entityType.startsWith('minecraft:') ? entityType : `minecraft:${entityType}`;

    const found = entities.find((e) => e.type === normalizedType);
    if (found) {
      throw new AssertionError(
        `Expected entity "${normalizedType}" not to exist`,
        undefined,
        normalizedType
      );
    }
    return this;
  }

  /**
   * 指定タイプのエンティティが近くにいるか確認
   */
  toBeNearby(entityType: string, radius: number): this {
    const entities = this.player.getEntities();
    const normalizedType = entityType.startsWith('minecraft:') ? entityType : `minecraft:${entityType}`;
    const playerPos = this.player.position;

    const nearby = entities.filter((e) => {
      if (e.type !== normalizedType) return false;
      const dx = e.position.x - playerPos.x;
      const dy = e.position.y - playerPos.y;
      const dz = e.position.z - playerPos.z;
      return Math.sqrt(dx * dx + dy * dy + dz * dz) <= radius;
    });

    if (nearby.length === 0) {
      throw new AssertionError(
        `Expected entity "${normalizedType}" to be within ${radius} blocks`,
        normalizedType,
        undefined
      );
    }
    return this;
  }

  /**
   * 指定タイプのエンティティが指定個数以上いるか確認
   */
  toHaveCount(entityType: string, minCount: number): this {
    const entities = this.player.getEntities();
    const normalizedType = entityType.startsWith('minecraft:') ? entityType : `minecraft:${entityType}`;

    const count = entities.filter((e) => e.type === normalizedType).length;

    if (count < minCount) {
      throw new AssertionError(
        `Expected at least ${minCount} entities of type "${normalizedType}", but found ${count}`,
        minCount,
        count
      );
    }
    return this;
  }

  /**
   * エンティティスポーンを待機
   */
  async toSpawn(
    entityType: string,
    options?: { timeout?: number; nearPlayer?: number }
  ): Promise<Entity> {
    const { timeout = 5000, nearPlayer } = options ?? {};
    const normalizedType = entityType.startsWith('minecraft:') ? entityType : `minecraft:${entityType}`;

    try {
      const [entity] = await this.player.waitFor('entity_spawn', {
        timeout,
        filter: (e: Entity) => {
          if (e.type !== normalizedType) return false;
          if (nearPlayer !== undefined) {
            const playerPos = this.player.position;
            const dx = e.position.x - playerPos.x;
            const dy = e.position.y - playerPos.y;
            const dz = e.position.z - playerPos.z;
            return Math.sqrt(dx * dx + dy * dy + dz * dz) <= nearPlayer;
          }
          return true;
        },
      });
      return entity;
    } catch {
      throw new AssertionError(
        `Timeout waiting for entity "${normalizedType}" to spawn`,
        normalizedType,
        undefined
      );
    }
  }
}

// スコアボードアサーション
export class ScoreboardAssertion {
  constructor(private player: Agent) {}

  /**
   * スコアボードの値を確認
   */
  toHaveValue(objective: string, expectedScore: number): this {
    const score = this.player.getScore(objective);

    if (score === null) {
      throw new AssertionError(
        `Player has no score in objective "${objective}"`,
        expectedScore,
        undefined
      );
    }

    if (score !== expectedScore) {
      throw new AssertionError(
        `Expected score in "${objective}" to be ${expectedScore}, but was ${score}`,
        expectedScore,
        score
      );
    }
    return this;
  }

  /**
   * スコアボードの値が指定値以上か確認
   */
  toHaveMinValue(objective: string, minScore: number): this {
    const score = this.player.getScore(objective);

    if (score === null) {
      throw new AssertionError(
        `Player has no score in objective "${objective}"`,
        `>= ${minScore}`,
        undefined
      );
    }

    if (score < minScore) {
      throw new AssertionError(
        `Expected score in "${objective}" to be at least ${minScore}, but was ${score}`,
        `>= ${minScore}`,
        score
      );
    }
    return this;
  }

  /**
   * スコアボードにオブジェクティブが存在するか確認
   */
  toHaveObjective(objective: string): this {
    const objectives = this.player.getScoreboardObjectives();

    if (!objectives.includes(objective)) {
      throw new AssertionError(
        `Expected scoreboard objective "${objective}" to exist`,
        objective,
        objectives
      );
    }
    return this;
  }

  /**
   * スコア変更を待機
   */
  async toReachValue(
    objective: string,
    targetScore: number,
    options?: { timeout?: number }
  ): Promise<number> {
    const { timeout = 10000 } = options ?? {};

    try {
      const [entry] = await this.player.waitFor('score_update', {
        timeout,
        filter: (e: ScoreboardEntry) =>
          e.objective === objective && e.score >= targetScore,
      });
      return entry.score;
    } catch {
      throw new AssertionError(
        `Timeout waiting for score in "${objective}" to reach ${targetScore}`,
        targetScore,
        this.player.getScore(objective)
      );
    }
  }
}
