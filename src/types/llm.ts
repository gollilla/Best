/**
 * LLMプロバイダーインターフェース
 */
export interface LLMProvider {
  /** プロバイダー名 */
  name: string;
  /** テキスト補完 */
  complete(prompt: string, options?: LLMOptions): Promise<string>;
  /** チャット形式 */
  chat(messages: LLMMessage[], options?: LLMOptions): Promise<string>;
}

/**
 * LLMオプション
 */
export interface LLMOptions {
  /** 温度パラメータ */
  temperature?: number;
  /** 最大トークン数 */
  maxTokens?: number;
  /** ツール定義 */
  tools?: LLMTool[];
}

/**
 * LLMメッセージ
 */
export interface LLMMessage {
  /** ロール */
  role: 'system' | 'user' | 'assistant' | 'tool';
  /** コンテンツ */
  content: string;
  /** ツール呼び出しID */
  toolCallId?: string;
}

/**
 * LLMツール定義
 */
export interface LLMTool {
  /** ツール名 */
  name: string;
  /** 説明 */
  description: string;
  /** パラメータスキーマ */
  parameters: Record<string, unknown>;
}

/**
 * パースされたアクション
 */
export interface ParsedAction {
  /** アクションタイプ */
  type:
    | 'connect'
    | 'disconnect'
    | 'command'
    | 'chat'
    | 'move'
    | 'interact'
    | 'wait'
    | 'assert_position'
    | 'assert_chat'
    | 'assert_form'
    | 'form_click'
    | 'form_submit'
    | 'form_close';
  /** パラメータ */
  params: Record<string, unknown>;
  /** 対象プレイヤー */
  player?: string;
  /** 説明 */
  description: string;
}

/**
 * LLMレスポンス（ツール呼び出し）
 */
export interface LLMToolCall {
  /** ツールID */
  id: string;
  /** ツール名 */
  name: string;
  /** 引数 */
  arguments: Record<string, unknown>;
}

/**
 * LLMプロバイダー設定
 */
export interface LLMProviderConfig {
  /** プロバイダー種別 */
  provider: 'anthropic' | 'openai';
  /** APIキー */
  apiKey?: string;
  /** モデル名 */
  model?: string;
  /** ベースURL（カスタムエンドポイント用） */
  baseURL?: string;
}
