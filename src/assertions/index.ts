import type { Agent } from '../core/client.js';
import { PositionAssertion } from './position.js';
import { ChatAssertion } from './chat.js';
import { CommandAssertion } from './command.js';
import { FormAssertion } from './form.js';
import { InventoryAssertion } from './inventory.js';
import {
  HealthAssertion,
  HungerAssertion,
  EffectAssertion,
  GamemodeAssertion,
  PermissionAssertion,
  TagAssertion,
} from './player.js';
import { BlockAssertion, EntityAssertion, ScoreboardAssertion } from './block.js';
import {
  TitleAssertion,
  SubtitleAssertion,
  ActionbarAssertion,
  SoundAssertion,
  ParticleAssertion,
} from './display.js';
import {
  ConnectionAssertion,
  TeleportAssertion,
  DimensionAssertion,
  DeathAssertion,
  RespawnAssertion,
} from './event.js';
import { TimingAssertion, SequenceAssertion, ConditionAssertion } from './timing.js';

export class AssertionError extends Error {
  constructor(
    message: string,
    public expected?: unknown,
    public actual?: unknown
  ) {
    super(message);
    this.name = 'AssertionError';
  }
}

export class AssertionContext {
  // 基本アサーション
  private _position: PositionAssertion;
  private _chat: ChatAssertion;
  private _form: FormAssertion;

  // プレイヤー状態系
  private _inventory: InventoryAssertion;
  private _health: HealthAssertion;
  private _hunger: HungerAssertion;
  private _effect: EffectAssertion;
  private _gamemode: GamemodeAssertion;
  private _permission: PermissionAssertion;
  private _tag: TagAssertion;

  // ワールド/ブロック系
  private _block: BlockAssertion;
  private _entity: EntityAssertion;
  private _scoreboard: ScoreboardAssertion;

  // UI/表示系
  private _title: TitleAssertion;
  private _subtitle: SubtitleAssertion;
  private _actionbar: ActionbarAssertion;
  private _sound: SoundAssertion;
  private _particle: ParticleAssertion;

  // イベント系
  private _connection: ConnectionAssertion;
  private _teleport: TeleportAssertion;
  private _dimension: DimensionAssertion;
  private _death: DeathAssertion;
  private _respawn: RespawnAssertion;

  // タイミング系
  private _timing: TimingAssertion;
  private _sequence: SequenceAssertion;
  private _condition: ConditionAssertion;

  constructor(private player: Agent) {
    // 基本アサーション
    this._position = new PositionAssertion(player);
    this._chat = new ChatAssertion(player);
    this._form = new FormAssertion(player);

    // プレイヤー状態系
    this._inventory = new InventoryAssertion(player);
    this._health = new HealthAssertion(player);
    this._hunger = new HungerAssertion(player);
    this._effect = new EffectAssertion(player);
    this._gamemode = new GamemodeAssertion(player);
    this._permission = new PermissionAssertion(player);
    this._tag = new TagAssertion(player);

    // ワールド/ブロック系
    this._block = new BlockAssertion(player);
    this._entity = new EntityAssertion(player);
    this._scoreboard = new ScoreboardAssertion(player);

    // UI/表示系
    this._title = new TitleAssertion(player);
    this._subtitle = new SubtitleAssertion(player);
    this._actionbar = new ActionbarAssertion(player);
    this._sound = new SoundAssertion(player);
    this._particle = new ParticleAssertion(player);

    // イベント系
    this._connection = new ConnectionAssertion(player);
    this._teleport = new TeleportAssertion(player);
    this._dimension = new DimensionAssertion(player);
    this._death = new DeathAssertion(player);
    this._respawn = new RespawnAssertion(player);

    // タイミング系
    this._timing = new TimingAssertion(player);
    this._sequence = new SequenceAssertion(player);
    this._condition = new ConditionAssertion(player);
  }

  // === 基本アサーション ===

  toBeConnected(): void {
    if (!this.player.isConnected) {
      throw new AssertionError(
        'Expected player to be connected',
        'connected',
        'disconnected'
      );
    }
  }

  toBeDisconnected(): void {
    if (this.player.isConnected) {
      throw new AssertionError(
        'Expected player to be disconnected',
        'disconnected',
        'connected'
      );
    }
  }

  get position(): PositionAssertion {
    return this._position;
  }

  get chat(): ChatAssertion {
    return this._chat;
  }

  command(output: import('../types/index.js').CommandOutput): CommandAssertion {
    return new CommandAssertion(output);
  }

  get form(): FormAssertion {
    return this._form;
  }

  // === プレイヤー状態系 ===

  get inventory(): InventoryAssertion {
    return this._inventory;
  }

  get health(): HealthAssertion {
    return this._health;
  }

  get hunger(): HungerAssertion {
    return this._hunger;
  }

  get effect(): EffectAssertion {
    return this._effect;
  }

  get gamemode(): GamemodeAssertion {
    return this._gamemode;
  }

  get permission(): PermissionAssertion {
    return this._permission;
  }

  get tag(): TagAssertion {
    return this._tag;
  }

  // === ワールド/ブロック系 ===

  get block(): BlockAssertion {
    return this._block;
  }

  get entity(): EntityAssertion {
    return this._entity;
  }

  get scoreboard(): ScoreboardAssertion {
    return this._scoreboard;
  }

  // === UI/表示系 ===

  get title(): TitleAssertion {
    return this._title;
  }

  get subtitle(): SubtitleAssertion {
    return this._subtitle;
  }

  get actionbar(): ActionbarAssertion {
    return this._actionbar;
  }

  get sound(): SoundAssertion {
    return this._sound;
  }

  get particle(): ParticleAssertion {
    return this._particle;
  }

  // === イベント系 ===

  get connection(): ConnectionAssertion {
    return this._connection;
  }

  get teleport(): TeleportAssertion {
    return this._teleport;
  }

  get dimension(): DimensionAssertion {
    return this._dimension;
  }

  get death(): DeathAssertion {
    return this._death;
  }

  get respawn(): RespawnAssertion {
    return this._respawn;
  }

  // === タイミング系 ===

  get timing(): TimingAssertion {
    return this._timing;
  }

  get sequence(): SequenceAssertion {
    return this._sequence;
  }

  get condition(): ConditionAssertion {
    return this._condition;
  }
}

// 基本アサーション
export { PositionAssertion } from './position.js';
export { ChatAssertion } from './chat.js';
export { CommandAssertion } from './command.js';
export { FormAssertion, ModalFormAssertion, ActionFormAssertion, CustomFormAssertion } from './form.js';

// プレイヤー状態系
export { InventoryAssertion } from './inventory.js';
export {
  HealthAssertion,
  HungerAssertion,
  EffectAssertion,
  GamemodeAssertion,
  PermissionAssertion,
  TagAssertion,
} from './player.js';

// ワールド/ブロック系
export { BlockAssertion, EntityAssertion, ScoreboardAssertion } from './block.js';

// UI/表示系
export {
  TitleAssertion,
  SubtitleAssertion,
  ActionbarAssertion,
  SoundAssertion,
  ParticleAssertion,
} from './display.js';

// イベント系
export {
  ConnectionAssertion,
  TeleportAssertion,
  DimensionAssertion,
  DeathAssertion,
  RespawnAssertion,
} from './event.js';

// タイミング系
export { TimingAssertion, SequenceAssertion, ConditionAssertion } from './timing.js';
