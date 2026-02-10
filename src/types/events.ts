import type {
  Position,
  CommandOutput,
  InventoryItem,
  Effect,
  Entity,
  ScoreboardEntry,
  TitleDisplay,
  SoundPlay,
  ParticleSpawn,
} from './index.js';
import type { Form } from './forms.js';

export interface ChatMessage {
  type: 'chat' | 'system' | 'whisper' | 'announcement' | 'raw' | 'tip' | 'json';
  sender?: string;
  message: string;
  timestamp: number;
  xuid?: string;
}

export interface ChunkPosition {
  x: number;
  z: number;
}

export interface BlockUpdate {
  position: Position;
  runtimeId: number;
}

export interface BlockBreakEvent {
  position: Position;
}

export interface InventoryUpdate {
  slot: number;
  item: InventoryItem;
}

export type ClientEvents = {
  // Connection events
  join: [];
  spawn: [];
  disconnect: [reason: string];
  error: [error: Error];

  // Game events
  chat: [message: ChatMessage];
  command_output: [output: CommandOutput];
  position_update: [position: Position];
  health_update: [health: number];
  hunger_update: [hunger: number];
  gamemode_update: [gamemode: number];

  // Form events
  form: [form: Form];

  // World events
  chunk_loaded: [position: ChunkPosition];
  block_update: [update: BlockUpdate];

  // Block break events
  block_break_start: [event: BlockBreakEvent];
  block_break_complete: [event: BlockBreakEvent];
  block_break_abort: [event: BlockBreakEvent];

  // Inventory events
  inventory_update: [update: InventoryUpdate];

  // Effect events
  effect_add: [effect: Effect];
  effect_remove: [effectId: string];

  // Entity events
  entity_spawn: [entity: Entity];
  entity_remove: [runtimeId: bigint];

  // Scoreboard events
  score_update: [entry: ScoreboardEntry];

  // Title events
  title: [display: TitleDisplay];

  // Sound events
  sound: [sound: SoundPlay];

  // Particle events
  particle: [particle: ParticleSpawn];

  // Dimension events
  dimension_change: [dimension: string];

  // Death/Respawn events
  death: [];
  respawn: [position: Position];

  // Raw packet events
  packet: [name: string, data: unknown];

  // Index signature for extensibility
  [key: string]: unknown[];
};

export type ClientEventName = keyof ClientEvents;
