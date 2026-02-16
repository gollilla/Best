# Best - Bedrock Edition Server Testing

統合版マインクラフト（Minecraft Bedrock Edition）のサーバー用テストライブラリです。

仮想プレイヤーとしてサーバーに接続し、自然言語によるシナリオ記載でAIエージェントが自動的にテストを実行・アサーションまで行います。

[WIP]

## 特徴

- **仮想プレイヤーエージェント**: Bedrockサーバーに接続する仮想プレイヤー
- **AIエージェント**: 自然言語によるシナリオ記載でAIエージェントが動作し、アサーションまで実行
- **プレイヤー偽装**: 本物のプレイヤーとして振る舞い、サーバーとの完全な対話が可能
- **包括的アサーションフレームワーク**: 50+のアサーションカテゴリ
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

### コマンド実行
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
- [x] 基本アサーション (Connection, Position, Chat, Command, CommandOutput, Form)
- [x] プレイヤー状態系アサーション (Inventory, Health, Hunger, Effect, Gamemode, Permission, Tag)
- [x] ワールド/ブロック系アサーション (Block, Entity, Scoreboard)
- [x] UI/表示系アサーション (Title, Subtitle, Actionbar, Sound, Particle, Form)
- [x] イベント系アサーション (Connection, Teleport, Dimension)
- [x] 汎用アサーション (真偽値, 等価性, Nil, 数値比較, 文字列, コレクション)

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

### Phase 5-7
- [x] シナリオランナー
- [x] LLM統合
- [ ] CLI

## アサーション一覧

### 基本アサーション
- **接続状態**: `ToBeConnected`, `ToBeDisconnected`
- **Position**: `ToBe`, `ToBeNear`, `ToReach`
- **Chat**: `ToReceive`, `NotToReceive`, `ToReceiveInOrder`, `ToContain`
- **Command**: `ToSucceed`, `ToFail`, `ToContain`
- **CommandOutput**: `ToReceive`, `ToReceiveAny`, `ToContain`, `ToMatch`, `ToReceiveWithStatusCode`

### プレイヤー状態系アサーション
- **Inventory**: `ToHaveItem`, `ToHaveItemCount`, `ToBeEmpty`
- **Health**: `ToBe`, `ToBeAbove`, `ToBeBelow`, `ToBeFull`
- **Hunger**: `ToBe`, `ToBeAbove`, `ToBeFull`
- **Effect**: `ToHave`, `NotToHave`, `ToHaveLevel`
- **Gamemode**: `ToBe`, `ToBeSurvival`, `ToBeCreative`
- **Permission**: `ToBeOperator`, `ToHaveLevel`, `ToBeAtLeast`
- **Tag**: `ToHave`, `NotToHave`

### ワールド/ブロック系アサーション
- **Block**: `ToBe`, `ToBeAt`, `ToBeAir`
- **Entity**: `ToExist`, `ToBeNearby`, `ToHaveCount`
- **Scoreboard**: `ToHaveValue`, `ToHaveObjective`, `ToHaveScore`, `ToHaveScoreAbove`, `ToHaveScoreBelow`, `ToHaveScoreBetween`, `ToHaveDisplaySlot`, `ToHaveFakePlayerScore`, `NotToHaveObjective`

### UI/表示系アサーション
- **Title**: `ToReceive`, `ToContain`
- **Subtitle**: `ToReceive`, `ToContain`
- **Actionbar**: `ToReceive`, `ToContain`
- **Sound**: `ToPlay`, `NotToPlay`
- **Particle**: `ToSpawn`
- **Form**: `ToReceive`, `ToReceiveWithTitle`, `ToBeModal`, `ToBeActionForm`, `ToBeCustomForm`, `ToHaveTitle`, `ToContainTitle`, `ToHaveButton`, `ToHaveButtons`, `ToHaveContent`

### イベント系アサーション
- **Connection**: `ToBeKicked`, `ToBeBanned`
- **Teleport**: `ToOccur`, `ToDestination`
- **Dimension**: `ToChangeTo`

### タイミング系アサーション
- **Timing**: `ToCompleteWithin`, `ToTimeout`
- **Sequence**: `ToOccurInOrder`
- **Condition**: `ToBeMetWithin`

### 汎用アサーション
- **真偽値**: `IsTrue`, `IsFalse`
- **等価性**: `Equal`, `NotEqual`
- **Nil**: `IsNil`, `NotNil`
- **数値比較**: `GreaterThan`, `LessThan`, `GreaterThanOrEqual`, `LessThanOrEqual`, `InRange`
- **文字列**: `Contains`, `NotContains`, `HasPrefix`, `HasSuffix`, `IsEmpty`, `NotEmpty`
- **コレクション**: `LengthEqual`, `IsEmptyCollection`, `NotEmptyCollection`, `ContainsElement`

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
## ライセンス

MIT License

## 参考

- [Gophertunnel](https://github.com/sandertv/gophertunnel) - Bedrock Editionプロトコル実装
