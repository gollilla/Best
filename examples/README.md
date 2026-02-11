# Best Testing Framework - Examples

このディレクトリには、Best Testing Frameworkの使い方を示すサンプルコードが含まれています。

## 包括的なテストサンプル

### main.go

すべての主要機能を網羅した、テストランナーを使用した完全なサンプルです。

**含まれる機能:**
- 接続とプレイヤー状態の確認
- コマンド実行とチャット送信
- タイトル/サブタイトル/アクションバーのアサーション
- スコアボード操作とアサーション（未実装部分はコメントアウト）
- **異常系テスト（エラーハンドリング、バリデーション、タイムアウト処理）**
- テストスイートの構造化
- BeforeAll/AfterAll フック
- **テストランナーによる自動結果出力（ConsoleReporter）**

### 実行方法

1. マインクラフトサーバーを起動します（デフォルト: localhost:19132）

2. サンプルを実行します:

```bash
cd examples
go run main.go
```

### 出力例

テストランナーが自動的に以下のような結果を出力します：

```
Running 5 test suite(s)...

接続とプレイヤー状態
    ✓ サーバーに接続されているべき (0ms)
    ✓ 有効な座標を持つべき (0ms)
    ✓ 体力が0より大きいべき (0ms)
    ✓ デフォルトでサバイバルモードであるべき (0ms)
    ✓ 最低でも権限レベル0を持つべき (0ms)

コマンド実行
    ✓ /helpコマンドを実行できるべき (1000ms)
    ✓ チャットメッセージを送信できるべき (0ms)

タイトル/サブタイトル表示
    ✓ タイトルメッセージを受信するべき (592ms)
    ✓ サブタイトルメッセージを受信するべき (605ms)
    ✓ アクションバーメッセージを受信するべき (595ms)
    ✓ タイトルに特定のテキストが含まれるべき (605ms)

異常系テスト
    ✓ 体力は有効な範囲内であるべき (0ms)
    ✓ 座標は有効な値であるべき (0ms)
    ✓ タイムアウトエラーを正しく処理するべき (50ms)

高度な機能 (スキップ)
    ○ このテストはスキップされます (skipped)

==================================================
Test Results:
==================================================
  Passed:  14
  Failed:  0
  Skipped: 1
  Duration: 6342ms
==================================================
```

> **Note**: `ConsoleReporter` が自動的に各テストの結果と実行時間を表示します。
> 手動で `log.Println` を使用する必要はありません！

## テストランナーの主要機能

### 自動結果出力

テストランナーはデフォルトで `ConsoleReporter` を使用し、以下の情報を自動的に出力します：

- テスト開始時のサマリー
- 各テストスイートの実行状況
- 各テストケースの結果（✓成功、✗失敗、○スキップ）と実行時間
- 最終的なテスト結果のサマリー（成功/失敗/スキップの数、合計時間）
- 失敗したテストの詳細（エラーメッセージとスタックトレース）

**重要**: 手動で `log.Println` を使用する必要はありません！

### テストスイートの定義

```go
best.Describe("テストスイート名", func() {
    best.It("テストケース名", func(ctx *best.TestContext) {
        // テストコード
    })
})
```

### フック

- `best.BeforeAll()` - すべてのテスト前に1回実行
- `best.AfterAll()` - すべてのテスト後に1回実行
- `best.BeforeEach()` - 各テスト前に実行
- `best.AfterEach()` - 各テスト後に実行

### テストのスキップ

```go
// テストスイート全体をスキップ
best.SkipDescribe("スキップされるテスト", func() { ... })

// 個別のテストをスキップ
best.SkipTest("スキップされるテスト", func(ctx *best.TestContext) { ... })
```

### 主要なアサーション

#### コマンド実行（フルエントAPI）
```go
// コマンドの成功をアサート
agent.Expect().Command("/help").ToSucceed()

// コマンドの失敗をアサート
agent.Expect().Command("/invalid").ToFail()

// 出力内容のアサート（チェーン可能）
agent.Expect().Command("/help").ToSucceed().And().ToContain("ban")

// ステータスコードのアサート
agent.Expect().Command("/help").ToHaveStatusCode(0)
```

#### プレイヤー状態
```go
agent.Expect().Health().ToBeAbove(0)
agent.Expect().Gamemode().ToBeSurvival()
agent.Expect().Permission().ToBeAtLeast(0)
```

#### タイトル/UI
```go
agent.Expect().Title().ToReceive("Hello", 3*time.Second)
agent.Expect().Title().ToReceiveSubtitle("Subtitle", 3*time.Second)
agent.Expect().Title().ToReceiveActionbar("Action", 3*time.Second)
agent.Expect().Title().ToContain("keyword", 3*time.Second)
```

#### スコアボード
```go
agent.Expect().Scoreboard().ToHaveObjective("score_name", 3*time.Second)
agent.Expect().Scoreboard().ToHaveScore("score_name", 100, 3*time.Second)
agent.Expect().Scoreboard().ToHaveScoreAbove("score_name", 50, 3*time.Second)
agent.Expect().Scoreboard().ToHaveScoreBetween("score_name", 0, 200, 3*time.Second)
agent.Expect().Scoreboard().NotToHaveObjective("score_name", 3*time.Second)
```

#### 異常系テスト
```go
// 値の範囲チェック
health := agent.Health()
if health < 0 || health > 20 {
    panic("体力が範囲外です")
}

// タイムアウトエラーのハンドリング
defer func() {
    if r := recover(); r != nil {
        // タイムアウトが発生（期待される動作）
    }
}()
agent.Expect().Title().ToReceive("絶対に送信されない", 50*time.Millisecond)
```

## カスタマイズ

サーバーの接続情報を変更する場合:

```go
agent = best.NewAgent(
    best.WithHost("your-server.com"),
    best.WithPort(19132),
    best.WithUsername("YourBotName"),
)
```

## トラブルシューティング

- **接続できない**: サーバーが起動していること、ポートが正しいことを確認してください
- **タイムアウトエラー**: アサーションのタイムアウト時間を調整してください
- **権限エラー**: テストによっては、オペレーター権限が必要な場合があります

## その他のリソース

- [メインドキュメント](../README.md)
- [テストガイド](../TESTING.md)
