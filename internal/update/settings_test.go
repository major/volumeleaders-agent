package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadSettingsFileReturnsDefaultsWhenMissing(t *testing.T) {
	t.Parallel()
	settings, err := LoadSettingsFile(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("LoadSettingsFile returned error: %v", err)
	}
	if !settings.CheckNotifications {
		t.Fatal("expected update notifications to default to enabled")
	}
}

func TestSaveAndLoadSettingsFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "settings.json")
	want := Settings{CheckNotifications: false}
	if err := SaveSettingsFile(path, want); err != nil {
		t.Fatalf("SaveSettingsFile returned error: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat settings file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("settings permissions = %o, want 600", got)
	}
	got, err := LoadSettingsFile(path)
	if err != nil {
		t.Fatalf("LoadSettingsFile returned error: %v", err)
	}
	if got != want {
		t.Fatalf("settings = %+v, want %+v", got, want)
	}
}

func TestSaveSettingsFileTightensExistingPermissions(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"check_notifications":true}`), 0o644); err != nil {
		t.Fatalf("write permissive settings file: %v", err)
	}
	if err := SaveSettingsFile(path, Settings{CheckNotifications: false}); err != nil {
		t.Fatalf("SaveSettingsFile returned error: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat settings file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("settings permissions = %o, want 600", got)
	}
}

func TestSaveSettingsFileOverwritesExistingFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := SaveSettingsFile(path, Settings{CheckNotifications: true}); err != nil {
		t.Fatalf("first SaveSettingsFile returned error: %v", err)
	}
	if err := SaveSettingsFile(path, Settings{CheckNotifications: false}); err != nil {
		t.Fatalf("second SaveSettingsFile returned error: %v", err)
	}
	got, err := LoadSettingsFile(path)
	if err != nil {
		t.Fatalf("LoadSettingsFile returned error: %v", err)
	}
	if got.CheckNotifications {
		t.Fatal("expected second save to replace existing settings")
	}
}

func TestSaveAndLoadStateFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "state.json")
	want := State{LastCheckedAt: time.Date(2026, 5, 5, 12, 0, 0, 0, time.UTC), LatestVersion: "0.8.1"}
	if err := SaveStateFile(path, want); err != nil {
		t.Fatalf("SaveStateFile returned error: %v", err)
	}
	got, err := LoadStateFile(path)
	if err != nil {
		t.Fatalf("LoadStateFile returned error: %v", err)
	}
	if !got.LastCheckedAt.Equal(want.LastCheckedAt) || got.LatestVersion != want.LatestVersion {
		t.Fatalf("state = %+v, want %+v", got, want)
	}
}
