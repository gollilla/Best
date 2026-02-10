import { readFileSync, existsSync } from 'fs';
import { resolve } from 'path';
import { glob } from '../utils/glob.js';
import { parseScenarioMarkdown, getAllPlayerNames } from './markdown.js';
import { ScenarioContext } from './context.js';
import { ScenarioExecutor } from './executor.js';
import type {
  ParsedScenario,
  ScenarioResult,
  ScenarioConfig,
} from '../types/scenario.js';
import type { LLMProvider } from '../types/llm.js';
import type { ClientOptions } from '../types/index.js';

export interface ScenarioRunnerOptions {
  /** LLMプロバイダー */
  llmProvider?: LLMProvider;
  /** クライアント接続オプション */
  clientOptions?: Partial<ClientOptions>;
  /** ステップのタイムアウト（ミリ秒） */
  stepTimeout?: number;
  /** 全体のタイムアウト（ミリ秒） */
  totalTimeout?: number;
  /** 詳細ログを出力するか */
  verbose?: boolean;
  /** 結果を自然言語でサマリーするか */
  generateSummary?: boolean;
}

/**
 * シナリオローダー・ランナー
 */
export class ScenarioRunner {
  private scenarios: ParsedScenario[] = [];
  private options: ScenarioRunnerOptions;

  constructor(options: ScenarioRunnerOptions = {}) {
    this.options = {
      stepTimeout: options.stepTimeout ?? 30000,
      totalTimeout: options.totalTimeout ?? 300000,
      verbose: options.verbose ?? false,
      ...options,
    };
  }

  /**
   * Markdownファイルからシナリオを読み込み
   */
  loadFile(filePath: string): this {
    const absolutePath = resolve(process.cwd(), filePath);

    if (!existsSync(absolutePath)) {
      throw new Error(`Scenario file not found: ${absolutePath}`);
    }

    const content = readFileSync(absolutePath, 'utf-8');
    const scenario = parseScenarioMarkdown(content);
    this.scenarios.push(scenario);

    return this;
  }

  /**
   * グロブパターンでシナリオファイルを読み込み
   */
  async loadGlob(pattern: string, cwd: string = process.cwd()): Promise<this> {
    const files = await glob(pattern, cwd);

    for (const file of files) {
      if (file.endsWith('.scenario.md') || file.endsWith('.md')) {
        this.loadFile(file);
      }
    }

    return this;
  }

  /**
   * Markdown文字列からシナリオを読み込み
   */
  loadMarkdown(markdown: string): this {
    const scenario = parseScenarioMarkdown(markdown);
    this.scenarios.push(scenario);
    return this;
  }

  /**
   * 読み込んだシナリオを取得
   */
  getScenarios(): ParsedScenario[] {
    return this.scenarios;
  }

  /**
   * 単一のシナリオを実行
   */
  async runScenario(scenario: ParsedScenario): Promise<ScenarioResult> {
    const context = new ScenarioContext(this.options.clientOptions);

    // プレイヤーを事前に作成
    const playerNames = getAllPlayerNames(scenario);
    for (const name of playerNames) {
      await context.getPlayer(name);
    }

    const executor = new ScenarioExecutor(context, {
      llmProvider: this.options.llmProvider,
      stepTimeout: this.options.stepTimeout,
      totalTimeout: this.options.totalTimeout,
      verbose: this.options.verbose,
      generateSummary: this.options.generateSummary,
    });

    try {
      const result = await executor.execute(scenario);
      return result;
    } finally {
      await context.cleanup();
    }
  }

  /**
   * 全てのシナリオを実行
   */
  async runAll(): Promise<ScenarioResult[]> {
    const results: ScenarioResult[] = [];

    for (const scenario of this.scenarios) {
      const result = await this.runScenario(scenario);
      results.push(result);
    }

    return results;
  }

  /**
   * シナリオをクリア
   */
  clear(): this {
    this.scenarios = [];
    return this;
  }
}

/**
 * シナリオを実行するヘルパー関数
 */
export async function runScenario(
  filePath: string,
  options?: ScenarioRunnerOptions
): Promise<ScenarioResult> {
  const runner = new ScenarioRunner(options);
  runner.loadFile(filePath);
  const results = await runner.runAll();
  return results[0];
}

/**
 * 複数シナリオを実行するヘルパー関数
 */
export async function runScenarios(
  pattern: string,
  options?: ScenarioRunnerOptions
): Promise<ScenarioResult[]> {
  const runner = new ScenarioRunner(options);
  await runner.loadGlob(pattern);
  return runner.runAll();
}

/**
 * マークダウン文字列からシナリオを実行するヘルパー関数
 */
export async function runScenarioFromMarkdown(
  markdown: string,
  options?: ScenarioRunnerOptions
): Promise<ScenarioResult> {
  const runner = new ScenarioRunner(options);
  runner.loadMarkdown(markdown);
  const results = await runner.runAll();
  return results[0];
}

// エクスポート
export { ScenarioContext } from './context.js';
export { ScenarioExecutor } from './executor.js';
export { parseScenarioMarkdown, getAllPlayerNames } from './markdown.js';
