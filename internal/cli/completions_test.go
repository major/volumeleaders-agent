package cli

import (
	"slices"
	"testing"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/trade"
)

// Completion tests avoid calling ConfigureCompletions (and therefore
// carapace.Gen) because carapace registers global cobra.OnInitialize callbacks
// whose internal double-checked locking has a data race under -race. The
// subprocess tests (TestJSONSchemaTreeCoversDomainLeafCommands et al.) build
// the real binary and implicitly verify that carapace integrates correctly.

func TestWalkCobraCommandsSkipsHiddenSubtrees(t *testing.T) {
	t.Parallel()
	root := &cobra.Command{Use: "root"}
	visible := &cobra.Command{Use: "visible", Run: func(*cobra.Command, []string) {}}
	hidden := &cobra.Command{Use: "hidden", Hidden: true}
	hiddenChild := &cobra.Command{Use: "child", Run: func(*cobra.Command, []string) {}}
	hidden.AddCommand(hiddenChild)
	root.AddCommand(visible, hidden)

	var visited []string
	walkCobraCommands(root, func(cmd *cobra.Command) {
		visited = append(visited, cmd.Name())
	})

	if !slices.Contains(visited, "visible") {
		t.Fatal("walkCobraCommands should visit visible commands")
	}
	if slices.Contains(visited, "hidden") {
		t.Fatal("walkCobraCommands should skip hidden commands")
	}
	if slices.Contains(visited, "child") {
		t.Fatal("walkCobraCommands should skip children of hidden commands")
	}
}

func TestIsLeafCommandIgnoresHiddenChildren(t *testing.T) {
	t.Parallel()

	t.Run("command with only hidden children is a leaf", func(t *testing.T) {
		t.Parallel()
		cmd := &cobra.Command{Use: "leaf", Run: func(*cobra.Command, []string) {}}
		hidden := &cobra.Command{Use: "hidden", Hidden: true}
		cmd.AddCommand(hidden)

		if !isLeafCommand(cmd) {
			t.Fatal("command with only hidden children should be a leaf")
		}
	})

	t.Run("command with visible children is not a leaf", func(t *testing.T) {
		t.Parallel()
		cmd := &cobra.Command{Use: "parent", Run: func(*cobra.Command, []string) {}}
		child := &cobra.Command{Use: "child", Run: func(*cobra.Command, []string) {}}
		cmd.AddCommand(child)

		if isLeafCommand(cmd) {
			t.Fatal("command with visible children should not be a leaf")
		}
	})

	t.Run("non-runnable command is not a leaf", func(t *testing.T) {
		t.Parallel()
		cmd := &cobra.Command{Use: "group"}

		if isLeafCommand(cmd) {
			t.Fatal("non-runnable command should not be a leaf")
		}
	})
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
