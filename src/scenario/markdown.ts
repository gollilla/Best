import type {
  ParsedScenario,
  PlayerDefinition,
  ScenarioStep,
} from '../types/scenario.js';

/**
 * Markdownシナリオをパースする
 *
 * 対応するフォーマット:
 * ```markdown
 * # シナリオ名
 *
 * ## プレイヤー
 * - Alice: 購入者
 * - Bob: 店主
 *
 * ## 前提条件
 * - サーバーが起動している
 *
 * ## ステップ
 * 1. Aliceがサーバーに接続する
 * 2. Bobがサーバーに接続する
 * 3. Aliceが `/shop` コマンドを実行する
 * 4. **確認**: Aliceにフォームが表示される
 * ```
 */
export function parseScenarioMarkdown(markdown: string): ParsedScenario {
  const lines = markdown.split('\n');
  let name = '';
  const players: PlayerDefinition[] = [];
  const preconditions: string[] = [];
  const steps: ScenarioStep[] = [];

  let currentSection: 'none' | 'players' | 'preconditions' | 'steps' = 'none';

  for (const line of lines) {
    const trimmed = line.trim();

    // タイトル（# で始まる）
    if (trimmed.startsWith('# ') && !trimmed.startsWith('## ')) {
      name = trimmed.slice(2).trim();
      continue;
    }

    // セクションヘッダー
    if (trimmed.startsWith('## ')) {
      const sectionName = trimmed.slice(3).trim().toLowerCase();
      if (
        sectionName === 'プレイヤー' ||
        sectionName === 'players' ||
        sectionName === 'player'
      ) {
        currentSection = 'players';
      } else if (
        sectionName === '前提条件' ||
        sectionName === 'preconditions' ||
        sectionName === 'prerequisites'
      ) {
        currentSection = 'preconditions';
      } else if (
        sectionName === 'ステップ' ||
        sectionName === 'steps' ||
        sectionName === 'step'
      ) {
        currentSection = 'steps';
      } else {
        currentSection = 'none';
      }
      continue;
    }

    // 空行やコメントはスキップ
    if (trimmed === '' || trimmed.startsWith('<!--')) {
      continue;
    }

    // セクション内のコンテンツをパース
    switch (currentSection) {
      case 'players':
        const player = parsePlayerLine(trimmed);
        if (player) {
          players.push(player);
        }
        break;

      case 'preconditions':
        const precondition = parseBulletLine(trimmed);
        if (precondition) {
          preconditions.push(precondition);
        }
        break;

      case 'steps':
        const step = parseStepLine(trimmed);
        if (step) {
          steps.push(step);
        }
        break;
    }
  }

  return {
    name,
    players,
    preconditions,
    steps,
    raw: markdown,
  };
}

/**
 * プレイヤー行をパース
 * 例: "- Alice: 購入者" → { name: "Alice", description: "購入者" }
 */
function parsePlayerLine(line: string): PlayerDefinition | null {
  // "- Name: Description" または "- Name" 形式
  const match = line.match(/^[-*]\s+([^:]+)(?::\s*(.*))?$/);
  if (!match) return null;

  return {
    name: match[1].trim(),
    description: match[2]?.trim(),
  };
}

/**
 * 箇条書き行をパース
 */
function parseBulletLine(line: string): string | null {
  const match = line.match(/^[-*]\s+(.+)$/);
  return match ? match[1].trim() : null;
}

/**
 * ステップ行をパース
 * 例: "1. Aliceがサーバーに接続する" → { type: 'action', description: '...', players: ['Alice'] }
 * 例: "**確認**: Aliceにフォームが表示される" → { type: 'assertion', ... }
 */
function parseStepLine(line: string): ScenarioStep | null {
  // 番号付きリスト "1. ..." または "- ..."
  const listMatch = line.match(/^(?:\d+\.\s*|[-*]\s+)(.+)$/);
  if (!listMatch) return null;

  const content = listMatch[1].trim();

  // **確認**: または **Assertion**: で始まる場合はアサーション
  const assertionMatch = content.match(
    /^\*\*(?:確認|検証|アサート|Assertion|Assert|Check)\*\*[：:]\s*(.+)$/i
  );
  if (assertionMatch) {
    const description = assertionMatch[1].trim();
    return {
      type: 'assertion',
      description,
      players: extractPlayers(description),
      raw: line,
    };
  }

  // 「X秒待つ」「wait X seconds」などは待機ステップ
  const waitMatch = content.match(
    /^(?:(\d+)\s*(?:秒|ミリ秒|ms|seconds?|milliseconds?)\s*(?:待つ|待機|wait)|(?:wait|待つ|待機)\s*(\d+)\s*(?:秒|ミリ秒|ms|seconds?|milliseconds?))/i
  );
  if (waitMatch) {
    const time = parseInt(waitMatch[1] || waitMatch[2], 10);
    const isMs = /ミリ秒|ms|milliseconds?/i.test(content);
    return {
      type: 'wait',
      description: content,
      raw: line,
    };
  }

  // それ以外はアクション
  return {
    type: 'action',
    description: content,
    players: extractPlayers(content),
    raw: line,
  };
}

/**
 * テキストからプレイヤー名を抽出
 * 例: "Aliceがサーバーに接続する" → ["Alice"]
 * 例: "AliceとBobがチャットする" → ["Alice", "Bob"]
 */
function extractPlayers(text: string): string[] {
  const players: string[] = [];

  // よくあるパターン: "Xが...", "Xは...", "Xに..."
  // 大文字で始まる名前、または日本語名を想定
  const patterns = [
    /^([A-Z][a-zA-Z0-9_]*|[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FFF]+)(?:が|は|に|を|と|から)/,
    /([A-Z][a-zA-Z0-9_]*|[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FFF]+)(?:と|,\s*)([A-Z][a-zA-Z0-9_]*|[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FFF]+)(?:が|は)/,
  ];

  for (const pattern of patterns) {
    const match = text.match(pattern);
    if (match) {
      for (let i = 1; i < match.length; i++) {
        if (match[i] && !players.includes(match[i])) {
          players.push(match[i]);
        }
      }
    }
  }

  // 「〜にメッセージ」「〜に向かって」などのパターン
  const toMatch = text.match(
    /([A-Z][a-zA-Z0-9_]*|[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FFF]+)(?:に向かって|にメッセージ|に対して)/
  );
  if (toMatch && !players.includes(toMatch[1])) {
    players.push(toMatch[1]);
  }

  return players;
}

/**
 * シナリオ内の全プレイヤー名を取得
 */
export function getAllPlayerNames(scenario: ParsedScenario): string[] {
  const names = new Set<string>();

  // プレイヤー定義から
  for (const player of scenario.players) {
    names.add(player.name);
  }

  // ステップから
  for (const step of scenario.steps) {
    if (step.players) {
      for (const player of step.players) {
        names.add(player);
      }
    }
  }

  return Array.from(names);
}
