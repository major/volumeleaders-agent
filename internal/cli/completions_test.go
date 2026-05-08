package cli

import (
	"slices"
	"testing"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/trade"
)

func TestConfigureCompletionsAddsCarapaceCommand(t *testing.T) {
	t.Parallel()
	root := NewRootCmd("test")
	SetupCLI(root)

	var found bool
	for _, sub := range root.Commands() {
		if sub.Name() == "_carapace" {
			found = true
			if !sub.Hidden {
				t.Fatal("_carapace command should be hidden")
			}
			break
		}
	}
	if !found {
		t.Fatal("expected _carapace command on root after SetupCLI")
	}
}

func TestCarapaceCommandExcludedFromSchemaTree(t *testing.T) {
	t.Parallel()
	root := NewRootCmd("test")
	SetupCLI(root)

	// walkCobraCommands skips hidden subtrees, so no _carapace descendant
	// should appear in schema/MCP-visible commands.
	var exposed []string
	walkCobraCommands(root, func(cmd *cobra.Command) {
		exposed = append(exposed, cmd.Name())
	})
	for _, name := range exposed {
		if name == "_carapace" || name == "spec" || name == "style" {
			t.Fatalf("walkCobraCommands should skip hidden _carapace subtree, but found %q", name)
		}
	}
}

func TestTradeListFlagCompletions(t *testing.T) {
	t.Parallel()
	root := NewRootCmd("test")
	SetupCLI(root)

	// Find trade list once for all subtests.
	var tradeList *cobra.Command
	walkCobraCommands(root, func(cmd *cobra.Command) {
		if cmd.Name() == "list" && cmd.Parent() != nil && cmd.Parent().Name() == "trade" {
			tradeList = cmd
		}
	})
	if tradeList == nil {
		t.Fatal("could not find 'trade list' command")
	}

	t.Run("format flag has enum completions", func(t *testing.T) {
		t.Parallel()
		formatFlag := tradeList.Flags().Lookup("format")
		if formatFlag == nil {
			t.Fatal("trade list missing --format flag")
		}
		enumValues := common.FlagEnum(formatFlag)
		if len(enumValues) == 0 {
			t.Fatal("--format flag has no enum annotation; completions would be empty")
		}
		if !slices.Contains(enumValues, "json") || !slices.Contains(enumValues, "csv") {
			t.Fatalf("--format enum values = %v, want at least json and csv", enumValues)
		}
	})

	t.Run("preset flag exists with completable names", func(t *testing.T) {
		t.Parallel()
		if tradeList.Flags().Lookup("preset") == nil {
			t.Fatal("trade list missing --preset flag")
		}
		names := trade.PresetNames()
		if len(names) == 0 {
			t.Fatal("PresetNames() returned empty slice")
		}
	})
}

func TestRegisterFlagCompletionsNoEnumFlags(t *testing.T) {
	t.Parallel()

	// A bare command with no flags should not panic.
	cmd := &cobra.Command{Use: "bare", Run: func(*cobra.Command, []string) {}}
	carapace.Gen(cmd)
	registerFlagCompletions(cmd)
}
