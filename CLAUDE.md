## プロジェクト概要
統合版マインクラフトのサーバ用テストライブラリ

## 機能
- サーバー参加
- AIエージェント
- プレイヤー偽装
- 自然言語によるシナリオ記載でAIエージェントが動作し、アサーションまで行ってくれる

## アサーション一覧

### 基本アサーション (実装済み)
- 接続状態アサート (`ToBeConnected`, `ToBeDisconnected`)
- コマンド実行アサート (`Command().ToSucceed`, `Command().ToFail`, `Command().ToContain`)
- Form表示アサート (`Form().ToReceive`, `Form().ToReceiveWithTitle`, `Form().ToBeModal`, `Form().ToBeActionForm`, `Form().ToBeCustomForm`, `Form().ToHaveTitle`, `Form().ToContainTitle`, `Form().ToHaveButton`, `Form().ToHaveButtons`, `Form().ToHaveContent`, Modal/Action/CustomForm対応)
- 座標アサート (`Position().ToBe`, `Position().ToBeNear`, `Position().ToReach`)
- チャット表示アサート (`Chat().ToReceive`, `Chat().NotToReceive`, `Chat().ToReceiveInOrder`)

### プレイヤー状態系アサーション
- インベントリアサート (`Inventory().ToHaveItem`, `Inventory().ToHaveItemCount`, `Inventory().ToBeEmpty`)
- 体力アサート (`Health().ToBe`, `Health().ToBeAbove`, `Health().ToBeBelow`, `Health().ToBeFull`)
- 満腹度アサート (`Hunger().ToBe`, `Hunger().ToBeAbove`, `Hunger().ToBeFull`)
- エフェクトアサート (`Effect().ToHave`, `Effect().NotToHave`, `Effect().ToHaveLevel`)
- ゲームモードアサート (`Gamemode().ToBe`, `Gamemode().ToBeSurvival`, `Gamemode().ToBeCreative`)
- 権限レベルアサート (`Permission().ToBeOperator`, `Permission().ToHaveLevel`, `Permission().ToBeAtLeast`)

### ワールド/ブロック系アサーション
- ブロックアサート (`Block().ToBe`, `Block().ToBeAt`, `Block().ToBeAir`)
- エンティティアサート (`Entity().ToExist`, `Entity().ToBeNearby`, `Entity().ToHaveCount`)
- スコアボードアサート (`Scoreboard().ToHaveValue`, `Scoreboard().ToHaveObjective`, `Scoreboard().ToHaveScore`, `Scoreboard().ToHaveScoreAbove`, `Scoreboard().ToHaveScoreBelow`, `Scoreboard().ToHaveScoreBetween`, `Scoreboard().ToHaveDisplaySlot`, `Scoreboard().ToHaveFakePlayerScore`, `Scoreboard().NotToHaveObjective`)
- タグアサート (`Tag().ToHave`, `Tag().NotToHave`)

### UI/表示系アサーション
- タイトル表示アサート (`Title().ToReceive`, `Title().ToContain`)
- サブタイトル表示アサート (`Subtitle().ToReceive`, `Subtitle().ToContain`)
- アクションバー表示アサート (`Actionbar().ToReceive`, `Actionbar().ToContain`)
- サウンド再生アサート (`Sound().ToPlay`, `Sound().NotToPlay`)
- パーティクルアサート (`Particle().ToSpawn`)

### イベント系アサーション
- キック/BANアサート (`Connection().ToBeKicked`, `Connection().ToBeBanned`)
- テレポートアサート (`Teleport().ToOccur`, `Teleport().ToDestination`)
- ディメンション移動アサート (`Dimension().ToChangeTo`)

### タイミング系アサーション
- タイムアウトアサート (`Timing().ToCompleteWithin`, `Timing().ToTimeout`)
- シーケンスアサート (`Sequence().ToOccurInOrder`)
- 条件待機アサート (`Condition().ToBeMetWithin`)

### 汎用アサーション (実装済み)
- 真偽値アサート (`IsTrue`, `IsFalse`)
- 等価性アサート (`Equal`, `NotEqual`)
- Nilアサート (`IsNil`, `NotNil`)
- 数値比較アサート (`GreaterThan`, `LessThan`, `GreaterThanOrEqual`, `LessThanOrEqual`, `InRange`)
- 文字列アサート (`Contains`, `NotContains`, `HasPrefix`, `HasSuffix`, `IsEmpty`, `NotEmpty`)
- コレクションアサート (`LengthEqual`, `IsEmptyCollection`, `NotEmptyCollection`, `ContainsElement`)
