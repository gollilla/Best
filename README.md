# Best

Bedrock Edition Server Testing - 統合版マインクラフトサーバー用テストライブラリ

## 特徴

- Agentによるサーバー接続・操作
- 豊富なアサーション (接続、座標、チャット、コマンド、フォーム)
- タスクランナーによる処理の順次実行
- チャット/コマンドハンドラー
- 自然言語によるシナリオテスト (Markdown形式)
- LLM統合 (OpenAI / Anthropic)
- TypeScript完全対応

## インストール

```bash
npm install @gollilla/best
```

## クイックスタート

### 1. 設定ファイルを作成

```typescript
// best.config.ts
import { defineConfig } from '@gollilla/best';

export default defineConfig({
  host: 'localhost',
  port: 19132,
  offline: true,
  timeout: 30000,
});
```

### 2. テストを作成

```typescript
// tests/example.test.ts
describe('接続テスト', () => {
  beforeEach(async ({ player }) => {
    await player.connect();
  });

  afterEach(async ({ player }) => {
    await player.disconnect();
  });

  test('サーバーに接続できる', async ({ player }) => {
    player.expect.toBeConnected();
  });

  test('チャットを送信できる', async ({ player }) => {
    player.chat('Hello, World!');
  });

  test('コマンドが実行できる', async ({ player }) => {
    const result = await player.command('/say Hello');
    player.expect.command(result).toSucceed();
  });
});
```

### 3. テストを実行

```bash
npx @gollilla/best
```

## Agent API

`Agent`はサーバーに接続する仮想プレイヤーです。アサーション、タスクランナー、高レベルアクションを内蔵しています。

### 基本的な使い方

```typescript
import { Agent, createAgent } from '@gollilla/best';

// 方法1: 直接作成
const agent = new Agent({
  host: 'localhost',
  port: 19132,
  username: 'TestBot',
  offline: true,
});
await agent.connect();

// 方法2: 設定ファイルから作成（自動接続）
const agent = await createAgentFromConfig('TestBot');
```

### アサーション

```typescript
// === 基本アサーション ===

// 接続状態
agent.expect.toBeConnected();
agent.expect.toBeDisconnected();

// 座標
agent.expect.position.toBe({ x: 100, y: 64, z: 100 });
agent.expect.position.toBeNear({ x: 100, y: 64, z: 100 }, 5);
await agent.expect.position.toReach({ x: 0, y: 64, z: 0 }, { timeout: 5000 });

// チャット
await agent.expect.chat.toReceive('Hello', { timeout: 5000 });
await agent.expect.chat.toReceive(/welcome/i);

// コマンド
const result = await agent.command('/say test');
agent.expect.command(result).toSucceed();

// フォーム
const form = await agent.expect.form.toReceive({ timeout: 5000 });
form.toHaveTitle('メニュー');
await form.clickButton(0);

// === プレイヤー状態系 ===

// インベントリ
agent.expect.inventory.toHaveItem('diamond');
agent.expect.inventory.toHaveItemCount('diamond', 10);
agent.expect.inventory.toHaveEnchantedItem('diamond_sword', 'sharpness', 5);

// 体力
agent.expect.health.toBeFull();
agent.expect.health.toBeAbove(10);
await agent.expect.health.toReach(20, { timeout: 5000 });

// エフェクト
agent.expect.effect.toHave('speed');
agent.expect.effect.toHaveLevel('strength', 2);

// ゲームモード
agent.expect.gamemode.toBeSurvival();
agent.expect.gamemode.toBeCreative();
await agent.expect.gamemode.toChangeTo('creative', { timeout: 5000 });

// 権限レベル
agent.expect.permission.toBeOperator();

// タグ
agent.expect.tag.toHave('vip');
agent.expect.tag.toHaveAll('team_red', 'player');

// === ワールド/ブロック系 ===

// ブロック
agent.expect.block.toBeAt({ x: 100, y: 64, z: 100 }, 'stone');
agent.expect.block.toBeAirAt({ x: 100, y: 65, z: 100 });
await agent.expect.block.toChangeTo({ x: 100, y: 64, z: 100 }, 'diamond_block');

// エンティティ
agent.expect.entity.toExist('zombie');
agent.expect.entity.toBeNearby('creeper', 10);
await agent.expect.entity.toSpawn('pig', { nearPlayer: 20 });

// スコアボード
agent.expect.scoreboard.toHaveValue('kills', 5);
agent.expect.scoreboard.toHaveMinValue('money', 100);

// === UI/表示系 ===

// タイトル
await agent.expect.title.toReceive('ゲーム開始');
await agent.expect.subtitle.toReceive(/準備完了/);
await agent.expect.actionbar.toReceive('残り時間: 60秒');

// サウンド
await agent.expect.sound.toPlay('mob.zombie.say');
await agent.expect.sound.notToPlay('mob.wither.spawn', { duration: 5000 });

// パーティクル
await agent.expect.particle.toSpawn('minecraft:heart', { nearPlayer: 5 });

// === イベント系 ===

// 接続
await agent.expect.connection.toBeKicked({ reason: /不正行為/ });
await agent.expect.connection.notToBeKicked({ duration: 5000 });

// テレポート
await agent.expect.teleport.toOccur({ minDistance: 10 });
await agent.expect.teleport.toDestination({ x: 0, y: 64, z: 0 });

// ディメンション
agent.expect.dimension.toBeOverworld();
await agent.expect.dimension.toChangeTo('nether');

// 死亡/リスポーン
await agent.expect.death.toOccur({ timeout: 10000 });
await agent.expect.respawn.toOccur();

// === タイミング系 ===

// 時間内完了
await agent.expect.timing.toCompleteWithin(
  async () => agent.command('/tp @s 0 64 0'),
  1000
);

// シーケンス
await agent.expect.sequence.toOccurInOrder([
  { event: 'chat', filter: (m) => m.message.includes('開始') },
  { event: 'form' },
  { event: 'chat', filter: (m) => m.message.includes('完了') },
]);

// 条件待機
await agent.expect.condition.toBeMetWithin(
  () => agent.health >= 20,
  5000
);
```

### 高レベルアクション

```typescript
// テレポート
await agent.goto({ x: 100, y: 64, z: 100 });

// 視線を向ける
await agent.lookAt({ x: 0, y: 64, z: 0 });

// ブロック設置
await agent.placeBlock({ x: 100, y: 65, z: 100 }, 'stone');

// ブロック破壊（実プレイヤー操作）
const result = await agent.breakBlock({ x: 100, y: 65, z: 100 });
if (result.success) {
  console.log('破壊成功');
} else {
  console.log('破壊失敗:', result.reason);
}

// ブロック破壊（オプション付き）
const controller = new AbortController();
await agent.breakBlock(
  { x: 100, y: 65, z: 100 },
  {
    toolMultiplier: 6.0,  // ツール倍率（素手=1.0, 鉄ピッケル≈6.0）
    signal: controller.signal,  // キャンセル用
    onProgress: (progress) => {
      console.log(`進捗: ${Math.floor(progress * 100)}%`);
    },
  }
);

// ブロック破壊（コマンド版 - クリエイティブモード用、即座に破壊）
await agent.breakBlockInstant({ x: 100, y: 65, z: 100 });

// チャット
agent.say('Hello!');
agent.chat('Hello!'); // sayと同じ
```

### ブロック破壊イベント

```typescript
// 破壊開始
agent.on('block_break_start', ({ position }) => {
  console.log(`破壊開始: ${position.x}, ${position.y}, ${position.z}`);
});

// 破壊完了
agent.on('block_break_complete', ({ position }) => {
  console.log(`破壊完了: ${position.x}, ${position.y}, ${position.z}`);
});

// 破壊中断
agent.on('block_break_abort', ({ position }) => {
  console.log(`破壊中断: ${position.x}, ${position.y}, ${position.z}`);
});
```

### ワールド情報

```typescript
// 指定座標のブロックを取得
const block = agent.getBlock(100, 64, 100);
console.log(block?.name); // "minecraft:stone"

// ブロック名だけ取得
const name = agent.getBlockName(100, 64, 100);

// 足元のブロック
const below = agent.getBlockBelow();

// 前方のブロック（視線方向）
const front = agent.getBlockInFront(2); // 2ブロック先

// 通行可能か確認
if (agent.isPassable(100, 65, 100)) {
  // 空気や水など
}

// 固体ブロックか確認
if (agent.isSolid(100, 64, 100)) {
  // 石や土など
}

// 周囲のブロック情報
const blocks = agent.getBlocksAround(2); // 半径2ブロック

// Worldオブジェクトに直接アクセス
console.log(agent.world.loadedChunkCount);

// ブロックの破壊データを取得
const breakData = agent.world.getBlockBreakData(100, 64, 100, 1.0);
console.log(breakData.baseTime);    // 破壊時間（秒）
console.log(breakData.instant);     // 即座に破壊可能か
console.log(breakData.unbreakable); // 破壊不可能か（岩盤など）
```

### タスクランナー

```typescript
agent.tasks
  .add('初期位置へ移動', async (a) => {
    await a.goto({ x: 0, y: 64, z: 0 });
  })
  .add('ブロック設置', async (a) => {
    await a.placeBlock({ x: 0, y: 65, z: 0 }, 'stone');
  })
  .add('確認', async (a) => {
    a.expect.position.toBeNear({ x: 0, y: 64, z: 0 }, 5);
  });

await agent.tasks.runAll();
```

### チャット/コマンドハンドラー

```typescript
// チャットパターンに反応
agent.onChat(/こんにちは/, (msg, reply) => {
  reply(`${msg.sender}さん、こんにちは！`);
});

// コマンドに反応 (!help で発火)
agent.onCommand('help', (args, reply) => {
  reply('使い方: !help, !status');
});

// ハンドラーを有効化
agent.startAgent();

// 停止
agent.stopAgent();
```

## シナリオテスト

Markdown形式で自然言語によるシナリオテストを記述できます。

### シナリオファイル

```markdown
<!-- scenarios/shop.scenario.md -->
# ショップ購入テスト

## プレイヤー
- Alice: 購入者
- Bob: 店主

## ステップ
1. Aliceがサーバーに接続する
2. Bobがサーバーに接続する
3. Aliceが `/shop` コマンドを実行する
4. **確認**: Aliceにフォームが表示される
5. Aliceがフォームで「ダイヤモンド」をクリックする
6. **確認**: Aliceが「購入完了」メッセージを受信する
```

### シナリオ設定

```typescript
// best.config.ts
import { defineConfig } from '@gollilla/best';

export default defineConfig({
  host: 'localhost',
  port: 19132,

  scenario: {
    match: ['scenarios/**/*.scenario.md'],
    stepTimeout: 30000,
    totalTimeout: 300000,

    // LLM設定（オプション）
    llm: {
      provider: 'anthropic',
      apiKey: process.env.ANTHROPIC_API_KEY,
      model: 'claude-sonnet-4-20250514',
    },
  },
});
```

### シナリオ実行

```bash
npx @gollilla/best scenario
npx @gollilla/best scenario scenarios/shop.scenario.md
npx @gollilla/best scenario --verbose
```

## 設定オプション

```typescript
interface BestConfig {
  // サーバー設定
  host: string;
  port?: number;           // デフォルト: 19132
  offline?: boolean;       // デフォルト: true
  timeout?: number;        // デフォルト: 30000

  // テスト設定
  testMatch?: string[];    // デフォルト: ['**/*.test.ts']
  retries?: number;        // デフォルト: 0
  bail?: boolean;          // デフォルト: false

  // シナリオ設定
  scenario?: {
    match?: string[];
    stepTimeout?: number;
    totalTimeout?: number;
    llm?: {
      provider: 'openai' | 'anthropic';
      apiKey: string;
      model?: string;
    };
  };
}
```

## CLI

```bash
npx @gollilla/best                              # テスト実行
npx @gollilla/best scenario                     # シナリオ実行
npx @gollilla/best scenario path/to/file.md    # 特定シナリオ
npx @gollilla/best scenario --verbose          # 詳細ログ
```

## ライセンス

MIT
