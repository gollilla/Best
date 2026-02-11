# Best Testing Guide

## テスト環境のセットアップ

### 1. Minecraft Bedrock Edition サーバー

**必要なもの:**
- Minecraft Bedrock Dedicated Server
- オフラインモードが有効な設定

**設定方法:**

1. [公式サイト](https://www.minecraft.net/en-us/download/server/bedrock)からBedrock Serverをダウンロード

2. `server.properties` を編集:
```properties
server-port=19132
online-mode=false
allow-cheats=true
gamemode=creative
```

3. サーバーを起動:
```bash
./bedrock_server
```

### 2. Go環境

```bash
# Go 1.21+ が必要
go version

# 依存関係のインストール
cd best-go
go mod download
```

## テストの実行

### 単体テスト

```bash
# 全パッケージのテスト
go test ./...

# 詳細出力
go test ./... -v

# カバレッジレポート
go test ./... -cover

# 特定のパッケージのみ
go test ./pkg/assertions/... -v
```

### 統合テスト

```bash
# Phase 1: 基本的な接続とイベント
cd examples/basic
go run main.go

# Phase 2: ワールド管理
cd examples/phase2
go run main.go

# Phase 3: アサーション
cd examples/assertions
go run main.go
```

## 現在のテストカバレッジ

### Phase 1: 基盤 ✅
- ✅ Agent接続/切断
- ✅ イベントシステム (Emitter, WaitFor)
- ✅ プレイヤー状態追跡
- ✅ チャットメッセージ
- ✅ コマンド実行 (Chat経由)

### Phase 2: ワールド管理 ✅
- ✅ ブロック追跡
- ✅ インベントリ更新
- ✅ エフェクト管理
- ✅ エンティティ追跡
- ✅ チャンク管理 (基本構造)

### Phase 3: アサーション (進行中)
- ✅ 接続アサーション
- ✅ 座標アサーション (8種類)
- ✅ チャットアサーション (4種類)
- ✅ コマンドアサーション (4種類)
- 🚧 インベントリアサーション
- 🚧 プレイヤー状態アサーション
- 🚧 ブロック/エンティティアサーション
- 🚧 UI/表示アサーション
- 🚧 イベントアサーション
- 🚧 タイミングアサーション

## テスト作成のガイドライン

### 単体テストの例

```go
func TestPositionAssertion_ToBe(t *testing.T) {
    agent := newMockAgent()
    assertion := &PositionAssertion{agent: agent}

    // 成功ケース
    err := assertion.ToBe(types.Position{X: 100, Y: 64, Z: 100})
    if err != nil {
        t.Errorf("Expected ToBe to pass, got error: %v", err)
    }

    // 失敗ケース
    err = assertion.ToBe(types.Position{X: 200, Y: 64, Z: 100})
    if err == nil {
        t.Error("Expected ToBe to fail for different position")
    }
}
```

### 統合テストの例

```go
func main() {
    agent := best.NewAgent(
        best.WithHost("localhost"),
        best.WithPort(19132),
        best.WithUsername("TestBot"),
    )

    if err := agent.Connect(); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer agent.Disconnect()

    // アサーションテスト
    if err := agent.Expect().ToBeConnected(); err != nil {
        log.Printf("Assertion failed: %v", err)
    }
}
```

## CI/CDでのテスト

### GitHub Actions (例)

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test ./... -v -cover

      - name: Build examples
        run: |
          cd examples/basic && go build
          cd ../phase2 && go build
          cd ../assertions && go build
```

## デバッグ

### ログ出力の有効化

```go
// イベントのデバッグログ
agent.Emitter().On(best.EventChat, func(data best.EventData) {
    msg := data.(*best.ChatMessage)
    log.Printf("Chat: %s from %s", msg.Message, msg.Sender)
})

// すべてのイベントをログ
agent.Emitter().On(best.EventPacket, func(data best.EventData) {
    log.Printf("Packet: %+v", data)
})
```

### よくある問題

1. **接続タイムアウト**
   - サーバーが起動しているか確認
   - ファイアウォール設定を確認

2. **アサーションタイムアウト**
   - context のタイムアウト時間を延長
   - イベントが正しく発行されているか確認

3. **イベントが受信されない**
   - パケットハンドラーが登録されているか確認
   - イベントリスナーが正しく設定されているか確認

## パフォーマンステスト

```bash
# ベンチマークテスト
go test ./pkg/events/... -bench=. -benchmem

# プロファイリング
go test ./pkg/assertions/... -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

## 次のステップ

1. Phase 3の残りのアサーションを実装
2. E2Eテストスイートの作成
3. パフォーマンステストの追加
4. CI/CDパイプラインの構築
