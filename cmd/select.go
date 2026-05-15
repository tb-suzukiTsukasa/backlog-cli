package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ts-suzuki/backlog-cli/internal/api"
	"golang.org/x/term"
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

	items := make([]string, len(projects))
	for i, p := range projects {
		items[i] = fmt.Sprintf("%-20s %s", p.ProjectKey, p.Name)
	}

	idx, err := selectFromList("プロジェクトを選択", items)
	if err != nil {
		return "", err
	}
	return projects[idx].ProjectKey, nil
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

	items := make([]string, len(repos))
	for i, r := range repos {
		items[i] = r.Name
	}

	idx, err := selectFromList("リポジトリを選択", items)
	if err != nil {
		return "", err
	}
	return repos[idx].Name, nil
}

// selectFromList は矢印キー+Enterで選択できるインタラクティブメニューを表示する。
// ターミナル以外（パイプ等）では番号入力にフォールバックする。
func selectFromList(prompt string, items []string) (int, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return selectFromListFallback(prompt, items)
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return selectFromListFallback(prompt, items)
	}
	defer term.Restore(fd, oldState)

	cur := 0
	printMenu(prompt, items, cur)

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return 0, err
		}

		switch {
		case n == 1 && buf[0] == '\r': // Enter
			// カーソルを消してメニューをクリア
			clearMenu(len(items) + 1)
			fmt.Fprintf(os.Stderr, "%s: %s\n", prompt, items[cur])
			return cur, nil

		case n == 1 && (buf[0] == 3 || buf[0] == 'q'): // Ctrl+C / q
			clearMenu(len(items) + 1)
			return 0, fmt.Errorf("キャンセルされました")

		case n == 3 && buf[0] == 0x1b && buf[1] == '[' && buf[2] == 'A': // 上矢印
			if cur > 0 {
				cur--
			}
			clearMenu(len(items) + 1)
			printMenu(prompt, items, cur)

		case n == 3 && buf[0] == 0x1b && buf[1] == '[' && buf[2] == 'B': // 下矢印
			if cur < len(items)-1 {
				cur++
			}
			clearMenu(len(items) + 1)
			printMenu(prompt, items, cur)
		}
	}
}

func printMenu(prompt string, items []string, cur int) {
	fmt.Fprintf(os.Stderr, "%s:\r\n", prompt)
	for i, item := range items {
		if i == cur {
			fmt.Fprintf(os.Stderr, "  \x1b[36m▶ %s\x1b[0m\r\n", item)
		} else {
			fmt.Fprintf(os.Stderr, "    %s\r\n", item)
		}
	}
}

func clearMenu(lines int) {
	for i := 0; i < lines; i++ {
		fmt.Fprint(os.Stderr, "\x1b[1A\x1b[2K")
	}
}

// selectFromListFallback は非TTY環境向けの番号入力フォールバック。
func selectFromListFallback(prompt string, items []string) (int, error) {
	fmt.Fprintf(os.Stderr, "%s:\n", prompt)
	for i, item := range items {
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, item)
	}
	fmt.Fprint(os.Stderr, "番号を入力: ")

	var n int
	if _, err := fmt.Fscan(os.Stdin, &n); err != nil || n < 1 || n > len(items) {
		return 0, fmt.Errorf("無効な選択です")
	}
	return n - 1, nil
}
