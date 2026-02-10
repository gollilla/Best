import type { Position } from '../types/index.js';
import { Vec3 } from 'vec3';

// CommonJS modules - require を使用
// eslint-disable-next-line @typescript-eslint/no-require-imports
const createRegistry = require('prismarine-registry');
// eslint-disable-next-line @typescript-eslint/no-require-imports
const createChunk = require('prismarine-chunk');

// Lazy initialization
let ChunkColumnClass: any = null;
let registry: any = null;

/**
 * ブロックの詳細情報
 */
export interface BlockInfo {
  name: string;
  stateId: number;
  hardness: number;
  diggable: boolean;
  material?: string;
  harvestTools?: Record<number, boolean>;
}

/**
 * ブロック破壊に必要な時間を計算するためのデータ
 */
export interface BlockBreakData {
  /** 基本破壊時間（秒） */
  baseTime: number;
  /** 即座に破壊可能か */
  instant: boolean;
  /** 破壊不可能か */
  unbreakable: boolean;
}

export interface BlockPosition {
  x: number;
  y: number;
  z: number;
}

export interface Block {
  name: string;
  stateId: number;
  hardness?: number;
  diggable?: boolean;
}

interface ChunkPosition {
  x: number;
  z: number;
}

/**
 * prismarine-chunk の初期化
 */
function initChunkLoader(version: string): void {
  if (ChunkColumnClass) return;

  // Format: "bedrock_1.21.80"
  const registryVersion = `bedrock_${version}`;
  registry = createRegistry(registryVersion);
  ChunkColumnClass = createChunk(registry);
}

/**
 * チャンクカラム（16x384x16）
 * prismarine-chunk のラッパー
 */
export class ChunkColumn {
  private column: any;
  readonly x: number;
  readonly z: number;

  constructor(x: number, z: number, column: any) {
    this.x = x;
    this.z = z;
    this.column = column;
  }

  /**
   * 指定座標のブロックを取得
   */
  getBlock(localX: number, y: number, localZ: number): Block | null {
    try {
      const block = this.column.getBlock(new Vec3(localX, y, localZ));
      if (!block) return null;

      return {
        name: block.name,
        stateId: block.stateId,
        hardness: block.hardness ?? undefined,
        diggable: block.diggable ?? true,
      };
    } catch {
      return null;
    }
  }

  /**
   * 指定座標のブロック状態IDを取得
   */
  getBlockStateId(localX: number, y: number, localZ: number): number | null {
    try {
      return this.column.getBlockStateId(new Vec3(localX, y, localZ));
    } catch {
      return null;
    }
  }
}

/**
 * ワールドデータを管理するクラス
 * チャンクデータとブロック情報を保持
 */
export class World {
  private chunks: Map<string, ChunkColumn> = new Map();
  private version: string;
  private initialized = false;

  constructor(version: string = '1.21.130') {
    this.version = version;
  }

  /**
   * 初期化（prismarine-chunk のロード）
   */
  initialize(): void {
    if (this.initialized) return;
    initChunkLoader(this.version);
    this.initialized = true;
  }

  /**
   * チャンクキーを生成
   */
  private getChunkKey(x: number, z: number): string {
    return `${x},${z}`;
  }

  /**
   * ブロック座標からチャンク座標を計算
   */
  private toChunkPosition(x: number, z: number): ChunkPosition {
    return {
      x: Math.floor(x / 16),
      z: Math.floor(z / 16),
    };
  }

  /**
   * level_chunk パケットからチャンクをデコード
   */
  async decodeChunk(
    chunkX: number,
    chunkZ: number,
    subChunkCount: number,
    payload: Buffer
  ): Promise<ChunkColumn> {
    this.initialize();

    const column = new ChunkColumnClass({ x: chunkX, z: chunkZ });
    await column.networkDecodeNoCache(payload, subChunkCount);

    const chunk = new ChunkColumn(chunkX, chunkZ, column);
    const key = this.getChunkKey(chunkX, chunkZ);
    this.chunks.set(key, chunk);

    return chunk;
  }

  /**
   * チャンクデータを取得
   */
  getChunk(chunkX: number, chunkZ: number): ChunkColumn | undefined {
    const key = this.getChunkKey(chunkX, chunkZ);
    return this.chunks.get(key);
  }

  /**
   * チャンクがロード済みか確認
   */
  hasChunk(chunkX: number, chunkZ: number): boolean {
    const key = this.getChunkKey(chunkX, chunkZ);
    return this.chunks.has(key);
  }

  /**
   * 指定座標のブロックを取得
   */
  getBlock(x: number, y: number, z: number): Block | null {
    const chunkPos = this.toChunkPosition(x, z);
    const chunk = this.getChunk(chunkPos.x, chunkPos.z);

    if (!chunk) {
      return null;
    }

    const localX = ((x % 16) + 16) % 16;
    const localZ = ((z % 16) + 16) % 16;

    return chunk.getBlock(localX, y, localZ);
  }

  /**
   * 指定座標のブロック名を取得
   */
  getBlockName(x: number, y: number, z: number): string | null {
    const block = this.getBlock(x, y, z);
    return block?.name ?? null;
  }

  /**
   * 指定座標が通行可能か（空気または通過可能ブロック）
   */
  isPassable(x: number, y: number, z: number): boolean {
    const block = this.getBlock(x, y, z);
    if (!block) return false;

    return this.isPassableBlock(block.name);
  }

  /**
   * ブロックが通行可能か判定
   */
  private isPassableBlock(name: string): boolean {
    const passableBlocks = [
      'air',
      'cave_air',
      'void_air',
      'water',
      'flowing_water',
      'lava',
      'flowing_lava',
      'grass',
      'tallgrass',
      'tall_grass',
      'deadbush',
      'seagrass',
      'tall_seagrass',
      'fire',
      'soul_fire',
      'snow',
      'vine',
      'lily_pad',
      'torch',
      'redstone_torch',
      'soul_torch',
      'rail',
      'powered_rail',
      'detector_rail',
      'activator_rail',
    ];

    if (
      name.includes('flower') ||
      name.includes('coral') ||
      name.includes('sapling')
    ) {
      return true;
    }

    return passableBlocks.includes(name);
  }

  /**
   * 指定座標が固体ブロックか
   */
  isSolid(x: number, y: number, z: number): boolean {
    const block = this.getBlock(x, y, z);
    if (!block) return true;

    return !this.isPassableBlock(block.name);
  }

  /**
   * 周囲のブロック情報を取得
   */
  getBlocksAround(
    center: Position,
    radius: number = 1
  ): Map<string, Block | null> {
    const blocks = new Map<string, Block | null>();

    for (let dx = -radius; dx <= radius; dx++) {
      for (let dy = -radius; dy <= radius; dy++) {
        for (let dz = -radius; dz <= radius; dz++) {
          const x = Math.floor(center.x) + dx;
          const y = Math.floor(center.y) + dy;
          const z = Math.floor(center.z) + dz;
          const key = `${x},${y},${z}`;
          blocks.set(key, this.getBlock(x, y, z));
        }
      }
    }

    return blocks;
  }

  /**
   * プレイヤーの足元のブロックを取得
   */
  getBlockBelow(position: Position): Block | null {
    return this.getBlock(
      Math.floor(position.x),
      Math.floor(position.y) - 1,
      Math.floor(position.z)
    );
  }

  /**
   * プレイヤーの前方のブロックを取得（yaw角度から計算）
   */
  getBlockInFront(
    position: Position,
    yaw: number,
    distance: number = 1
  ): Block | null {
    const radians = (yaw * Math.PI) / 180;
    const dx = -Math.sin(radians) * distance;
    const dz = Math.cos(radians) * distance;

    return this.getBlock(
      Math.floor(position.x + dx),
      Math.floor(position.y),
      Math.floor(position.z + dz)
    );
  }

  /**
   * ロード済みチャンク数を取得
   */
  get loadedChunkCount(): number {
    return this.chunks.size;
  }

  /**
   * ワールドデータをクリア
   */
  clear(): void {
    this.chunks.clear();
  }

  /**
   * ブロックの破壊データを取得
   * @param x X座標
   * @param y Y座標
   * @param z Z座標
   * @param toolMultiplier ツールによる倍率（デフォルト: 1.0 = 素手）
   */
  getBlockBreakData(
    x: number,
    y: number,
    z: number,
    toolMultiplier: number = 1.0
  ): BlockBreakData {
    const block = this.getBlock(x, y, z);

    if (!block) {
      return { baseTime: 0, instant: true, unbreakable: false };
    }

    // 空気や破壊不可ブロックの処理
    if (block.name === 'air' || block.name === 'cave_air' || block.name === 'void_air') {
      return { baseTime: 0, instant: true, unbreakable: false };
    }

    // 破壊不可能ブロック
    const unbreakableBlocks = [
      'bedrock',
      'command_block',
      'chain_command_block',
      'repeating_command_block',
      'structure_block',
      'jigsaw',
      'barrier',
      'invisible_bedrock',
      'end_portal',
      'end_portal_frame',
      'end_gateway',
    ];
    if (unbreakableBlocks.includes(block.name)) {
      return { baseTime: Infinity, instant: false, unbreakable: true };
    }

    // 硬さがない場合はデフォルト値を使用
    const hardness = block.hardness ?? this.getDefaultHardness(block.name);

    // 硬さが0以下は即座に破壊
    if (hardness <= 0) {
      return { baseTime: 0, instant: true, unbreakable: false };
    }

    // 破壊時間の計算（秒）
    // 基本式: hardness * 1.5 / toolMultiplier
    const baseTime = (hardness * 1.5) / toolMultiplier;

    return {
      baseTime,
      instant: baseTime <= 0.05, // 1tick以下は即座に破壊
      unbreakable: false,
    };
  }

  /**
   * ブロック名からデフォルトの硬さを取得
   */
  private getDefaultHardness(name: string): number {
    // 一般的なブロックの硬さマップ
    const hardnessMap: Record<string, number> = {
      // 即座に破壊
      grass: 0,
      tallgrass: 0,
      tall_grass: 0,
      deadbush: 0,
      flower: 0,
      torch: 0,
      redstone_torch: 0,
      soul_torch: 0,
      snow: 0.1,
      vine: 0.2,

      // 柔らかいブロック
      dirt: 0.5,
      sand: 0.5,
      gravel: 0.6,
      clay: 0.6,
      soul_sand: 0.5,
      farmland: 0.6,
      grass_block: 0.6,

      // 木材系
      oak_planks: 2.0,
      spruce_planks: 2.0,
      birch_planks: 2.0,
      jungle_planks: 2.0,
      acacia_planks: 2.0,
      dark_oak_planks: 2.0,
      oak_log: 2.0,
      spruce_log: 2.0,
      birch_log: 2.0,
      jungle_log: 2.0,
      acacia_log: 2.0,
      dark_oak_log: 2.0,

      // 石系
      stone: 1.5,
      cobblestone: 2.0,
      mossy_cobblestone: 2.0,
      stone_bricks: 1.5,
      deepslate: 3.0,
      cobbled_deepslate: 3.5,

      // 鉱石系
      coal_ore: 3.0,
      iron_ore: 3.0,
      gold_ore: 3.0,
      diamond_ore: 3.0,
      emerald_ore: 3.0,
      lapis_ore: 3.0,
      redstone_ore: 3.0,
      copper_ore: 3.0,

      // 金属ブロック
      iron_block: 5.0,
      gold_block: 3.0,
      diamond_block: 5.0,
      emerald_block: 5.0,
      netherite_block: 50.0,

      // その他
      obsidian: 50.0,
      crying_obsidian: 50.0,
      ancient_debris: 30.0,
      ender_chest: 22.5,
    };

    // 完全一致を探す
    if (hardnessMap[name] !== undefined) {
      return hardnessMap[name];
    }

    // パターンマッチング
    if (name.includes('planks')) return 2.0;
    if (name.includes('log') || name.includes('wood')) return 2.0;
    if (name.includes('ore')) return 3.0;
    if (name.includes('deepslate')) return 3.0;
    if (name.includes('stone')) return 1.5;
    if (name.includes('brick')) return 2.0;
    if (name.includes('wool')) return 0.8;
    if (name.includes('glass')) return 0.3;
    if (name.includes('leaves')) return 0.2;
    if (name.includes('flower') || name.includes('sapling')) return 0;
    if (name.includes('coral')) return 0;

    // デフォルト
    return 1.5;
  }
}
