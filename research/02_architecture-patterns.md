# 調査: アーキテクチャパターン

- カテゴリ: architecture-patterns
- 作成日: 2026-04-30
- 紐づく達成条件: C1, C2, C3

## 1. 調査の目的

設定管理・認証フロー・ディレクトリ構成などのアーキテクチャ設計を確定するために、gh/aws-cli/kubectl などの参考実装から採用すべきパターンを整理する。

## 2. 調査方法・ソース

- 使用ツール: 公開情報（gh/aws-cli/kubectlの設計ドキュメント・ソース）
- 主要ソース: GitHub CLI, aws-cli, kubectl の実装パターン

## 3. 発見・要点

### 3.1 設定ファイルの保存場所

| ツール | 場所 | 形式 |
|---|---|---|
| gh (GitHub CLI) | `~/.config/gh/config.yml` | YAML |
| aws-cli | `~/.aws/config` + `~/.aws/credentials` | INI |
| kubectl | `~/.kube/config` | YAML |

**推奨**: `~/.config/backlog/config.yaml` (XDG Base Directory準拠) + フォールバック `~/.backlog/`

### 3.2 複数プロファイル対応

- gh: `hosts:` セクションで複数GitHubアカウント管理
- aws-cli: `[profile name]` セクション + `AWS_PROFILE` 環境変数
- kubectl: contexts + `KUBECONFIG` 環境変数

**推奨パターン**:
```yaml
# ~/.config/backlog/config.yaml
default_space: myteam
spaces:
  myteam:
    host: myteam.backlog.com
    auth_method: api-key
  client:
    host: client.backlog.jp
    auth_method: api-key
```

+ `BACKLOG_SPACE` 環境変数でオーバーライド可能

### 3.3 認証フロー

**推奨フロー**:
1. `bk auth login` で対話型セットアップ（ホスト名・APIキーを入力）
2. APIキーを Keychain に保存（macOS Keychain / Linux Secret Service）
3. CI/CD 向けに `BACKLOG_API_KEY` 環境変数でオーバーライド可能
4. チームメンバーは全員が自分のAPIキーを設定（共有キー禁止）

### 3.4 推奨ディレクトリ構成

```
backlog-cli/
├── main.go
├── cmd/
│   ├── root.go          # ルートコマンド・グローバルフラグ
│   ├── auth.go          # bk auth login/logout/status
│   ├── issue.go         # bk issue list/view/create/edit/comment
│   └── pr.go            # bk pr list/view/create/comment
├── internal/
│   ├── config/          # 設定ファイル読み込み（viper）
│   ├── auth/            # Keychain統合
│   ├── api/             # Backlog APIクライアントラッパー
│   └── output/          # table/JSON フォーマッタ
└── go.mod
```

### 3.5 コマンド設計ベストプラクティス

- 動詞統一: `list`, `view`, `create`, `edit`, `delete`, `comment`（gh互換）
- グローバルフラグ: `--space` (プロファイル指定), `--debug`, `--json`
- エラー: stderr + exit code 1
- 詳細ログ: `--debug` または `BACKLOG_DEBUG=true`
- 機械可読出力: `--json` フラグ（パイプライン対応）

## 4. 達成条件への影響

| 達成条件 | 影響 | 補足 |
|---------|------|------|
| C1 (Issue CRUD) | 達成しやすくなる | ディレクトリ構成とコマンド設計が明確になる |
| C2 (PR操作) | 達成しやすくなる | 同一パターンでpr.goを追加できる |
| C3 (4ステップ以内) | 達成しやすくなる | `bk auth login` 1コマンドで設定完了する設計 |

## 5. 推奨アクション

- 設定ファイル: XDG準拠の `~/.config/backlog/config.yaml`
- 認証: `bk auth login` → Keychain保存 → `BACKLOG_API_KEY` 環境変数でオーバーライド
- コマンド動詞: gh CLI互換（list/view/create/edit/delete/comment）
- 複数スペース: 最初から `spaces:` 構造で設計し `--space` フラグで切り替え

## 6. 未解決事項

- Keychain未対応環境（一部Linux）でのフォールバック方針
- `bk issue create` の対話型プロンプト（必須か否か）

## 7. 参考資料

- https://github.com/cli/cli (GitHub CLI)
- https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html
- https://kubernetes.io/docs/reference/kubectl/
