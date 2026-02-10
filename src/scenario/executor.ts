import type { Agent } from '../agent/index.js';
import type { ScenarioContext } from './context.js';
import type {
  ParsedScenario,
  ScenarioStep,
  ScenarioResult,
  ScenarioStepResult,
} from '../types/scenario.js';
import type { LLMProvider, ParsedAction } from '../types/llm.js';
import { LLMProcessor } from '../llm/index.js';

export interface ExecutorOptions {
  /** LLMプロバイダー */
  llmProvider?: LLMProvider;
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
 * シナリオ実行エンジン
 */
export class ScenarioExecutor {
  constructor(
    private context: ScenarioContext,
    private options: ExecutorOptions = {}
  ) {}

  /**
   * シナリオを実行
   */
  async execute(scenario: ParsedScenario): Promise<ScenarioResult> {
    const result: ScenarioResult = {
      name: scenario.name,
      steps: [],
      passed: true,
      duration: 0,
    };

    const startTime = Date.now();

    try {
      for (const step of scenario.steps) {
        const stepResult = await this.executeStep(step);
        result.steps.push(stepResult);

        if (stepResult.status === 'failed') {
          result.passed = false;
          result.error = stepResult.error;
          break;
        }
      }
    } catch (error) {
      result.passed = false;
      result.error = error as Error;
    }

    result.duration = Date.now() - startTime;

    // 自然言語サマリーを生成
    if (this.options.generateSummary && this.options.llmProvider) {
      try {
        const processor = new LLMProcessor(this.options.llmProvider);
        result.summary = await processor.summarizeResult(
          scenario.name,
          result.steps,
          result.passed,
          result.duration
        );
      } catch (summaryError) {
        if (this.options.verbose) {
          console.warn('[Executor] Failed to generate summary:', summaryError);
        }
      }
    }

    return result;
  }

  /**
   * 単一のステップを実行
   */
  private async executeStep(step: ScenarioStep): Promise<ScenarioStepResult> {
    const stepStart = Date.now();

    if (this.options.verbose) {
      console.log(`[Step] ${step.description}`);
    }

    try {
      // LLMでステップを解析してアクションに変換
      const actions = await this.parseStepToActions(step);

      // 各アクションを実行
      for (const action of actions) {
        await this.executeAction(action);
      }

      return {
        description: step.description,
        status: 'passed',
        duration: Date.now() - stepStart,
      };
    } catch (error) {
      return {
        description: step.description,
        status: 'failed',
        duration: Date.now() - stepStart,
        error: error as Error,
      };
    }
  }

  /**
   * ステップをLLMで解析してアクションに変換
   */
  private async parseStepToActions(step: ScenarioStep): Promise<ParsedAction[]> {
    // LLMがない場合は簡易パーサーを使用
    if (!this.options.llmProvider) {
      return this.parseStepSimple(step);
    }

    // LLMでステップを解析
    const prompt = this.buildParsePrompt(step);
    const response = await this.options.llmProvider.complete(prompt);

    try {
      return JSON.parse(response);
    } catch {
      // JSONパースに失敗した場合は簡易パーサーにフォールバック
      return this.parseStepSimple(step);
    }
  }

  /**
   * LLM用のプロンプトを構築
   */
  private buildParsePrompt(step: ScenarioStep): string {
    return `
以下のテストステップを解析し、実行可能なアクションに変換してください。

ステップ: "${step.description}"
タイプ: ${step.type}

利用可能なアクションタイプ:
- connect: サーバーに接続
- disconnect: サーバーから切断
- command: コマンドを実行 (params: { command: string })
- chat: チャットメッセージを送信 (params: { message: string })
- move: 座標に移動 (params: { x: number, y: number, z: number })
- wait: 待機 (params: { ms: number })
- assert_position: 位置を検証 (params: { x, y, z, tolerance? })
- assert_chat: チャット受信を検証 (params: { pattern: string, timeout? })
- assert_form: フォーム表示を検証 (params: { type?: 'modal'|'action'|'form' })
- form_click: フォームのボタンをクリック (params: { button: number | string })
- form_submit: フォームを送信 (params: { values: any[] })
- form_close: フォームを閉じる

JSONで回答してください。形式:
[
  { "type": "アクションタイプ", "player": "プレイヤー名", "params": {...}, "description": "説明" }
]
`;
  }

  /**
   * 簡易パーサー（LLMなしの場合）
   */
  private parseStepSimple(step: ScenarioStep): ParsedAction[] {
    const text = step.description;
    const player = step.players?.[0] ?? 'default';

    // 接続パターン
    if (/サーバーに接続|connect|ログイン/i.test(text)) {
      return [{ type: 'connect', player, params: {}, description: text }];
    }

    // 切断パターン
    if (/切断|disconnect|ログアウト/i.test(text)) {
      return [{ type: 'disconnect', player, params: {}, description: text }];
    }

    // コマンド実行パターン
    const cmdMatch = text.match(/[「`]?\/?([^」`]+)[」`]?\s*(?:コマンド|command)?を?実行/i);
    if (cmdMatch) {
      const command = cmdMatch[1].startsWith('/') ? cmdMatch[1] : `/${cmdMatch[1]}`;
      return [
        { type: 'command', player, params: { command }, description: text },
      ];
    }

    // 別のコマンドパターン
    const cmdMatch2 = text.match(/\/([^\s]+(?:\s+[^\s]+)*)/);
    if (cmdMatch2) {
      return [
        {
          type: 'command',
          player,
          params: { command: `/${cmdMatch2[1]}` },
          description: text,
        },
      ];
    }

    // チャットパターン
    const chatMatch = text.match(/[「"]([^」"]+)[」"]\s*と?(?:言う|発言|チャット|say)/i);
    if (chatMatch) {
      return [
        { type: 'chat', player, params: { message: chatMatch[1] }, description: text },
      ];
    }

    // 移動パターン
    const moveMatch = text.match(
      /座標?\s*\(?\s*(-?\d+\.?\d*)\s*[,、]\s*(-?\d+\.?\d*)\s*[,、]\s*(-?\d+\.?\d*)\s*\)?/
    );
    if (moveMatch) {
      return [
        {
          type: 'move',
          player,
          params: {
            x: parseFloat(moveMatch[1]),
            y: parseFloat(moveMatch[2]),
            z: parseFloat(moveMatch[3]),
          },
          description: text,
        },
      ];
    }

    // 待機パターン
    const waitMatch = text.match(/(\d+)\s*(?:秒|ミリ秒|ms|seconds?)/i);
    if (waitMatch) {
      const time = parseInt(waitMatch[1], 10);
      const isMs = /ミリ秒|ms|milliseconds?/i.test(text);
      return [
        {
          type: 'wait',
          player,
          params: { ms: isMs ? time : time * 1000 },
          description: text,
        },
      ];
    }

    // フォーム表示確認パターン
    if (/フォーム.*表示|表示.*フォーム/i.test(text)) {
      return [
        { type: 'assert_form', player, params: {}, description: text },
      ];
    }

    // フォームクリックパターン
    const clickMatch = text.match(
      /フォーム.*[「"]([^」"]+)[」"].*クリック|[「"]([^」"]+)[」"].*ボタン.*クリック/i
    );
    if (clickMatch) {
      return [
        {
          type: 'form_click',
          player,
          params: { button: clickMatch[1] || clickMatch[2] },
          description: text,
        },
      ];
    }

    // チャット受信確認パターン
    const receiveMatch = text.match(/[「"]([^」"]+)[」"].*(?:受信|メッセージ|チャット)/i);
    if (receiveMatch) {
      return [
        {
          type: 'assert_chat',
          player,
          params: { pattern: receiveMatch[1] },
          description: text,
        },
      ];
    }

    // 位置確認パターン
    const posMatch = text.match(
      /位置.*\(?\s*(-?\d+\.?\d*)\s*[,、]\s*(-?\d+\.?\d*)\s*[,、]\s*(-?\d+\.?\d*)\s*\)?/
    );
    if (posMatch) {
      return [
        {
          type: 'assert_position',
          player,
          params: {
            x: parseFloat(posMatch[1]),
            y: parseFloat(posMatch[2]),
            z: parseFloat(posMatch[3]),
          },
          description: text,
        },
      ];
    }

    // 認識できない場合は空の配列
    console.warn(`[Executor] Unknown step pattern: ${text}`);
    return [];
  }

  /**
   * アクションを実行
   */
  private async executeAction(action: ParsedAction): Promise<void> {
    const playerName = action.player ?? 'default';
    const agent = await this.context.getPlayer(playerName);

    if (this.options.verbose) {
      console.log(`  [Action] ${action.type}: ${JSON.stringify(action.params)}`);
    }

    switch (action.type) {
      case 'connect':
        if (!agent.isConnected) {
          await agent.connect();
        }
        break;

      case 'disconnect':
        if (agent.isConnected) {
          await agent.disconnect();
        }
        break;

      case 'command':
        await agent.command(action.params.command as string);
        break;

      case 'chat':
        agent.say(action.params.message as string);
        break;

      case 'move':
        await agent.goto({
          x: action.params.x as number,
          y: action.params.y as number,
          z: action.params.z as number,
        });
        break;

      case 'wait':
        await new Promise((resolve) =>
          setTimeout(resolve, action.params.ms as number)
        );
        break;

      case 'assert_position':
        agent.expect.position.toBeNear(
          {
            x: action.params.x as number,
            y: action.params.y as number,
            z: action.params.z as number,
          },
          (action.params.tolerance as number) ?? 5
        );
        break;

      case 'assert_chat':
        await agent.expect.chat.toReceive(action.params.pattern as string, {
          timeout: (action.params.timeout as number) ?? 10000,
        });
        break;

      case 'assert_form':
        await agent.expect.form.toReceive({
          timeout: (action.params.timeout as number) ?? 10000,
        });
        break;

      case 'form_click':
        const form = agent.getPendingForm();
        if (!form) {
          throw new Error('No pending form to click');
        }
        const button = action.params.button;
        if (typeof button === 'number') {
          agent.respondToForm(form.id, button);
        } else {
          // テキストでボタンを検索（ActionFormの場合）
          if (form.type === 'action') {
            const index = form.buttons.findIndex((b) =>
              b.text.includes(button as string)
            );
            if (index === -1) {
              throw new Error(`Button not found: ${button}`);
            }
            agent.respondToForm(form.id, index);
          } else if (form.type === 'modal') {
            // ModalFormの場合
            if (form.button1.includes(button as string)) {
              agent.respondToForm(form.id, true);
            } else if (form.button2.includes(button as string)) {
              agent.respondToForm(form.id, false);
            } else {
              throw new Error(`Button not found: ${button}`);
            }
          }
        }
        break;

      case 'form_submit':
        const customForm = agent.getPendingForm();
        if (!customForm) {
          throw new Error('No pending form to submit');
        }
        agent.respondToForm(customForm.id, action.params.values as (string | number | boolean)[]);
        break;

      case 'form_close':
        const pendingForm = agent.getPendingForm();
        if (pendingForm) {
          agent.closeForm(pendingForm.id);
        }
        break;

      default:
        console.warn(`[Executor] Unknown action type: ${action.type}`);
    }
  }
}
