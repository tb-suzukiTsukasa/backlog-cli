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

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
