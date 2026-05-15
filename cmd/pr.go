package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ts-suzuki/backlog-cli/internal/api"
	"github.com/ts-suzuki/backlog-cli/internal/auth"
	"github.com/ts-suzuki/backlog-cli/internal/config"
	gitpkg "github.com/ts-suzuki/backlog-cli/internal/git"
	"github.com/ts-suzuki/backlog-cli/internal/output"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Manage pull requests",
}

// ------ bk pr list ------

var (
	prListStatus string
	prListLimit  int
	prListProject string
	prListRepo   string
)

var prListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pull requests",
	RunE:  runPRList,
}

func init() {
	rootCmd.AddCommand(prCmd)
	prCmd.AddCommand(prListCmd)
	prCmd.AddCommand(prViewCmd)
	prCmd.AddCommand(prCreateCmd)
	prCmd.AddCommand(prCommentCmd)

	prListCmd.Flags().StringVar(&prListStatus, "status", "open", "フィルタするステータス (open/closed/merged/all)")
	prListCmd.Flags().IntVar(&prListLimit, "limit", 30, "取得する最大件数")
	prListCmd.Flags().StringVar(&prListProject, "project", "", "プロジェクトキー（省略時は git remote から自動検出）")
	prListCmd.Flags().StringVar(&prListRepo, "repo", "", "リポジトリ名（省略時は git remote から自動検出）")

	prCreateCmd.Flags().StringVar(&prCreateTitle, "title", "", "PR タイトル（省略時はインタラクティブ入力）")
	prCreateCmd.Flags().StringVar(&prCreateBody, "body", "", "PR 本文")
	prCreateCmd.Flags().StringVar(&prCreateBase, "base", "", "マージ先ブランチ（省略時はデフォルトブランチ）")
	prCreateCmd.Flags().StringVar(&prCreateReviewer, "reviewer", "", "レビュアーのユーザー名（カンマ区切り）")
	prCreateCmd.Flags().StringVar(&prCreateProject, "project", "", "プロジェクトキー")
	prCreateCmd.Flags().StringVar(&prCreateRepo, "repo", "", "リポジトリ名")

	prViewCmd.Flags().StringVar(&prViewProject, "project", "", "プロジェクトキー（省略時は git remote から自動検出）")
	prViewCmd.Flags().StringVar(&prViewRepo, "repo", "", "リポジトリ名（省略時は git remote から自動検出）")

	prCommentCmd.Flags().StringVar(&prCommentBody, "body", "", "コメント本文（省略時は $EDITOR を起動）")
	prCommentCmd.Flags().StringVar(&prCommentProject, "project", "", "プロジェクトキー")
	prCommentCmd.Flags().StringVar(&prCommentRepo, "repo", "", "リポジトリ名")
}

func statusToIDs(status string) ([]int, error) {
	switch strings.ToLower(status) {
	case "open":
		return []int{1}, nil
	case "closed":
		return []int{2}, nil
	case "merged":
		return []int{3}, nil
	case "all", "":
		return nil, nil
	default:
		return nil, fmt.Errorf("不明なステータス %q: open/closed/merged/all のいずれかを指定してください", status)
	}
}

func statusName(id int) string {
	switch id {
	case 1:
		return "Open"
	case 2:
		return "Closed"
	case 3:
		return "Merged"
	default:
		return "Unknown"
	}
}

func runPRList(cmd *cobra.Command, _ []string) error {
	client, projectKey, repoName, err := setupClientAndRepo(prListProject, prListRepo)
	if err != nil {
		return err
	}

	statusIDs, err := statusToIDs(prListStatus)
	if err != nil {
		return err
	}

	prs, err := client.ListPullRequests(context.Background(), projectKey, repoName, prListLimit, api.PRListOptions{
		StatusIDs: statusIDs,
	})
	if err != nil {
		return err
	}

	if flagJSON {
		return output.PrintJSON(os.Stdout, prs)
	}

	rows := make([]output.PRRow, 0, len(prs))
	for _, pr := range prs {
		status := ""
		if pr.Status != nil {
			status = statusName(pr.Status.ID)
		}
		author := ""
		if pr.CreatedUser != nil {
			author = pr.CreatedUser.Name
		}
		rows = append(rows, output.PRRow{
			Number: pr.Number,
			Title:  pr.Summary,
			Branch: pr.Branch,
			Status: status,
			Author: author,
		})
	}
	output.PrintPRTable(os.Stdout, rows)
	return nil
}

// ------ bk pr view ------

var (
	prViewProject string
	prViewRepo    string
)

var prViewCmd = &cobra.Command{
	Use:   "view <number>",
	Short: "View a pull request",
	Args:  cobra.ExactArgs(1),
	RunE:  runPRView,
}

func runPRView(cmd *cobra.Command, args []string) error {
	var number int
	if _, err := fmt.Sscanf(args[0], "%d", &number); err != nil {
		return fmt.Errorf("PR 番号は数値で指定してください: %q", args[0])
	}

	client, projectKey, repoName, err := setupClientAndRepo(prViewProject, prViewRepo)
	if err != nil {
		return err
	}

	pr, err := client.GetPullRequest(context.Background(), projectKey, repoName, number)
	if err != nil {
		return err
	}

	if flagJSON {
		return output.PrintJSON(os.Stdout, pr)
	}

	detail := prToDetail(pr, projectKey, repoName)
	output.PrintPRDetail(os.Stdout, detail)
	return nil
}

func prToDetail(pr *api.PullRequest, projectKey, repoName string) output.PRDetail {
	status := ""
	if pr.Status != nil {
		status = statusName(pr.Status.ID)
	}
	author := ""
	if pr.CreatedUser != nil {
		author = pr.CreatedUser.Name
	}
	return output.PRDetail{
		Number:       pr.Number,
		Title:        pr.Summary,
		Status:       status,
		Branch:       pr.Branch,
		Base:         pr.Base,
		BaseCommit:   pr.BaseCommit,
		BranchCommit: pr.BranchCommit,
		MergeCommit:  pr.MergeCommit,
		Author:       author,
		Description:  pr.Description,
	}
}

// ------ bk pr create ------

var (
	prCreateTitle    string
	prCreateBody     string
	prCreateBase     string
	prCreateReviewer string
	prCreateProject  string
	prCreateRepo     string
)

var prCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a pull request",
	RunE:  runPRCreate,
}

func runPRCreate(cmd *cobra.Command, _ []string) error {
	client, projectKey, repoName, err := setupClientAndRepo(prCreateProject, prCreateRepo)
	if err != nil {
		return err
	}

	// カレントブランチをソースブランチとして使用
	branch, err := gitpkg.GetCurrentBranch()
	if err != nil {
		return err
	}

	// タイトル未指定時はインタラクティブ入力
	title := prCreateTitle
	if title == "" {
		fmt.Fprint(os.Stderr, "PR タイトル: ")
		reader := bufio.NewReader(os.Stdin)
		title, _ = reader.ReadString('\n')
		title = strings.TrimSpace(title)
		if title == "" {
			return fmt.Errorf("PR タイトルを入力してください")
		}
	}

	// --base 未指定時はデフォルトブランチを取得
	base := prCreateBase
	if base == "" {
		base = gitpkg.GetDefaultBranch()
	}

	// レビュアー名 → userID 変換
	var notifiedUserIDs []int
	if prCreateReviewer != "" {
		names := strings.Split(prCreateReviewer, ",")
		ids, err := resolveUserIDs(context.Background(), client, names)
		if err != nil {
			return err
		}
		notifiedUserIDs = ids
	}

	pr, err := client.CreatePullRequest(context.Background(), projectKey, repoName, title, prCreateBody, base, branch, api.PRCreateOptions{
		NotifiedUserIDs: notifiedUserIDs,
	})
	if err != nil {
		return err
	}

	fmt.Printf("PR #%d を作成しました: %s\n", pr.Number, pr.Summary)
	return nil
}

// resolveUserIDs はユーザー名一覧をユーザーID一覧に変換する。
func resolveUserIDs(ctx context.Context, client *api.Client, names []string) ([]int, error) {
	users, err := client.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("ユーザー一覧の取得に失敗しました: %w", err)
	}

	nameToID := make(map[string]int, len(users))
	for _, u := range users {
		if u.UserID != "" {
			nameToID[u.UserID] = u.ID
		}
		if u.Name != "" {
			nameToID[u.Name] = u.ID
		}
	}

	var ids []int
	for _, name := range names {
		name = strings.TrimSpace(name)
		if id, ok := nameToID[name]; ok {
			ids = append(ids, id)
		} else {
			return nil, fmt.Errorf("ユーザー %q が見つかりません", name)
		}
	}
	return ids, nil
}

// ------ bk pr comment ------

var (
	prCommentBody    string
	prCommentProject string
	prCommentRepo    string
)

var prCommentCmd = &cobra.Command{
	Use:   "comment <number>",
	Short: "Add a comment to a pull request",
	Args:  cobra.ExactArgs(1),
	RunE:  runPRComment,
}

func runPRComment(cmd *cobra.Command, args []string) error {
	var number int
	if _, err := fmt.Sscanf(args[0], "%d", &number); err != nil {
		return fmt.Errorf("PR 番号は数値で指定してください: %q", args[0])
	}

	client, projectKey, repoName, err := setupClientAndRepo(prCommentProject, prCommentRepo)
	if err != nil {
		return err
	}

	body := prCommentBody
	if body == "" {
		var editorErr error
		body, editorErr = editWithEditor()
		if editorErr != nil {
			return editorErr
		}
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("コメント本文を入力してください")
	}

	if err := client.AddPullRequestComment(context.Background(), projectKey, repoName, number, body); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "PR #%d にコメントを追加しました\n", number)
	return nil
}

func editWithEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	tmpFile, err := os.CreateTemp("", "bk-pr-comment-*.md")
	if err != nil {
		return "", fmt.Errorf("一時ファイルの作成に失敗しました: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	parts := strings.Fields(editor)
	editorCmd := exec.Command(parts[0], append(parts[1:], tmpPath)...)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		return "", fmt.Errorf("エディタの起動に失敗しました: %w", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("一時ファイルの読み込みに失敗しました: %w", err)
	}
	return string(data), nil
}

// ------ 共通ヘルパー ------

// setupClientAndRepo は設定・認証・git remote を解決して api.Client とプロジェクト/リポジトリを返す。
func setupClientAndRepo(projectFlag, repoFlag string) (*api.Client, string, string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, "", "", err
	}

	sc, spaceName, err := cfg.ResolveSpace(flagSpace)
	if err != nil {
		return nil, "", "", err
	}

	apiKey, err := auth.GetAPIKey(spaceName)
	if err != nil {
		return nil, "", "", err
	}

	client, err := api.NewClient(sc.Host, apiKey)
	if err != nil {
		return nil, "", "", err
	}

	projectKey := projectFlag
	repoName := repoFlag
	if projectKey == "" || repoName == "" {
		repo, detectErr := gitpkg.DetectRepo()
		if detectErr == nil {
			if projectKey == "" {
				projectKey = repo.Project
			}
			if repoName == "" {
				repoName = repo.Repo
			}
		}
	}

	// git remote 検出失敗またはフラグ未指定の場合は対話的に選択
	if projectKey == "" {
		var err error
		projectKey, err = selectProject(client)
		if err != nil {
			return nil, "", "", err
		}
	}
	if repoName == "" {
		var err error
		repoName, err = selectRepo(client, projectKey)
		if err != nil {
			return nil, "", "", err
		}
	}

	return client, projectKey, repoName, nil
}

