import bedrock from 'bedrock-protocol';
import { TypedEventEmitter } from './events.js';
import { createInitialState } from './state.js';
import { World, type Block, type BlockBreakData } from './world.js';
import { AssertionContext } from '../assertions/index.js';
import type {
  ClientEvents,
  ClientOptions,
  PlayerState,
  Position,
  CommandOutput,
  ChatMessage,
  Form,
  FormResponse,
  InventoryItem,
  Effect,
  Entity,
  TitleDisplay,
  SoundPlay,
  ParticleSpawn,
} from '../types/index.js';

type BedrockClient = ReturnType<typeof bedrock.createClient>;

export interface AgentOptions {
  /** コマンドプレフィックス */
  commandPrefix?: string;
}

export interface Task {
  name: string;
  execute: (agent: Agent) => Promise<void>;
  cancel?: () => void;
}

type ChatHandler = (
  message: ChatMessage,
  reply: (msg: string) => void
) => void | Promise<void>;

type CommandHandler = (
  args: string[],
  reply: (msg: string) => void
) => void | Promise<void>;

/**
 * Agent: マインクラフト統合版サーバーに接続する仮想プレイヤー
 *
 * 使用例:
 * ```typescript
 * const agent = new Agent({ host: 'localhost', username: 'TestBot' });
 * await agent.connect();
 * await agent.goto({ x: 100, y: 64, z: 100 });
 * agent.expect.position.toBeNear({ x: 100, y: 64, z: 100 }, 5);
 * ```
 */
export class Agent extends TypedEventEmitter<ClientEvents> {
  readonly username: string;
  private readonly options: Required<ClientOptions>;
  private client: BedrockClient | null = null;
  private _state: PlayerState;
  private _isConnected = false;
  private pendingForms: Map<number, Form> = new Map();
  private commandCallbacks: Map<
    string,
    { resolve: (output: CommandOutput) => void; reject: (err: Error) => void }
  > = new Map();

  // Agent機能
  private chatHandlers: Array<{
    pattern: string | RegExp;
    handler: ChatHandler;
  }> = [];
  private commandHandlers: Map<string, CommandHandler> = new Map();
  private _tasks: TaskRunner;
  private _isAgentRunning = false;
  private _expect: AssertionContext;
  private _world: World;
  private commandPrefix: string;

  // プレイヤー状態
  private _inventory: InventoryItem[] = [];
  private _effects: Effect[] = [];
  private _entities: Map<bigint, Entity> = new Map();
  private _scores: Map<string, number> = new Map();
  private _scoreboardObjectives: string[] = [];
  private _tags: string[] = [];
  private _hunger = 20;
  private _permissionLevel = 0;

  constructor(options: ClientOptions, agentOptions?: AgentOptions) {
    super();
    this.options = {
      host: options.host,
      port: options.port ?? 19132,
      username: options.username,
      offline: options.offline ?? true,
      timeout: options.timeout ?? 30000,
      version: options.version ?? '1.21.130',
    };
    this.username = this.options.username;
    this._state = createInitialState();

    // Agent機能の初期化
    this.commandPrefix = agentOptions?.commandPrefix ?? '!';
    this._tasks = new TaskRunner(this);
    this._expect = new AssertionContext(this);
    this._world = new World(this.options.version);
  }

  get state(): Readonly<PlayerState> {
    return this._state;
  }

  get position(): Position {
    return { ...this._state.position };
  }

  get health(): number {
    return this._state.health;
  }

  get gamemode(): number {
    return this._state.gamemode;
  }

  get isConnected(): boolean {
    return this._isConnected;
  }

  // === プレイヤー状態取得 ===

  /**
   * インベントリを取得
   */
  getInventory(): InventoryItem[] {
    return [...this._inventory];
  }

  /**
   * 満腹度を取得
   */
  getHunger(): number {
    return this._hunger;
  }

  /**
   * エフェクト一覧を取得
   */
  getEffects(): Effect[] {
    return [...this._effects];
  }

  /**
   * 権限レベルを取得
   */
  getPermissionLevel(): number {
    return this._permissionLevel;
  }

  /**
   * タグ一覧を取得
   */
  getTags(): string[] {
    return [...this._tags];
  }

  /**
   * エンティティ一覧を取得
   */
  getEntities(): Entity[] {
    return Array.from(this._entities.values());
  }

  /**
   * スコアを取得
   */
  getScore(objective: string): number | null {
    return this._scores.get(objective) ?? null;
  }

  /**
   * スコアボードオブジェクティブ一覧を取得
   */
  getScoreboardObjectives(): string[] {
    return [...this._scoreboardObjectives];
  }

  /**
   * アサーションコンテキスト
   */
  get expect(): AssertionContext {
    return this._expect;
  }

  /**
   * タスクランナー
   */
  get tasks(): TaskRunner {
    return this._tasks;
  }

  /**
   * ワールドデータ
   */
  get world(): World {
    return this._world;
  }

  async connect(): Promise<void> {
    if (this._isConnected) {
      throw new Error('Already connected');
    }

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('Connection timeout'));
      }, this.options.timeout);

      this.client = bedrock.createClient({
        host: this.options.host,
        port: this.options.port,
        username: this.options.username,
        offline: this.options.offline,
        version: this.options.version as '1.21.50',
      });

      this.client.on('error', (err: Error) => {
        clearTimeout(timeout);
        this.emit('error', err);
        reject(err);
      });

      this.client.on('close', () => {
        this._isConnected = false;
        this.emit('disconnect', 'Connection closed');
      });

      this.client.on('kick', (packet: { message: string }) => {
        this._isConnected = false;
        this.emit('disconnect', packet.message || 'Kicked');
      });

      this.client.on('join', () => {
        this._isConnected = true;
        this.emit('join');
      });

      this.client.on('spawn', () => {
        clearTimeout(timeout);
        this.emit('spawn');
        resolve();
      });

      this.setupPacketHandlers();
      this.setupAgentListeners();
    });
  }

  async disconnect(): Promise<void> {
    if (!this.client) return;

    this.stopAgent();
    this.client.close();
    this.client = null;
    this._isConnected = false;
    this._state = createInitialState();
  }

  private setupPacketHandlers(): void {
    if (!this.client) return;

    // Chat messages
    this.client.on('text', (packet: {
      type: string;
      needs_translation: boolean;
      source_name: string;
      message: string;
      xuid: string;
    }) => {
      const chatMessage: ChatMessage = {
        type: this.mapTextType(packet.type),
        sender: packet.source_name,
        message: packet.message,
        timestamp: Date.now(),
        xuid: packet.xuid,
      };
      this.emit('chat', chatMessage);
    });

    // Player position
    this.client.on('move_player', (packet: {
      runtime_id: bigint;
      position: { x: number; y: number; z: number };
      pitch: number;
      yaw: number;
      on_ground: boolean;
    }) => {
      if (packet.runtime_id === this._state.runtimeEntityId) {
        this._state.position = {
          x: packet.position.x,
          y: packet.position.y,
          z: packet.position.z,
        };
        this._state.rotation = {
          pitch: packet.pitch,
          yaw: packet.yaw,
        };
        this._state.isOnGround = packet.on_ground;
        this.emit('position_update', this._state.position);
      }
    });

    // Start game (initial state)
    this.client.on('start_game', (packet: {
      runtime_entity_id: bigint;
      player_position: { x: number; y: number; z: number };
      player_gamemode: number;
    }) => {
      this._state.runtimeEntityId = packet.runtime_entity_id;
      this._state.position = {
        x: packet.player_position.x,
        y: packet.player_position.y,
        z: packet.player_position.z,
      };
      this._state.gamemode = packet.player_gamemode;
    });

    // Health update
    this.client.on('update_attributes', (packet: {
      runtime_entity_id: bigint;
      attributes: Array<{ name: string; current: number }>;
    }) => {
      if (packet.runtime_entity_id === this._state.runtimeEntityId) {
        const health = packet.attributes.find(
          (attr) => attr.name === 'minecraft:health'
        );
        if (health) {
          this._state.health = health.current;
          this.emit('health_update', health.current);
        }
      }
    });

    // Gamemode update
    this.client.on('set_player_game_type', (packet: { gamemode: number }) => {
      this._state.gamemode = packet.gamemode;
      this.emit('gamemode_update', packet.gamemode);
    });

    // Form handling
    this.client.on('modal_form_request', (packet: {
      form_id: number;
      data: string;
    }) => {
      try {
        const formData = JSON.parse(packet.data);
        const form = this.parseForm(packet.form_id, formData);
        this.pendingForms.set(packet.form_id, form);
        this.emit('form', form);
      } catch {
        // Ignore invalid form data
      }
    });

    // Command output
    this.client.on('command_output', (packet: {
      origin: { uuid: string };
      output_type: number;
      success_count: number;
      output: Array<{ message_id: string; parameters: string[] }>;
    }) => {
      const callback = this.commandCallbacks.get(packet.origin.uuid);
      if (callback) {
        this.commandCallbacks.delete(packet.origin.uuid);
        const output: CommandOutput = {
          command: '',
          success: packet.success_count > 0,
          output: packet.output
            .map((o) => o.parameters.join(' ') || o.message_id)
            .join('\n'),
          statusCode: packet.output_type,
        };
        callback.resolve(output);
      }
    });

    // Level chunk (ワールドデータ)
    this.client.on('level_chunk', (packet: {
      x: number;
      z: number;
      sub_chunk_count: number;
      cache_enabled: boolean;
      blobs?: Array<{ hash: bigint; payload: Buffer }>;
      payload: Buffer;
    }) => {
      // チャンクをデコードしてワールドに保存
      this._world
        .decodeChunk(packet.x, packet.z, packet.sub_chunk_count, packet.payload)
        .then(() => {
          this.emit('chunk_loaded', { x: packet.x, z: packet.z });
        })
        .catch((err) => {
          // デコード失敗は無視（1.18+のサブチャンク方式の場合など）
          console.debug(`Chunk decode failed at ${packet.x},${packet.z}:`, err.message);
        });
    });

    // Block update (ブロック変更)
    this.client.on('update_block', (packet: {
      position: { x: number; y: number; z: number };
      block_runtime_id: number;
      flags: number;
      layer: number;
    }) => {
      // ブロック更新イベントを発火
      this.emit('block_update', {
        position: packet.position,
        runtimeId: packet.block_runtime_id,
      });
    });

    // Inventory update
    this.client.on('inventory_content', (packet: {
      window_id: number;
      input: Array<{
        network_id: number;
        count?: number;
        metadata?: number;
        block_runtime_id?: number;
        extra?: { nbt?: { value?: { Name?: { value: string }; Damage?: { value: number }; ench?: { value: Array<{ id: { value: number }; lvl: { value: number } }> } } } };
      }>;
    }) => {
      // プレイヤーインベントリ (window_id: 0)
      if (packet.window_id === 0) {
        this._inventory = packet.input
          .map((item, slot) => {
            if (!item || item.network_id === 0) return null;
            const inventoryItem: InventoryItem = {
              id: `minecraft:item_${item.network_id}`,
              count: item.count ?? 1,
              slot,
              damage: item.metadata,
            };
            // NBTからエンチャント情報を取得
            if (item.extra?.nbt?.value?.ench) {
              inventoryItem.enchantments = item.extra.nbt.value.ench.value.map((e) => ({
                id: `minecraft:enchantment_${e.id.value}`,
                level: e.lvl.value,
              }));
            }
            return inventoryItem;
          })
          .filter((item): item is InventoryItem => item !== null);
      }
    });

    // Inventory slot update
    this.client.on('inventory_slot', (packet: {
      window_id: number;
      slot: number;
      item: {
        network_id: number;
        count?: number;
        metadata?: number;
      };
    }) => {
      if (packet.window_id === 0) {
        const existingIndex = this._inventory.findIndex((i) => i.slot === packet.slot);
        if (packet.item.network_id === 0) {
          if (existingIndex >= 0) {
            this._inventory.splice(existingIndex, 1);
          }
        } else {
          const newItem: InventoryItem = {
            id: `minecraft:item_${packet.item.network_id}`,
            count: packet.item.count ?? 1,
            slot: packet.slot,
            damage: packet.item.metadata,
          };
          if (existingIndex >= 0) {
            this._inventory[existingIndex] = newItem;
          } else {
            this._inventory.push(newItem);
          }
          this.emit('inventory_update', { slot: packet.slot, item: newItem });
        }
      }
    });

    // Hunger update (update_attributes から)
    this.client.on('update_attributes', (packet: {
      runtime_entity_id: bigint;
      attributes: Array<{ name: string; current: number }>;
    }) => {
      if (packet.runtime_entity_id === this._state.runtimeEntityId) {
        const hunger = packet.attributes.find(
          (attr) => attr.name === 'minecraft:player.hunger'
        );
        if (hunger) {
          this._hunger = hunger.current;
          this.emit('hunger_update', hunger.current);
        }
      }
    });

    // Effect add
    this.client.on('mob_effect', (packet: {
      runtime_entity_id: bigint;
      event_id: number;
      effect_id: number;
      amplifier: number;
      particles: boolean;
      duration: number;
    }) => {
      if (packet.runtime_entity_id === this._state.runtimeEntityId) {
        const effect: Effect = {
          id: `minecraft:effect_${packet.effect_id}`,
          amplifier: packet.amplifier,
          duration: packet.duration,
          visible: packet.particles,
        };
        if (packet.event_id === 1) {
          // Add effect
          this._effects = this._effects.filter((e) => e.id !== effect.id);
          this._effects.push(effect);
          this.emit('effect_add', effect);
        } else if (packet.event_id === 2) {
          // Remove effect
          this._effects = this._effects.filter((e) => e.id !== effect.id);
          this.emit('effect_remove', effect.id);
        }
      }
    });

    // Entity spawn
    this.client.on('add_entity', (packet: {
      runtime_id: bigint;
      entity_type: string;
      position: { x: number; y: number; z: number };
      metadata?: Array<{ key: string; value: unknown }>;
    }) => {
      const entity: Entity = {
        runtimeId: packet.runtime_id,
        type: packet.entity_type,
        position: packet.position,
      };
      // name_tag メタデータを取得
      const nameTag = packet.metadata?.find((m) => m.key === 'name_tag');
      if (nameTag && typeof nameTag.value === 'string') {
        entity.nameTag = nameTag.value;
      }
      this._entities.set(packet.runtime_id, entity);
      this.emit('entity_spawn', entity);
    });

    // Entity remove
    this.client.on('remove_entity', (packet: { entity_id_self: bigint }) => {
      this._entities.delete(packet.entity_id_self);
      this.emit('entity_remove', packet.entity_id_self);
    });

    // Scoreboard objective
    this.client.on('set_display_objective', (packet: {
      display_slot: string;
      objective_name: string;
      display_name: string;
      criteria_name: string;
      sort_order: number;
    }) => {
      if (!this._scoreboardObjectives.includes(packet.objective_name)) {
        this._scoreboardObjectives.push(packet.objective_name);
      }
    });

    // Scoreboard score
    this.client.on('set_score', (packet: {
      action: number;
      entries: Array<{
        scoreboard_id: bigint;
        objective_name: string;
        score: number;
        entry_type?: number;
        entity_unique_id?: bigint;
        custom_name?: string;
      }>;
    }) => {
      for (const entry of packet.entries) {
        if (packet.action === 0) {
          // Set score
          this._scores.set(entry.objective_name, entry.score);
          this.emit('score_update', {
            objective: entry.objective_name,
            score: entry.score,
            displayName: entry.custom_name,
          });
        } else if (packet.action === 1) {
          // Remove score
          this._scores.delete(entry.objective_name);
        }
      }
    });

    // Title display
    this.client.on('set_title', (packet: {
      type: number;
      text: string;
      fade_in_time: number;
      stay_time: number;
      fade_out_time: number;
    }) => {
      const typeMap: Record<number, TitleDisplay['type']> = {
        0: 'title',
        1: 'subtitle',
        2: 'actionbar',
      };
      const displayType = typeMap[packet.type];
      if (displayType) {
        const display: TitleDisplay = {
          type: displayType,
          text: packet.text,
          fadeIn: packet.fade_in_time,
          stay: packet.stay_time,
          fadeOut: packet.fade_out_time,
        };
        this.emit('title', display);
      }
    });

    // Sound play
    this.client.on('play_sound', (packet: {
      sound_name: string;
      position: { x: number; y: number; z: number };
      volume: number;
      pitch: number;
    }) => {
      const sound: SoundPlay = {
        name: packet.sound_name,
        position: packet.position,
        volume: packet.volume,
        pitch: packet.pitch,
      };
      this.emit('sound', sound);
    });

    // Level sound event (別のサウンドパケット)
    this.client.on('level_sound_event', (packet: {
      sound_id: number;
      position: { x: number; y: number; z: number };
      extra_data: number;
      entity_type: string;
      is_baby_mob: boolean;
      is_global: boolean;
    }) => {
      const sound: SoundPlay = {
        name: `sound_${packet.sound_id}`,
        position: packet.position,
        volume: 1,
        pitch: 1,
      };
      this.emit('sound', sound);
    });

    // Particle spawn
    this.client.on('spawn_particle_effect', (packet: {
      dimension_id: number;
      entity_id: bigint;
      position: { x: number; y: number; z: number };
      particle_name: string;
      molang_variables_json?: string;
    }) => {
      const particle: ParticleSpawn = {
        name: packet.particle_name,
        position: packet.position,
      };
      this.emit('particle', particle);
    });

    // Dimension change
    this.client.on('change_dimension', (packet: {
      dimension: number;
      position: { x: number; y: number; z: number };
      respawn: boolean;
    }) => {
      const dimensionMap: Record<number, string> = {
        0: 'overworld',
        1: 'nether',
        2: 'the_end',
      };
      const dimension = dimensionMap[packet.dimension] ?? 'overworld';
      this._state.dimension = dimension;
      this.emit('dimension_change', dimension);
    });

    // Death
    this.client.on('death_info', () => {
      this.emit('death');
    });

    // Respawn
    this.client.on('respawn', (packet: {
      position: { x: number; y: number; z: number };
      state: number;
      runtime_entity_id: bigint;
    }) => {
      this._state.position = packet.position;
      this.emit('respawn', packet.position);
    });

    // Player permission level
    this.client.on('adventure_settings', (packet: {
      permission_level: number;
    }) => {
      this._permissionLevel = packet.permission_level;
    });

    // All packets (for debugging)
    this.client.on('packet', (packet: { name: string; params: unknown }) => {
      this.emit('packet', packet.name, packet.params);
    });
  }

  private setupAgentListeners(): void {
    this.on('chat', async (message) => {
      if (!this._isAgentRunning) return;
      if (message.sender === this.username) return;

      const reply = (msg: string) => this.chat(msg);

      // Check for commands
      if (message.message.startsWith(this.commandPrefix)) {
        const parts = message.message.slice(this.commandPrefix.length).split(' ');
        const cmd = parts[0].toLowerCase();
        const args = parts.slice(1);

        const handler = this.commandHandlers.get(cmd);
        if (handler) {
          try {
            await handler(args, reply);
          } catch (err) {
            console.error(`Command error: ${err}`);
          }
          return;
        }
      }

      // Check for chat patterns
      for (const { pattern, handler } of this.chatHandlers) {
        const matches =
          typeof pattern === 'string'
            ? message.message.includes(pattern)
            : pattern.test(message.message);

        if (matches) {
          try {
            await handler(message, reply);
          } catch (err) {
            console.error(`Chat handler error: ${err}`);
          }
        }
      }
    });
  }

  private mapTextType(type: string): ChatMessage['type'] {
    const typeMap: Record<string, ChatMessage['type']> = {
      chat: 'chat',
      system: 'system',
      whisper: 'whisper',
      announcement: 'announcement',
      raw: 'raw',
      tip: 'tip',
      json_whisper: 'whisper',
      json: 'json',
    };
    return typeMap[type] ?? 'raw';
  }

  private parseForm(
    id: number,
    data: {
      type: string;
      title: string;
      content?: string | Array<{ type: string; text?: string; [key: string]: unknown }>;
      button1?: string;
      button2?: string;
      buttons?: Array<{ text: string; image?: { type: string; data: string } }>;
    }
  ): Form {
    if (data.type === 'modal') {
      return {
        id,
        type: 'modal',
        title: data.title,
        content: data.content as string,
        button1: data.button1!,
        button2: data.button2!,
      };
    } else if (data.type === 'form') {
      return {
        id,
        type: 'action',
        title: data.title,
        content: typeof data.content === 'string' ? data.content : '',
        buttons: (data.buttons ?? []).map((btn) => ({
          text: btn.text,
          image: btn.image
            ? { type: btn.image.type as 'path' | 'url', data: btn.image.data }
            : undefined,
        })),
      };
    } else {
      return {
        id,
        type: 'form',
        title: data.title,
        content: (data.content as Array<{ type: string; text?: string; [key: string]: unknown }>) ?? [],
      } as Form;
    }
  }

  // === 基本アクション ===

  chat(message: string): void {
    if (!this.client) {
      throw new Error('Not connected');
    }

    this.client.queue('text', {
      type: 'chat',
      needs_translation: false,
      source_name: this.username,
      xuid: '',
      platform_chat_id: '',
      filtered_message: '',
      message,
    });
  }

  async command(cmd: string): Promise<CommandOutput> {
    if (!this.client) {
      throw new Error('Not connected');
    }

    const command = cmd.startsWith('/') ? cmd.slice(1) : cmd;
    const uuid = crypto.randomUUID();

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        this.commandCallbacks.delete(uuid);
        reject(new Error(`Command timeout: ${cmd}`));
      }, this.options.timeout);

      this.commandCallbacks.set(uuid, {
        resolve: (output) => {
          clearTimeout(timeout);
          resolve({ ...output, command: cmd });
        },
        reject: (err) => {
          clearTimeout(timeout);
          reject(err);
        },
      });

      this.client!.queue('command_request', {
        command,
        origin: {
          type: 0,
          uuid,
          request_id: uuid,
        },
        internal: false,
        version: 52,
      });
    });
  }

  getPendingForm(id?: number): Form | undefined {
    if (id !== undefined) {
      return this.pendingForms.get(id);
    }
    // Return first pending form
    for (const form of this.pendingForms.values()) {
      return form;
    }
    return undefined;
  }

  respondToForm(formId: number, response: FormResponse): void {
    if (!this.client) {
      throw new Error('Not connected');
    }

    this.pendingForms.delete(formId);

    this.client.queue('modal_form_response', {
      form_id: formId,
      data: response === null ? '' : JSON.stringify(response),
      cancel_reason: response === null ? 0 : undefined,
    });
  }

  closeForm(formId: number): void {
    this.respondToForm(formId, null);
  }

  sendPacket(name: string, payload: Record<string, unknown>): void {
    if (!this.client) {
      throw new Error('Not connected');
    }
    this.client.queue(name, payload);
  }

  onPacket(name: string, handler: (data: unknown) => void): void {
    if (!this.client) {
      throw new Error('Not connected');
    }
    this.client.on(name, handler);
  }

  // === Agent機能: チャット/コマンドハンドラー ===

  /**
   * チャットパターンハンドラーを登録
   */
  onChat(pattern: string | RegExp, handler: ChatHandler): this {
    this.chatHandlers.push({ pattern, handler });
    return this;
  }

  /**
   * コマンドハンドラーを登録
   */
  onCommand(command: string, handler: CommandHandler): this {
    this.commandHandlers.set(command.toLowerCase(), handler);
    return this;
  }

  /**
   * エージェントを開始（チャット/コマンドハンドリングを有効化）
   */
  startAgent(): this {
    this._isAgentRunning = true;
    return this;
  }

  /**
   * エージェントを停止
   */
  stopAgent(): this {
    this._isAgentRunning = false;
    this._tasks.cancel();
    return this;
  }

  /**
   * エージェントが動作中かどうか
   */
  get isAgentRunning(): boolean {
    return this._isAgentRunning;
  }

  // === 高レベルアクション ===

  /**
   * 指定座標に移動（テレポート）
   */
  async goto(position: Position): Promise<void> {
    await this.command(`/tp @s ${position.x} ${position.y} ${position.z}`);
  }

  /**
   * 指定座標を見る
   */
  async lookAt(position: Position): Promise<void> {
    const current = this.position;
    const dx = position.x - current.x;
    const dy = position.y - current.y;
    const dz = position.z - current.z;

    const yaw = -Math.atan2(dx, dz) * (180 / Math.PI);
    const distance = Math.sqrt(dx * dx + dz * dz);
    const pitch = -Math.atan2(dy, distance) * (180 / Math.PI);

    this.sendPacket('move_player', {
      runtime_id: this.state.runtimeEntityId,
      position: {
        x: current.x,
        y: current.y,
        z: current.z,
      },
      pitch,
      yaw,
      head_yaw: yaw,
      mode: 0,
      on_ground: this.state.isOnGround,
      riding_eid: 0n,
      teleport_cause: 0,
      teleport_source_entity_type: 0,
      tick: 0n,
    });
  }

  /**
   * ブロックを破壊（実プレイヤー操作）
   * @param position 破壊するブロックの座標
   * @param options オプション
   */
  async breakBlock(
    position: Position,
    options: {
      /** ツール倍率（デフォルト: 1.0 = 素手） */
      toolMultiplier?: number;
      /** 破壊をキャンセルするためのAbortSignal */
      signal?: AbortSignal;
      /** 進捗コールバック（0.0 - 1.0） */
      onProgress?: (progress: number) => void;
    } = {}
  ): Promise<{ success: boolean; reason?: string }> {
    if (!this.client) {
      throw new Error('Not connected');
    }

    const { toolMultiplier = 1.0, signal, onProgress } = options;

    const blockX = Math.floor(position.x);
    const blockY = Math.floor(position.y);
    const blockZ = Math.floor(position.z);

    // ブロックの破壊データを取得
    const breakData = this._world.getBlockBreakData(
      blockX,
      blockY,
      blockZ,
      toolMultiplier
    );

    // 破壊不可能なブロック
    if (breakData.unbreakable) {
      return { success: false, reason: 'Block is unbreakable' };
    }

    const blockPosition = { x: blockX, y: blockY, z: blockZ };

    // プレイヤーの向きを計算（ブロックを向く）
    await this.lookAt(position);

    // StartBreak パケットを送信
    this.sendPacket('player_action', {
      action: 'start_break',
      position: blockPosition,
      result_position: blockPosition,
      face: this.calculateBlockFace(position),
    });

    this.emit('block_break_start', { position: blockPosition });

    // 即座に破壊可能な場合
    if (breakData.instant) {
      // StopBreak パケットを送信
      this.sendPacket('player_action', {
        action: 'stop_break',
        position: blockPosition,
        result_position: blockPosition,
        face: 0,
      });

      onProgress?.(1.0);
      this.emit('block_break_complete', { position: blockPosition });
      return { success: true };
    }

    // 破壊時間（ミリ秒）
    const breakTimeMs = breakData.baseTime * 1000;
    const startTime = Date.now();
    const tickInterval = 50; // 50ms = 1 tick

    // 破壊進行ループ
    return new Promise((resolve) => {
      const progressInterval = setInterval(() => {
        // キャンセルチェック
        if (signal?.aborted) {
          clearInterval(progressInterval);
          // AbortBreak パケットを送信
          this.sendPacket('player_action', {
            action: 'abort_break',
            position: blockPosition,
            result_position: blockPosition,
            face: 0,
          });
          this.emit('block_break_abort', { position: blockPosition });
          resolve({ success: false, reason: 'Aborted' });
          return;
        }

        const elapsed = Date.now() - startTime;
        const progress = Math.min(elapsed / breakTimeMs, 1.0);

        // 進捗コールバック
        onProgress?.(progress);

        // 破壊進行パケットを送信（クラック表示用）
        this.sendPacket('level_event', {
          event: 3600, // block_start_break
          position: {
            x: blockX + 0.5,
            y: blockY + 0.5,
            z: blockZ + 0.5,
          },
          data: Math.floor(65535 / (breakTimeMs / tickInterval)),
        });

        // 破壊完了
        if (progress >= 1.0) {
          clearInterval(progressInterval);

          // StopBreak パケットを送信
          this.sendPacket('player_action', {
            action: 'stop_break',
            position: blockPosition,
            result_position: blockPosition,
            face: 0,
          });

          // 破壊完了イベント
          this.sendPacket('level_event', {
            event: 2001, // particle_destroy_block
            position: {
              x: blockX + 0.5,
              y: blockY + 0.5,
              z: blockZ + 0.5,
            },
            data: 0,
          });

          this.emit('block_break_complete', { position: blockPosition });
          resolve({ success: true });
        }
      }, tickInterval);
    });
  }

  /**
   * プレイヤーからブロックへの面を計算
   */
  private calculateBlockFace(blockPosition: Position): number {
    const playerPos = this.position;
    const dx = blockPosition.x - playerPos.x;
    const dy = blockPosition.y - playerPos.y;
    const dz = blockPosition.z - playerPos.z;

    const absDx = Math.abs(dx);
    const absDy = Math.abs(dy);
    const absDz = Math.abs(dz);

    // 最も大きい差分の方向を面とする
    if (absDy >= absDx && absDy >= absDz) {
      return dy > 0 ? 0 : 1; // 0: bottom, 1: top
    } else if (absDz >= absDx) {
      return dz > 0 ? 2 : 3; // 2: north, 3: south
    } else {
      return dx > 0 ? 4 : 5; // 4: west, 5: east
    }
  }

  /**
   * ブロックを破壊（コマンド版 - クリエイティブモード用）
   */
  async breakBlockInstant(position: Position): Promise<void> {
    await this.command(
      `/setblock ${Math.floor(position.x)} ${Math.floor(position.y)} ${Math.floor(position.z)} air destroy`
    );
  }

  /**
   * ブロックを設置
   */
  async placeBlock(position: Position, blockId: string): Promise<void> {
    const block = blockId.startsWith('minecraft:') ? blockId : `minecraft:${blockId}`;
    await this.command(
      `/setblock ${Math.floor(position.x)} ${Math.floor(position.y)} ${Math.floor(position.z)} ${block}`
    );
  }

  /**
   * チャットメッセージを送信
   */
  say(message: string): void {
    this.chat(message);
  }

  // === ワールド情報 ===

  /**
   * 指定座標のブロックを取得
   */
  getBlock(x: number, y: number, z: number): Block | null {
    return this._world.getBlock(x, y, z);
  }

  /**
   * 指定座標のブロック名を取得
   */
  getBlockName(x: number, y: number, z: number): string | null {
    return this._world.getBlockName(x, y, z);
  }

  /**
   * 指定座標が通行可能か
   */
  isPassable(x: number, y: number, z: number): boolean {
    return this._world.isPassable(x, y, z);
  }

  /**
   * 指定座標が固体ブロックか
   */
  isSolid(x: number, y: number, z: number): boolean {
    return this._world.isSolid(x, y, z);
  }

  /**
   * 足元のブロックを取得
   */
  getBlockBelow(): Block | null {
    return this._world.getBlockBelow(this.position);
  }

  /**
   * 前方のブロックを取得
   */
  getBlockInFront(distance: number = 1): Block | null {
    return this._world.getBlockInFront(this.position, this.state.rotation.yaw, distance);
  }

  /**
   * 周囲のブロック情報を取得
   */
  getBlocksAround(radius: number = 1): Map<string, Block | null> {
    return this._world.getBlocksAround(this.position, radius);
  }
}

/**
 * TaskRunner: タスクキュー管理
 */
export class TaskRunner {
  private queue: Task[] = [];
  private _current: Task | null = null;
  private _isCancelled = false;
  private _isRunning = false;

  constructor(private agent: Agent) {}

  add(task: Task): this;
  add(name: string, fn: (agent: Agent) => Promise<void>): this;
  add(
    taskOrName: Task | string,
    fn?: (agent: Agent) => Promise<void>
  ): this {
    if (typeof taskOrName === 'string') {
      this.queue.push({
        name: taskOrName,
        execute: fn!,
      });
    } else {
      this.queue.push(taskOrName);
    }
    return this;
  }

  async runAll(): Promise<void> {
    if (this._isRunning) {
      throw new Error('Task runner is already running');
    }

    this._isRunning = true;
    this._isCancelled = false;

    try {
      while (this.queue.length > 0 && !this._isCancelled) {
        this._current = this.queue.shift()!;

        try {
          await this._current.execute(this.agent);
        } catch (err) {
          console.error(`Task "${this._current.name}" failed:`, err);
          throw err;
        }
      }
    } finally {
      this._current = null;
      this._isRunning = false;
    }
  }

  async runNext(): Promise<boolean> {
    if (this.queue.length === 0) {
      return false;
    }

    this._current = this.queue.shift()!;

    try {
      await this._current.execute(this.agent);
      return true;
    } catch (err) {
      console.error(`Task "${this._current.name}" failed:`, err);
      throw err;
    } finally {
      this._current = null;
    }
  }

  cancel(): void {
    this._isCancelled = true;
    if (this._current?.cancel) {
      this._current.cancel();
    }
  }

  clear(): void {
    this.queue = [];
  }

  get current(): Task | null {
    return this._current;
  }

  get remaining(): number {
    return this.queue.length;
  }

  get isRunning(): boolean {
    return this._isRunning;
  }
}

// Built-in task factories
export const tasks = {
  wait: (ms: number): Task => ({
    name: `wait ${ms}ms`,
    execute: () => new Promise((resolve) => setTimeout(resolve, ms)),
  }),

  sequence: (...taskList: Task[]): Task => ({
    name: 'sequence',
    execute: async (agent) => {
      const runner = new TaskRunner(agent);
      for (const task of taskList) {
        runner.add(task);
      }
      await runner.runAll();
    },
  }),

  repeat: (task: Task, times: number): Task => ({
    name: `repeat ${task.name} ${times}x`,
    execute: async (agent) => {
      for (let i = 0; i < times; i++) {
        await task.execute(agent);
      }
    },
  }),

  repeatUntil: (
    task: Task,
    condition: () => boolean | Promise<boolean>
  ): Task => ({
    name: `repeatUntil ${task.name}`,
    execute: async (agent) => {
      while (!(await condition())) {
        await task.execute(agent);
      }
    },
  }),
};

/**
 * Agentを作成
 */
export function createAgent(options: ClientOptions, agentOptions?: AgentOptions): Agent {
  return new Agent(options, agentOptions);
}
