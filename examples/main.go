package main

import (
	"os"
	"time"

	"github.com/gollilla/best"
)

func main() {
	// テストランナーを作成（デフォルトでConsoleReporterを使用し、自動的に結果を出力）
	best.NewRunner(nil)

	// エージェント（テスト間で共有）
	var agent *best.Agent

	// グローバルフック: すべてのテスト前に実行
	best.BeforeAll(func(ctx *best.TestContext) {
		agent = best.NewAgent(
			best.WithHost("localhost"),
			best.WithPort(19132),
			best.WithUsername("BestTestBot"),
		)
		if err := agent.Connect(); err != nil {
			panic(err)
		}
		time.Sleep(2 * time.Second)
	})

	// グローバルフック: すべてのテスト後に実行
	best.AfterAll(func(ctx *best.TestContext) {
		if agent != nil {
			agent.Disconnect()
		}
	})

	// ==========================================
	// テストスイート1: 接続とプレイヤー状態
	// ==========================================
	best.Describe("接続とプレイヤー状態", func() {
		best.It("サーバーに接続されているべき", func(ctx *best.TestContext) {
			if !agent.IsConnected() {
				panic("プレイヤーが接続されていません")
			}
		})

		best.It("有効な座標を持つべき", func(ctx *best.TestContext) {
			pos := agent.Position()
			// 座標が取得できることを確認
			_ = pos
		})

		best.It("体力が0より大きいべき", func(ctx *best.TestContext) {
			agent.Expect().Health().ToBeAbove(0)
		})

		best.It("デフォルトでサバイバルモードであるべき", func(ctx *best.TestContext) {
			agent.Expect().Gamemode().ToBeSurvival()
		})

		best.It("最低でも権限レベル0を持つべき", func(ctx *best.TestContext) {
			agent.Expect().Permission().ToBeAtLeast(0)
		})
	})

	// ==========================================
	// テストスイート2: コマンド実行
	// ==========================================
	best.Describe("コマンド実行", func() {
		best.It("/helpコマンドを実行できるべき", func(ctx *best.TestContext) {
			// フルエントAPIでコマンド実行とアサーション
			agent.Expect().Command("/help").ToSucceed()
		})

		best.It("/helpコマンドの出力に特定の文字列が含まれるべき", func(ctx *best.TestContext) {
			// チェーンでアサーションを追加
			agent.Expect().Command("/help").ToSucceed().And().ToContain("ban")
		})

		best.It("チャットメッセージを送信できるべき", func(ctx *best.TestContext) {
			if err := agent.Chat("Hello from Best Framework!"); err != nil {
				panic(err)
			}
		})
	})

	// ==========================================
	// テストスイート3: タイトル/サブタイトル
	// ==========================================
	best.Describe("タイトル/サブタイトル表示", func() {
		best.It("タイトルメッセージを受信するべき", func(ctx *best.TestContext) {
			go func() {
				time.Sleep(500 * time.Millisecond)
				agent.Command("/title @s title Hello World")
			}()

			agent.Expect().Title().ToReceive("Hello World", 3*time.Second)
		})

		best.It("サブタイトルメッセージを受信するべき", func(ctx *best.TestContext) {
			go func() {
				time.Sleep(500 * time.Millisecond)
				agent.Command("/title @s subtitle Welcome to the server")
			}()

			agent.Expect().Title().ToReceiveSubtitle("Welcome to the server", 3*time.Second)
		})

		best.It("アクションバーメッセージを受信するべき", func(ctx *best.TestContext) {
			go func() {
				time.Sleep(500 * time.Millisecond)
				agent.Command("/title @s actionbar Loading...")
			}()

			agent.Expect().Title().ToReceiveActionbar("Loading...", 3*time.Second)
		})

		best.It("タイトルに特定のテキストが含まれるべき", func(ctx *best.TestContext) {
			go func() {
				time.Sleep(500 * time.Millisecond)
				agent.Command("/title @s title Test Message Here")
			}()

			agent.Expect().Title().ToContain("Message", 3*time.Second)
		})
	})

	// ==========================================
	// テストスイート4: スコアボード 未実装
	// ==========================================
	/**best.Describe("スコアボード操作", func() {
		best.It("スコアボード目標を作成できるべき", func(ctx *best.TestContext) {
			agent.Command("/scoreboard objectives add test_score dummy Test Score")
			agent.Expect().Scoreboard().ToHaveObjective("test_score", 3*time.Second)
		})

		best.It("スコアを設定できるべき", func(ctx *best.TestContext) {
			agent.Command("/scoreboard players set @s test_score 100")
			agent.Expect().Scoreboard().ToHaveScore("test_score", 100, 3*time.Second)
		})

		best.It("スコアを増加できるべき", func(ctx *best.TestContext) {
			agent.Command("/scoreboard players add @s test_score 50")
			agent.Expect().Scoreboard().ToHaveScoreAbove("test_score", 100, 3*time.Second)
		})

		best.It("スコアが範囲内にあるべき", func(ctx *best.TestContext) {
			agent.Expect().Scoreboard().ToHaveScoreBetween("test_score", 100, 200, 2*time.Second)
		})

		best.It("スコアボード目標を削除できるべき", func(ctx *best.TestContext) {
			agent.Command("/scoreboard objectives remove test_score")
			agent.Expect().Scoreboard().NotToHaveObjective("test_score", 3*time.Second)
		})
	})**/

	// ==========================================
	// テストスイート5: 異常系テスト（エラーハンドリングの例）
	// ==========================================
	best.Describe("異常系テスト", func() {
		best.It("無効なコマンドは失敗するべき", func(ctx *best.TestContext) {
			// フルエントAPIで失敗をアサート
			agent.Expect().Command("/this_command_does_not_exist").ToFail()
		})

		best.It("体力は有効な範囲内であるべき", func(ctx *best.TestContext) {
			health := agent.Health()
			// 体力は0以上、最大値以下であるべき
			if health < 0 {
				panic("体力が負の値です")
			}
			if health > 20 { // サバイバルモードの最大体力
				panic("体力が最大値を超えています")
			}
		})

		best.It("座標は有効な値であるべき", func(ctx *best.TestContext) {
			pos := agent.Position()
			// NaNやInfinityでないことを確認
			if pos.X != pos.X || pos.Y != pos.Y || pos.Z != pos.Z {
				panic("座標がNaNです")
			}
		})

		best.It("タイムアウトエラーを正しく処理するべき", func(ctx *best.TestContext) {
			// 意図的に短いタイムアウトを設定して、タイムアウトをテスト
			defer func() {
				if r := recover(); r != nil {
					// panicが発生したことを確認（タイムアウトによる）
					// これは期待される動作なので、テストは成功
				} else {
					// panicが発生しなかった場合は、予期しない状況
					panic("タイムアウトが発生しませんでした")
				}
			}()

			// 非常に短いタイムアウトで受信を試みる（タイムアウトが期待される）
			agent.Expect().Title().ToReceive("絶対に送信されないタイトル", 50*time.Millisecond)
		})
	})

	// ==========================================
	// スキップされるテストスイートの例
	// ==========================================
	best.SkipDescribe("高度な機能 (スキップ)", func() {
		best.It("このテストはスキップされます", func(ctx *best.TestContext) {
			// このテストは実行されません
		})
	})

	// テスト実行（ConsoleReporterが自動的に結果を出力）
	result, err := best.Run()
	if err != nil {
		panic(err)
	}

	// 失敗があれば終了コード1で終了
	if result.Failed > 0 {
		os.Exit(1)
	}
}
