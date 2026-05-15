package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// PRRow はテーブル出力用の PR 行データ。
type PRRow struct {
	Number int
	Title  string
	Branch string
	Status string
	Author string
}

// PrintPRTable は PR 一覧をテーブル形式で w に出力する。
func PrintPRTable(w io.Writer, rows []PRRow) {
	if w == nil {
		w = os.Stdout
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "#\tタイトル\tブランチ\tステータス\t作成者")
	fmt.Fprintln(tw, "-\t------\t------\t--------\t----")
	for _, r := range rows {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n",
			r.Number,
			truncate(r.Title, 50),
			r.Branch,
			r.Status,
			r.Author,
		)
	}
	tw.Flush()
}

// PRDetail は PR 詳細表示用のデータ。
type PRDetail struct {
	Number       int
	Title        string
	Status       string
	Branch       string
	Base         string
	BaseCommit   string
	BranchCommit string
	MergeCommit  string
	Author       string
	Reviewers    []string
	CommentCount int
	URL          string
	Description  string
}

// PrintPRDetail は PR 詳細を整形表示する。
func PrintPRDetail(w io.Writer, pr PRDetail) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "PR #%d: %s\n", pr.Number, pr.Title)
	fmt.Fprintf(w, "ステータス: %s\n", pr.Status)
	fmt.Fprintf(w, "ブランチ:   %s → %s\n", pr.Branch, pr.Base)
	if pr.BranchCommit != "" {
		fmt.Fprintf(w, "コミット:   %s → %s\n", pr.BranchCommit, pr.BaseCommit)
	}
	if pr.MergeCommit != "" {
		fmt.Fprintf(w, "マージ済:   %s\n", pr.MergeCommit)
	}
	fmt.Fprintf(w, "作成者:     %s\n", pr.Author)
	if len(pr.Reviewers) > 0 {
		fmt.Fprintf(w, "レビュアー: %s\n", strings.Join(pr.Reviewers, ", "))
	}
	fmt.Fprintf(w, "コメント数: %d\n", pr.CommentCount)
	if pr.URL != "" {
		fmt.Fprintf(w, "URL:        %s\n", pr.URL)
	}
	if pr.Description != "" {
		fmt.Fprintf(w, "\n%s\n", pr.Description)
	}
}

// IssueRow はテーブル出力用の課題行データ。
type IssueRow struct {
	Key      string
	Title    string
	Status   string
	Priority string
	Assignee string
}

// PrintIssueTable は課題一覧をテーブル形式で w に出力する。
func PrintIssueTable(w io.Writer, rows []IssueRow) {
	if w == nil {
		w = os.Stdout
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "キー\tタイトル\tステータス\t優先度\t担当者")
	fmt.Fprintln(tw, "---\t------\t--------\t-----\t----")
	for _, r := range rows {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			r.Key,
			truncate(r.Title, 50),
			r.Status,
			r.Priority,
			r.Assignee,
		)
	}
	tw.Flush()
}

// IssueDetail は課題詳細表示用のデータ。
type IssueDetail struct {
	Key         string
	Title       string
	Status      string
	Priority    string
	IssueType   string
	Assignee    string
	Author      string
	Description string
}

// PrintIssueDetail は課題詳細を整形表示する。
func PrintIssueDetail(w io.Writer, issue IssueDetail) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "課題 %s: %s\n", issue.Key, issue.Title)
	fmt.Fprintf(w, "ステータス: %s\n", issue.Status)
	fmt.Fprintf(w, "種別:       %s\n", issue.IssueType)
	fmt.Fprintf(w, "優先度:     %s\n", issue.Priority)
	fmt.Fprintf(w, "担当者:     %s\n", issue.Assignee)
	fmt.Fprintf(w, "登録者:     %s\n", issue.Author)
	if issue.Description != "" {
		fmt.Fprintf(w, "\n%s\n", issue.Description)
	}
}

// ProjectRow はテーブル出力用のプロジェクト行データ。
type ProjectRow struct {
	Key  string
	Name string
}

// PrintProjectTable はプロジェクト一覧をテーブル形式で w に出力する。
func PrintProjectTable(w io.Writer, rows []ProjectRow) {
	if w == nil {
		w = os.Stdout
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "キー\t名前")
	fmt.Fprintln(tw, "---\t----")
	for _, r := range rows {
		fmt.Fprintf(tw, "%s\t%s\n", r.Key, r.Name)
	}
	tw.Flush()
}

// RepoRow はテーブル出力用のリポジトリ行データ。
type RepoRow struct {
	Name        string
	Description string
}

// PrintRepoTable はリポジトリ一覧をテーブル形式で w に出力する。
func PrintRepoTable(w io.Writer, rows []RepoRow) {
	if w == nil {
		w = os.Stdout
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "名前\t説明")
	fmt.Fprintln(tw, "----\t----")
	for _, r := range rows {
		fmt.Fprintf(tw, "%s\t%s\n", r.Name, truncate(r.Description, 60))
	}
	tw.Flush()
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
