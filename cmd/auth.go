package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ts-suzuki/backlog-cli/internal/auth"
	"github.com/ts-suzuki/backlog-cli/internal/config"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Backlog authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with a Backlog space",
	RunE:  runAuthLogin,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from a Backlog space",
	RunE:  runAuthLogout,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	RunE:  runAuthStatus,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
}

func runAuthLogin(cmd *cobra.Command, _ []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Fprint(os.Stderr, "Backlog スペース名 (https://◆ここ◆.backlog.com): ")
	spaceName, _ := reader.ReadString('\n')
	spaceName = strings.TrimSpace(spaceName)
	if spaceName == "" {
		return fmt.Errorf("スペース名を入力してください")
	}

	fmt.Fprint(os.Stderr, "ホスト名 (例: xxxxx.backlog.com または xxxxx.backlog.jp): ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		return fmt.Errorf("ホスト名を入力してください")
	}

	fmt.Fprint(os.Stderr, "API キー: ")
	apiKeyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return fmt.Errorf("APIキーの読み取りに失敗しました: %w", err)
	}
	apiKey := strings.TrimSpace(string(apiKeyBytes))
	if apiKey == "" {
		return fmt.Errorf("APIキーを入力してください")
	}

	// 設定ファイルへの保存
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{
			Spaces: map[string]config.SpaceConfig{},
		}
	}
	if cfg.Spaces == nil {
		cfg.Spaces = map[string]config.SpaceConfig{}
	}
	cfg.Spaces[spaceName] = config.SpaceConfig{Host: host, AuthMethod: "api-key"}
	if cfg.DefaultSpace == "" {
		cfg.DefaultSpace = spaceName
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("設定の保存に失敗しました: %w", err)
	}

	if err := auth.SetAPIKey(spaceName, apiKey); err != nil {
		return fmt.Errorf("APIキーの保存に失敗しました: %w", err)
	}

	fmt.Fprintf(os.Stderr, "スペース %q の認証が完了しました\n", spaceName)
	return nil
}

func runAuthLogout(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	_, spaceName, err := cfg.ResolveSpace(flagSpace)
	if err != nil {
		return err
	}

	if err := auth.DeleteAPIKey(spaceName); err != nil {
		return fmt.Errorf("スペース %q のAPIキー削除に失敗しました: %w", spaceName, err)
	}

	fmt.Fprintf(os.Stderr, "スペース %q からログアウトしました\n", spaceName)
	return nil
}

func runAuthStatus(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Spaces) == 0 {
		fmt.Fprintln(os.Stderr, "認証済みスペースがありません。`bk auth login` を実行してください")
		return nil
	}

	fmt.Printf("デフォルトスペース: %s\n", cfg.DefaultSpace)
	fmt.Println("\n登録済みスペース:")
	for name, sc := range cfg.Spaces {
		marker := "  "
		if name == cfg.DefaultSpace {
			marker = "* "
		}
		fmt.Printf("%s%s (%s)\n", marker, name, sc.Host)
	}
	return nil
}
