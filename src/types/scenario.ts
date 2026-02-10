import type { Agent } from '../agent/index.js';

/**
 * シナリオのステップ
 */
export interface ScenarioStep {
  /** ステップの種類 */
  type: 'action' | 'assertion' | 'wait';
  /** ステップの説明 */
  description: string;
  /** 対象のプレイヤー名（複数可） */
  players?: string[];
  /** ステップを実行する関数 */
  execute?: (agents: Map<string, Agent>) => Promise<void>;
  /** 元のMarkdownテキスト */
  raw?: string;
}

/**
 * シナリオの実行結果
 */
export interface ScenarioResult {
  /** シナリオ名 */
  name: string;
  /** 各ステップの結果 */
  steps: ScenarioStepResult[];
  /** 成功したかどうか */
  passed: boolean;
  /** 実行時間（ミリ秒） */
  duration: number;
  /** エラー（失敗時） */
  error?: Error;
  /** 自然言語による結果サマリー */
  summary?: string;
}

/**
 * ステップの実行結果
 */
export interface ScenarioStepResult {
  /** ステップの説明 */
  description: string;
  /** ステータス */
  status: 'passed' | 'failed' | 'skipped';
  /** 実行時間（ミリ秒） */
  duration: number;
  /** エラー（失敗時） */
  error?: Error;
}

/**
 * パースされたシナリオ
 */
export interface ParsedScenario {
  /** シナリオ名（# タイトル） */
  name: string;
  /** プレイヤー定義（## プレイヤー） */
  players: PlayerDefinition[];
  /** 前提条件（## 前提条件） */
  preconditions: string[];
  /** ステップ（## ステップ） */
  steps: ScenarioStep[];
  /** 元のMarkdownテキスト */
  raw: string;
}

/**
 * プレイヤー定義
 */
export interface PlayerDefinition {
  /** プレイヤー名 */
  name: string;
  /** 説明・役割 */
  description?: string;
}

/**
 * シナリオ設定
 */
export interface ScenarioConfig {
  /** シナリオファイルのマッチパターン */
  match?: string[];
  /** LLM設定 */
  llm?: {
    /** プロバイダー */
    provider: 'anthropic' | 'openai';
    /** APIキー */
    apiKey?: string;
    /** モデル名 */
    model?: string;
  };
  /** ステップのタイムアウト（ミリ秒） */
  stepTimeout?: number;
  /** 全体のタイムアウト（ミリ秒） */
  totalTimeout?: number;
}
