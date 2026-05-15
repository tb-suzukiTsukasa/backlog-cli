# 調査: 配布方法

- カテゴリ: distribution-channel
- 作成日: 2026-04-30
- 紐づく達成条件: C3

## 1. 調査の目的

チームメンバーが4ステップ以内でインストールできる配布方法を選定する。段階的な展開戦略（最初は最小コスト・後で拡充）を決定する。

## 2. 調査方法・ソース

- 使用ツール: 公開情報（goreleaser公式、Homebrew公式、Go公式）
- 主要ソース: goreleaser docs, Homebrew docs

## 3. 発見・要点

### 3.1 配布方法の比較

| 方法 | 手順数 | セットアップコスト | 更新フロー | 条件 |
|---|---|---|---|---|
| `go install` | 1〜2 | 低 | 手動（再実行） | Go環境が必要 |
| GitHub Releases（手動DL） | 3〜4 | なし | 手動DL | なし |
| Homebrew tap | 2〜3 | 中（tap管理） | `brew upgrade` で自動 | macOS/Linux |
| aqua / mise | 1〜2 | 低〜中 | 自動更新 | ツールマネージャー導入済み |

### 3.2 goreleaser による GitHub Releases 自動化

- `.goreleaser.yaml` 1ファイルで Linux/macOS/Windows の全バイナリを自動ビルド
- GitHub Actions + tag push で自動リリース
- Homebrew tap への自動 formula 更新も設定可能
- **セットアップコスト**: `.goreleaser.yaml` 作成 + GitHub Actions設定（1〜2時間）

### 3.3 段階的展開の推奨

**フェーズ1（初版リリース直後）**:
```bash
# Go環境がある場合（エンジニア向け）
go install github.com/ts-suzuki/backlog-cli@latest

# Go環境がない場合
# GitHub Releases から OS対応バイナリをダウンロードして PATH に置く
```
→ 手順数: 1〜2ステップ（4ステップ以内を満たす）

**フェーズ2（チーム展開安定後）**:
```bash
brew tap ts-suzuki/backlog-cli
brew install backlog-cli
```
→ 手順数: 2ステップ、`brew upgrade` で自動更新

**フェーズ3（オプション）**:
- aqua / mise 対応（バージョン固定管理が必要なチーム向け）

### 3.4 インストール後の認証セットアップ（合計手順）

```bash
# Step 1: インストール
go install github.com/ts-suzuki/backlog-cli@latest

# Step 2: 認証設定（ホスト名とAPIキーを入力）
bk auth login

# Step 3: 動作確認
bk issue list
```
→ 合計3ステップ（C3達成条件「4ステップ以内」を満たす）

## 4. 達成条件への影響

| 達成条件 | 影響 | 補足 |
|---------|------|------|
| C3 (4ステップ以内) | 達成可能 | go install + bk auth login の2ステップ構成で実現できる |

## 5. 推奨アクション

- **初版**: goreleaser + GitHub Releases を整備。`go install` をプライマリ手段とする
- **次段階**: Homebrew tap を追加（`brew install` 対応でGo環境不要にする）
- README に「3ステップのクイックスタート」を明記し、C3達成を担保する

## 6. 未解決事項

- プライベートリポジトリ vs パブリックリポジトリ（チーム内配布 vs OSS公開）の方針
- go install はモジュールがパブリックでないと動作しない点の対処

## 7. 参考資料

- https://goreleaser.com/
- https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap
- https://pkg.go.dev/cmd/go#hdr-Add_dependencies_to_current_module_and_install_them
