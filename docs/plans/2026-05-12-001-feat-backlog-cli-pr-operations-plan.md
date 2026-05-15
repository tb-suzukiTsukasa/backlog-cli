---
title: "feat: Build Backlog CLI for PR operations"
type: feat
status: completed
date: 2026-05-12
origin: requirements_specification.md
---

# feat: Build Backlog CLI for PR operations

## Overview

新しいGoアプリケーション `backlog-cli`（コマンド名: `bk`）を構築する。ターミナルからBacklogのプルリクエスト操作（一覧・詳細表示・作成・コメント）を完結できるCLIツール。`gh` CLIと同等のUXを提供し、チームメンバーが3ステップでセットアップできる設計とする。

## Problem Frame

開発作業中にBacklogのPR確認・作成を行う際、ブラウザへ切り替えるコンテキストスイッチが生産性を低下させている。git操作と一体化したPRワークフローをターミナルから完結させることで、開発フローを中断させずにBacklogを活用できるようにする。

（see origin: requirements_specification.md § 1.1）

## Requirements Trace

- R1. PR操作（一覧・詳細・作成・コメント投稿）がCLIコマンドで完結する（C1）
- R2. カレントブランチからPRを作成できる（タイトル・本文・レビュアー指定）（C2）
- R3. チームメンバーが4ステップ以内にCLIを利用開始できる（C3）

## Scope Boundaries

- Issue操作（作成・一覧・更新・コメント）は対象外（see origin: requirements_specification.md § 1.4）
- Wiki・プロジェクト操作は対象外
- PRのマージ・クローズ・編集・削除は対象外（初版スコープ外）
- OAuth 2.0認証は対象外（v1はAPIキー認証のみ）
- GUIダッシュボード・Webインターフェースは対象外

## Context & Research

### Relevant Code and Patterns

- `research/00_existing-implementations.md`: simochee/backlog-cli のコマンド動詞設計（list/view/create/edit）を参考
- `research/01_tech-stack.md`: Cobra + kenzo0107/backlog + viper + go-keyring の採用方針
- `research/02_architecture-patterns.md`: `cmd/internal` ディレクトリ構成、`~/.config/backlog/config.yaml`（XDG準拠）、`gh` CLI互換コマンド設計
- `research/03_security-considerations.md`: go-keyring によるKeychain統合、`BACKLOG_API_KEY` 環境変数フォールバック、認証情報の設定ファイル平文保存禁止
- `research/04_distribution-channel.md`: goreleaser + GitHub Releases、`go install` をプライマリ手段

### Institutional Learnings

- なし（新規プロジェクト）

### External References

- Backlog Pull Request API: `GET /api/v2/projects/:projectIdOrKey/git/repositories/:repoIdOrName/pullRequests`（一覧）
- Backlog Pull Request API: `POST /api/v2/projects/:projectIdOrKey/git/repositories/:repoIdOrName/pullRequests`（作成）
- kenzo0107/backlog: https://github.com/kenzo0107/backlog
- cobra: https://github.com/spf13/cobra
- zalando/go-keyring: https://github.com/zalando/go-keyring
- goreleaser: https://goreleaser.com/

## Key Technical Decisions

- **コマンド名 `bk`**: 短く覚えやすい。既存コマンドとの衝突リスクは低い
- **kenzo0107/backlog を第一SDK候補**: PR APIの網羅性が高い。create/comment のカバレッジは実装時に確認し、不足なら nattokin/go-backlog に切り替える
- **Keychainファースト認証**: ローカル開発では go-keyring でKeychain統合、CI/CDでは `BACKLOG_API_KEY` 環境変数、最後の手段としてファイルフォールバック（パーミッション600）
- **git remote URL解析でプロジェクト/リポジトリを自動検出**: Backlog git URLパターン（HTTPS: `https://<space>.backlog.com/git/<project>/<repo>.git`、SSH: `git@<space>.git.backlog.com:<project>/<repo>.git`）からプロジェクトキーとリポジトリ名を抽出。`--project` / `--repo` フラグで明示オーバーライド可能
- **XDG準拠設定パス**: `$XDG_CONFIG_HOME/backlog/config.yaml` → フォールバック `~/.config/backlog/config.yaml`
- **`--json` フラグでパイプライン対応**: 全PRコマンドで機械可読出力を提供（see origin: requirements_specification.md § 4 simochee/backlog-cli参考）
- **リポジトリをパブリック公開**: `go install github.com/ts-suzuki/backlog-cli@latest` を動作させるために必要

## Open Questions

### Resolved During Planning

- **コマンド動詞設計**: `gh` 互換（list/view/create/comment）で統一
- **複数スペース対応**: `spaces:` マップ + `default_space:` + `--space` フラグ。初版から設計に組み込む（see origin: research/00_existing-implementations.md §3.7 「単一スペース前提設計は後から変更すると破壊的」）
- **Keychainフォールバック（Linux）**: go-keyring が失敗した場合は `~/.config/backlog/credentials.yaml`（パーミッション600）にフォールバックし、警告を stderr に出力
- **配布形態**: goreleaser + GitHub Releases がプライマリ。リポジトリはパブリック公開前提
- **フラグ・環境変数・設定ファイルの優先順位**: `--space` フラグ > `BACKLOG_SPACE` 環境変数 > `default_space` 設定値。`BACKLOG_API_KEY` 環境変数は「現在アクティブなスペース（--space または BACKLOG_SPACE で解決されたスペース）」のAPIキーとして適用される
- **`bk pr create` のベースブランチ指定**: `--base` フラグ（省略時はデフォルトブランチを `GET /api/v2/projects/:projectIdOrKey/git/repositories/:repoIdOrName` で取得し自動設定）
- **レビュアーのユーザー名→ID変換**: `GET /api/v2/users` でユーザー一覧を取得し、`--reviewer` に指定されたユーザー名と照合してuserIdを解決。存在しないユーザー名はエラー
- **バイナリ名**: `bk` に確定（goreleaser の `builds.binary` に明示指定）
- **APIキー入力のセキュリティ**: `bk auth login` でのAPIキー入力は `golang.org/x/term.ReadPassword` 相当で端末エコーを無効にする
- **`$EDITOR` 一時ファイルのクリーンアップ**: `bk pr comment` での一時ファイルは `defer os.Remove` で確実に削除する

### Deferred to Implementation

- **kenzo0107/backlog のPR create/comment APIカバレッジ**: SDK実装を確認後に nattokin/go-backlog への切り替えを判断
- **`bk auth login` でAPIキー検証するエンドポイント**: `GET /api/v2/space` を想定（実装時に確認）
- **Linux Secret Service 未利用環境でのフォールバックUI**: 警告メッセージの文言は実装時に決定
- **レビュアー指定の入力方式**: `--reviewer` フラグ（カンマ区切りユーザー名）を基本とし、インタラクティブ選択の要否は実装時に判断

## High-Level Technical Design

> *これは意図するアプローチを示す方向性ガイダンスであり、実装仕様ではありません。実装エージェントはコンテキストとして参照し、コードをそのまま再現しないでください。*

```
bk pr list / view / create / comment
         │
         ▼
    cmd/pr.go  (Cobra サブコマンド)
         │
         ├─ internal/config/   viper: ~/.config/backlog/config.yaml
         │       └─ SpaceConfig { host, auth_method }
         │
         ├─ internal/auth/     APIキー解決
         │       └─ BACKLOG_API_KEY env → Keychain → credentials.yaml
         │
         ├─ internal/git/      git remote URL 解析
         │       └─ project key + repo name の自動検出
         │
         ├─ internal/api/      kenzo0107/backlog SDK ラッパー
         │       └─ ListPullRequests / GetPullRequest / CreatePullRequest / AddComment
         │
         └─ internal/output/   フォーマッタ
                 └─ --json フラグで table ↔ JSON 切り替え
```

## Implementation Units

- [ ] **Unit 1: プロジェクトスキャフォールディング**

**Goal:** go.mod 初期化、ディレクトリ構成の確立、Cobra ルートコマンドの実装、基本ビルドパイプラインの骨格設定

**Requirements:** R3（シングルバイナリ配布の前提となる基盤）

**Dependencies:** なし

**Files:**
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `go.mod`
- Create: `Makefile`
- Create: `.goreleaser.yaml`（基本骨格）
- Create: `.github/workflows/ci.yml`（lint + test + build）

**Approach:**
- `main.go` は最小限（`cmd.Execute()` を呼ぶのみ）
- グローバルフラグ: `--space`（プロファイル指定）、`--debug`、`--json`（全コマンド共通）
- エラー出力: stderr + exit code 1
- `bk --version` でバージョン情報を表示（goreleaser がビルド時に埋め込む）
- `.goreleaser.yaml` のターゲット: linux/amd64, darwin/amd64, darwin/arm64, windows/amd64

**Patterns to follow:**
- `research/02_architecture-patterns.md` §3.4 ディレクトリ構成

**Test scenarios:**
- Test expectation: none -- スキャフォールディングのみ。`go build ./...` のビルド成功と `goreleaser check` をVerificationとする

**Verification:**
- `go build -o bk .` が成功する
- `./bk --help` および `./bk --version` が正常終了する
- `goreleaser check` でエラーがない

---

- [ ] **Unit 2: 設定管理**

**Goal:** viper ベースの YAML 設定ファイル管理、複数スペース対応、環境変数オーバーライド

**Requirements:** R3（チームメンバーが容易に設定できる基盤）

**Dependencies:** Unit 1

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Approach:**
- 設定ファイルパス: `$XDG_CONFIG_HOME/backlog/config.yaml` → フォールバック `~/.config/backlog/config.yaml`
- 設定構造体: `Config { DefaultSpace string; Spaces map[string]SpaceConfig }`
- `SpaceConfig { Host string; AuthMethod string }` （AuthMethod は現在 "api-key" 固定）
- スペース解決の優先順位: `--space` フラグ > `BACKLOG_SPACE` 環境変数 > `default_space` 設定値
- 設定ファイルが存在しない場合: エラーではなく「`bk auth login` を実行してください」というガイドメッセージを返す
- viper でYAML読み込み、カスタム型へデシリアライズ

**Patterns to follow:**
- `research/02_architecture-patterns.md` §3.1, §3.2 設定ファイル構造例

**Test scenarios:**
- Happy path: 有効な config.yaml を読み込み、`default_space` の `SpaceConfig` を返す
- Happy path: 複数スペースが定義されている場合に指定スペースの設定を返す
- Edge case: 設定ファイルが存在しない場合、ガイドエラーを返す（exit code 1、stderr にガイドメッセージ）
- Edge case: `BACKLOG_SPACE` 環境変数が設定されている場合、`default_space` より優先される（`--space` フラグが最優先）
- Edge case: 指定されたスペースが `spaces` マップに存在しない場合、エラーを返す

**Verification:**
- テストが全て通る
- 設定ファイルなし状態で `bk pr list` を実行すると、ガイドメッセージが stderr に出力される

---

- [ ] **Unit 3: 認証サブシステム**

**Goal:** `bk auth login/logout/status` コマンド実装、go-keyring による APIキー保存、環境変数フォールバック

**Requirements:** R3（C3: `bk auth login` 1コマンドで認証設定完了）

**Dependencies:** Unit 2

**Files:**
- Create: `cmd/auth.go`
- Create: `internal/auth/keychain.go`
- Create: `internal/auth/keychain_test.go`

**Approach:**
- `bk auth login`: ホスト名（例: `myteam.backlog.com`）と APIキーをインタラクティブ入力。APIキー入力は `golang.org/x/term.ReadPassword` で端末エコーを無効にし画面に表示しない。`GET /api/v2/space` でAPIキーを検証後に Keychain に保存
- `bk auth logout`: 指定スペースの APIキーを Keychain から削除
- `bk auth status`: 認証済みスペース一覧とデフォルトスペースを表示
- APIキー解決優先順位: `BACKLOG_API_KEY` 環境変数 → Keychain（go-keyring）→ `~/.config/backlog/credentials.yaml`（パーミッション600）
- Keychain サービス名: `backlog-cli`、アカウント名: スペース名（例: `myteam`）
- go-keyring が失敗した場合（Linux Secret Service未起動等）はファイルフォールバックに切り替え、警告を stderr に出力

**Patterns to follow:**
- `research/03_security-considerations.md` §3.1 認証方式比較、§3.3 設定ファイルパーミッション

**Test scenarios:**
- Happy path: APIキーが Keychain に保存され、`bk auth status` で確認できる
- Happy path: `BACKLOG_API_KEY` 環境変数が設定されている場合は Keychain より優先される
- Happy path: `bk auth logout` で Keychain からAPIキーが削除される
- Edge case: `bk auth login` でAPIキーが無効な場合（API検証失敗）、Keychainに保存せずエラー終了
- Error path: `bk auth logout` で未ログインスペースを指定した場合、明確なエラーメッセージを返す
- Integration: APIキー解決後に `internal/api.NewClient` の初期化が成功する

**Verification:**
- `bk auth login` → `bk auth status` のフローが動作する
- `BACKLOG_API_KEY=xxx bk auth status` で環境変数のAPIキーが認識される

---

- [ ] **Unit 4: Backlog API クライアントラッパー**

**Goal:** kenzo0107/backlog SDK を内部型でラップし、PR 操作の統一インターフェースを提供

**Requirements:** R1（PR操作APIの基盤）、R2（PR作成の基盤）

**Dependencies:** Unit 3

**Files:**
- Create: `internal/api/client.go`
- Create: `internal/api/pr.go`
- Create: `internal/api/client_test.go`
- Create: `internal/api/pr_test.go`

**Approach:**
- `NewClient(host, apiKey string) *Client`: SDK クライアントを初期化
- PR操作メソッド群: `ListPullRequests`, `GetPullRequest`, `CreatePullRequest`, `AddPullRequestComment`
- ユーザー解決メソッド: `ListUsers` — `--reviewer` のユーザー名→ユーザーID変換に使用（`GET /api/v2/users`）
- リポジトリ情報メソッド: `GetRepository` — `--base` 省略時のデフォルトブランチ取得に使用（`GET /api/v2/projects/:projectIdOrKey/git/repositories/:repoIdOrName`）
- パラメータは内部オプション型（`PRListOptions`, `PRCreateOptions` 等）で定義し、SDK型に変換する
- エラーは SDK エラーを内部エラー型にラップして返す（HTTPステータスコードを保持）
- `ListPullRequests` は count（最大100）/offset を `PRListOptions` で受け取る。`bk pr list` 側に `--limit` フラグ（デフォルト30）を設け、100超の全件取得は自動ページネーションループを実装する
- SDKの PR create/comment カバレッジが不足している場合は nattokin/go-backlog に切り替える（実装時評価）

**Patterns to follow:**
- `research/01_tech-stack.md` §3.2 SDK選定根拠

**Test scenarios:**
- Happy path: `NewClient` でクライアントが正常に初期化される
- Happy path: `ListPullRequests` が PR 一覧を返す（SDKをモック）
- Happy path: `GetPullRequest` が指定IDのPR詳細を返す（SDKをモック）
- Happy path: `CreatePullRequest` がPRを作成し、作成されたPRを返す（SDKをモック）
- Happy path: `AddPullRequestComment` がコメントを追加する（SDKをモック）
- Happy path: `ListUsers` がユーザー一覧を返す（SDKをモック）
- Error path: 認証エラー（HTTP 401）時に適切なエラーメッセージを返す
- Error path: 存在しないPRのID指定（HTTP 404）時に NotFound エラーを返す
- Error path: HTTP 429 レート制限時にリトライ案内を含むエラーメッセージを返す
- Edge case: `ListPullRequests` で count=100 ちょうど返された場合（次ページが存在しうるページ境界）に自動ページネーションが継続して呼ばれる

**Verification:**
- テストが全て通る
- SDK の PR create/comment メソッドが存在しない場合は nattokin/go-backlog へ切り替えを記録

---

- [ ] **Unit 5: PR コマンド群**

**Goal:** `bk pr list`, `bk pr view <id>`, `bk pr create`, `bk pr comment <id>` の完全実装

**Requirements:** R1（C1: PR操作CLI）、R2（C2: カレントブランチからPR作成）

**Dependencies:** Unit 4

**Files:**
- Create: `cmd/pr.go`
- Create: `internal/git/remote.go`（git remote URL 解析）
- Create: `internal/git/remote_test.go`
- Create: `internal/output/table.go`
- Create: `internal/output/json.go`
- Create: `internal/output/output_test.go`

**Approach:**
- **プロジェクト/リポジトリ自動検出** (`internal/git/remote.go`): `git remote get-url origin` でURLを取得し、以下パターンを解析:
  - HTTPS: `https://<space>.backlog.com/git/<project>/<repo>.git`
  - SSH: `git@<space>.git.backlog.com:<project>/<repo>.git`（コロン区切りに注意）
  - 解析失敗時は `--project` / `--repo` フラグでの明示指定をガイドするエラーを返す
- **`bk pr list`**: `--status`（open/closed/merged/all）フィルタ。デフォルトはopen。`--limit`（デフォルト30、最大100超は自動ページネーション）。テーブル出力 or `--json` でJSON配列
- **`bk pr view <id>`**: タイトル・ステータス・本文・作成者・レビュアー・コメント数・URL を整形表示
- **`bk pr create`**: カレントブランチ名を `branch`（source）フィールドに使用。`--base`（マージ先ブランチ、省略時はリポジトリのデフォルトブランチをAPIで取得）、`--title`（必須。未指定時インタラクティブ入力）、`--body`（任意）、`--reviewer`（カンマ区切りユーザー名 → `ListUsers` で userIdに変換）フラグ。作成後にPR URLを標準出力
- **`bk pr comment <id>`**: `--body` フラグでコメント本文を指定。未指定時は `$EDITOR` を起動し、終了後に一時ファイルを `defer os.Remove` で必ず削除する

**Patterns to follow:**
- `research/00_existing-implementations.md` §3.2 コマンド設計（gh互換動詞）
- `research/02_architecture-patterns.md` §3.5 グローバルフラグ・エラー・出力方針

**Test scenarios:**
- Happy path (git remote): Backlog HTTPS URL から project key と repo name を正しく抽出する
- Happy path (git remote): Backlog SSH URL（`git@<space>.git.backlog.com:<project>/<repo>.git` 形式）から project key と repo name を正しく抽出する
- Edge case (git remote): 非Backlog URL（GitHub等）が設定されている場合、エラーを返す
- Edge case (git remote): `--project` / `--repo` フラグが git remote URL 解析より優先される
- Happy path (list): PR一覧がテーブル形式で stdout に出力される
- Happy path (list): `--json` フラグで JSON 配列が出力される
- Happy path (list): `--status closed` が API リクエストパラメータに反映される
- Happy path (view): PR詳細が整形表示される（タイトル・ステータス・作成者・レビュアー）
- Happy path (create): カレントブランチ名が `branch` フィールドに設定されてPRが作成される
- Happy path (create): `--reviewer` で指定したユーザー名が userIdに変換されレビュアーとして設定される
- Happy path (create): `--base` 省略時にリポジトリAPIでデフォルトブランチを取得して `base` に設定される
- Happy path (create): ブランチ名にスラッシュ（例: `feature/foo`）を含む場合もPRが作成される
- Edge case (create): git リポジトリ外で実行した場合にエラーを返す
- Error path (create): `--reviewer` に存在しないユーザー名を指定した場合、エラーを返しPRを作成しない
- Error path (create): Backlog APIがエラーを返した場合（例: 対象ブランチが存在しない）、わかりやすいメッセージを stderr に表示する
- Happy path (comment): `--body` でコメントが投稿される
- Edge case (comment): `$EDITOR` で作成した一時ファイルがコマンド終了後に削除されている
- Integration: `bk pr create` → `bk pr view <id>` → `bk pr comment <id>` のフローが一貫して動作する（統合テストまたは手動確認）

**Verification:**
- テストが全て通る
- 実際のBacklogスペースに対して `bk pr list` が動作する（手動E2E確認）
- `bk pr create` でBacklog側にPRが作成されることを確認

---

- [ ] **Unit 6: 配布パイプライン**

**Goal:** goreleaser による自動バイナリビルドとGitHub Releasesへの公開、CIパイプライン整備

**Requirements:** R3（C3: `go install` または GitHub Releases バイナリで4ステップ以内のインストール）

**Dependencies:** Unit 1〜5の全完成後

**Files:**
- Modify: `.goreleaser.yaml`（Unit 1の骨格を完成させる）
- Create: `.github/workflows/release.yml`（tag push でトリガー）
- Modify: `.github/workflows/ci.yml`（全ユニット実装後に更新）

**Approach:**
- goreleaser: linux/amd64, darwin/amd64, darwin/arm64, windows/amd64 の4バイナリをビルド
- GitHub Actions release.yml: `v*` タグのpushをトリガーに goreleaser を実行
- CI (ci.yml): PRおよび main ブランチpushで `go vet`, `staticcheck`, `go test ./...` を実行
- バイナリ名は `bk`（goreleaser の `builds.binary` に明示指定。`archives` のアーカイブ名は `backlog-cli` でも可）
- goreleaser の `builds.ldflags` でバージョン情報をバイナリに埋め込む

**Patterns to follow:**
- `research/04_distribution-channel.md` §3.2, §3.3 goreleaser設定・段階的展開

**Test scenarios:**
- Test expectation: none -- 設定ファイルのみ。`goreleaser check` と `goreleaser release --snapshot --clean` で検証する

**Verification:**
- `goreleaser check` でエラーがない
- `goreleaser release --snapshot --clean` でローカルビルドが成功し、4プラットフォームのバイナリが生成される
- tag push をトリガーに GitHub Actions が起動し、GitHub Releases に公開されることを確認

## System-Wide Impact

- **依存グラフ**: 全PRコマンド → api client → auth（APIキー解決） → config（スペース設定）。設定・認証の初期化は全コマンドの前提
- **エラー伝播**: SDK エラーは内部エラー型にラップして stderr + exit code 1 で出力。HTTPステータスコードを保持してエラーメッセージをユーザーフレンドリーに変換する
- **ステートライフサイクルリスク**: APIキー解決はリクエストごとではなく、コマンド起動時に1回実行してクライアントに注入する設計
- **変更しない不変条件**: Issue/Wiki/Project 操作コマンドは実装しない（スコープ外）。PR以外のAPI呼び出しは発生しない
- **インテグレーションカバレッジ**: Unit 5 の統合テストシナリオは、モックなしで実際のBacklog APIに接続する手動確認で補う

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| kenzo0107/backlog の PR create/comment APIがカバーされていない | Unit 4 実装前にSDKのメソッドリストを確認し、不足なら nattokin/go-backlog に切り替える |
| Linux Keychain未対応環境での認証失敗 | ファイルフォールバック（credentials.yaml 600）を実装し、警告メッセージを表示 |
| git remote URL のパターン未対応（カスタムドメイン・Backlog Enterprise等） | `--project` / `--repo` フラグによる明示指定をガイドするエラーメッセージで対処 |
| パブリックリポジトリ公開に伴うセキュリティリスク | APIキーがソースコードに入らない設計（Keychain/env var/600ファイル）の徹底。コミット前に `BACKLOG_API_KEY` の有無を確認 |

## Sources & References

- **Origin document:** [requirements_specification.md](../../requirements_specification.md)
- Related research: [research/00_existing-implementations.md](../../research/00_existing-implementations.md)
- Related research: [research/01_tech-stack.md](../../research/01_tech-stack.md)
- Related research: [research/02_architecture-patterns.md](../../research/02_architecture-patterns.md)
- Related research: [research/03_security-considerations.md](../../research/03_security-considerations.md)
- Related research: [research/04_distribution-channel.md](../../research/04_distribution-channel.md)
- Related issue: https://github.com/tsukasaSUZUKI-721/LIFE/issues/29
- External: https://developer.nulab.com/ja/docs/backlog/
- Reference impl: https://github.com/simochee/backlog-cli
