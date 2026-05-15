package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "bk",
	Short: "Backlog CLI - manage Backlog from your terminal",
	Long:  "A command-line interface for Backlog. Manage pull requests without leaving the terminal.",
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
}

var (
	flagSpace string
	flagDebug bool
	flagJSON  bool
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagSpace, "space", "", "Backlog space profile to use")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
}
