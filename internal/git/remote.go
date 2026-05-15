package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// BacklogRepo は git remote URL から抽出したプロジェクト・リポジトリ情報。
type BacklogRepo struct {
	SpaceName string
	Host      string
	Project   string
	Repo      string
}

var (
	// HTTPS: https://<space>.backlog.com/git/<project>/<repo>.git
	// または https://<space>.backlog.jp/git/<project>/<repo>.git
	httpsPattern = regexp.MustCompile(`^https?://([^/]+\.backlog(?:\.com|\.jp))/git/([^/]+)/([^/]+?)(?:\.git)?$`)

	// SSH: git@<space>.git.backlog.com:<project>/<repo>.git
	// または git@<space>.git.backlog.jp:<project>/<repo>.git
	sshPattern = regexp.MustCompile(`^git@([^:]+\.git\.backlog(?:\.com|\.jp)):([^/]+)/([^/]+?)(?:\.git)?$`)
)

// DetectRepo は git remote origin の URL から BacklogRepo を抽出する。
// --project / --repo フラグが指定されている場合はそちらが優先されるため、呼び出し側で判断する。
func DetectRepo() (*BacklogRepo, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return nil, fmt.Errorf("git remote の取得に失敗しました。git リポジトリ内で実行してください: %w", err)
	}
	rawURL := strings.TrimSpace(string(out))
	return parseBacklogURL(rawURL)
}

// GetCurrentBranch はカレントブランチ名を返す。
func GetCurrentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("カレントブランチの取得に失敗しました: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetDefaultBranch は git remote show origin からデフォルトブランチを取得する。
// 失敗した場合は "master" を返す。
func GetDefaultBranch() string {
	out, err := exec.Command("git", "remote", "show", "origin").Output()
	if err != nil {
		return "master"
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "HEAD branch:") {
			branch := strings.TrimSpace(strings.TrimPrefix(line, "HEAD branch:"))
			if branch != "" && branch != "(unknown)" {
				return branch
			}
		}
	}
	return "master"
}

func parseBacklogURL(rawURL string) (*BacklogRepo, error) {
	if m := httpsPattern.FindStringSubmatch(rawURL); len(m) == 4 {
		host := m[1]
		spaceName := extractSpaceName(host)
		return &BacklogRepo{
			SpaceName: spaceName,
			Host:      host,
			Project:   m[2],
			Repo:      m[3],
		}, nil
	}

	if m := sshPattern.FindStringSubmatch(rawURL); len(m) == 4 {
		// SSH ホストは "<space>.git.backlog.com" → 実際のホストは "<space>.backlog.com"
		sshHost := m[1]
		host := strings.Replace(sshHost, ".git.backlog", ".backlog", 1)
		spaceName := extractSpaceName(host)
		return &BacklogRepo{
			SpaceName: spaceName,
			Host:      host,
			Project:   m[2],
			Repo:      m[3],
		}, nil
	}

	return nil, fmt.Errorf("Backlog の git remote URL が検出できませんでした。\n"+
		"URL: %q\n"+
		"`--project` / `--repo` フラグで明示指定してください", rawURL)
}

func extractSpaceName(host string) string {
	// "myteam.backlog.com" → "myteam"
	parts := strings.SplitN(host, ".", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return host
}
