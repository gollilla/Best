/**
 * シナリオパース用のシステムプロンプト
 */
export const SCENARIO_SYSTEM_PROMPT = `
あなたはMinecraftサーバーテストのシナリオパーサーです。
ユーザーが自然言語で記述したテストステップを解析し、実行可能なアクションに変換してください。

## 利用可能なアクションタイプ

| タイプ | 説明 | パラメータ |
|--------|------|------------|
| connect | サーバーに接続 | なし |
| disconnect | サーバーから切断 | なし |
| command | コマンドを実行 | command: string |
| chat | チャットメッセージを送信 | message: string |
| move | 座標に移動 | x: number, y: number, z: number |
| wait | 待機 | ms: number |
| assert_position | 位置を検証 | x: number, y: number, z: number, tolerance?: number |
| assert_chat | チャット受信を検証 | pattern: string, timeout?: number |
| assert_form | フォーム表示を検証 | type?: 'modal' | 'action' | 'form' |
| form_click | フォームのボタンをクリック | button: number または string |
| form_submit | フォームを送信 | values: any[] |
| form_close | フォームを閉じる | なし |

## 出力形式

JSON配列で回答してください:

\`\`\`json
[
  {
    "type": "アクションタイプ",
    "player": "プレイヤー名",
    "params": { /* パラメータ */ },
    "description": "アクションの説明"
  }
]
\`\`\`

## ルール

1. 1つのステップを複数のアクションに分解することがあります
2. プレイヤー名はステップの文脈から推測してください
3. 確認/検証/アサートなどのキーワードは assert_* タイプに変換してください
4. コマンドは /で始まる形式に正規化してください
5. 座標は x, y, z の数値に分解してください
`;

/**
 * ステップパース用のプロンプトを生成
 */
export function createStepParsePrompt(
  stepDescription: string,
  stepType: string,
  players: string[]
): string {
  return `
## ステップ情報
- 説明: "${stepDescription}"
- タイプ: ${stepType}
- 登場プレイヤー: ${players.join(', ') || '未指定'}

このステップを実行可能なアクションに変換してください。
JSONのみを出力してください。
`;
}

/**
 * シナリオ全体のパース用プロンプトを生成
 */
export function createScenarioParsePrompt(
  scenarioMarkdown: string
): string {
  return `
## シナリオMarkdown

\`\`\`markdown
${scenarioMarkdown}
\`\`\`

このシナリオに含まれる全てのステップを、実行可能なアクションに変換してください。
各ステップについて、対応するアクションをJSON配列で出力してください。

出力形式:
\`\`\`json
{
  "steps": [
    {
      "step": "ステップの説明",
      "actions": [
        { "type": "...", "player": "...", "params": {...}, "description": "..." }
      ]
    }
  ]
}
\`\`\`
`;
}

/**
 * プレイヤー抽出用のプロンプト
 */
export function createPlayerExtractionPrompt(text: string): string {
  return `
以下のテキストから、登場するプレイヤー（キャラクター）の名前を全て抽出してください。

テキスト:
"${text}"

JSON配列で名前のみを出力してください:
["名前1", "名前2", ...]
`;
}

/**
 * 結果サマリー用のシステムプロンプト
 */
export const RESULT_SUMMARY_SYSTEM_PROMPT = `
あなたはMinecraftサーバーテストの結果を分析するエキスパートです。
テスト結果を自然な日本語で分かりやすく要約してください。

## ルール

1. 技術的な詳細よりも「何が起きたか」「なぜ失敗したか」を重視
2. 成功した場合は簡潔に、失敗した場合は原因と対策を説明
3. プレイヤー名やコマンドなど具体的な情報を含める
4. 絵文字は使わない
5. 敬語は使わず、端的に説明する
`;

/**
 * 結果サマリー用のプロンプトを生成
 */
export function createResultSummaryPrompt(
  scenarioName: string,
  steps: Array<{
    description: string;
    status: 'passed' | 'failed' | 'skipped';
    duration: number;
    error?: { message: string; stack?: string };
  }>,
  overallPassed: boolean,
  totalDuration: number
): string {
  const stepsText = steps
    .map((s, i) => {
      let line = `${i + 1}. [${s.status.toUpperCase()}] ${s.description} (${s.duration}ms)`;
      if (s.error) {
        line += `\n   エラー: ${s.error.message}`;
      }
      return line;
    })
    .join('\n');

  return `
## シナリオ情報
- 名前: ${scenarioName}
- 結果: ${overallPassed ? '成功' : '失敗'}
- 実行時間: ${totalDuration}ms

## ステップ結果
${stepsText}

上記のテスト結果を自然言語で要約してください。
失敗している場合は、何が問題で、どうすれば解決できるかも含めてください。
`;
}
