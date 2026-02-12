package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gollilla/best/pkg/agent"
	"github.com/gollilla/best/pkg/types"
)

// registerBuiltinActions registers all builtin actions
func registerBuiltinActions(r *Registry) {
	// connect - Connect to server
	r.RegisterAction("connect", ActionDefinition{
		Description: "サーバーに接続する",
		Parameters:  []ParameterDef{},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		return a.Connect()
	})

	// disconnect - Disconnect from server
	r.RegisterAction("disconnect", ActionDefinition{
		Description: "サーバーから切断する",
		Parameters:  []ParameterDef{},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		return a.Disconnect()
	})

	// command - Execute a command
	r.RegisterAction("command", ActionDefinition{
		Description: "コマンドを実行する",
		Parameters: []ParameterDef{
			{Name: "cmd", Type: "string", Required: true, Description: "実行するコマンド（例: /help）"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		cmd, ok := params["cmd"].(string)
		if !ok {
			return fmt.Errorf("cmd parameter is required and must be a string")
		}
		return a.Command(cmd)
	})

	// chat - Send a chat message
	r.RegisterAction("chat", ActionDefinition{
		Description: "チャットメッセージを送信する",
		Parameters: []ParameterDef{
			{Name: "message", Type: "string", Required: true, Description: "送信するメッセージ"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		message, ok := params["message"].(string)
		if !ok {
			return fmt.Errorf("message parameter is required and must be a string")
		}
		return a.Chat(message)
	})

	// wait - Wait for a specified duration
	r.RegisterAction("wait", ActionDefinition{
		Description: "指定時間待機する",
		Parameters: []ParameterDef{
			{Name: "duration", Type: "duration", Required: true, Description: "待機時間（例: 2s, 500ms, 1m）"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		durationStr, ok := params["duration"].(string)
		if !ok {
			return fmt.Errorf("duration parameter is required and must be a string")
		}
		d, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("invalid duration format: %v", err)
		}
		select {
		case <-time.After(d):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	// goto - Teleport to a position
	r.RegisterAction("goto", ActionDefinition{
		Description: "指定座標にテレポートする",
		Parameters: []ParameterDef{
			{Name: "x", Type: "number", Required: true, Description: "X座標"},
			{Name: "y", Type: "number", Required: true, Description: "Y座標"},
			{Name: "z", Type: "number", Required: true, Description: "Z座標"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		x, _ := getFloat(params, "x")
		y, _ := getFloat(params, "y")
		z, _ := getFloat(params, "z")
		return a.Goto(types.Position{X: x, Y: y, Z: z})
	})

	// move_relative - Move relative to current position
	r.RegisterAction("move_relative", ActionDefinition{
		Description: "現在位置から相対的に移動する（テレポート）",
		Parameters: []ParameterDef{
			{Name: "dx", Type: "number", Required: false, Description: "X方向の移動量", Default: "0"},
			{Name: "dy", Type: "number", Required: false, Description: "Y方向の移動量", Default: "0"},
			{Name: "dz", Type: "number", Required: false, Description: "Z方向の移動量", Default: "0"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		current := a.Position()
		// Save current position for assertion
		r.SetLastPosition(current)

		dx, _ := getFloat(params, "dx")
		dy, _ := getFloat(params, "dy")
		dz, _ := getFloat(params, "dz")
		newPos := types.Position{
			X: current.X + dx,
			Y: current.Y + dy,
			Z: current.Z + dz,
		}

		fmt.Printf("        [move_relative] 移動前: (%.2f, %.2f, %.2f) → 移動後: (%.2f, %.2f, %.2f)\n",
			current.X, current.Y, current.Z, newPos.X, newPos.Y, newPos.Z)

		return a.Goto(newPos)
	})

	// look_at - Look at a position
	r.RegisterAction("look_at", ActionDefinition{
		Description: "指定座標を向く",
		Parameters: []ParameterDef{
			{Name: "x", Type: "number", Required: true, Description: "X座標"},
			{Name: "y", Type: "number", Required: true, Description: "Y座標"},
			{Name: "z", Type: "number", Required: true, Description: "Z座標"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		x, _ := getFloat(params, "x")
		y, _ := getFloat(params, "y")
		z, _ := getFloat(params, "z")
		return a.LookAt(types.Position{X: x, Y: y, Z: z})
	})

	// wait_for_spawn - Wait for player to spawn
	r.RegisterAction("wait_for_spawn", ActionDefinition{
		Description: "プレイヤーのスポーン完了まで待機する",
		Parameters: []ParameterDef{
			{Name: "timeout", Type: "number", Required: false, Description: "タイムアウト秒数", Default: "30"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		timeout := 30.0
		if t, ok := getFloat(params, "timeout"); ok {
			timeout = t
		}
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
		return a.WaitForSpawn(timeoutCtx)
	})

	// submit_form - Submit a form response
	r.RegisterAction("submit_form", ActionDefinition{
		Description: "フォームに回答を送信する",
		Parameters: []ParameterDef{
			{Name: "button_index", Type: "number", Required: false, Description: "選択するボタンのインデックス（0から）"},
			{Name: "button_text", Type: "string", Required: false, Description: "選択するボタンのテキスト"},
			{Name: "modal_response", Type: "boolean", Required: false, Description: "ModalFormの場合: true=Button1, false=Button2"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		form := a.GetLastForm()
		if form == nil {
			return fmt.Errorf("受信したフォームがありません")
		}

		var response types.FormResponse

		switch f := form.(type) {
		case *types.ModalForm:
			// ModalForm expects boolean response
			if modalResp, ok := params["modal_response"].(bool); ok {
				response = modalResp
			} else if idx, ok := getFloat(params, "button_index"); ok {
				response = idx == 0 // 0 = Button1 (true), 1 = Button2 (false)
			} else {
				response = true // Default: Button1
			}

		case *types.ActionForm:
			// ActionForm expects button index
			if idx, ok := getFloat(params, "button_index"); ok {
				response = int(idx)
			} else if buttonText, ok := params["button_text"].(string); ok {
				found := false
				for i, btn := range f.Buttons {
					if btn.Text == buttonText || strings.Contains(btn.Text, buttonText) {
						response = i
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("ボタン '%s' が見つかりません", buttonText)
				}
			} else {
				response = 0 // Default: first button
			}

		case *types.CustomForm:
			// CustomForm expects array of values - for now just return empty array
			response = []interface{}{}

		default:
			return fmt.Errorf("不明なフォームタイプです")
		}

		return a.SubmitForm(form.GetID(), response)
	})

	// close_form - Close a form without submitting
	r.RegisterAction("close_form", ActionDefinition{
		Description: "フォームを閉じる（キャンセル）",
		Parameters:  []ParameterDef{},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		form := a.GetLastForm()
		if form == nil {
			return fmt.Errorf("受信したフォームがありません")
		}
		// Send null to cancel
		return a.SubmitForm(form.GetID(), nil)
	})
}

// registerBuiltinAssertions registers all builtin assertions
func registerBuiltinAssertions(r *Registry) {
	// assert_connected - Assert that the agent is connected
	r.RegisterAssertion("assert_connected", AssertionDefinition{
		Description: "サーバーに接続されていることを確認する",
		Parameters:  []ParameterDef{},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		if !a.IsConnected() {
			return fmt.Errorf("プレイヤーが接続されていません")
		}
		return nil
	})

	// assert_disconnected - Assert that the agent is disconnected
	r.RegisterAssertion("assert_disconnected", AssertionDefinition{
		Description: "サーバーから切断されていることを確認する",
		Parameters:  []ParameterDef{},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		if a.IsConnected() {
			return fmt.Errorf("プレイヤーがまだ接続されています")
		}
		return nil
	})

	// assert_chat - Assert that a chat message is received
	r.RegisterAssertion("assert_chat", AssertionDefinition{
		Description: "指定パターンのチャットメッセージを受信することを確認する",
		Parameters: []ParameterDef{
			{Name: "pattern", Type: "string", Required: true, Description: "期待するパターン（部分一致）"},
			{Name: "timeout", Type: "number", Required: false, Description: "タイムアウト秒数", Default: "5"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		pattern, ok := params["pattern"].(string)
		if !ok {
			return fmt.Errorf("pattern parameter is required and must be a string")
		}

		timeout := 5.0
		if t, ok := getFloat(params, "timeout"); ok {
			timeout = t
		}

		timeoutDuration := time.Duration(timeout) * time.Second
		a.Expect().Chat().ToReceive(pattern, timeoutDuration, nil)
		return nil
	})

	// assert_health_above - Assert that health is above a value
	r.RegisterAssertion("assert_health_above", AssertionDefinition{
		Description: "体力が指定値より大きいことを確認する",
		Parameters: []ParameterDef{
			{Name: "value", Type: "number", Required: true, Description: "期待する最小体力値"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		value, ok := getFloat(params, "value")
		if !ok {
			return fmt.Errorf("value parameter is required and must be a number")
		}
		a.Expect().Health().ToBeAbove(float32(value))
		return nil
	})

	// assert_health_below - Assert that health is below a value
	r.RegisterAssertion("assert_health_below", AssertionDefinition{
		Description: "体力が指定値より小さいことを確認する",
		Parameters: []ParameterDef{
			{Name: "value", Type: "number", Required: true, Description: "期待する最大体力値"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		value, ok := getFloat(params, "value")
		if !ok {
			return fmt.Errorf("value parameter is required and must be a number")
		}
		a.Expect().Health().ToBeBelow(float32(value))
		return nil
	})

	// assert_gamemode - Assert the current gamemode
	r.RegisterAssertion("assert_gamemode", AssertionDefinition{
		Description: "ゲームモードを確認する",
		Parameters: []ParameterDef{
			{Name: "mode", Type: "string", Required: true, Description: "期待するゲームモード（survival, creative, adventure, spectator）"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		mode, ok := params["mode"].(string)
		if !ok {
			return fmt.Errorf("mode parameter is required and must be a string")
		}

		switch mode {
		case "survival":
			a.Expect().Gamemode().ToBeSurvival()
		case "creative":
			a.Expect().Gamemode().ToBeCreative()
		case "adventure":
			a.Expect().Gamemode().ToBeAdventure()
		case "spectator":
			a.Expect().Gamemode().ToBeSpectator()
		default:
			return fmt.Errorf("unknown gamemode: %s", mode)
		}
		return nil
	})

	// assert_position_near - Assert that the player is near a position
	r.RegisterAssertion("assert_position_near", AssertionDefinition{
		Description: "プレイヤーが指定座標の近くにいることを確認する",
		Parameters: []ParameterDef{
			{Name: "x", Type: "number", Required: true, Description: "X座標"},
			{Name: "y", Type: "number", Required: true, Description: "Y座標"},
			{Name: "z", Type: "number", Required: true, Description: "Z座標"},
			{Name: "distance", Type: "number", Required: false, Description: "許容距離", Default: "5"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		x, _ := getFloat(params, "x")
		y, _ := getFloat(params, "y")
		z, _ := getFloat(params, "z")
		distance := 5.0
		if d, ok := getFloat(params, "distance"); ok {
			distance = d
		}

		pos := types.Position{X: x, Y: y, Z: z}
		a.Expect().Position().ToBeNear(pos, distance)
		return nil
	})

	// assert_inventory_has_item - Assert that inventory contains an item
	r.RegisterAssertion("assert_inventory_has_item", AssertionDefinition{
		Description: "インベントリに指定アイテムがあることを確認する",
		Parameters: []ParameterDef{
			{Name: "item", Type: "string", Required: true, Description: "アイテム名またはID"},
			{Name: "count", Type: "number", Required: false, Description: "期待する個数"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		item, ok := params["item"].(string)
		if !ok {
			return fmt.Errorf("item parameter is required and must be a string")
		}

		if count, ok := getFloat(params, "count"); ok {
			a.Expect().Inventory().ToHaveItemCount(item, int32(count))
		} else {
			a.Expect().Inventory().ToHaveItem(item)
		}
		return nil
	})

	// assert_inventory_empty - Assert that inventory is empty
	r.RegisterAssertion("assert_inventory_empty", AssertionDefinition{
		Description: "インベントリが空であることを確認する",
		Parameters:  []ParameterDef{},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		a.Expect().Inventory().ToBeEmpty()
		return nil
	})

	// assert_title - Assert that a title is received
	r.RegisterAssertion("assert_title", AssertionDefinition{
		Description: "タイトルが表示されることを確認する",
		Parameters: []ParameterDef{
			{Name: "text", Type: "string", Required: true, Description: "期待するタイトルテキスト"},
			{Name: "timeout", Type: "number", Required: false, Description: "タイムアウト秒数", Default: "5"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		text, ok := params["text"].(string)
		if !ok {
			return fmt.Errorf("text parameter is required and must be a string")
		}

		timeout := 5.0
		if t, ok := getFloat(params, "timeout"); ok {
			timeout = t
		}

		timeoutDuration := time.Duration(timeout) * time.Second
		a.Expect().Title().ToReceive(text, timeoutDuration)
		return nil
	})

	// assert_effect_has - Assert that player has an effect
	r.RegisterAssertion("assert_effect_has", AssertionDefinition{
		Description: "プレイヤーが指定エフェクトを持っていることを確認する",
		Parameters: []ParameterDef{
			{Name: "effect", Type: "string", Required: true, Description: "エフェクト名"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		effect, ok := params["effect"].(string)
		if !ok {
			return fmt.Errorf("effect parameter is required and must be a string")
		}
		a.Expect().Effect().ToHave(effect)
		return nil
	})

	// assert_permission_operator - Assert that player is an operator
	// Checks by running /op command and verifying usage info is returned (not permission denied)
	r.RegisterAssertion("assert_permission_operator", AssertionDefinition{
		Description: "プレイヤーがオペレーター権限を持っていることを確認する",
		Parameters:  []ParameterDef{},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		// Run /op command to check if player has operator permission
		a.Command("/op")
		// If we get usage info or any response (not permission denied), we're an operator
		a.Expect().CommandOutput().ToReceiveAny(3 * time.Second)
		return nil
	})

	// assert_tag_has - Assert that player has a tag
	r.RegisterAssertion("assert_tag_has", AssertionDefinition{
		Description: "プレイヤーが指定タグを持っていることを確認する",
		Parameters: []ParameterDef{
			{Name: "tag", Type: "string", Required: true, Description: "タグ名"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		tag, ok := params["tag"].(string)
		if !ok {
			return fmt.Errorf("tag parameter is required and must be a string")
		}
		a.Expect().Tag().ToHave(tag)
		return nil
	})

	// assert_form_received - Assert that a form is received
	r.RegisterAssertion("assert_form_received", AssertionDefinition{
		Description: "フォームを受信することを確認する",
		Parameters: []ParameterDef{
			{Name: "timeout", Type: "number", Required: false, Description: "タイムアウト秒数", Default: "5"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		timeout := 5.0
		if t, ok := getFloat(params, "timeout"); ok {
			timeout = t
		}

		timeoutDuration := time.Duration(timeout) * time.Second
		a.Expect().Form().ToReceive(timeoutDuration)
		return nil
	})

	// assert_position_moved - Assert that player moved by expected amount from last position
	r.RegisterAssertion("assert_position_moved", AssertionDefinition{
		Description: "move_relative実行前の位置から指定量移動したことを確認する",
		Parameters: []ParameterDef{
			{Name: "dx", Type: "number", Required: false, Description: "期待するX方向の移動量", Default: "0"},
			{Name: "dy", Type: "number", Required: false, Description: "期待するY方向の移動量", Default: "0"},
			{Name: "dz", Type: "number", Required: false, Description: "期待するZ方向の移動量", Default: "0"},
			{Name: "tolerance", Type: "number", Required: false, Description: "許容誤差", Default: "1"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		lastPos := r.GetLastPosition()
		if lastPos == nil {
			return fmt.Errorf("前回の位置が記録されていません（先にmove_relativeを実行してください）")
		}

		dx, _ := getFloat(params, "dx")
		dy, _ := getFloat(params, "dy")
		dz, _ := getFloat(params, "dz")
		tolerance := 1.0
		if t, ok := getFloat(params, "tolerance"); ok {
			tolerance = t
		}

		expectedPos := types.Position{
			X: lastPos.X + dx,
			Y: lastPos.Y + dy,
			Z: lastPos.Z + dz,
		}

		a.Expect().Position().ToBeNear(expectedPos, tolerance)
		return nil
	})

	// assert_hunger_above - Assert that hunger is above a value
	r.RegisterAssertion("assert_hunger_above", AssertionDefinition{
		Description: "満腹度が指定値より大きいことを確認する",
		Parameters: []ParameterDef{
			{Name: "value", Type: "number", Required: true, Description: "期待する最小満腹度値"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		value, ok := getFloat(params, "value")
		if !ok {
			return fmt.Errorf("value parameter is required and must be a number")
		}
		hunger := a.GetHunger()
		if hunger <= float32(value) {
			return fmt.Errorf("満腹度が %v 以下です（実際: %v）", value, hunger)
		}
		return nil
	})

	// assert_hunger_below - Assert that hunger is below a value
	r.RegisterAssertion("assert_hunger_below", AssertionDefinition{
		Description: "満腹度が指定値より小さいことを確認する",
		Parameters: []ParameterDef{
			{Name: "value", Type: "number", Required: true, Description: "期待する最大満腹度値"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		value, ok := getFloat(params, "value")
		if !ok {
			return fmt.Errorf("value parameter is required and must be a number")
		}
		hunger := a.GetHunger()
		if hunger >= float32(value) {
			return fmt.Errorf("満腹度が %v 以上です（実際: %v）", value, hunger)
		}
		return nil
	})

	// assert_scoreboard - Assert scoreboard value
	r.RegisterAssertion("assert_scoreboard", AssertionDefinition{
		Description: "スコアボードの値を確認する",
		Parameters: []ParameterDef{
			{Name: "objective", Type: "string", Required: true, Description: "オブジェクティブ名"},
			{Name: "value", Type: "number", Required: true, Description: "期待する値"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		objective, ok := params["objective"].(string)
		if !ok {
			return fmt.Errorf("objective parameter is required and must be a string")
		}
		expected, ok := getFloat(params, "value")
		if !ok {
			return fmt.Errorf("value parameter is required and must be a number")
		}

		actual, found := a.GetScore(objective)
		if !found {
			return fmt.Errorf("スコアボード '%s' が見つかりません", objective)
		}
		if int32(expected) != actual {
			return fmt.Errorf("スコアボード '%s' の値が一致しません（期待: %v, 実際: %v）", objective, int32(expected), actual)
		}
		return nil
	})

	// assert_entity_nearby - Assert that an entity exists nearby
	r.RegisterAssertion("assert_entity_nearby", AssertionDefinition{
		Description: "指定タイプのエンティティが近くにいることを確認する",
		Parameters: []ParameterDef{
			{Name: "type", Type: "string", Required: true, Description: "エンティティタイプ（例: minecraft:zombie）"},
			{Name: "distance", Type: "number", Required: false, Description: "検索距離", Default: "10"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		entityType, ok := params["type"].(string)
		if !ok {
			return fmt.Errorf("type parameter is required and must be a string")
		}
		distance := 10.0
		if d, ok := getFloat(params, "distance"); ok {
			distance = d
		}

		playerPos := a.Position()
		entities := a.GetEntities()
		for _, e := range entities {
			if e.Type == entityType {
				dx := e.Position.X - playerPos.X
				dy := e.Position.Y - playerPos.Y
				dz := e.Position.Z - playerPos.Z
				dist := dx*dx + dy*dy + dz*dz
				if dist <= distance*distance {
					return nil
				}
			}
		}
		return fmt.Errorf("エンティティ '%s' が距離 %v 以内に見つかりません", entityType, distance)
	})

	// assert_form_title - Assert form has specific title
	r.RegisterAssertion("assert_form_title", AssertionDefinition{
		Description: "フォームのタイトルを確認する",
		Parameters: []ParameterDef{
			{Name: "title", Type: "string", Required: true, Description: "期待するタイトル（部分一致）"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		title, ok := params["title"].(string)
		if !ok {
			return fmt.Errorf("title parameter is required and must be a string")
		}
		form := a.GetLastForm()
		if form == nil {
			return fmt.Errorf("受信したフォームがありません")
		}
		formTitle := form.GetTitle()
		if formTitle != title && !strings.Contains(formTitle, title) {
			return fmt.Errorf("フォームタイトルが一致しません（期待: %s, 実際: %s）", title, formTitle)
		}
		return nil
	})

	// assert_form_has_button - Assert form has a button with specific text
	r.RegisterAssertion("assert_form_has_button", AssertionDefinition{
		Description: "フォームに指定テキストのボタンがあることを確認する",
		Parameters: []ParameterDef{
			{Name: "text", Type: "string", Required: true, Description: "ボタンテキスト（部分一致）"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		text, ok := params["text"].(string)
		if !ok {
			return fmt.Errorf("text parameter is required and must be a string")
		}
		form := a.GetLastForm()
		if form == nil {
			return fmt.Errorf("受信したフォームがありません")
		}

		// Check buttons based on form type
		switch f := form.(type) {
		case *types.ActionForm:
			for _, btn := range f.Buttons {
				if btn.Text == text || strings.Contains(btn.Text, text) {
					return nil
				}
			}
		case *types.ModalForm:
			if f.Button1 == text || strings.Contains(f.Button1, text) ||
				f.Button2 == text || strings.Contains(f.Button2, text) {
				return nil
			}
		default:
			return fmt.Errorf("このフォームタイプにはボタンがありません")
		}
		return fmt.Errorf("ボタン '%s' がフォームに見つかりません", text)
	})

	// assert_permission_level - Assert player has specific permission level
	r.RegisterAssertion("assert_permission_level", AssertionDefinition{
		Description: "プレイヤーの権限レベルを確認する",
		Parameters: []ParameterDef{
			{Name: "level", Type: "number", Required: true, Description: "期待する権限レベル（0-4）"},
		},
	}, func(ctx context.Context, a *agent.Agent, params map[string]interface{}) error {
		expected, ok := getFloat(params, "level")
		if !ok {
			return fmt.Errorf("level parameter is required and must be a number")
		}
		actual := a.GetPermissionLevel()
		if int32(expected) != actual {
			return fmt.Errorf("権限レベルが一致しません（期待: %v, 実際: %v）", int32(expected), actual)
		}
		return nil
	})
}

// getFloat extracts a float64 from params, handling both float64 and int types
func getFloat(params map[string]interface{}, key string) (float64, bool) {
	val, ok := params[key]
	if !ok {
		return 0, false
	}

	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}

