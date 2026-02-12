# Best - Bedrock Edition Server Testing

統合版マインクラフト（Minecraft Bedrock Edition）のサーバー用テストライブラリのGo実装です。

## 特徴

- **仮想プレイヤーエージェント**: Bedrockサーバーに接続する仮想プレイヤー
- **包括的アサーションフレームワーク**: 15+のアサーションカテゴリ
- **イベント駆動アーキテクチャ**: チャネルベースの非同期イベント処理
- **テストランナー**: describe/test/itスタイルのテストフレームワーク
- **複数サーバー対応**: PowerNukkitX、PocketMine-MP、BDS等に対応
- **Gophertunnel**: 最新のBedrock Editionプロトコルサポート

## インストール

```bash
go get github.com/gollilla/best
```

## クイックスタート

### 設定ファイル

`best.config.yml` を作成して、サーバー接続情報を設定できます：

```yaml
server:
  host: localhost
  port: 19132

agent:
  username: TestBot
  timeout: 30
  commandPrefix: "/"
  # コマンド送信方式: "text" (デフォルト) または "request"
  # PNX: "request" を推奨
  # PMMP/BDS: "text" を推奨
  commandSendMethod: text
  # コマンドレスポンス待機タイムアウト（秒）
  commandTimeout: 5
```

設定ファイルを使ってAgentを作成：

```go
// 設定ファイルを読み込み
cfg, err := best.LoadConfig()
if err != nil {
    cfg = best.DefaultConfig()
}

// 設定からAgentを作成
agent := best.NewAgentFromConfig(cfg)
agent.Connect()
```

### 注意事項
エージェントの利用にあたり、各サーバーソフトウェアの設定にてXbox認証を無効化する必要があります

### 基本的な接続

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/gollilla/best"
)

func main() {
    // エージェント作成
    agent := best.NewAgent(
        best.WithHost("localhost"),
        best.WithPort(19132),
        best.WithUsername("TestBot"),
    )

    // イベントリスナー登録
    agent.Emitter().On(best.EventChat, func(data best.EventData) {
        msg := data.(*best.ChatMessage)
        fmt.Printf("[%s]: %s\n", msg.Sender, msg.Message)
    })

    // 接続
    if err := agent.Connect(); err != nil {
        panic(err)
    }
    defer agent.Disconnect()

    // スポーン待機
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    agent.WaitForSpawn(ctx)

    // チャット送信
    agent.Chat("Hello, world!")

    // コマンド実行（レスポンスはアサーションで待機）
    agent.Command("/help")
    agent.Expect().Chat().ToReceive("help", 3*time.Second, nil)
}
```

### コマンド実行

コマンドの実行とレスポンスの待機は分離されています：

```go
// コマンド送信
agent.Command("/say hello")

// レスポンスを待機（サーバーによって異なる）
// PNX: CommandOutputパケットで応答
agent.Expect().CommandOutput().ToContain("hello", 3*time.Second)

// PMMP/BDS: Chatパケットで応答
agent.Expect().Chat().ToReceive("hello", 3*time.Second, nil)
```

### テストランナー

Jest/Mocha風のテストフレームワークを提供します：

```go
package main

import (
    "time"
    "github.com/gollilla/best"
)

func main() {
    runner := best.NewRunner(nil)
    var agent *best.Agent

    best.BeforeAll(func(ctx *best.TestContext) {
        // 名前だけ指定！設定は best.config.yml から自動読み込み
        agent = best.CreateAgent("TestBot")
        agent.Connect()
    })

    best.AfterAll(func(ctx *best.TestContext) {
        agent.Disconnect()
    })

    best.Describe("Server Tests", func() {
        best.It("should connect", func(ctx *best.TestContext) {
            if !agent.IsConnected() {
                panic("not connected")
            }
        })

        best.It("should receive title", func(ctx *best.TestContext) {
            go func() {
                agent.Command("/title @s title Hello")
            }()
            agent.Expect().Title().ToReceive("Hello", 3*time.Second)
        })

        best.It("should execute command", func(ctx *best.TestContext) {
            agent.Command("/help")
            // PNXの場合
            agent.Expect().CommandOutput().ToReceiveAny(3*time.Second)
        })
    })

    runner.Run()
}
```

**3つの方法でAgentを作成**:

```go
// 方法1: 最もシンプル - 名前だけ指定（推奨）
agent := best.CreateAgent("TestBot")  // best.config.yml の設定を使用

// 方法2: 名前 + オプション（設定を一部上書き）
agent := best.CreateAgent("Player1", best.WithHost("example.com"))

// 方法3: 完全に手動
agent := best.NewAgent(
    best.WithHost("localhost"),
    best.WithPort(19132),
    best.WithUsername("TestBot"),
)
```

**複数Agent**:

```go
var agent1, agent2 *best.Agent

best.BeforeAll(func(ctx *best.TestContext) {
    agent1 = best.CreateAgent("Player1")
    agent2 = best.CreateAgent("Player2")
    agent1.Connect()
    agent2.Connect()
})
```

**特徴**:
- 最小限のコード - 名前だけ指定すれば動く
- 設定ファイルで接続情報を一元管理
- 必要に応じてオプションで上書き可能
- 複数Agentの同時使用に対応
- Jest/Mocha風のdescribe/test/it構文

## サーバー別設定

### PowerNukkitX (PNX)

```yaml
agent:
  commandSendMethod: request  # CommandRequestパケットを使用
```

```go
// PNXはCommandOutputパケットでレスポンスを返す
agent.Command("/help")
agent.Expect().CommandOutput().ToContain("help", 3*time.Second)
```

### PocketMine-MP / BDS

```yaml
agent:
  commandSendMethod: text  # Textパケットを使用（デフォルト）
```

```go
// PMMMはTextパケット（チャット）でレスポンスを返す
agent.Command("/help")
agent.Expect().Chat().ToReceive("help", 3*time.Second, nil)
```

## 設定項目一覧

### server

| 項目 | 型 | デフォルト | 説明 |
|------|-----|-----------|------|
| `host` | string | `localhost` | サーバーホスト |
| `port` | int | `19132` | ポート番号 |
| `version` | string | - | Minecraftバージョン（省略可） |

### agent

| 項目 | 型 | デフォルト | 説明 |
|------|-----|-----------|------|
| `username` | string | `TestBot` | ユーザー名 |
| `timeout` | int | `30` | 接続タイムアウト（秒） |
| `commandPrefix` | string | `/` | コマンドプレフィックス |
| `commandSendMethod` | string | `text` | コマンド送信方式（`text` or `request`） |
| `commandTimeout` | int | `5` | コマンドレスポンス待機タイムアウト（秒） |

## プロジェクト構造

```
best/
├── cmd/best/              # CLIエントリーポイント
├── pkg/
│   ├── agent/             # Agentコア実装
│   ├── events/            # イベントシステム
│   ├── protocol/          # Gophertunnelラッパー
│   ├── state/             # プレイヤー状態管理
│   ├── assertions/        # アサーションフレームワーク
│   ├── runner/            # テストランナー
│   ├── scenario/          # シナリオランナー
│   ├── llm/               # LLM統合
│   ├── config/            # 設定管理
│   ├── types/             # 共通型定義
│   └── utils/             # ユーティリティ
├── examples/              # 使用例
│   ├── pnx/               # PowerNukkitX用テスト例
│   ├── pmmp/              # PocketMine-MP用テスト例
│   └── ...
├── best.go                # メインパッケージ
├── best.config.example.yml # 設定ファイルのテンプレート
└── go.mod
```

## 実装状況

### Phase 1: 基盤- [x] Go moduleセットアップ
- [x] イベントシステム (pkg/events/)
- [x] プロトコル層 (pkg/protocol/)
- [x] プレイヤー状態 (pkg/state/)
- [x] 基本Agent (pkg/agent/)

### Phase 2: コアアクション- [x] Agentアクション (Chat, Command, Goto等)
- [x] ワールド管理 (pkg/world/)
- [x] 追加パケットハンドラー (40+パケット対応)
- [x] タスクランナー

### Phase 3: アサーション- [x] AssertionContext
- [x] 基本アサーション (Position, Chat, Command, CommandOutput)
- [x] インベントリアサーション
- [x] プレイヤー状態アサーション (Health, Hunger, Effect, Gamemode, Permission, Tag)
- [x] UI/表示アサーション (Title, Subtitle, Actionbar, Scoreboard(TODO))
- [x] ブロック/エンティティアサーション

### Phase 4: テストランナー- [x] TestRunner (describe/test/it)
- [x] フック (BeforeAll, AfterAll, BeforeEach, AfterEach)
- [x] Skip/Only機能
- [x] Reporter (ConsoleReporter)
- [x] グローバル関数API
- [x] リトライロジック
- [x] タイムアウト処理
- [x] 設定ファイルサポート (best.config.yml)
- [x] 柔軟なAgent管理（複数Agent対応）
- [x] 複数サーバー種類対応 (PNX, PMMP, BDS)

### Phase 5-7（予定）
- [ ] シナリオランナー
- [ ] LLM統合
- [ ] CLI

## アサーション一覧

### 基本アサーション
- **Position**: `toBe`, `toBeNear`, `toReach`
- **Chat**: `toReceive`, `notToReceive`, `toReceiveInOrder`, `toContain`
- **CommandOutput**: `toReceive`, `toReceiveAny`, `toContain`, `toMatch`, `toReceiveWithStatusCode`

### プレイヤー状態アサーション
- **Health**: `toBe`, `toBeAbove`, `toBeBelow`, `toBeFull`
- **Hunger**: `toBe`, `toBeAbove`, `toBeFull`
- **Effect**: `toHave`, `notToHave`, `toHaveLevel`
- **Gamemode**: `toBe`, `toBeSurvival`, `toBeCreative`
- **Permission**: `toBeOperator`, `toHaveLevel`
- **Tag**: `toHave`, `notToHave`

### インベントリアサーション
- **Inventory**: `toHaveItem`, `toHaveItemCount`, `toBeEmpty`

### UI/表示アサーション
- **Title**: `toReceive`, `toReceiveSubtitle`, `toReceiveActionbar`, `toContain`
- **Scoreboard**: `toHaveObjective`, `toHaveScore`, `toHaveScoreAbove`
- **Form**: `toReceive`, `toReceiveWithTitle`, `toBeModal`, `toBeActionForm`, `toBeCustomForm`

詳細は [pkg/assertions/](pkg/assertions/) を参照してください。

## 例の実行

```bash
# PowerNukkitX用テスト
cd examples/pnx
go run main.go

# PocketMine-MP用テスト
cd examples/pmmp
go run main.go
```

**注意**:
- 各exampleディレクトリに移動してから実行してください（設定ファイルを読み込むため）
- Minecraftサーバーが `localhost:19132` で実行されている必要があります

## ライセンス

MIT License

## 参考

- [Gophertunnel](https://github.com/sandertv/gophertunnel) - Bedrock Editionプロトコル実装
