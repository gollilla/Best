import { BaseLLMProvider } from './index.js';
import type { LLMOptions, LLMMessage } from '../../types/llm.js';

export interface AnthropicProviderOptions {
  /** APIキー */
  apiKey: string;
  /** モデル名 */
  model?: string;
  /** ベースURL */
  baseURL?: string;
}

/**
 * Anthropic (Claude) プロバイダー
 */
export class AnthropicProvider extends BaseLLMProvider {
  name = 'anthropic';
  private apiKey: string;
  private model: string;
  private baseURL: string;

  constructor(options: AnthropicProviderOptions) {
    super();
    this.apiKey = options.apiKey;
    this.model = options.model ?? 'claude-sonnet-4-20250514';
    this.baseURL = options.baseURL ?? 'https://api.anthropic.com';
  }

  async complete(prompt: string, options?: LLMOptions): Promise<string> {
    return this.chat([{ role: 'user', content: prompt }], options);
  }

  async chat(messages: LLMMessage[], options?: LLMOptions): Promise<string> {
    const systemMessage = messages.find((m) => m.role === 'system');
    const otherMessages = messages.filter((m) => m.role !== 'system');

    const requestBody: Record<string, unknown> = {
      model: this.model,
      max_tokens: options?.maxTokens ?? 4096,
      messages: otherMessages.map((m) => ({
        role: m.role === 'assistant' ? 'assistant' : 'user',
        content: m.content,
      })),
    };

    if (systemMessage) {
      requestBody.system = systemMessage.content;
    }

    if (options?.temperature !== undefined) {
      requestBody.temperature = options.temperature;
    }

    if (options?.tools && options.tools.length > 0) {
      requestBody.tools = options.tools.map((t) => ({
        name: t.name,
        description: t.description,
        input_schema: t.parameters,
      }));
    }

    const response = await fetch(`${this.baseURL}/v1/messages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-api-key': this.apiKey,
        'anthropic-version': '2023-06-01',
      },
      body: JSON.stringify(requestBody),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Anthropic API error: ${response.status} - ${error}`);
    }

    const data = (await response.json()) as {
      content: Array<{ type: string; text?: string }>;
    };

    // テキストコンテンツを抽出
    const textContent = data.content.find((c) => c.type === 'text');
    return textContent?.text ?? '';
  }
}
