# 調査: セキュリティ考慮事項

- カテゴリ: security-considerations
- 作成日: 2026-04-30
- 紐づく達成条件: C3

## 1. 調査の目的

APIキーをどこに・どのように保存するかを決定し、チーム内共有時の安全な認証フローを設計する。特に「共有キーを使わない前提でのチーム展開方法」を明確にする。

## 2. 調査方法・ソース

- 使用ツール: Backlog公式ドキュメント、OSS実装パターン調査
- 主要ソース: Backlog API docs, zalando/go-keyring, gh CLI設計

## 3. 発見・要点

### 3.1 APIキーの安全な保存方法

| 保存方法 | セキュリティ | CI/CD対応 | 推奨 |
|---|---|---|---|
| OS Keychain | 暗号化 + Touch ID | 手動 | ローカル開発向け |
| 環境変数 | セッション内のみ | 対応 | CI/CD向け |
| 設定ファイル平文 | ファイル盗難リスク | 対応可 | 最終手段のみ |
| コマンドライン引数 | シェル履歴に残る | 不可 | 禁止 |

**Go実装**: `zalando/go-keyring`（macOS Keychain / Linux Secret Service / Windows Credential Manager をクロスプラットフォームで統一）

### 3.2 チーム共有キーを避けるべき理由

- 責任追跡が不可能（誰が何をしたか特定できない）
- メンバー離脱時にチーム全体でキーを再発行する必要がある
- Backlog APIキーはユーザーアカウントに紐づくため設計上も個人キーが前提

**推奨**: 各メンバーが自身のBacklogアカウントでAPIキーを生成し、`bk auth login` で登録

### 3.3 設定ファイルパーミッション

- 設定ファイル（非機密）: `~/.config/backlog/config.yaml` → パーミッション `644`
- 認証情報（機密）: Keychain優先。フォールバックでファイル保存する場合は `600`
- CI環境では `BACKLOG_API_KEY` 環境変数を使用（ファイル不要）

### 3.4 Backlog APIキーの特性

- **有効期限**: なし（手動取消まで有効）
- **スコープ**: 細粒度スコープなし（APIアクセスのon/offのみ）
- **生成場所**: Backlog WebUI → 個人設定 → API → 追加
- **重要**: 生成後1回しか表示されない → 即座に安全な場所にコピー

### 3.5 OAuth 2.0 対応（将来検討）

- Backlog は OAuth 2.0 (Authorization Code Flow) にも対応
- `access_token` 有効期限: 1時間（`refresh_token` で自動更新）
- APIキーよりセキュリティは高いが、実装コストも高い
- 初版はAPIキー認証のみでOK、v2でOAuth対応を検討

## 4. 達成条件への影響

| 達成条件 | 影響 | 補足 |
|---------|------|------|
| C3 (4ステップ以内) | 前提条件が明確になる | `bk auth login` 1コマンドで完了する設計にすれば達成可能 |

## 5. 推奨アクション

- Keychain統合に `zalando/go-keyring` を採用
- `BACKLOG_API_KEY` 環境変数をフォールバックとして必ず対応（CI/CD用）
- ドキュメントに「チーム展開手順」（各自がAPIキーを取得して `bk auth login` を実行）を明記
- OAuth 2.0 対応は v2 以降に先送り

## 6. 未解決事項

- Linux環境でKeychain（Secret Service）が使えない場合のフォールバック（ファイル保存のみにするか）
- `bk auth login` で入力したAPIキーが正しいか検証するエンドポイントの確認

## 7. 参考資料

- https://developer.nulab.com/ja/docs/backlog/api/1/get-space/（認証確認用エンドポイント候補）
- https://github.com/zalando/go-keyring
- https://backlog.com/ja/help/usersguide/personal-settings/userguide188/（APIキー生成方法）
