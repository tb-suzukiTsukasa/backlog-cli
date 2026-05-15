# 引き継ぎ（Claude Code 向け）: Backlog API 自前 HTTP クライアント

## ステータス

**完了（2026-05-14 時点）**  
`github.com/kenzo0107/backlog` は除去済み。`internal/api` は `net/http` で Backlog REST API を直接呼び出す。`go.mod` に当該 SDK は含まれない。

直近の差分（前セッション成果）:

- `types.go` の `PullRequest` に **`BaseCommit` / `BranchCommit` / `MergeCommit`** を追加。
- `bk pr view` で上記コミット SHA を表示（`internal/output/table.go` 経由）。
- 全テスト PASS、`go build ./...` 通過。

**リポジトリの git 状態**: ブランチは `main`、**まだ 1 件もコミットされていない**（全ファイル untracked）。初回コミットの粒度・タイミングは引き継ぎ先の判断に委ねる。

このドキュメントは、**以降の機能追加・バグ修正・CI 連携**を引き継ぐエージェント向けのコンテキストである。

---

## 背景・目的（経緯）

社内展開や出所の明確化のため、外部 SDK への依存をやめ、`net/http` による自前実装に統一した。

---

## Backlog API の仕様（変更時は必ず照合すること）

- ベース URL: `https://{host}/api/v2`
- 認証: URL クエリ `?apiKey=xxx`（Bearer ではない）
- GET: `GET /path?apiKey=xxx&param=value`
- POST: `POST /path?apiKey=xxx` + `Content-Type: application/x-www-form-urlencoded`
- レスポンス: JSON（コメント POST など本文を使わない場合はデコードしない）

### 実装で使っているエンドポイント

| メソッド | パス（`/api/v2` 以降） | 用途 |
|---|---|---|
| GET | `/users` | ユーザー一覧 |
| GET | `/projects/{projectKey}/git/repositories/{repoName}` | リポジトリ情報 |
| GET | `/projects/.../pullRequests` | PR 一覧 |
| GET | `/projects/.../pullRequests/{number}` | PR 詳細 |
| POST | `/projects/.../pullRequests` | PR 作成 |
| POST | `/projects/.../pullRequests/{number}/comments` | PR コメント |

---

## コードマップ（触る場所の目安）

```
internal/api/
  types.go       ← User, PRStatus, PullRequest, GitRepository（JSON タグ付き）
  client.go      ← Client, NewClient, get/post/do, convertError, ListUsers, GetRepository
  pr.go          ← PRListOptions, PRCreateOptions, List/Get/Create PR, AddPullRequestComment
  client_test.go ← mockHTTPClient, newMockClient（httpClient 直注入）
  pr_test.go     ← PR 系のモック HTTP テスト

cmd/
  pr.go          ← Cobra: pr list/view/create/comment。型は api.* のみ。view は commit SHA も表示
  auth.go        ← Cobra: auth login/logout/status
  root.go        ← ルートコマンド・グローバルフラグ（--json / --space / --debug）

internal/output/
  table.go       ← PR 詳細テーブル整形（commit SHA 列を含む）
```

CLI エントリ: リポジトリルートの `main.go` → `cmd/`。

認証（API キー解決）: `internal/auth/keychain.go` — 優先順位は **`BACKLOG_API_KEY` 環境変数 → OS キーチェーン → credentials.yaml（600）**。

---

## 実装メモ（仕様とコードの対応）

### 型（`types.go`）

- トップレベルの PR フィールド（`Number`, `Summary`, `Branch`, `BaseCommit`, `BranchCommit`, `MergeCommit` など）は**値型**。
- `PullRequest.Status` と `CreatedUser` のみ **ポインタ**（JSON で欠落しうるため）。`cmd/pr.go` では nil チェック後に参照すること。
- `BaseCommit` / `BranchCommit` / `MergeCommit` は Backlog の PR 詳細 API レスポンスにそのまま載るフィールド。マージ前の PR では `MergeCommit` が空文字になる前提で扱う（表示側で空ならスキップ）。

### HTTP クライアント（`client.go`）

- `httpDoer` で `Do` を抽象化し、テストでモック注入。
- `post` / `get` 失敗時は `convertError(statusCode, body)`（401/403/404/429 とその他）。
- **`do` でレスポンス `out == nil` のとき**は JSON をデコードせず、`io.Copy(io.Discard, resp.Body)` でボディを捨てる（接続再利用・EOF 対策）。

### PR 一覧（`pr.go`）

- `statusId[]` は `url.Values.Add("statusId[]", ...)` で複数値（`Set` 不可）。
- `limit` 超は `count` 最大 100 でループ（ページネーション）。最終ページが `count` 未満なら打ち切り。

### PR 作成

- フォーム: `summary`, `description`, `base`, `branch`, `notifiedUserId[]`（繰り返し）。

### テスト

- `newMockClient` は SDK を経由せず `Client{host, apiKey, httpClient: mock}` を返す。

---

## 検証コマンド

```bash
go build ./...
go test -race ./...
```

実 API（要: `bk auth login` 済み、または `BACKLOG_API_KEY` + 設定済みホスト）:

```bash
./bk pr list --project <PROJECT_KEY> --repo <REPO_NAME> --limit 10
./bk pr view <NUMBER> --project <PROJECT_KEY> --repo <REPO_NAME>
```

---

## 既知の制約・セキュリティ（後続タスクの判断材料）

- Backlog の **`apiKey` がクエリに載る**仕様のため、プロキシやログ基盤によっては URL にキーが残るリスクをインフラ側と相談する必要がある（CLI だけでは解消不可）。
- CI で使う場合は **`BACKLOG_API_KEY` またはシークレットストア**と、**専用・最小権限の API キー**を推奨（`internal/auth` の優先順位を参照）。
- `research/` や `docs/plans/` に **旧 SDK 名（kenzo0107/backlog）**が残っている場合がある。コード真実は `go.mod` と `internal/api` に従うこと。

---

## 未着手の領域（次の引き継ぎ先へ）

- **`bk issue` 系コマンド**（`list` / `view` / `comment` / `update`）が未実装（`cmd/issue.go` が存在しない）。Backlog 側に PR#2（`feat/issue-commands`）が Open のまま残っている可能性があるため、内容を再利用するか改めて実装するかは判断が必要。
- **`bk pr view --diff`**: Backlog API には PR の diff エンドポイントが**存在しない**ことが確認済み。実装する場合は PR の `baseCommit` / `branchCommit` を取得して、ローカルの `git diff <base>..<branch>` を呼ぶ方針が現実的。
- **CI**: `.github/workflows/ci.yml` 骨格は配置済み。実行確認やゴールデンパス E2E はこれから。

---

## リリース・バイナリ

`.goreleaser.yaml` で `bk` を linux/darwin/windows（amd64/arm64、Windows は amd64 のみビルド設定）向けにビルドする想定。社内配布はチェックサム検証・配布元の統一を推奨。

---

## Claude Code に渡すときの一言プロンプト例

「`docs/handoff-remove-backlog-sdk.md` を読んで。SDK 除去は完了済み。`internal/api` の HTTP クライアントと `cmd/pr.go` を前提に、〈ここにやりたいこと〉を実装して。変更後は `go test -race ./...` を通すこと。」
