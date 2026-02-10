import type { LLMProvider, LLMProviderConfig, ParsedAction } from '../types/llm.js';
import type { ScenarioStep } from '../types/scenario.js';
import { AnthropicProvider, OpenAIProvider, MockLLMProvider } from './providers/index.js';
import {
  SCENARIO_SYSTEM_PROMPT,
  RESULT_SUMMARY_SYSTEM_PROMPT,
  createStepParsePrompt,
  createPlayerExtractionPrompt,
  createResultSummaryPrompt,
} from './prompts.js';

/**
 * LLMプロセッサ
 */
export class LLMProcessor {
  constructor(private provider: LLMProvider) {}

  /**
   * ステップをアクションに変換
   */
  async parseStep(step: ScenarioStep): Promise<ParsedAction[]> {
    const prompt = createStepParsePrompt(
      step.description,
      step.type,
      step.players ?? []
    );

    const response = await this.provider.chat([
      { role: 'system', content: SCENARIO_SYSTEM_PROMPT },
      { role: 'user', content: prompt },
    ]);

    try {
      // JSONを抽出
      const jsonMatch = response.match(/\[[\s\S]*\]/);
      if (jsonMatch) {
        return JSON.parse(jsonMatch[0]);
      }
      return [];
    } catch {
      console.warn('[LLMProcessor] Failed to parse response as JSON');
      return [];
    }
  }

  /**
   * テキストからプレイヤー名を抽出
   */
  async extractPlayers(text: string): Promise<string[]> {
    const prompt = createPlayerExtractionPrompt(text);
    const response = await this.provider.complete(prompt);

    try {
      const jsonMatch = response.match(/\[[\s\S]*\]/);
      if (jsonMatch) {
        return JSON.parse(jsonMatch[0]);
      }
      return [];
    } catch {
      return [];
    }
  }

  /**
   * シナリオ結果を自然言語でサマリー
   */
  async summarizeResult(
    scenarioName: string,
    steps: Array<{
      description: string;
      status: 'passed' | 'failed' | 'skipped';
      duration: number;
      error?: Error;
    }>,
    passed: boolean,
    duration: number
  ): Promise<string> {
    const stepsForPrompt = steps.map((s) => ({
      description: s.description,
      status: s.status,
      duration: s.duration,
      error: s.error ? { message: s.error.message, stack: s.error.stack } : undefined,
    }));

    const prompt = createResultSummaryPrompt(
      scenarioName,
      stepsForPrompt,
      passed,
      duration
    );

    const response = await this.provider.chat([
      { role: 'system', content: RESULT_SUMMARY_SYSTEM_PROMPT },
      { role: 'user', content: prompt },
    ]);

    return response.trim();
  }

  /**
   * プロバイダーを取得
   */
  getProvider(): LLMProvider {
    return this.provider;
  }
}

/**
 * 設定からLLMプロバイダーを作成
 */
export function createLLMProvider(config: LLMProviderConfig): LLMProvider {
  switch (config.provider) {
    case 'anthropic':
      if (!config.apiKey) {
        throw new Error('Anthropic API key is required');
      }
      return new AnthropicProvider({
        apiKey: config.apiKey,
        model: config.model,
        baseURL: config.baseURL,
      });

    case 'openai':
      if (!config.apiKey) {
        throw new Error('OpenAI API key is required');
      }
      return new OpenAIProvider({
        apiKey: config.apiKey,
        model: config.model,
        baseURL: config.baseURL,
      });

    default:
      throw new Error(`Unknown LLM provider: ${config.provider}`);
  }
}

/**
 * モックプロバイダーを作成（テスト用）
 */
export function createMockProvider(): MockLLMProvider {
  return new MockLLMProvider();
}

// エクスポート
export { AnthropicProvider, OpenAIProvider, MockLLMProvider } from './providers/index.js';
export {
  SCENARIO_SYSTEM_PROMPT,
  RESULT_SUMMARY_SYSTEM_PROMPT,
  createStepParsePrompt,
  createScenarioParsePrompt,
  createPlayerExtractionPrompt,
  createResultSummaryPrompt,
} from './prompts.js';
