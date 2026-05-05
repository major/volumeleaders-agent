package update

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/major/volumeleaders-agent/internal/cli/testutil"
)

func TestUpdateConfigShowsDefaultSettings(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	cmd := NewCmd("0.8.1")

	stdout, _, err := testutil.ExecuteCommand(t, cmd, context.Background(), "config")
	if err != nil {
		t.Fatalf("update config returned error: %v", err)
	}
	var result SettingsResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode update config output: %v", err)
	}
	if !result.CheckNotifications {
		t.Fatal("expected check_notifications to default to true")
	}
	wantPath := filepath.Join(configHome, "volumeleaders-agent", "update-settings.json")
	if result.Path != wantPath {
		t.Fatalf("path = %q, want %q", result.Path, wantPath)
	}
}

func TestUpdateConfigPersistsNotificationSetting(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	cmd := NewCmd("0.8.1")

	stdout, _, err := testutil.ExecuteCommand(t, cmd, context.Background(), "config", "--check-notifications=false")
	if err != nil {
		t.Fatalf("update config returned error: %v", err)
	}
	var result SettingsResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode update config output: %v", err)
	}
	if result.CheckNotifications {
		t.Fatal("expected check_notifications to be false")
	}
}

func TestUpdateConfigRepairsCorruptSettingsWhenSaving(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	settingsPath := filepath.Join(configHome, "volumeleaders-agent", "update-settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o700); err != nil {
		t.Fatalf("create settings directory: %v", err)
	}
	if err := os.WriteFile(settingsPath, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("write corrupt settings: %v", err)
	}
	cmd := NewCmd("0.8.1")

	stdout, _, err := testutil.ExecuteCommand(t, cmd, context.Background(), "config", "--check-notifications=true")
	if err != nil {
		t.Fatalf("update config returned error: %v", err)
	}
	var result SettingsResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode update config output: %v", err)
	}
	if !result.CheckNotifications {
		t.Fatal("expected check_notifications to be true")
	}
}
