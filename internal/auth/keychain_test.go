package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func setupCredsDir(t *testing.T) string {
	t.Helper()
	xdgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgDir)
	return xdgDir
}

func TestGetAPIKey_EnvVar(t *testing.T) {
	setupCredsDir(t)
	t.Setenv("BACKLOG_API_KEY", "env-api-key")

	key, err := GetAPIKey("myteam")
	if err != nil {
		t.Fatalf("GetAPIKey() error = %v", err)
	}
	if key != "env-api-key" {
		t.Errorf("key = %q, want %q", key, "env-api-key")
	}
}

func TestGetAPIKey_FromFile(t *testing.T) {
	xdgDir := setupCredsDir(t)
	t.Setenv("BACKLOG_API_KEY", "")

	dir := filepath.Join(xdgDir, "backlog")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	content := "credentials:\n  myteam: file-api-key\n"
	if err := os.WriteFile(filepath.Join(dir, "credentials.yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// Keychain は単体テストでは使えないので、ファイルフォールバックで検証
	key, err := getFromFile("myteam")
	if err != nil {
		t.Fatalf("getFromFile() error = %v", err)
	}
	if key != "file-api-key" {
		t.Errorf("key = %q, want %q", key, "file-api-key")
	}
}

func TestSaveToFile_And_DeleteFromFile(t *testing.T) {
	setupCredsDir(t)

	if err := saveToFile("myteam", "saved-key"); err != nil {
		t.Fatalf("saveToFile() error = %v", err)
	}

	key, err := getFromFile("myteam")
	if err != nil {
		t.Fatalf("getFromFile() after save error = %v", err)
	}
	if key != "saved-key" {
		t.Errorf("key = %q, want %q", key, "saved-key")
	}

	if err := deleteFromFile("myteam"); err != nil {
		t.Fatalf("deleteFromFile() error = %v", err)
	}

	_, err = getFromFile("myteam")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestGetFromFile_NotFound(t *testing.T) {
	setupCredsDir(t)

	_, err := getFromFile("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent space, got nil")
	}
}
