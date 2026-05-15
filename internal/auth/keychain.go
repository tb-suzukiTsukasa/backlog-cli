package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const keychainService = "backlog-cli"

// GetAPIKey はスペース名に対応する APIキーを解決する。
// 優先順位: BACKLOG_API_KEY 環境変数 → Keychain → credentials.yaml
func GetAPIKey(spaceName string) (string, error) {
	if key := os.Getenv("BACKLOG_API_KEY"); key != "" {
		return key, nil
	}

	key, err := keyring.Get(keychainService, spaceName)
	if err == nil {
		return key, nil
	}

	return getFromFile(spaceName)
}

// SetAPIKey は APIキーを Keychain に保存する。
// Keychain が使えない場合は credentials.yaml（パーミッション600）にフォールバックする。
func SetAPIKey(spaceName, apiKey string) error {
	if err := keyring.Set(keychainService, spaceName, apiKey); err != nil {
		fmt.Fprintln(os.Stderr, "警告: Keychain への保存に失敗しました。ファイルにフォールバックします:", err)
		return saveToFile(spaceName, apiKey)
	}
	return nil
}

// DeleteAPIKey はスペース名に対応する APIキーを Keychain と credentials.yaml から削除する。
func DeleteAPIKey(spaceName string) error {
	keychainErr := keyring.Delete(keychainService, spaceName)
	fileErr := deleteFromFile(spaceName)
	if keychainErr != nil && fileErr != nil {
		return fmt.Errorf("スペース %q のAPIキーが見つかりません", spaceName)
	}
	return nil
}

// credentialsFile はファイルフォールバック用の credentials.yaml パスを返す。
func credentialsFile() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "backlog", "credentials.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "backlog", "credentials.yaml")
}

type credentialsFile_ struct {
	Credentials map[string]string `yaml:"credentials"`
}

func loadCredentialsFile() (*credentialsFile_, error) {
	data, err := os.ReadFile(credentialsFile())
	if os.IsNotExist(err) {
		return &credentialsFile_{Credentials: map[string]string{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var cf credentialsFile_
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return nil, err
	}
	if cf.Credentials == nil {
		cf.Credentials = map[string]string{}
	}
	return &cf, nil
}

func saveCredentialsFile(cf *credentialsFile_) error {
	path := credentialsFile()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cf)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func getFromFile(spaceName string) (string, error) {
	cf, err := loadCredentialsFile()
	if err != nil {
		return "", fmt.Errorf("APIキーが見つかりません: %w", err)
	}
	key, ok := cf.Credentials[spaceName]
	if !ok {
		return "", fmt.Errorf("スペース %q のAPIキーが設定されていません。`bk auth login` を実行してください", spaceName)
	}
	return key, nil
}

func saveToFile(spaceName, apiKey string) error {
	cf, err := loadCredentialsFile()
	if err != nil {
		cf = &credentialsFile_{Credentials: map[string]string{}}
	}
	cf.Credentials[spaceName] = apiKey
	return saveCredentialsFile(cf)
}

func deleteFromFile(spaceName string) error {
	cf, err := loadCredentialsFile()
	if err != nil {
		return err
	}
	if _, ok := cf.Credentials[spaceName]; !ok {
		return fmt.Errorf("not found")
	}
	delete(cf.Credentials, spaceName)
	return saveCredentialsFile(cf)
}
