import { BaseLLMProvider } from './index.js';
import type { LLMOptions, LLMMessage } from '../../types/llm.js';

export interface OpenAIProviderOptions {
  /** APIキー */
  apiKey: string;
  /** モデル名 */
  model?: string;
  /** ベースURL */
  baseURL?: string;
}

/**
 * OpenAI (GPT) プロバイダー
 */
export class OpenAIProvider extends BaseLLMProvider {
  name = 'openai';
  private apiKey: string;
  private model: string;
  private baseURL: string;

  constructor(options: OpenAIProviderOptions) {
    super();
    this.apiKey = options.apiKey;
    this.model = options.model ?? 'gpt-4o';
    this.baseURL = options.baseURL ?? 'https://api.openai.com/v1';
  }

  async complete(prompt: string, options?: LLMOptions): Promise<string> {
    return this.chat([{ role: 'user', content: prompt }], options);
  }

  async chat(messages: LLMMessage[], options?: LLMOptions): Promise<string> {
    const requestBody: Record<string, unknown> = {
      model: this.model,
      messages: messages.map((m) => ({
        role: m.role,
        content: m.content,
      })),
    };

    if (options?.temperature !== undefined) {
      requestBody.temperature = options.temperature;
    }

    if (options?.maxTokens) {
      requestBody.max_tokens = options.maxTokens;
    }

    if (options?.tools && options.tools.length > 0) {
      requestBody.tools = options.tools.map((t) => ({
        type: 'function',
        function: {
          name: t.name,
          description: t.description,
          parameters: t.parameters,
        },
      }));
    }

    const response = await fetch(`${this.baseURL}/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${this.apiKey}`,
      },
      body: JSON.stringify(requestBody),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`OpenAI API error: ${response.status} - ${error}`);
    }

    const data = (await response.json()) as {
      choices: Array<{ message: { content: string } }>;
    };

    return data.choices[0]?.message?.content ?? '';
  }
}
