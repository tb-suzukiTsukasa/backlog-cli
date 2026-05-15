package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const githubRepo = "tb-suzukiTsukasa/backlog-cli"

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade bk to the latest version",
	Args:  cobra.NoArgs,
	RunE:  runUpgrade,
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(_ *cobra.Command, _ []string) error {
	fmt.Fprintln(os.Stderr, "最新バージョンを確認しています...")

	latest, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("バージョン情報の取得に失敗しました: %w", err)
	}

	// "dev" ビルドでも強制アップグレード可能にする（version は vX.Y.Z 形式）
	current := strings.TrimPrefix(version, "v")
	latestClean := strings.TrimPrefix(latest, "v")

	if current == latestClean && current != "dev" {
		fmt.Fprintf(os.Stderr, "すでに最新バージョン (%s) です\n", latest)
		return nil
	}

	fmt.Fprintf(os.Stderr, "アップグレードします: %s → %s\n", version, latest)

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("実行ファイルのパス取得に失敗しました: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("シンボリックリンクの解決に失敗しました: %w", err)
	}

	if err := downloadAndReplace(latest, execPath); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "アップグレード完了: %s\n", latest)
	return nil
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func fetchLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API エラー (status %d)", resp.StatusCode)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("リリースが見つかりません")
	}
	return rel.TagName, nil
}

func downloadAndReplace(ver, destPath string) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// GoReleaser のアーカイブ名に合わせる
	archiveName := fmt.Sprintf("backlog-cli_%s_%s.tar.gz", goos, goarch)
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", githubRepo, ver, archiveName)

	fmt.Fprintf(os.Stderr, "ダウンロード中: %s\n", archiveName)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("ダウンロードに失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ダウンロードエラー (status %d): %s", resp.StatusCode, url)
	}

	newBin, err := extractBinary(resp.Body)
	if err != nil {
		return fmt.Errorf("アーカイブの展開に失敗しました: %w", err)
	}

	// 既存バイナリをバックアップしてから置き換え
	tmpPath := destPath + ".tmp"
	if err := os.WriteFile(tmpPath, newBin, 0o755); err != nil {
		return fmt.Errorf("一時ファイルの書き込みに失敗しました: %w", err)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("バイナリの置き換えに失敗しました（sudo が必要かもしれません）: %w", err)
	}

	return nil
}

// extractBinary は .tar.gz から "bk" バイナリの内容を返す。
func extractBinary(r io.Reader) ([]byte, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		name := filepath.Base(hdr.Name)
		if name == "bk" || name == "bk.exe" {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("アーカイブ内に bk バイナリが見つかりません")
}
