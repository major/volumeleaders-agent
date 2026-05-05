package update

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"
)

const notificationTimeout = 2 * time.Second

// NotifyIfDue checks for a newer release when notification settings and cache state allow it.
func NotifyIfDue(ctx context.Context, currentVersion, commandPath string) {
	if shouldSkipNotification(currentVersion, commandPath) {
		return
	}
	settings, err := LoadSettings()
	if err != nil {
		slog.Debug("Skipping update notification", "error", err)
		return
	}
	if !settings.CheckNotifications {
		return
	}
	state, err := LoadState()
	if err != nil {
		slog.Debug("Skipping update notification", "error", err)
		return
	}
	if time.Since(state.LastCheckedAt) < DefaultCheckInterval {
		return
	}
	checkCtx, cancel := context.WithTimeout(ctx, notificationTimeout)
	defer cancel()
	result, err := CheckLatest(checkCtx, currentVersion)
	state.LastCheckedAt = time.Now().UTC()
	if result.LatestVersion != "" {
		state.LatestVersion = result.LatestVersion
	}
	if saveErr := SaveState(state); saveErr != nil {
		slog.Debug("Unable to save update check state", "error", saveErr)
	}
	if err != nil {
		slog.Debug("Update check failed", "error", err)
		return
	}
	if result.UpdateAvailable {
		slog.Warn("Update available", "current", currentVersion, "latest", result.LatestVersion, "command", "volumeleaders-agent update")
	}
}

func shouldSkipNotification(currentVersion, commandPath string) bool {
	if currentVersion == "" || currentVersion == "dev" {
		return true
	}
	if os.Getenv("CI") != "" {
		return true
	}
	if strings.HasPrefix(commandPath, "volumeleaders-agent update") {
		return true
	}
	return !isTerminal(os.Stdout) || !isTerminal(os.Stderr)
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
