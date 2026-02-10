import { Agent, createAgentSync } from '../agent/index.js';
import type { ClientOptions } from '../types/index.js';

/**
 * ScenarioContext: シナリオ実行時にプレイヤーを動的に生成・管理
 *
 * シナリオMarkdownで「Alice」と書かれていたら、`getPlayer('Alice')` で自動生成
 */
export class ScenarioContext {
  private players: Map<string, Agent> = new Map();
  private clientOptions: Partial<ClientOptions>;

  constructor(clientOptions?: Partial<ClientOptions>) {
    this.clientOptions = clientOptions ?? {};
  }

  /**
   * プレイヤーを取得（なければ自動生成）
   *
   * @param name プレイヤー名
   * @returns Agent インスタンス
   */
  async getPlayer(name: string): Promise<Agent> {
    // 既存のプレイヤーがあれば返す
    const existing = this.players.get(name);
    if (existing) {
      return existing;
    }

    // 新しいプレイヤーを作成
    const agent = createAgentSync(name, undefined, this.clientOptions);
    this.players.set(name, agent);

    return agent;
  }

  /**
   * プレイヤーが存在するか確認
   */
  hasPlayer(name: string): boolean {
    return this.players.has(name);
  }

  /**
   * 全プレイヤーを取得
   */
  getAllPlayers(): Agent[] {
    return Array.from(this.players.values());
  }

  /**
   * プレイヤー名の一覧を取得
   */
  getPlayerNames(): string[] {
    return Array.from(this.players.keys());
  }

  /**
   * 全プレイヤーを接続
   */
  async connectAll(): Promise<void> {
    const promises = Array.from(this.players.values()).map((agent) =>
      agent.connect()
    );
    await Promise.all(promises);
  }

  /**
   * 指定したプレイヤーを接続
   */
  async connectPlayer(name: string): Promise<void> {
    const agent = await this.getPlayer(name);
    if (!agent.isConnected) {
      await agent.connect();
    }
  }

  /**
   * 全プレイヤーを切断
   */
  async disconnectAll(): Promise<void> {
    const promises = Array.from(this.players.values()).map((agent) =>
      agent.disconnect()
    );
    await Promise.all(promises);
  }

  /**
   * クリーンアップ（全プレイヤー切断 + マップクリア）
   */
  async cleanup(): Promise<void> {
    await this.disconnectAll();
    this.players.clear();
  }

  /**
   * 接続中のプレイヤー数
   */
  get connectedCount(): number {
    return Array.from(this.players.values()).filter((a) => a.isConnected).length;
  }

  /**
   * 総プレイヤー数
   */
  get playerCount(): number {
    return this.players.size;
  }
}
