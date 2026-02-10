import { Agent, AgentOptions } from '../core/client.js';
import type { ClientOptions } from '../types/index.js';
import { loadConfig } from '../config/index.js';

// 設定キャッシュ
let cachedConfig: ClientOptions | null = null;

/**
 * 設定ファイルからAgentを作成し、自動接続
 *
 * @param name プレイヤー名（省略時は自動生成）
 * @param options 追加オプション
 * @returns Agent インスタンス（接続済み）
 *
 * @example
 * ```typescript
 * // 名前を指定
 * const agent = await createAgent('TestBot');
 *
 * // 名前を自動生成
 * const agent = await createAgent();
 * ```
 */
export async function createAgent(
  name?: string,
  options?: AgentOptions
): Promise<Agent> {
  // 設定をキャッシュから取得またはロード
  if (!cachedConfig) {
    const config = await loadConfig();
    cachedConfig = {
      host: config.host,
      port: config.port ?? 19132,
      username: name ?? `Player_${Math.random().toString(36).slice(2, 8)}`,
      offline: config.offline ?? true,
      timeout: config.timeout ?? 30000,
    };
  }

  const clientOptions: ClientOptions = {
    ...cachedConfig,
    username: name ?? `Player_${Math.random().toString(36).slice(2, 8)}`,
  };

  const agent = new Agent(clientOptions, options);
  await agent.connect();
  return agent;
}

/**
 * 同期版のAgent作成（接続は別途行う）
 */
export function createAgentSync(
  name?: string,
  options?: AgentOptions,
  clientOptions?: Partial<ClientOptions>
): Agent {
  const defaultOptions: ClientOptions = {
    host: clientOptions?.host ?? 'localhost',
    port: clientOptions?.port ?? 19132,
    username: name ?? `Player_${Math.random().toString(36).slice(2, 8)}`,
    offline: clientOptions?.offline ?? true,
    timeout: clientOptions?.timeout ?? 30000,
  };

  return new Agent(defaultOptions, options);
}

// core/client.ts から re-export
export { Agent, TaskRunner, tasks } from '../core/client.js';
export type { AgentOptions, Task } from '../core/client.js';
