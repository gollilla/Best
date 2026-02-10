export * from './events.js';
export * from './forms.js';

export interface ClientOptions {
  host: string;
  port?: number;
  username: string;
  offline?: boolean;
  timeout?: number;
  version?: string;
}

export interface Position {
  x: number;
  y: number;
  z: number;
}

export interface Rotation {
  pitch: number;
  yaw: number;
}

export interface PlayerState {
  position: Position;
  rotation: Rotation;
  health: number;
  gamemode: number;
  dimension: string;
  isOnGround: boolean;
  runtimeEntityId: bigint;
}

export interface CommandOutput {
  command: string;
  success: boolean;
  output: string;
  statusCode: number;
}

export interface ServerInfo {
  host: string;
  port: number;
  version: string;
}

// インベントリ関連
export interface InventoryItem {
  id: string;
  count: number;
  slot: number;
  damage?: number;
  enchantments?: Enchantment[];
  customName?: string;
  lore?: string[];
}

export interface Enchantment {
  id: string;
  level: number;
}

// エフェクト関連
export interface Effect {
  id: string;
  amplifier: number;
  duration: number;
  visible: boolean;
}

// エンティティ関連
export interface Entity {
  runtimeId: bigint;
  type: string;
  position: Position;
  nameTag?: string;
}

// スコアボード関連
export interface ScoreboardEntry {
  objective: string;
  score: number;
  displayName?: string;
}

// タイトル表示関連
export interface TitleDisplay {
  type: 'title' | 'subtitle' | 'actionbar';
  text: string;
  fadeIn?: number;
  stay?: number;
  fadeOut?: number;
}

// サウンド関連
export interface SoundPlay {
  name: string;
  position: Position;
  volume: number;
  pitch: number;
}

// パーティクル関連
export interface ParticleSpawn {
  name: string;
  position: Position;
  data?: number;
}

// ゲームモード
export type GameMode = 'survival' | 'creative' | 'adventure' | 'spectator';

export const GameModeMap: Record<number, GameMode> = {
  0: 'survival',
  1: 'creative',
  2: 'adventure',
  3: 'spectator',
};

// 権限レベル
export type PermissionLevel = 'visitor' | 'member' | 'operator' | 'custom';

// ディメンション
export type Dimension = 'overworld' | 'nether' | 'the_end';
