package main

import (
	"os"
	"time"

	"github.com/gollilla/best"
	"github.com/gollilla/best/pkg/assertions"
)

func main() {
	// テストランナーを作成（デフォルトでConsoleReporterを使用し、自動的に結果を出力）
	best.NewRunner(nil)

	// エージェント（テスト間で共有）
	var agent *best.Agent

	// グローバルフック: すべてのテスト前に実行
	best.BeforeAll(func(ctx *best.TestContext) {
		agent = best.CreateAgent("BestTestBot")
		if err := agent.Connect(); err != nil {
			panic(err)
		}
		time.Sleep(2 * time.Second)
	})

	// グローバルフック: すべてのテスト後に実行
	best.AfterAll(func(ctx *best.TestContext) {
		if agent != nil {
			// Disconnect() internally waits for server-side cleanup
			agent.Disconnect()
		}
	})

	// ==========================================
	// テストスイート1: 接続とプレイヤー状態
	// ==========================================
	best.Describe("接続とプレイヤー状態", func() {
		best.It("サーバーに接続されているべき", func(ctx *best.TestContext) {
			assertions.IsTrue(agent.IsConnected(), "プレイヤーが接続されているべき")
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

			agent.Expect().Subtitle().ToReceive("Welcome to the server", 3*time.Second)
		})

		best.It("アクションバーメッセージを受信するべき", func(ctx *best.TestContext) {
			go func() {
				time.Sleep(500 * time.Millisecond)
				agent.Command("/title @s actionbar Loading...")
			}()

			agent.Expect().Actionbar().ToReceive("Loading...", 3*time.Second)
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
			// 体力は0以上、最大値以下であるべき（サバイバルモードの最大体力は20）
			assertions.InRange(float64(health), 0, 20, "体力は0から20の範囲内であるべき")
		})

		best.It("座標は有効な値であるべき", func(ctx *best.TestContext) {
			pos := agent.Position()
			// NaNやInfinityでないことを確認
			isNaN := pos.X != pos.X || pos.Y != pos.Y || pos.Z != pos.Z
			assertions.IsFalse(isNaN, "座標がNaNであってはならない")
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
	// テストスイート6: マルチエージェント（2体のエージェント）
	// ==========================================
	best.Describe("マルチエージェント連携", func() {
		var agent1 *best.Agent
		var agent2 *best.Agent

		// スイート開始時に両方のエージェントを1回だけ接続
		best.BeforeAll(func(ctx *best.TestContext) {
			agent1 = best.CreateAgent("BestTestBot")
			if err := agent1.Connect(); err != nil {
				panic(err)
			}
			time.Sleep(2 * time.Second)

			agent2 = best.CreateAgent("BestTestBot2")
			if err := agent2.Connect(); err != nil {
				panic(err)
			}
			time.Sleep(2 * time.Second)
		})

		// スイート終了時に両方のエージェントを切断
		best.AfterAll(func(ctx *best.TestContext) {
			if agent2 != nil {
				agent2.Disconnect()
			}
			if agent1 != nil {
				agent1.Disconnect()
			}
		})

		best.It("2体のエージェントが同時に接続されているべき", func(ctx *best.TestContext) {
			assertions.IsTrue(agent1.IsConnected(), "エージェント1が接続されているべき")
			assertions.IsTrue(agent2.IsConnected(), "エージェント2が接続されているべき")
		})

		best.It("エージェント1からエージェント2にチャットメッセージを送信できるべき", func(ctx *best.TestContext) {
			// エージェント2がチャットメッセージを受信することを期待
			go func() {
				time.Sleep(500 * time.Millisecond)
				agent1.Chat("Hello from Agent 1!")
			}()

			agent2.Expect().Chat().ToReceive("Hello from Agent 1!", 3*time.Second, nil)
		})

		best.It("エージェント2からエージェント1にチャットメッセージを送信できるべき", func(ctx *best.TestContext) {
			// エージェント1がチャットメッセージを受信することを期待
			go func() {
				time.Sleep(500 * time.Millisecond)
				agent2.Chat("Hello from Agent 2!")
			}()

			agent1.Expect().Chat().ToReceive("Hello from Agent 2!", 3*time.Second, nil)
		})

		best.It("片方のエージェントがコマンドを実行した結果を、もう片方が確認できるべき", func(ctx *best.TestContext) {
			// エージェント1がグローバルなメッセージを送信
			go func() {
				time.Sleep(500 * time.Millisecond)
				agent1.Command("/say Test message from Agent 1")
			}()

			// エージェント2がそのメッセージを受信することを期待
			agent2.Expect().Chat().ToReceive("Test message from Agent 1", 3*time.Second, nil)
		})

		best.It("両方のエージェントがコマンドを実行できるべき", func(ctx *best.TestContext) {
			// サーバーが同時に同じコマンドを実行すると片方にしか応答しないため、
			// 異なるコマンドを順番に実行する
			agent1.Expect().Command("/help").ToSucceed()
			agent2.Expect().Command("/list").ToSucceed()
		})

		best.It("エージェント1がエージェント2に対してコマンドを実行できるべき", func(ctx *best.TestContext) {
			// エージェント1からエージェント2に対する/giveコマンドを実行
			agent2.Expect().Command("/clear").ToSucceed()
			time.Sleep(1 * time.Second) // インベントリ更新を待つ
			agent2.Expect().Inventory().ToBeEmpty()

			agent1.Expect().Command("/give BestTestBot2 diamond 5").ToSucceed()
			time.Sleep(1 * time.Second) // インベントリ更新を待つ
			// アイテム名（"diamond", "minecraft:diamond"）、NetworkID（"335"）、完全なID（"item:335"）のいずれでもOK
			agent2.Expect().Inventory().ToHaveItemCount("diamond", 5)
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
