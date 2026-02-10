## プロジェクト概要
統合版マインクラフトのサーバ用テストライブラリ

## 機能
- サーバー参加
- AIエージェント
- プレイヤー偽装
- 自然言語によるシナリオ記載でAIエージェントが動作し、アサーションまで行ってくれる

## アサーション一覧

### 基本アサーション (実装済み)
- 接続状態アサート (`toBeConnected`, `toBeDisconnected`)
- コマンド実行アサート (`command().toSucceed`, `command().toFail`, `command().toContain`)
- Form表示アサート (`form.toReceive`, Modal/Action/CustomForm対応)
- 座標アサート (`position.toBe`, `position.toBeNear`, `position.toReach`)
- チャット表示アサート (`chat.toReceive`, `chat.notToReceive`, `chat.toReceiveInOrder`)

### プレイヤー状態系アサーション
- インベントリアサート (`inventory.toHaveItem`, `inventory.toHaveItemCount`, `inventory.toBeEmpty`)
- 体力アサート (`health.toBe`, `health.toBeAbove`, `health.toBeBelow`, `health.toBeFull`)
- 満腹度アサート (`hunger.toBe`, `hunger.toBeAbove`, `hunger.toBeFull`)
- エフェクトアサート (`effect.toHave`, `effect.notToHave`, `effect.toHaveLevel`)
- ゲームモードアサート (`gamemode.toBe`, `gamemode.toBeSurvival`, `gamemode.toBeCreative`)
- 権限レベルアサート (`permission.toBeOperator`, `permission.toHaveLevel`)

### ワールド/ブロック系アサーション
- ブロックアサート (`block.toBe`, `block.toBeAt`, `block.toBeAir`)
- エンティティアサート (`entity.toExist`, `entity.toBeNearby`, `entity.toHaveCount`)
- スコアボードアサート (`scoreboard.toHaveValue`, `scoreboard.toHaveObjective`)
- タグアサート (`tag.toHave`, `tag.notToHave`)

### UI/表示系アサーション
- タイトル表示アサート (`title.toReceive`, `subtitle.toReceive`, `actionbar.toReceive`)
- サウンド再生アサート (`sound.toPlay`, `sound.notToPlay`)
- パーティクルアサート (`particle.toSpawn`)

### イベント系アサーション
- キック/BANアサート (`connection.toBeKicked`, `connection.toBeBanned`)
- テレポートアサート (`teleport.toOccur`, `teleport.toDestination`)
- ディメンション移動アサート (`dimension.toChangeTo`)

### タイミング系アサーション
- タイムアウトアサート (`timing.toCompleteWithin`, `timing.toTimeout`)
- シーケンスアサート (`sequence.toOccurInOrder`)
- 条件待機アサート (`condition.toBeMetWithin`)
