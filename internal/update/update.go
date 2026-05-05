package update

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	selfupdate "github.com/creativeprojects/go-selfupdate"
)

const (
	repositorySlug       = "major/volumeleaders-agent"
	DefaultCheckInterval = 24 * time.Hour
)

// CheckResult describes the latest available release for this binary.
type CheckResult struct {
	CurrentVersion  string    `json:"current_version"`
	LatestVersion   string    `json:"latest_version,omitempty"`
	UpdateAvailable bool      `json:"update_available"`
	AssetName       string    `json:"asset_name,omitempty"`
	ReleaseURL      string    `json:"release_url,omitempty"`
	CheckedAt       time.Time `json:"checked_at"`
}

// InstallResult describes the result of a self-update attempt.
type InstallResult struct {
	PreviousVersion string `json:"previous_version"`
	CurrentVersion  string `json:"current_version"`
	Updated         bool   `json:"updated"`
	AssetName       string `json:"asset_name,omitempty"`
	ReleaseURL      string `json:"release_url,omitempty"`
}

// CheckLatest detects the latest GitHub release matching the current OS and architecture.
func CheckLatest(ctx context.Context, currentVersion string) (CheckResult, error) {
	updater, err := newUpdater()
	if err != nil {
		return CheckResult{}, fmt.Errorf("check latest: %w", err)
	}
	release, found, err := updater.DetectLatest(ctx, selfupdate.ParseSlug(repositorySlug))
	if err != nil {
		return CheckResult{}, fmt.Errorf("detect latest release: %w", err)
	}
	checkedAt := time.Now().UTC()
	if !found {
		return CheckResult{CurrentVersion: currentVersion, CheckedAt: checkedAt}, nil
	}
	latestVersion := release.Version()
	return CheckResult{
		CurrentVersion:  currentVersion,
		LatestVersion:   latestVersion,
		UpdateAvailable: isNewerVersion(latestVersion, currentVersion),
		AssetName:       release.AssetName,
		ReleaseURL:      release.URL,
		CheckedAt:       checkedAt,
	}, nil
}

// InstallLatest downloads, verifies, and applies the latest GitHub release.
func InstallLatest(ctx context.Context, currentVersion string, force bool) (InstallResult, error) {
	updater, err := newUpdater()
	if err != nil {
		return InstallResult{}, fmt.Errorf("install latest: %w", err)
	}
	release, found, err := updater.DetectLatest(ctx, selfupdate.ParseSlug(repositorySlug))
	if err != nil {
		return InstallResult{}, fmt.Errorf("detect latest release: %w", err)
	}
	if !found {
		return InstallResult{}, fmt.Errorf("no release asset found for this OS and architecture")
	}
	latestVersion := release.Version()
	result := InstallResult{
		PreviousVersion: currentVersion,
		CurrentVersion:  currentVersion,
		AssetName:       release.AssetName,
		ReleaseURL:      release.URL,
	}
	if !force && !isNewerVersion(latestVersion, currentVersion) {
		return result, nil
	}
	executablePath, err := selfupdate.ExecutablePath()
	if err != nil {
		return InstallResult{}, fmt.Errorf("locate executable path: %w", err)
	}
	if err := updater.UpdateTo(ctx, release, executablePath); err != nil {
		return InstallResult{}, fmt.Errorf("apply verified update: %w", err)
	}
	result.CurrentVersion = latestVersion
	result.Updated = true
	return result, nil
}

func newUpdater() (*selfupdate.Updater, error) {
	updater, err := selfupdate.NewUpdater(selfupdate.Config{Validator: GoReleaserChecksumValidator{}})
	if err != nil {
		return nil, fmt.Errorf("create updater: %w", err)
	}
	return updater, nil
}

func isNewerVersion(latestVersion, currentVersion string) bool {
	if latestVersion == "" {
		return false
	}
	latest, err := semver.NewVersion(latestVersion)
	if err != nil {
		return false
	}
	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		return true
	}
	return latest.GreaterThan(current)
}
