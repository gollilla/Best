import type { LLMProvider, LLMOptions, LLMMessage } from '../../types/llm.js';

export { LLMProvider };

/**
 * LLMプロバイダーの基底クラス
 */
export abstract class BaseLLMProvider implements LLMProvider {
  abstract name: string;
  abstract complete(prompt: string, options?: LLMOptions): Promise<string>;
  abstract chat(messages: LLMMessage[], options?: LLMOptions): Promise<string>;
}

export { AnthropicProvider } from './anthropic.js';
export { OpenAIProvider } from './openai.js';
export { MockLLMProvider } from './mock.js';
