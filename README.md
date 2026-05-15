# backlog-cli (`bk`)

ターミナルから [Backlog](https://backlog.com/) の **Git プルリクエスト** と **認証設定** を操作する CLI です。Backlog REST API を `net/http` で直接呼び出します。

## できること

### 認証・設定（`bk auth`）

| サブコマンド | 内容 |
|--------------|------|
| `login` | 対話形式でスペース名・API ホスト・API キーを入力し、`config.yaml` にホスト等を保存、API キーは OS キーチェーン（失敗時は `credentials.yaml`）に保存 |
| `logout` | 指定スペースの API キーをキーチェーン／`credentials.yaml` から削除（`--space` で対象を切り替え可能） |
| `status` | 登録済みスペース一覧とデフォルトスペースを表示 |

設定ファイルは **`$XDG_CONFIG_HOME/backlog/config.yaml`**（未設定時は **`~/.config/backlog/config.yaml`**）です。

API キーの参照優先順位は次のとおりです。

1. 環境変数 **`BACKLOG_API_KEY`**（設定されていれば最優先）
2. OS キーチェーン（サービス名 `backlog-cli`）
3. **`~/.config/backlog/credentials.yaml`**（パーミッション 600 想定のフォールバック）

どのスペースの設定を使うかは **`--space` フラグ**、環境変数 **`BACKLOG_SPACE`**、設定の **`default_space`** の順で決まります。

### プルリクエスト（`bk pr`）

| サブコマンド | 内容 |
|--------------|------|
| `list` | PR 一覧を表形式（または `--json` で JSON）で表示。`--status` で open / closed / merged / all、`--limit` で最大件数 |
| `view <number>` | 指定 PR の詳細（タイトル・ステータス・ブランチ・ベース・**Base / Branch / Merge のコミット SHA**・作成者・説明など）を表示 |
| `create` | **カレント Git ブランチ**をソースとして PR を作成。`--title` 省略時は対話入力。`--base` 省略時は `origin` のデフォルトブランチ（取得失敗時は `master`）。`--reviewer` で通知ユーザー（カンマ区切り、ユーザー名または表示名で解決） |
| `comment <number>` | PR にコメント。`--body` 省略時は **`$EDITOR`**（未設定なら `vi`）で本文を編集 |

`--project` / `--repo` を省略した場合は、**カレントディレクトリの Git `origin`** から Backlog の HTTPS / SSH URL を解析し、プロジェクトキーとリポジトリ名を自動検出します（Backlog 以外の remote では失敗し、その場合はフラグで明示指定が必要です）。

### 共通オプション（ルート）

- `--space` … 利用する Backlog スペース（プロファイル名）
- `--json` … 出力を JSON に（`pr list` / `pr view` で利用）
- `--debug` … デバッグ出力を有効化
- `--version` … ビルド済みバイナリのバージョン情報

## まだできないこと（現状の範囲外）

- **課題（issue）** の一覧・参照・更新・コメントなどのサブコマンドは **未実装**です（`bk issue` はありません）。
- PR の **diff を API だけで取得**する機能は Backlog 側の都合で想定しておらず、CLI にも **`pr view --diff` のようなコマンドはありません**。

## 必要環境

- **Go 1.26.1** 以降（`go.mod` に準拠）
- `pr` 関連で auto-detect や `create` を使う場合は **`git`** が PATH にあり、Backlog 用の `origin` と作業ツリーが必要です

## インストール

リポジトリを clone したうえで:

```bash
make build   # カレントディレクトリに bk を生成
```

または:

```bash
go build -o bk .
```

`go install` で入れる場合はモジュールパスに合わせて実行してください（例: `go install github.com/ts-suzuki/backlog-cli@latest` ※公開・タグ運用はリポジトリ方針に依存します）。

## 使い方の例

```bash
bk auth login
bk auth status

# プロジェクト・リポジトリは git origin から推測（Backlog の remote であること）
bk pr list --status open --limit 20
bk pr view 42
bk pr create --title "fix: ..." --body "詳細はここに"
bk pr comment 42 --body "LGTM"

# 別スペース・明示指定
bk --space myteam pr list --project MYPROJ --repo my-repo --limit 10
```

## 開発・テスト

```bash
make test
go test -race ./...
```

## セキュリティの注意

Backlog の API は仕様上 **`apiKey` がクエリパラメータに含まれる**ことがあります。プロキシやアクセスログの扱いは運用環境に依存するため、CI では **専用・最小権限の API キー** とシークレット管理を推奨します。

## 関連ドキュメント

- SDK 除去後の実装・API マップ: [docs/handoff-remove-backlog-sdk.md](docs/handoff-remove-backlog-sdk.md)
