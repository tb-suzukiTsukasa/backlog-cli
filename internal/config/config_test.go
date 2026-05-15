package config

import (
	"os"
	"path/filepath"
	"testing"
)

// setupConfig は一時ディレクトリに XDG_CONFIG_HOME/backlog/config.yaml を作成し、
// XDG_CONFIG_HOME を設定して Load() が読み込める状態にする。
func setupConfig(t *testing.T, content string) {
	t.Helper()
	xdgDir := t.TempDir()
	dir := filepath.Join(xdgDir, "backlog")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", xdgDir)
}

func TestLoad_ValidConfig(t *testing.T) {
	setupConfig(t, `
default_space: myteam
spaces:
  myteam:
    host: myteam.backlog.com
    auth_method: api-key
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DefaultSpace != "myteam" {
		t.Errorf("DefaultSpace = %q, want %q", cfg.DefaultSpace, "myteam")
	}
	sc, ok := cfg.Spaces["myteam"]
	if !ok {
		t.Fatal("spaces[myteam] not found")
	}
	if sc.Host != "myteam.backlog.com" {
		t.Errorf("Host = %q, want %q", sc.Host, "myteam.backlog.com")
	}
}

func TestLoad_MultipleSpaces(t *testing.T) {
	setupConfig(t, `
default_space: myteam
spaces:
  myteam:
    host: myteam.backlog.com
    auth_method: api-key
  client:
    host: client.backlog.jp
    auth_method: api-key
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Spaces) != 2 {
		t.Errorf("len(Spaces) = %d, want 2", len(cfg.Spaces))
	}
	if _, ok := cfg.Spaces["client"]; !ok {
		t.Error("spaces[client] not found")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing config file, got nil")
	}
}

func TestResolveSpace_FlagPriority(t *testing.T) {
	cfg := &Config{
		DefaultSpace: "default",
		Spaces: map[string]SpaceConfig{
			"default": {Host: "default.backlog.com", AuthMethod: "api-key"},
			"flagged": {Host: "flagged.backlog.com", AuthMethod: "api-key"},
		},
	}
	t.Setenv("BACKLOG_SPACE", "env-space")

	sc, name, err := cfg.ResolveSpace("flagged")
	if err != nil {
		t.Fatalf("ResolveSpace() error = %v", err)
	}
	if name != "flagged" {
		t.Errorf("name = %q, want %q", name, "flagged")
	}
	if sc.Host != "flagged.backlog.com" {
		t.Errorf("Host = %q, want %q", sc.Host, "flagged.backlog.com")
	}
}

func TestResolveSpace_EnvOverridesDefault(t *testing.T) {
	cfg := &Config{
		DefaultSpace: "default",
		Spaces: map[string]SpaceConfig{
			"default": {Host: "default.backlog.com", AuthMethod: "api-key"},
			"envsp":   {Host: "envsp.backlog.com", AuthMethod: "api-key"},
		},
	}
	t.Setenv("BACKLOG_SPACE", "envsp")

	_, name, err := cfg.ResolveSpace("")
	if err != nil {
		t.Fatalf("ResolveSpace() error = %v", err)
	}
	if name != "envsp" {
		t.Errorf("name = %q, want %q", name, "envsp")
	}
}

func TestResolveSpace_UnknownSpace(t *testing.T) {
	cfg := &Config{
		DefaultSpace: "myteam",
		Spaces: map[string]SpaceConfig{
			"myteam": {Host: "myteam.backlog.com", AuthMethod: "api-key"},
		},
	}
	_, _, err := cfg.ResolveSpace("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent space, got nil")
	}
}
