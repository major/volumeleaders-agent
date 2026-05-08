package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/trade"
)

// ConfigureCompletions registers carapace-powered shell completions for all
// commands. It walks the command tree and registers ActionValues for every flag
// that carries enum annotations, plus preset name completions for --preset.
//
// Call this after SetupCLI. It is separate because carapace registers global
// cobra.OnInitialize callbacks whose internal double-checked locking has a
// data race under Go's race detector when multiple command trees exist in
// parallel tests.
func ConfigureCompletions(root *cobra.Command) {
	carapace.Gen(root)
	walkCobraCommands(root, func(cmd *cobra.Command) {
		registerFlagCompletions(cmd)
	})
}

// registerFlagCompletions inspects a single command's flags and registers
// carapace completions for flags that have known enum values or preset names.
func registerFlagCompletions(cmd *cobra.Command) {
	completions := carapace.ActionMap{}

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if values := common.FlagEnum(flag); len(values) > 0 {
			completions[flag.Name] = carapace.ActionValues(values...)
		}
	})

	if flag := cmd.Flags().Lookup("preset"); flag != nil {
		completions["preset"] = carapace.ActionValues(trade.PresetNames()...)
	}

	if len(completions) > 0 {
		carapace.Gen(cmd).FlagCompletion(completions)
	}
}
