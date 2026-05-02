package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
)

// NewRootCmd returns the root cobra command for volumeleaders-agent.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "volumeleaders-agent",
		Short:         "CLI tool for querying VolumeLeaders institutional trade data",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			prettyFlag, _ := cmd.Flags().GetBool("pretty")
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
			cmd.SetContext(context.WithValue(cmd.Context(), common.PrettyJSONKey, prettyFlag))
			return nil
		},
	}
	cmd.PersistentFlags().Bool("pretty", false, "Pretty-print JSON output with indentation")
	return cmd
}
