package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ts-suzuki/backlog-cli/internal/api"
)

// selectProject はプロジェクト一覧を表示してユーザーに選択させ、選択されたプロジェクトキーを返す。
func selectProject(client *api.Client) (string, error) {
	projects, err := client.ListProjects(context.Background())
	if err != nil {
		return "", err
	}
	if len(projects) == 0 {
		return "", fmt.Errorf("プロジェクトが見つかりません")
	}

	fmt.Fprintln(os.Stderr, "プロジェクトを選択してください:")
	for i, p := range projects {
		fmt.Fprintf(os.Stderr, "  %d) %s  %s\n", i+1, p.ProjectKey, p.Name)
	}
	fmt.Fprint(os.Stderr, "番号を入力: ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	n, err := strconv.Atoi(line)
	if err != nil || n < 1 || n > len(projects) {
		return "", fmt.Errorf("無効な選択です: %q", line)
	}
	return projects[n-1].ProjectKey, nil
}

// selectRepo はリポジトリ一覧を表示してユーザーに選択させ、選択されたリポジトリ名を返す。
// リポジトリが1件のみの場合は自動選択する。
func selectRepo(client *api.Client, projectKey string) (string, error) {
	repos, err := client.ListRepositories(context.Background(), projectKey)
	if err != nil {
		return "", err
	}
	if len(repos) == 0 {
		return "", fmt.Errorf("プロジェクト %q にリポジトリが見つかりません", projectKey)
	}
	if len(repos) == 1 {
		fmt.Fprintf(os.Stderr, "リポジトリを自動選択しました: %s\n", repos[0].Name)
		return repos[0].Name, nil
	}

	fmt.Fprintln(os.Stderr, "リポジトリを選択してください:")
	for i, r := range repos {
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, r.Name)
	}
	fmt.Fprint(os.Stderr, "番号を入力: ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	n, err := strconv.Atoi(line)
	if err != nil || n < 1 || n > len(repos) {
		return "", fmt.Errorf("無効な選択です: %q", line)
	}
	return repos[n-1].Name, nil
}
