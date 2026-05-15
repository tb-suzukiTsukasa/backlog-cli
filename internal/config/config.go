package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type SpaceConfig struct {
	Host       string `mapstructure:"host"`
	AuthMethod string `mapstructure:"auth_method"`
}

type Config struct {
	DefaultSpace string                 `mapstructure:"default_space"`
	Spaces       map[string]SpaceConfig `mapstructure:"spaces"`
}

const keychainService = "backlog-cli"

func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "backlog")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "backlog")
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir())

	v.SetEnvPrefix("BACKLOG")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("設定ファイルが見つかりません。`bk auth login` を実行してセットアップしてください")
		}
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("設定ファイルのパースに失敗しました: %w", err)
	}
	return &cfg, nil
}

// ResolveSpace はスペース解決の優先順位に従って SpaceConfig を返す。
// 優先順位: flagSpace > BACKLOG_SPACE env > default_space 設定値
func (c *Config) ResolveSpace(flagSpace string) (*SpaceConfig, string, error) {
	spaceName := flagSpace
	if spaceName == "" {
		spaceName = os.Getenv("BACKLOG_SPACE")
	}
	if spaceName == "" {
		spaceName = c.DefaultSpace
	}
	if spaceName == "" {
		return nil, "", fmt.Errorf("スペースが指定されていません。`--space` フラグか BACKLOG_SPACE 環境変数を設定してください")
	}

	sc, ok := c.Spaces[spaceName]
	if !ok {
		return nil, "", fmt.Errorf("スペース %q が設定ファイルに存在しません。`bk auth login` で登録してください", spaceName)
	}
	return &sc, spaceName, nil
}

// Save は設定を YAML ファイルに書き込む。
func Save(cfg *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("設定ディレクトリの作成に失敗しました: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(filepath.Join(dir, "config.yaml"))
	v.Set("default_space", cfg.DefaultSpace)
	v.Set("spaces", cfg.Spaces)

	return v.WriteConfigAs(filepath.Join(dir, "config.yaml"))
}
