# 調査: 既存OSS実装

- カテゴリ: existing-implementations
- 作成日: 2026-04-30
- 紐づく達成条件: C1, C2, C3

## 1. 調査の目的

参考実装 simochee/backlog-cli と他のOSS実装を分析し、コマンド設計・認証方式・設定ファイル形式の先行事例を把握する。独自実装で採用・回避すべき設計判断を明確にする。

## 2. 調査方法・ソース

- 使用ツール: GitHub検索、リポジトリ構造読み取り
- 主要ソース: https://github.com/simochee/backlog-cli

## 3. 発見・要点

### 3.1 simochee/backlog-cli の概要

- **言語**: TypeScript (Bun runtime) ※Go実装ではない
- **CLIフレームワーク**: citty（最小限フレームワーク）
- **コマンド数**: 108コマンド（auth/issue/pr/project/wiki等を全網羅）

### 3.2 コマンド設計

- `auth` (login, logout, refresh, status, switch, token)
- `issue` (list, view, create, edit, delete, close, reopen, comment, status)
- `pr` (list, view, create, edit, delete, close, reopen, comment, merge, status)
- 他: project, repo, notification, wiki, user, team 等
- **`gh` (GitHub CLI) 互換の動詞設計**（list/view/create/edit/delete）

### 3.3 認証方式

- APIキー認証 + OAuth 2.0 Bearer token（アクセストークン + リフレッシュトークン）
- ホスト（`.backlog.com` / `.backlog.jp`）ごとに独立した認証設定

### 3.4 設定ファイル

- ファイル名: `.backlogrc`（ホームディレクトリ直下）
- 形式: JSON
- 複数スペース対応: `spaces: []` 配列 + `defaultSpace` フィールド
- 認証情報も同じファイルに保存

### 3.5 インストール

```bash
npm install -g @simochee/backlog-cli
```

### 3.6 参考にできる設計

- `gh` 互換のサブコマンド動詞（list/view/create/edit）を採用すれば学習コストが低い
- `--json` フラグでパイプライン対応
- 複数スペース + switch コマンドで複数環境対応
- シェル補完（Bash/Zsh/Fish）を最初から用意

### 3.7 避けるべき設計

- 認証トークンを設定ファイルに平文保存（セキュリティリスク）
- 単一スペース前提設計（後から複数対応すると破壊的変更が必要）

### 3.8 他のGo実装

- **shufo/backlog-cli**: Go + Cobra採用、GitHub CLIライク、MIT ライセンス。機能は限定的だが設計参考になる

## 4. 達成条件への影響

| 達成条件 | 影響 | 補足 |
|---------|------|------|
| C1 (Issue CRUD) | 達成しやすくなる | simocheeのコマンド設計をそのまま参考にできる |
| C2 (PR操作) | 達成しやすくなる | pr サブコマンドの設計パターンが明確 |
| C3 (4ステップ以内) | 前提が変わる | npm install ではなく go install / brew が必要。別途調査で補完 |

## 5. 推奨アクション

- コマンド動詞は `gh` / simochee 互換で設計する（`bk issue list`, `bk pr create`）
- 設定ファイルは JSONより YAML/TOMLを検討（Go CLIのデファクト）
- 認証情報はファイル平文保存ではなく Keychain 統合を優先する
- 複数スペース対応は最初から設計に組み込む

## 6. 未解決事項

- shufo/backlog-cli の実装詳細（Backlog SDK使用有無、メンテナンス状況）
- simochee/backlog-cli のOpenAPI client自動生成パターンをGoでも使えるか

## 7. 参考資料

- https://github.com/simochee/backlog-cli
- https://github.com/shufo/backlog-cli (Go実装)
