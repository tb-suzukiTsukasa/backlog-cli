package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ts-suzuki/backlog-cli/internal/output"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage git repositories",
}

var repoListProject string

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List git repositories in a project",
	Args:  cobra.NoArgs,
	RunE:  runRepoList,
}

func init() {
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(repoListCmd)

	repoListCmd.Flags().StringVar(&repoListProject, "project", "", "プロジェクトキー（省略時は対話的に選択）")
}

func runRepoList(_ *cobra.Command, _ []string) error {
	client, err := setupIssueClient()
	if err != nil {
		return err
	}

	projectKey := repoListProject
	if projectKey == "" {
		projectKey, err = selectProject(client)
		if err != nil {
			return err
		}
	}

	repos, err := client.ListRepositories(context.Background(), projectKey)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		fmt.Fprintf(os.Stderr, "プロジェクト %q にリポジトリが見つかりません\n", projectKey)
		return nil
	}

	if flagJSON {
		return output.PrintJSON(os.Stdout, repos)
	}

	rows := make([]output.RepoRow, 0, len(repos))
	for _, r := range repos {
		rows = append(rows, output.RepoRow{Name: r.Name, Description: r.Description})
	}
	output.PrintRepoTable(os.Stdout, rows)
	return nil
}
