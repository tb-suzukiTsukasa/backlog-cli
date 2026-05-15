package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/ts-suzuki/backlog-cli/internal/output"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Args:  cobra.NoArgs,
	RunE:  runProjectList,
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
}

func runProjectList(_ *cobra.Command, _ []string) error {
	client, err := setupIssueClient()
	if err != nil {
		return err
	}

	projects, err := client.ListProjects(context.Background())
	if err != nil {
		return err
	}

	if flagJSON {
		return output.PrintJSON(os.Stdout, projects)
	}

	rows := make([]output.ProjectRow, 0, len(projects))
	for _, p := range projects {
		rows = append(rows, output.ProjectRow{Key: p.ProjectKey, Name: p.Name})
	}
	output.PrintProjectTable(os.Stdout, rows)
	return nil
}
