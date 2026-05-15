# 調査: 技術スタック

- カテゴリ: tech-stack
- 作成日: 2026-04-30
- 紐づく達成条件: C1, C2

## 1. 調査の目的

GoでBacklog CLIを実装する際に採用するCLIフレームワーク・Backlog SDK・補助ライブラリを確定する。選定基準は「メンテナンス継続性・サブコマンド対応・補完機能・チーム実績」。

## 2. 調査方法・ソース

- 使用ツール: GitHub検索、pkg.go.dev
- 主要ソース: GitHub ("backlog go", "go cli framework")

## 3. 発見・要点

### 3.1 Go CLIフレームワーク比較

| フレームワーク | Stars | 特徴 | 採用OSS |
|---|---|---|---|
| **spf13/cobra** | 40k+ | 業界標準、Viper連携、自動補完 | kubernetes, docker, gh |
| urfave/cli | 22k+ | 軽量・シンプル、依存ゼロ | 小規模ツール向け |
| kong | 2k+ | 構造体ベース、型安全 | マイナー |

**Backlog CLIへの推奨: cobra**
- kubernetes/docker/GitHub CLI が採用する業界標準
- サブコマンドのネスト、自動ヘルプ生成、シェル補完が揃っている
- Viper（設定管理）との組み合わせが鉄板

### 3.2 Backlog Go SDK

| SDK | 最新版 | 状態 | カバレッジ |
|---|---|---|---|
| **nattokin/go-backlog** | v0.13.0 (2026-04) | アクティブ | Issue/Wiki/Project/User/Activity |
| kenzo0107/backlog | v1.1.0 (2025-07) | アクティブ | ほぼ全REST API対応 |

**推奨: kenzo0107/backlog**（API網羅性が高い）または **nattokin/go-backlog**（最新更新）
- いずれも MIT ライセンス
- SDK が存在するため直接 HTTP クライアントを書く必要はない

### 3.3 補助ライブラリ

| 用途 | ライブラリ | 備考 |
|---|---|---|
| 設定管理 | spf13/viper | YAML/JSON/TOML/env変数を統一管理 |
| テーブル出力 | olekukonko/tablewriter | Issue一覧などの整形表示 |
| Keychain | zalando/go-keyring | macOS/Linux/Windows対応 |
| カラー出力 | fatih/color | ターミナルカラー |

## 4. 達成条件への影響

| 達成条件 | 影響 | 補足 |
|---------|------|------|
| C1 (Issue CRUD) | 達成しやすくなる | kenzo0107/backlog SDK でIssue API が網羅されている |
| C2 (PR操作) | 達成しやすくなる | SDK にPullRequest API が含まれている |

## 5. 推奨アクション

- CLIフレームワーク: **cobra** を採用
- Backlog SDK: **kenzo0107/backlog** を第一候補とし、PR API カバレッジを確認後に決定
- 設定管理: **viper** でYAML設定ファイルを管理
- 認証情報保存: **zalando/go-keyring** でKeychain統合

## 6. 未解決事項

- kenzo0107/backlog がPRの全操作（create/comment/merge）を網羅しているか要確認
- Go の最低バージョン要件（1.21 以上推奨）

## 7. 参考資料

- https://github.com/spf13/cobra
- https://github.com/kenzo0107/backlog
- https://github.com/nattokin/go-backlog
- https://github.com/spf13/viper
- https://github.com/zalando/go-keyring
