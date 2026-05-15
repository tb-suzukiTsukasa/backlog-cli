package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ts-suzuki/backlog-cli/internal/api"
	"github.com/ts-suzuki/backlog-cli/internal/auth"
	"github.com/ts-suzuki/backlog-cli/internal/config"
	"github.com/ts-suzuki/backlog-cli/internal/output"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage issues",
}

// ------ bk issue list ------

var (
	issueListStatus  string
	issueListLimit   int
	issueListProject string
)

var issueListCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues",
	RunE:  runIssueList,
}

// ------ bk issue view ------

var issueViewCmd = &cobra.Command{
	Use:   "view <issue-key>",
	Short: "View an issue",
	Args:  cobra.ExactArgs(1),
	RunE:  runIssueView,
}

// ------ bk issue comment ------

var (
	issueCommentBody    string
	issueCommentProject string
)

var issueCommentCmd = &cobra.Command{
	Use:   "comment <issue-key>",
	Short: "Add a comment to an issue",
	Args:  cobra.ExactArgs(1),
	RunE:  runIssueComment,
}

func init() {
	rootCmd.AddCommand(issueCmd)
	issueCmd.AddCommand(issueListCmd)
	issueCmd.AddCommand(issueViewCmd)
	issueCmd.AddCommand(issueCommentCmd)

	issueListCmd.Flags().StringVar(&issueListStatus, "status", "open", "フィルタするステータス (open/in-progress/resolved/closed/all)")
	issueListCmd.Flags().IntVar(&issueListLimit, "limit", 30, "取得する最大件数")
	issueListCmd.Flags().StringVar(&issueListProject, "project", "", "プロジェクトキー（必須）")
	_ = issueListCmd.MarkFlagRequired("project")

	issueCommentCmd.Flags().StringVar(&issueCommentBody, "body", "", "コメント本文（省略時は $EDITOR を起動）")
}

func issueStatusToIDs(status string) ([]int, error) {
	switch strings.ToLower(status) {
	case "open":
		return []int{1}, nil
	case "in-progress":
		return []int{2}, nil
	case "resolved":
		return []int{3}, nil
	case "closed":
		return []int{4}, nil
	case "all", "":
		return nil, nil
	default:
		return nil, fmt.Errorf("不明なステータス %q: open/in-progress/resolved/closed/all のいずれかを指定してください", status)
	}
}

func setupIssueClient() (*api.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	sc, spaceName, err := cfg.ResolveSpace(flagSpace)
	if err != nil {
		return nil, err
	}
	apiKey, err := auth.GetAPIKey(spaceName)
	if err != nil {
		return nil, err
	}
	return api.NewClient(sc.Host, apiKey)
}

func runIssueList(cmd *cobra.Command, _ []string) error {
	client, err := setupIssueClient()
	if err != nil {
		return err
	}

	project, err := client.GetProject(context.Background(), issueListProject)
	if err != nil {
		return fmt.Errorf("プロジェクト %q の取得に失敗しました: %w", issueListProject, err)
	}

	statusIDs, err := issueStatusToIDs(issueListStatus)
	if err != nil {
		return err
	}

	issues, err := client.ListIssues(context.Background(), project.ID, issueListLimit, api.IssueListOptions{
		StatusIDs: statusIDs,
	})
	if err != nil {
		return err
	}

	if flagJSON {
		return output.PrintJSON(os.Stdout, issues)
	}

	rows := make([]output.IssueRow, 0, len(issues))
	for _, issue := range issues {
		status := ""
		if issue.Status != nil {
			status = issue.Status.Name
		}
		priority := ""
		if issue.Priority != nil {
			priority = issue.Priority.Name
		}
		assignee := ""
		if issue.Assignee != nil {
			assignee = issue.Assignee.Name
		}
		rows = append(rows, output.IssueRow{
			Key:      issue.IssueKey,
			Title:    issue.Summary,
			Status:   status,
			Priority: priority,
			Assignee: assignee,
		})
	}
	output.PrintIssueTable(os.Stdout, rows)
	return nil
}

func runIssueView(cmd *cobra.Command, args []string) error {
	client, err := setupIssueClient()
	if err != nil {
		return err
	}

	issue, err := client.GetIssue(context.Background(), args[0])
	if err != nil {
		return err
	}

	if flagJSON {
		return output.PrintJSON(os.Stdout, issue)
	}

	status := ""
	if issue.Status != nil {
		status = issue.Status.Name
	}
	priority := ""
	if issue.Priority != nil {
		priority = issue.Priority.Name
	}
	issueType := ""
	if issue.IssueType != nil {
		issueType = issue.IssueType.Name
	}
	assignee := ""
	if issue.Assignee != nil {
		assignee = issue.Assignee.Name
	}
	author := ""
	if issue.CreatedUser != nil {
		author = issue.CreatedUser.Name
	}

	output.PrintIssueDetail(os.Stdout, output.IssueDetail{
		Key:         issue.IssueKey,
		Title:       issue.Summary,
		Status:      status,
		Priority:    priority,
		IssueType:   issueType,
		Assignee:    assignee,
		Author:      author,
		Description: issue.Description,
	})
	return nil
}

func runIssueComment(cmd *cobra.Command, args []string) error {
	client, err := setupIssueClient()
	if err != nil {
		return err
	}

	body := issueCommentBody
	if body == "" {
		fmt.Fprint(os.Stderr, "コメント本文: ")
		reader := bufio.NewReader(os.Stdin)
		body, _ = reader.ReadString('\n')
		body = strings.TrimSpace(body)
	}
	if body == "" {
		return fmt.Errorf("コメント本文を入力してください")
	}

	if err := client.AddIssueComment(context.Background(), args[0], body); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "課題 %s にコメントを追加しました\n", args[0])
	return nil
}
