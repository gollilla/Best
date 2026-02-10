import { BaseLLMProvider } from './index.js';
import type { LLMOptions, LLMMessage, ParsedAction } from '../../types/llm.js';

/**
 * テスト用のモックLLMプロバイダー
 */
export class MockLLMProvider extends BaseLLMProvider {
  name = 'mock';
  private responses: Map<string, string> = new Map();
  private defaultResponse: string = '[]';

  constructor() {
    super();
  }

  /**
   * 特定のパターンに対するレスポンスを設定
   */
  setResponse(pattern: string, response: string): this {
    this.responses.set(pattern, response);
    return this;
  }

  /**
   * デフォルトレスポンスを設定
   */
  setDefaultResponse(response: string): this {
    this.defaultResponse = response;
    return this;
  }

  /**
   * アクションリストをデフォルトレスポンスとして設定
   */
  setDefaultActions(actions: ParsedAction[]): this {
    this.defaultResponse = JSON.stringify(actions);
    return this;
  }

  async complete(prompt: string, _options?: LLMOptions): Promise<string> {
    // パターンマッチング
    for (const [pattern, response] of this.responses) {
      if (prompt.includes(pattern)) {
        return response;
      }
    }
    return this.defaultResponse;
  }

  async chat(messages: LLMMessage[], options?: LLMOptions): Promise<string> {
    const lastMessage = messages[messages.length - 1];
    return this.complete(lastMessage.content, options);
  }

  /**
   * レスポンスをクリア
   */
  clear(): this {
    this.responses.clear();
    this.defaultResponse = '[]';
    return this;
  }
}
