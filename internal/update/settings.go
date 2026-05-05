package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	appName          = "volumeleaders-agent"
	settingsFileName = "update-settings.json"
	stateFileName    = "update-state.json"
)

// Settings stores user-controlled update notification preferences.
type Settings struct {
	CheckNotifications bool `json:"check_notifications"`
}

// State stores updater cache data that should not affect user preferences.
type State struct {
	LastCheckedAt time.Time `json:"last_checked_at"`
	LatestVersion string    `json:"latest_version,omitempty"`
}

// DefaultSettings returns updater settings for users without a settings file.
func DefaultSettings() Settings {
	return Settings{CheckNotifications: true}
}

// SettingsPath returns the platform-specific path for updater settings.
func SettingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config directory: %w", err)
	}
	return filepath.Join(dir, appName, settingsFileName), nil
}

// StatePath returns the platform-specific path for updater cache state.
func StatePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("user cache directory: %w", err)
	}
	return filepath.Join(dir, appName, stateFileName), nil
}

// LoadSettings reads updater settings, returning defaults when the file does not exist.
func LoadSettings() (Settings, error) {
	path, err := SettingsPath()
	if err != nil {
		return Settings{}, err
	}
	return LoadSettingsFile(path)
}

// LoadSettingsFile reads updater settings from path, returning defaults on cache miss.
func LoadSettingsFile(path string) (Settings, error) {
	settings := DefaultSettings()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return settings, nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("read update settings: %w", err)
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}, fmt.Errorf("decode update settings: %w", err)
	}
	return settings, nil
}

// SaveSettings writes updater settings to path using user-only permissions.
func SaveSettings(settings Settings) error {
	path, err := SettingsPath()
	if err != nil {
		return err
	}
	return SaveSettingsFile(path, settings)
}

// SaveSettingsFile writes updater settings to path using user-only permissions.
func SaveSettingsFile(path string, settings Settings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("encode update settings: %w", err)
	}
	if err := writeJSONFile(path, data); err != nil {
		return fmt.Errorf("write update settings: %w", err)
	}
	return nil
}

// LoadState reads updater cache state, returning an empty state when it does not exist.
func LoadState() (State, error) {
	path, err := StatePath()
	if err != nil {
		return State{}, err
	}
	return LoadStateFile(path)
}

// LoadStateFile reads updater cache state from path.
func LoadStateFile(path string) (State, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return State{}, nil
	}
	if err != nil {
		return State{}, fmt.Errorf("read update state: %w", err)
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, fmt.Errorf("decode update state: %w", err)
	}
	return state, nil
}

// SaveState writes updater cache state using user-only permissions.
func SaveState(state State) error {
	path, err := StatePath()
	if err != nil {
		return err
	}
	return SaveStateFile(path, state)
}

// SaveStateFile writes updater cache state using user-only permissions.
func SaveStateFile(path string, state State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode update state: %w", err)
	}
	if err := writeJSONFile(path, data); err != nil {
		return fmt.Errorf("write update state: %w", err)
	}
	return nil
}

func writeJSONFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("secure temporary file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temporary file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}
	if err := replaceFile(tmpPath, path); err != nil {
		return fmt.Errorf("replace file: %w", err)
	}
	return nil
}

func replaceFile(sourcePath, destinationPath string) error {
	if err := os.Rename(sourcePath, destinationPath); err == nil {
		return nil
	}
	// Windows does not allow os.Rename to replace an existing destination. The
	// updater settings and state files are small, non-critical JSON files, so the
	// portable fallback is to remove the old file and retry the rename.
	if err := os.Remove(destinationPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove existing file: %w", err)
	}
	if err := os.Rename(sourcePath, destinationPath); err != nil {
		return fmt.Errorf("rename replacement file: %w", err)
	}
	return nil
}
