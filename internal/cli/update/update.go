package update

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	updater "github.com/major/volumeleaders-agent/internal/update"
)

const updateTimeout = 5 * time.Minute

type configOptions struct {
	CheckNotifications bool
}

type installOptions struct {
	Force bool
}

// SettingsResult describes persisted updater notification settings.
type SettingsResult struct {
	CheckNotifications bool   `json:"check_notifications"`
	Path               string `json:"path"`
}

// NewCmd returns the update command group.
func NewCmd(currentVersion string) *cobra.Command {
	installOpts := &installOptions{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update volumeleaders-agent",
		Args:  cobra.NoArgs,
		Long:  "Download the latest GitHub release for the current platform, verify it against the release checksum file, and replace the running binary atomically. Automatic update notifications are enabled by default, cached for one day, skipped in CI and non-interactive output, and can be disabled with update config.",
		Example: `volumeleaders-agent update
volumeleaders-agent update --force`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), updateTimeout)
			defer cancel()
			slog.Info("Checking GitHub releases for updates")
			result, err := updater.InstallLatest(ctx, currentVersion, installOpts.Force)
			if err != nil {
				return fmt.Errorf("install latest release: %w", err)
			}
			if result.Updated {
				slog.Info("Installed update", "version", result.CurrentVersion, "asset", result.AssetName)
			} else {
				slog.Info("Already running the latest release", "version", result.CurrentVersion)
			}
			return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), result)
		},
	}
	cmd.Flags().BoolVar(&installOpts.Force, "force", false, "Install the latest release even when the current version is already latest")
	common.AnnotateFlagGroup(cmd, "force", "Update")
	common.WrapValidation(cmd, installOpts)
	cmd.AddCommand(newCheckCmd(currentVersion), newConfigCmd())
	return cmd
}

func newCheckCmd(currentVersion string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "check",
		Short:   "Check for available updates",
		Args:    cobra.NoArgs,
		Long:    "Check the latest GitHub release for the current platform and report whether it is newer than the running binary. This command only reports status and never modifies the installed binary.",
		Example: "volumeleaders-agent update check",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			result, err := updater.CheckLatest(ctx, currentVersion)
			if err != nil {
				return fmt.Errorf("check for updates: %w", err)
			}
			return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), result)
		},
	}
	return cmd
}

func newConfigCmd() *cobra.Command {
	opts := &configOptions{CheckNotifications: true}
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show or change update settings",
		Args:  cobra.NoArgs,
		Long:  "Show updater notification settings, or persist a new automatic notification preference when --check-notifications is set. This updater-specific settings file only controls update checks and does not enable general CLI config loading.",
		Example: `volumeleaders-agent update config
volumeleaders-agent update config --check-notifications=false`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			settings, err := updater.LoadSettings()
			if err != nil {
				if cmd.Flags().Changed("check-notifications") {
					settings = updater.DefaultSettings()
				} else {
					return fmt.Errorf("load update settings: %w", err)
				}
			}
			if cmd.Flags().Changed("check-notifications") {
				settings.CheckNotifications = opts.CheckNotifications
				if err := updater.SaveSettings(settings); err != nil {
					return fmt.Errorf("save update settings: %w", err)
				}
			}
			path, err := updater.SettingsPath()
			if err != nil {
				return fmt.Errorf("resolve update settings path: %w", err)
			}
			return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), SettingsResult{CheckNotifications: settings.CheckNotifications, Path: path})
		},
	}
	cmd.Flags().BoolVar(&opts.CheckNotifications, "check-notifications", true, "Set automatic update notification preference; true enables notifications, false disables them, omitted only displays current settings")
	common.AnnotateFlagGroup(cmd, "check-notifications", "Update")
	common.WrapValidation(cmd, opts)
	return cmd
}
