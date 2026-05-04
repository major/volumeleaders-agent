package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/testutil"
)

func TestRootPersistentPreRunStoresPrettyFlagInContext(t *testing.T) {
	t.Parallel()
	var got bool
	rootCmd := NewRootCmd("test")
	rootCmd.AddCommand(&cobra.Command{
		Use: "child",
		RunE: func(cmd *cobra.Command, _ []string) error {
			got, _ = cmd.Context().Value(common.PrettyJSONKey).(bool)
			return nil
		},
	})

	_, _, err := testutil.ExecuteCommand(t, rootCmd, context.Background(), "--pretty", "child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatal("expected pretty flag to be stored as true in command context")
	}
}

func TestRootSilenceErrorsPreventsCobraErrorOutput(t *testing.T) {
	t.Parallel()
	rootCmd := NewRootCmd("test")
	rootCmd.AddCommand(&cobra.Command{
		Use: "fail",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("boom")
		},
	})

	_, stderr, err := testutil.ExecuteCommand(t, rootCmd, context.Background(), "fail")
	if err == nil {
		t.Fatal("expected command error")
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestRootSilenceUsagePreventsUsageOutputOnError(t *testing.T) {
	t.Parallel()
	rootCmd := NewRootCmd("test")
	rootCmd.AddCommand(&cobra.Command{
		Use: "fail",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("boom")
		},
	})

	stdout, stderr, err := testutil.ExecuteCommand(t, rootCmd, context.Background(), "fail")
	if err == nil {
		t.Fatal("expected command error")
	}
	combinedOutput := stdout + stderr
	if strings.Contains(combinedOutput, "Usage:") {
		t.Fatalf("expected no usage output, got stdout=%q stderr=%q", stdout, stderr)
	}
}

// walkCommands recursively visits cmd and all of its subcommands, calling fn
// on each one. This lets structural tests assert properties across the entire
// command tree without duplicating the traversal logic.
func walkCommands(cmd *cobra.Command, fn func(*cobra.Command)) {
	fn(cmd)
	for _, sub := range cmd.Commands() {
		walkCommands(sub, fn)
	}
}

func TestRootTraverseChildrenEnabled(t *testing.T) {
	t.Parallel()
	cmd := NewRootCmd("test")
	if !cmd.TraverseChildren {
		t.Fatal("expected TraverseChildren = true on root command")
	}
}

func TestRootHasCommandGroups(t *testing.T) {
	t.Parallel()
	cmd := NewRootCmd("test")

	groups := cmd.Groups()
	if len(groups) != 6 {
		t.Fatalf("expected 6 command groups, got %d", len(groups))
	}

	expectedGroups := []struct {
		id string
	}{
		{"trading"},
		{"volume"},
		{"market"},
		{"alerts"},
		{"watchlists"},
		{"reference"},
	}
	for _, tt := range expectedGroups {
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()
			var found bool
			for _, g := range groups {
				if g.ID == tt.id {
					found = true
					if g.Title == "" {
						t.Fatalf("group %q has empty Title", tt.id)
					}
					break
				}
			}
			if !found {
				t.Fatalf("expected group %q not found", tt.id)
			}
		})
	}
}

func TestAllTopLevelCommandsHaveGroupID(t *testing.T) {
	t.Parallel()
	cmd := NewRootCmd("test")

	validGroups := []string{"trading", "volume", "market", "alerts", "watchlists", "reference"}
	builtins := []string{"help", "completion"}

	for _, sub := range cmd.Commands() {
		t.Run(sub.Name(), func(t *testing.T) {
			t.Parallel()
			if slices.Contains(builtins, sub.Name()) {
				return
			}
			if sub.GroupID == "" {
				t.Fatalf("command %q has empty GroupID", sub.Name())
			}
			if !slices.Contains(validGroups, sub.GroupID) {
				t.Fatalf("command %q has unexpected GroupID %q", sub.Name(), sub.GroupID)
			}
		})
	}
}

func TestAllCommandsHaveArgsValidator(t *testing.T) {
	t.Parallel()
	cmd := NewRootCmd("test")

	walkCommands(cmd, func(c *cobra.Command) {
		t.Run(c.CommandPath(), func(t *testing.T) {
			t.Parallel()
			if c.Args == nil {
				t.Fatalf("command %q has nil Args validator", c.CommandPath())
			}
		})
	})
}

func TestAllCommandsHaveLongDescription(t *testing.T) {
	t.Parallel()
	cmd := NewRootCmd("test")

	builtins := []string{"help", "completion"}

	walkCommands(cmd, func(c *cobra.Command) {
		t.Run(c.CommandPath(), func(t *testing.T) {
			t.Parallel()
			if slices.Contains(builtins, c.Name()) {
				return
			}
			if c.Long == "" {
				t.Fatalf("command %q has empty Long description", c.CommandPath())
			}
			if c.Long == c.Short {
				t.Fatalf("command %q Long equals Short; Long must add value", c.CommandPath())
			}
			for _, ch := range []string{"#", "`", "[", "]"} {
				if strings.Contains(c.Long, ch) {
					t.Fatalf("command %q Long contains Markdown character %q", c.CommandPath(), ch)
				}
			}
		})
	})
}

func TestAllLeafCommandsHaveExamples(t *testing.T) {
	t.Parallel()
	cmd := NewRootCmd("test")

	builtins := []string{"help", "completion", "bash", "fish", "powershell", "zsh", "config-keys", "env-vars"}

	walkCommands(cmd, func(c *cobra.Command) {
		t.Run(c.CommandPath(), func(t *testing.T) {
			if slices.Contains(builtins, c.Name()) || !c.Runnable() {
				return
			}
			if c.Example == "" {
				t.Fatalf("leaf command %q has empty Example", c.CommandPath())
			}
			if !strings.Contains(c.Example, "volumeleaders-agent ") {
				t.Fatalf("leaf command %q Example should include binary name, got %q", c.CommandPath(), c.Example)
			}
		})
	})
}

func TestKnownCommandAliases(t *testing.T) {
	t.Parallel()
	cmd := NewRootCmd("volumeleaders-agent")

	aliasesByCommand := map[string][]string{
		"volumeleaders-agent trade list":        {"ls"},
		"volumeleaders-agent alert configs":     {"ls"},
		"volumeleaders-agent alert delete":      {"rm"},
		"volumeleaders-agent alert create":      {"new"},
		"volumeleaders-agent watchlist configs": {"ls"},
		"volumeleaders-agent watchlist delete":  {"rm"},
		"volumeleaders-agent watchlist create":  {"new"},
	}

	walkCommands(cmd, func(c *cobra.Command) {
		expectedAliases, ok := aliasesByCommand[c.CommandPath()]
		if !ok {
			return
		}
		delete(aliasesByCommand, c.CommandPath())
		t.Run(c.CommandPath(), func(t *testing.T) {
			assertStringSet(t, c.Aliases, expectedAliases)
		})
	})
	if len(aliasesByCommand) > 0 {
		missing := make([]string, 0, len(aliasesByCommand))
		for commandPath := range aliasesByCommand {
			missing = append(missing, commandPath)
		}
		slices.Sort(missing)
		t.Fatalf("alias expectations did not match commands: %v", missing)
	}
}

func TestWorkflowRecoveryGuidanceIsDiscoverable(t *testing.T) {
	t.Parallel()
	cmd := NewRootCmd("volumeleaders-agent")

	for _, section := range []string{"RECOVERY PLAYBOOK", "COMMAND SEQUENCES", "Authentication failed", "Date validation failed", "Pagination validation failed"} {
		if !strings.Contains(cmd.Long, section) {
			t.Fatalf("root Long description missing %q", section)
		}
	}

	commands := map[string][]string{
		"volumeleaders-agent trade list":           {"PREREQUISITES:", "RECOVERY:", "NEXT STEPS:"},
		"volumeleaders-agent trade levels":         {"PREREQUISITES:", "RECOVERY:", "NEXT STEPS:"},
		"volumeleaders-agent trade level-touches":  {"PREREQUISITES:", "RECOVERY:", "NEXT STEPS:"},
		"volumeleaders-agent volume institutional": {"PREREQUISITES:", "RECOVERY:", "NEXT STEPS:"},
		"volumeleaders-agent market earnings":      {"PREREQUISITES:", "RECOVERY:", "NEXT STEPS:"},
	}
	walkCommands(cmd, func(c *cobra.Command) {
		sections, ok := commands[c.CommandPath()]
		if !ok {
			return
		}
		delete(commands, c.CommandPath())
		t.Run(c.CommandPath(), func(t *testing.T) {
			t.Parallel()
			for _, section := range sections {
				if !strings.Contains(c.Long, section) {
					t.Fatalf("command %q Long description missing %q", c.CommandPath(), section)
				}
			}
		})
	})
	if len(commands) > 0 {
		missing := make([]string, 0, len(commands))
		for commandPath := range commands {
			missing = append(missing, commandPath)
		}
		slices.Sort(missing)
		t.Fatalf("workflow recovery guidance expectations did not match commands: %v", missing)
	}
}

func TestNoAliasConflictsWithinParentScope(t *testing.T) {
	t.Parallel()
	cmd := NewRootCmd("test")

	walkCommands(cmd, func(parent *cobra.Command) {
		children := parent.Commands()
		if len(children) == 0 {
			return
		}
		t.Run(parent.CommandPath(), func(t *testing.T) {
			t.Parallel()
			seen := make(map[string]string) // name/alias -> owning command
			for _, child := range children {
				name := child.Name()
				if owner, ok := seen[name]; ok {
					t.Fatalf("duplicate name %q: used by both %q and %q", name, owner, child.CommandPath())
				}
				seen[name] = child.CommandPath()

				for _, alias := range child.Aliases {
					if owner, ok := seen[alias]; ok {
						t.Fatalf("alias %q of %q conflicts with %q", alias, child.CommandPath(), owner)
					}
					seen[alias] = child.CommandPath()
				}
			}
		})
	})
}

func TestNoShortFlagConflictsWithinCommand(t *testing.T) {
	t.Parallel()
	cmd := NewRootCmd("test")

	walkCommands(cmd, func(c *cobra.Command) {
		t.Run(c.CommandPath(), func(t *testing.T) {
			seen := make(map[string]string) // shorthand -> owning flag
			check := func(flags *pflag.FlagSet) {
				flags.VisitAll(func(flag *pflag.Flag) {
					if flag.Shorthand == "" {
						return
					}
					if owner, ok := seen[flag.Shorthand]; ok {
						t.Fatalf("command %q has duplicate shorthand -%s on --%s and --%s", c.CommandPath(), flag.Shorthand, owner, flag.Name)
					}
					seen[flag.Shorthand] = flag.Name
				})
			}

			check(c.LocalFlags())
			check(c.InheritedFlags())
		})
	})
}

func TestNoArgsCommandsRejectPositionalArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"trade sentiment", []string{"trade", "sentiment", "bogus"}},
		{"trade presets", []string{"trade", "presets", "bogus"}},
		{"trade preset-tickers", []string{"trade", "preset-tickers", "bogus"}},
		{"trade alerts", []string{"trade", "alerts", "--date", "2025-01-01", "bogus"}},
		{"trade cluster-alerts", []string{"trade", "cluster-alerts", "--date", "2025-01-01", "bogus"}},
		{"market snapshots", []string{"market", "snapshots", "bogus"}},
		{"market earnings", []string{"market", "earnings", "bogus"}},
		{"market exhaustion", []string{"market", "exhaustion", "bogus"}},
		{"alert configs", []string{"alert", "configs", "bogus"}},
		{"alert delete", []string{"alert", "delete", "bogus"}},
		{"alert create", []string{"alert", "create", "bogus"}},
		{"alert edit", []string{"alert", "edit", "bogus"}},
		{"watchlist configs", []string{"watchlist", "configs", "bogus"}},
		{"watchlist tickers", []string{"watchlist", "tickers", "bogus"}},
		{"watchlist delete", []string{"watchlist", "delete", "bogus"}},
		{"watchlist add-ticker", []string{"watchlist", "add-ticker", "bogus"}},
		{"watchlist create", []string{"watchlist", "create", "bogus"}},
		{"watchlist edit", []string{"watchlist", "edit", "bogus"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewRootCmd("test")
			_, _, err := testutil.ExecuteCommand(t, cmd, context.Background(), tt.args...)
			if err == nil {
				t.Fatalf("expected error for args %v, got nil", tt.args)
			}
		})
	}
}

func TestArbitraryArgsCommandsAcceptPositionalArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"trade list", []string{"trade", "list", "AAPL"}},
		{"trade clusters", []string{"trade", "clusters", "AAPL"}},
		{"trade cluster-bombs", []string{"trade", "cluster-bombs", "AAPL"}},
		{"trade levels", []string{"trade", "levels", "AAPL"}},
		{"trade level-touches", []string{"trade", "level-touches", "AAPL"}},
		{"volume institutional", []string{"volume", "institutional", "--date", "2025-01-01", "AAPL"}},
		{"volume ah-institutional", []string{"volume", "ah-institutional", "--date", "2025-01-01", "AAPL"}},
		{"volume total", []string{"volume", "total", "--date", "2025-01-01", "AAPL"}},
	}

	rejectMsgs := []string{"unknown command", "does not accept"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewRootCmd("test")
			_, _, err := testutil.ExecuteCommand(t, cmd, context.Background(), tt.args...)
			if err != nil {
				msg := err.Error()
				for _, reject := range rejectMsgs {
					if strings.Contains(msg, reject) {
						t.Fatalf("command rejected positional arg: %v", err)
					}
				}
			}
		})
	}
}

// buildBinary compiles the CLI binary into a temp directory and returns the
// path. Tests that need SetupCLI (which registers process-global cobra
// callbacks) use this to run the binary as a subprocess, avoiding data races
// in parallel tests.
func buildBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	binary := filepath.Join(dir, "volumeleaders-agent")
	cmd := exec.Command("go", "build", "-o", binary, "../../cmd/volumeleaders-agent")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return binary
}

func TestJSONSchemaTreeProducesValidJSON(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	out, err := exec.Command(binary, "--jsonschema=tree").CombinedOutput()
	if err != nil {
		t.Fatalf("--jsonschema=tree failed: %v\nOutput: %s", err, out)
	}

	var schemas []map[string]any
	if jsonErr := json.Unmarshal(out, &schemas); jsonErr != nil {
		t.Fatalf("output is not valid JSON array: %v\nOutput: %s", jsonErr, out)
	}
	if len(schemas) == 0 {
		t.Fatal("--jsonschema=tree produced empty schema array")
	}

	// Collect all titles from the schema array.
	titles := make(map[string]bool)
	for _, s := range schemas {
		if title, ok := s["title"].(string); ok {
			titles[title] = true
		}
	}

	expectedTitles := []string{
		"volumeleaders-agent trade list", "volumeleaders-agent trade sentiment",
		"volumeleaders-agent market earnings", "volumeleaders-agent volume institutional",
		"volumeleaders-agent alert create", "volumeleaders-agent watchlist create",
	}
	for _, expected := range expectedTitles {
		if !titles[expected] {
			t.Errorf("missing expected command title %q; found titles: %v", expected, titles)
		}
	}
}

func TestJSONSchemaTreeCoversDomainLeafCommands(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	schemas := jsonSchemaTree(t, binary)
	titles := schemaTitles(schemas)

	root := NewRootCmd("volumeleaders-agent")
	expectedTitles := make([]string, 0, 39)
	walkCommands(root, func(c *cobra.Command) {
		if !isDomainLeafCommand(c) {
			return
		}
		expectedTitles = append(expectedTitles, c.CommandPath())
	})
	slices.Sort(expectedTitles)
	if len(expectedTitles) != 39 {
		t.Fatalf("expected current command tree to have 39 domain leaf commands, got %d: %v", len(expectedTitles), expectedTitles)
	}

	missing := make([]string, 0)
	for _, title := range expectedTitles {
		if !titles[title] {
			missing = append(missing, title)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("--jsonschema=tree missing domain leaf command schemas: %v", missing)
	}
}

func TestJSONSchemaSubcommandProducesValidJSON(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	out, err := exec.Command(binary, "trade", "list", "--jsonschema").CombinedOutput()
	if err != nil {
		t.Fatalf("trade list --jsonschema failed: %v\nOutput: %s", err, out)
	}

	var schema map[string]any
	if jsonErr := json.Unmarshal(out, &schema); jsonErr != nil {
		t.Fatalf("output is not valid JSON object: %v\nOutput: %s", jsonErr, out)
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("schema missing 'properties' key")
	}

	expectedFlags := []string{"tickers", "start-date", "end-date", "min-dollars", "format", "start"}
	for _, flag := range expectedFlags {
		if _, exists := props[flag]; !exists {
			t.Errorf("trade list --jsonschema missing expected flag %q", flag)
		}
	}
}

func TestJSONSchemaSubcommandIncludesFlagUsabilityMetadata(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	out, err := exec.Command(binary, "trade", "list", "--jsonschema").CombinedOutput()
	if err != nil {
		t.Fatalf("trade list --jsonschema failed: %v\nOutput: %s", err, out)
	}

	var schema map[string]any
	if jsonErr := json.Unmarshal(out, &schema); jsonErr != nil {
		t.Fatalf("output is not valid JSON object: %v\nOutput: %s", jsonErr, out)
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("schema missing 'properties' key")
	}

	groupedFlags := map[string]string{
		"tickers":    "Input",
		"start-date": "Dates",
		"min-volume": "Ranges",
		"conditions": "Filters",
		"premarket":  "Sessions",
		"start":      "Pagination",
		"format":     "Output",
	}
	for flag, expectedGroup := range groupedFlags {
		t.Run("group "+flag, func(t *testing.T) {
			t.Parallel()
			flagSchema, ok := props[flag].(map[string]any)
			if !ok {
				t.Fatalf("flag %q schema is not an object", flag)
			}
			group, ok := flagSchema["x-structcli-group"].(string)
			if !ok {
				t.Fatalf("flag %q missing x-structcli-group", flag)
			}
			if group != expectedGroup {
				t.Fatalf("flag %q group = %q, want %q", flag, group, expectedGroup)
			}
		})
	}

	shortFlags := map[string]string{
		"tickers":    "t",
		"start-date": "s",
		"end-date":   "e",
		"days":       "d",
		"format":     "f",
	}
	for flag, expectedShort := range shortFlags {
		t.Run("short "+flag, func(t *testing.T) {
			t.Parallel()
			flagSchema, ok := props[flag].(map[string]any)
			if !ok {
				t.Fatalf("flag %q schema is not an object", flag)
			}
			short, ok := flagSchema["x-structcli-shorthand"].(string)
			if !ok {
				t.Fatalf("flag %q missing x-structcli-shorthand", flag)
			}
			if short != expectedShort {
				t.Fatalf("flag %q shorthand = %q, want %q", flag, short, expectedShort)
			}
		})
	}

	groups, ok := schema["x-structcli-groups"].(map[string]any)
	if !ok {
		t.Fatal("schema missing top-level x-structcli-groups map")
	}
	for _, expectedGroup := range []string{"Dates", "Filters", "Input", "Output", "Pagination", "Ranges", "Sessions"} {
		if _, ok := groups[expectedGroup]; !ok {
			t.Fatalf("schema groups missing %q: %v", expectedGroup, groups)
		}
	}
}

func TestJSONSchemaEnumValuesPresent(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)
	schema := commandJSONSchema(t, binary, "trade", "list", "--jsonschema")
	props := schemaProperties(t, schema)

	tests := []struct {
		flag string
		want []string
	}{
		{flag: "format", want: []string{"csv", "json", "tsv"}},
		{flag: "order-dir", want: []string{"asc", "desc"}},
		{flag: "group-by", want: []string{"day", "ticker", "ticker,day"}},
	}
	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			t.Parallel()
			flagSchema := schemaProperty(t, props, tt.flag)
			enums, ok := flagSchema["enum"].([]any)
			if !ok {
				t.Fatalf("flag %q missing enum array: %v", tt.flag, flagSchema)
			}
			got := make([]string, 0, len(enums))
			for _, enumValue := range enums {
				value, ok := enumValue.(string)
				if !ok {
					t.Fatalf("flag %q enum contains non-string value %v", tt.flag, enumValue)
				}
				got = append(got, value)
			}
			assertStringSet(t, got, tt.want)
		})
	}
}

func TestJSONSchemaRequiredFlags(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "alert delete", args: []string{"alert", "delete", "--jsonschema"}, want: []string{"key"}},
		{name: "alert create", args: []string{"alert", "create", "--jsonschema"}, want: []string{"name"}},
		{name: "alert edit", args: []string{"alert", "edit", "--jsonschema"}, want: []string{"key"}},
		{name: "watchlist delete", args: []string{"watchlist", "delete", "--jsonschema"}, want: []string{"key"}},
		{name: "watchlist add-ticker", args: []string{"watchlist", "add-ticker", "--jsonschema"}, want: []string{"ticker", "watchlist-key"}},
		{name: "watchlist create", args: []string{"watchlist", "create", "--jsonschema"}, want: []string{"name"}},
		{name: "watchlist edit", args: []string{"watchlist", "edit", "--jsonschema"}, want: []string{"key"}},
		{name: "trade alerts", args: []string{"trade", "alerts", "--jsonschema"}, want: []string{"date"}},
		{name: "trade cluster-alerts", args: []string{"trade", "cluster-alerts", "--jsonschema"}, want: []string{"date"}},
		{name: "trade preset-tickers", args: []string{"trade", "preset-tickers", "--jsonschema"}, want: []string{"preset"}},
		{name: "volume institutional", args: []string{"volume", "institutional", "--jsonschema"}, want: []string{"date"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			schema := commandJSONSchema(t, binary, tt.args...)
			requiredValues, ok := schema["required"].([]any)
			if !ok {
				t.Fatalf("schema for %q missing required array", tt.name)
			}
			got := make([]string, 0, len(requiredValues))
			for _, value := range requiredValues {
				required, ok := value.(string)
				if !ok {
					t.Fatalf("schema for %q contains non-string required value %v", tt.name, value)
				}
				got = append(got, required)
			}
			assertStringSet(t, got, tt.want)
		})
	}
}

func TestJSONSchemaDefaultValues(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	tests := []struct {
		name string
		args []string
		flag string
		want any
	}{
		{name: "trade list min dollars", args: []string{"trade", "list", "--jsonschema"}, flag: "min-dollars", want: float64(500000)},
		{name: "trade list group by", args: []string{"trade", "list", "--jsonschema"}, flag: "group-by", want: "ticker"},
		{name: "trade clusters min dollars", args: []string{"trade", "clusters", "--jsonschema"}, flag: "min-dollars", want: float64(10000000)},
		{name: "trade levels count", args: []string{"trade", "levels", "--jsonschema"}, flag: "trade-level-count", want: float64(10)},
		{name: "trade level touches length", args: []string{"trade", "level-touches", "--jsonschema"}, flag: "length", want: float64(50)},
		{name: "trade level touches rank", args: []string{"trade", "level-touches", "--jsonschema"}, flag: "trade-level-rank", want: float64(5)},
		{name: "watchlist tickers key", args: []string{"watchlist", "tickers", "--jsonschema"}, flag: "watchlist-key", want: float64(-1)},
		{name: "volume institutional length", args: []string{"volume", "institutional", "--jsonschema"}, flag: "length", want: float64(100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			schema := commandJSONSchema(t, binary, tt.args...)
			props := schemaProperties(t, schema)
			flagSchema := schemaProperty(t, props, tt.flag)
			got, ok := flagSchema["default"]
			if !ok {
				t.Fatalf("flag %q missing default in schema: %v", tt.flag, flagSchema)
			}
			if got != tt.want {
				t.Fatalf("flag %q default = %#v, want %#v", tt.flag, got, tt.want)
			}
		})
	}
}

func TestTradeCommandsWithoutUserSelectableLength(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	tests := []struct {
		name string
		args []string
	}{
		{name: "trade list", args: []string{"trade", "list", "--jsonschema"}},
		{name: "trade clusters", args: []string{"trade", "clusters", "--jsonschema"}},
		{name: "trade cluster bombs", args: []string{"trade", "cluster-bombs", "--jsonschema"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			schema := commandJSONSchema(t, binary, tt.args...)
			props := schemaProperties(t, schema)
			if _, ok := props["length"]; ok {
				t.Fatalf("schema for %q exposes length flag: %v", tt.name, props["length"])
			}
		})
	}
}

func TestJSONSchemaFlagGroupsAcrossCommands(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "alert create",
			args: []string{"alert", "create", "--jsonschema"},
			want: []string{"After-Hours Filters", "Basic", "Closing Filters", "Cluster Filters", "Total Filters", "Trade Filters"},
		},
		{
			name: "watchlist create",
			args: []string{"watchlist", "create", "--jsonschema"},
			want: []string{"Basic", "Filters", "Print Types", "Ranges", "RSI", "Sessions", "Venues"},
		},
		{
			name: "volume institutional",
			args: []string{"volume", "institutional", "--jsonschema"},
			want: []string{"Dates", "Input", "Output", "Pagination"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			schema := commandJSONSchema(t, binary, tt.args...)
			groups, ok := schema["x-structcli-groups"].(map[string]any)
			if !ok {
				t.Fatalf("schema for %q missing x-structcli-groups", tt.name)
			}
			got := make([]string, 0, len(groups))
			for group := range groups {
				got = append(got, group)
			}
			assertStringSet(t, got, tt.want)
		})
	}
}

func TestJSONSchemaRepresentativeFlagsHaveDescriptions(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	for _, args := range [][]string{
		{"trade", "list", "--jsonschema"},
		{"alert", "create", "--jsonschema"},
		{"watchlist", "create", "--jsonschema"},
	} {
		args := args
		t.Run(strings.Join(args[:len(args)-1], " "), func(t *testing.T) {
			t.Parallel()
			schema := commandJSONSchema(t, binary, args...)
			props := schemaProperties(t, schema)
			for flag, value := range props {
				flagSchema, ok := value.(map[string]any)
				if !ok {
					t.Fatalf("flag %q schema is not an object", flag)
				}
				description, ok := flagSchema["description"].(string)
				if !ok || description == "" {
					t.Fatalf("flag %q missing non-empty description in schema: %v", flag, flagSchema)
				}
			}
		})
	}
}

func TestJSONSchemaIncludesDiscreteValueEnums(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	tests := []struct {
		name string
		args []string
		flag string
		want []string
	}{
		{
			name: "trade list tri-state filter",
			args: []string{"trade", "list", "--jsonschema"},
			flag: "dark-pools",
			want: []string{"-1", "0", "1"},
		},
		{
			name: "trade list session filter",
			args: []string{"trade", "list", "--jsonschema"},
			flag: "premarket",
			want: []string{"-1", "0", "1"},
		},
		{
			name: "alert create ticker group",
			args: []string{"alert", "create", "--jsonschema"},
			flag: "ticker-group",
			want: []string{"AllTickers", "SelectedTickers"},
		},
		{
			name: "watchlist create security type",
			args: []string{"watchlist", "create", "--jsonschema"},
			flag: "security-type",
			want: []string{"-1", "1", "26", "4"},
		},
		{
			name: "watchlist create relative size",
			args: []string{"watchlist", "create", "--jsonschema"},
			flag: "min-relative-size",
			want: []string{"0", "5", "10", "25", "50", "100"},
		},
		{
			name: "watchlist create RSI toggle",
			args: []string{"watchlist", "create", "--jsonschema"},
			flag: "rsi-overbought-daily",
			want: []string{"-1", "0", "1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			schema := commandJSONSchema(t, binary, tc.args...)
			props := schemaProperties(t, schema)
			flagSchema, ok := props[tc.flag].(map[string]any)
			if !ok {
				t.Fatalf("flag %q schema is not an object", tc.flag)
			}

			got := schemaEnumValues(t, flagSchema)
			if !slices.Equal(got, sortedStrings(tc.want)) {
				t.Fatalf("flag %q enum = %v, want %v", tc.flag, got, sortedStrings(tc.want))
			}
		})
	}
}

func schemaEnumValues(t *testing.T, flagSchema map[string]any) []string {
	t.Helper()
	rawEnum, ok := flagSchema["enum"].([]any)
	if !ok {
		t.Fatalf("schema missing enum: %v", flagSchema)
	}

	values := make([]string, 0, len(rawEnum))
	for _, rawValue := range rawEnum {
		values = append(values, fmt.Sprint(rawValue))
	}
	slices.Sort(values)
	return values
}

func sortedStrings(values []string) []string {
	clone := slices.Clone(values)
	slices.Sort(clone)
	return clone
}

func TestOutputSchemaTreeProducesCommandContracts(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	out, err := exec.Command(binary, "outputschema").CombinedOutput()
	if err != nil {
		t.Fatalf("outputschema failed: %v\nOutput: %s", err, out)
	}

	var contracts []map[string]any
	if jsonErr := json.Unmarshal(out, &contracts); jsonErr != nil {
		t.Fatalf("outputschema output is not valid JSON array: %v\nOutput: %s", jsonErr, out)
	}
	if len(contracts) != 39 {
		t.Fatalf("expected 39 output contracts, got %d", len(contracts))
	}

	byCommand := make(map[string]map[string]any, len(contracts))
	for _, contract := range contracts {
		command, ok := contract["command"].(string)
		if !ok || command == "" {
			t.Fatalf("contract missing command string: %v", contract)
		}
		byCommand[command] = contract
	}
	for _, command := range []string{"report list", "report top-100-rank", "report leveraged-etfs", "report phantom-trades", "trade list", "market exhaustion", "alert configs", "watchlist create"} {
		if _, ok := byCommand[command]; !ok {
			t.Fatalf("missing output contract for %q", command)
		}
	}
}

func TestOutputSchemaSubcommandDescribesVariantsAndFields(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	out, err := exec.Command(binary, "outputschema", "trade", "list").CombinedOutput()
	if err != nil {
		t.Fatalf("outputschema trade list failed: %v\nOutput: %s", err, out)
	}

	var contract map[string]any
	if jsonErr := json.Unmarshal(out, &contract); jsonErr != nil {
		t.Fatalf("outputschema trade list is not valid JSON object: %v\nOutput: %s", jsonErr, out)
	}
	if got := contract["command"]; got != "trade list" {
		t.Fatalf("command = %v, want trade list", got)
	}

	schema := nestedMap(t, contract, "schema")
	if got := schema["type"]; got != "array" {
		t.Fatalf("schema.type = %v, want array", got)
	}
	items := nestedMap(t, schema, "items")
	if got := items["model"]; got != "TradeListRow" {
		t.Fatalf("schema.items.model = %v, want TradeListRow", got)
	}
	properties := nestedMap(t, items, "properties")
	for _, field := range []string{"Ticker", "Dollars", "DarkPool"} {
		if _, ok := properties[field]; !ok {
			t.Fatalf("TradeListRow schema missing field %q", field)
		}
	}

	variants, ok := contract["variants"].([]any)
	if !ok {
		t.Fatalf("contract variants missing or wrong type: %v", contract["variants"])
	}
	variantModels := make(map[string]bool)
	for _, raw := range variants {
		variant, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("variant is not an object: %v", raw)
		}
		variantSchema := nestedMap(t, variant, "schema")
		model := schemaModel(variantSchema)
		variantModels[model] = true
	}
	for _, model := range []string{"Trade", "TradeSummary"} {
		if !variantModels[model] {
			t.Fatalf("trade list contract missing variant model %q; got %v", model, variantModels)
		}
	}
}

func TestOutputSchemaMarketExhaustionProperties(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	out, err := exec.Command(binary, "outputschema", "market", "exhaustion").CombinedOutput()
	if err != nil {
		t.Fatalf("outputschema market exhaustion failed: %v\nOutput: %s", err, out)
	}

	var contract map[string]any
	if jsonErr := json.Unmarshal(out, &contract); jsonErr != nil {
		t.Fatalf("outputschema market exhaustion is not valid JSON object: %v\nOutput: %s", jsonErr, out)
	}
	schema := nestedMap(t, contract, "schema")
	if got := schema["model"]; got != "MarketExhaustion" {
		t.Fatalf("model = %v, want MarketExhaustion", got)
	}
	properties := nestedMap(t, schema, "properties")
	for _, field := range []string{"date_key", "rank", "rank_30d", "rank_90d", "rank_365d"} {
		fieldSchema := nestedMap(t, properties, field)
		if got := fieldSchema["type"]; got != "integer" {
			t.Fatalf("field %q type = %v, want integer", field, got)
		}
	}
}

func TestOutputSchemaUnknownCommandFails(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	cmd := exec.Command(binary, "outputschema", "bogus")
	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected exit error, got: %v", err)
	}
}

func nestedMap(t *testing.T, parent map[string]any, key string) map[string]any {
	t.Helper()
	child, ok := parent[key].(map[string]any)
	if !ok {
		t.Fatalf("%q missing or not an object: %v", key, parent[key])
	}
	return child
}

func schemaModel(schema map[string]any) string {
	if model, ok := schema["model"].(string); ok {
		return model
	}
	items, ok := schema["items"].(map[string]any)
	if !ok {
		return ""
	}
	model, _ := items["model"].(string)
	return model
}

func TestMCPToolsListExposesLeafCommands(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	cmd := exec.Command(binary, "--mcp")
	cmd.Stdin = strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--mcp tools/list failed: %v\nOutput: %s", err, out)
	}

	var response struct {
		Result struct {
			Tools []struct {
				Name        string         `json:"name"`
				Description string         `json:"description"`
				InputSchema map[string]any `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if jsonErr := json.Unmarshal(out, &response); jsonErr != nil {
		t.Fatalf("MCP response is not valid JSON: %v\nOutput: %s", jsonErr, out)
	}
	if response.Error != nil {
		t.Fatalf("MCP tools/list returned error: %s\nOutput: %s", response.Error.Message, out)
	}

	toolSchemas := make(map[string]map[string]any, len(response.Result.Tools))
	for _, tool := range response.Result.Tools {
		if tool.Description == "" {
			t.Fatalf("tool %q has empty description", tool.Name)
		}
		toolSchemas[tool.Name] = tool.InputSchema
	}

	for _, expected := range []string{"trade-list", "trade-sentiment", "volume-institutional", "market-snapshots", "watchlist-configs"} {
		if _, ok := toolSchemas[expected]; !ok {
			t.Fatalf("MCP tools/list missing %q; got %v", expected, mapsKeys(toolSchemas))
		}
	}
	for _, parent := range []string{"trade", "volume", "market", "alert", "watchlist"} {
		if _, ok := toolSchemas[parent]; ok {
			t.Fatalf("MCP tools/list exposed non-leaf parent command %q", parent)
		}
	}
	for _, completionTool := range []string{"completion-bash", "completion-fish", "completion-powershell", "completion-zsh"} {
		if _, ok := toolSchemas[completionTool]; ok {
			t.Fatalf("MCP tools/list exposed shell completion tool %q", completionTool)
		}
	}

	tradeListSchema := toolSchemas["trade-list"]
	props, ok := tradeListSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("trade-list inputSchema missing properties: %v", tradeListSchema)
	}
	for _, expectedFlag := range []string{"tickers", "days", "format", "start"} {
		if _, ok := props[expectedFlag]; !ok {
			t.Fatalf("trade-list inputSchema missing flag %q", expectedFlag)
		}
	}
	if _, ok := props["length"]; ok {
		t.Fatalf("trade-list inputSchema should not expose removed flag \"length\"")
	}
}

func TestMCPToolsCallReportsStructuredValidationError(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	cmd := exec.Command(binary, "--mcp")
	cmd.Stdin = strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"volume-institutional","arguments":{"format":"json"}}}` + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--mcp tools/call failed: %v\nOutput: %s", err, out)
	}

	var response struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if jsonErr := json.Unmarshal(out, &response); jsonErr != nil {
		t.Fatalf("MCP response is not valid JSON: %v\nOutput: %s", jsonErr, out)
	}
	if response.Error != nil {
		t.Fatalf("MCP tools/call returned protocol error: %s\nOutput: %s", response.Error.Message, out)
	}
	if !response.Result.IsError {
		t.Fatalf("expected tool call to return IsError=true; output: %s", out)
	}
	if len(response.Result.Content) == 0 || !strings.Contains(response.Result.Content[0].Text, "required flag") {
		t.Fatalf("expected structured required flag error, got: %s", out)
	}
}

func mapsKeys[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func TestHelpOutputDisplaysFlagGroups(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	tests := []struct {
		name           string
		args           []string
		expectedGroups []string
	}{
		{
			name:           "trade list help",
			args:           []string{"trade", "list", "--help"},
			expectedGroups: []string{"Dates Flags:", "Filters Flags:", "Input Flags:", "Output Flags:", "Pagination Flags:", "Ranges Flags:", "Sessions Flags:"},
		},
		{
			name:           "alert create help",
			args:           []string{"alert", "create", "--help"},
			expectedGroups: []string{"After-Hours Filters Flags:", "Basic Flags:", "Closing Filters Flags:", "Cluster Filters Flags:", "Total Filters Flags:", "Trade Filters Flags:"},
		},
		{
			name:           "watchlist create help",
			args:           []string{"watchlist", "create", "--help"},
			expectedGroups: []string{"Basic Flags:", "Filters Flags:", "Print Types Flags:", "Ranges Flags:", "RSI Flags:", "Sessions Flags:", "Venues Flags:"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, err := exec.Command(binary, tc.args...).CombinedOutput()
			if err != nil {
				t.Fatalf("%v failed: %v\nOutput: %s", tc.args, err, out)
			}
			helpText := string(out)
			for _, expected := range tc.expectedGroups {
				if !strings.Contains(helpText, expected) {
					t.Fatalf("help output missing %q\nOutput: %s", expected, helpText)
				}
			}
		})
	}
}

func TestHelpTopicsAreAvailableAsReferenceCommands(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	helpOut, err := exec.Command(binary, "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("--help failed: %v\nOutput: %s", err, helpOut)
	}
	helpText := string(helpOut)
	for _, expected := range []string{"Reference:", "config-keys List all configuration file keys", "env-vars    List all environment variable bindings"} {
		if !strings.Contains(helpText, expected) {
			t.Errorf("--help missing %q\nOutput: %s", expected, helpText)
		}
	}

	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "environment variables topic",
			args:     []string{"env-vars"},
			contains: "Environment Variables",
		},
		{
			name:     "configuration keys topic",
			args:     []string{"config-keys"},
			contains: "Configuration Keys",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, cmdErr := exec.Command(binary, tc.args...).CombinedOutput()
			if cmdErr != nil {
				t.Fatalf("%v failed: %v\nOutput: %s", tc.args, cmdErr, out)
			}
			if !strings.Contains(string(out), tc.contains) {
				t.Errorf("%v output missing %q\nOutput: %s", tc.args, tc.contains, out)
			}
		})
	}
}

func TestStructuredErrorExitCodes(t *testing.T) {
	t.Parallel()
	binary := buildBinary(t)

	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{
			name:     "unknown flag returns exit code 12",
			args:     []string{"trade", "list", "--nonexistent-flag"},
			wantCode: 12,
		},
		{
			name:     "missing required flag returns exit code 10",
			args:     []string{"alert", "delete"},
			wantCode: 10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(binary, tc.args...)
			err := cmd.Run()
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("expected exit error, got: %v", err)
			}
			if exitErr.ExitCode() != tc.wantCode {
				t.Errorf("exit code = %d, want %d", exitErr.ExitCode(), tc.wantCode)
			}
		})
	}
}
func isDomainLeafCommand(cmd *cobra.Command) bool {
	if !cmd.Runnable() || len(cmd.Commands()) > 0 {
		return false
	}
	for _, name := range []string{"help", "completion", "bash", "fish", "powershell", "zsh", "config-keys", "env-vars", "outputschema"} {
		if cmd.Name() == name {
			return false
		}
	}
	return true
}

func jsonSchemaTree(t *testing.T, binary string) []map[string]any {
	t.Helper()
	out, err := exec.Command(binary, "--jsonschema=tree").CombinedOutput()
	if err != nil {
		t.Fatalf("--jsonschema=tree failed: %v\nOutput: %s", err, out)
	}
	var schemas []map[string]any
	if jsonErr := json.Unmarshal(out, &schemas); jsonErr != nil {
		t.Fatalf("output is not valid JSON array: %v\nOutput: %s", jsonErr, out)
	}
	return schemas
}

func commandJSONSchema(t *testing.T, binary string, args ...string) map[string]any {
	t.Helper()
	out, err := exec.Command(binary, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("%v failed: %v\nOutput: %s", args, err, out)
	}
	var schema map[string]any
	if jsonErr := json.Unmarshal(out, &schema); jsonErr != nil {
		t.Fatalf("%v output is not valid JSON object: %v\nOutput: %s", args, jsonErr, out)
	}
	return schema
}

func schemaTitles(schemas []map[string]any) map[string]bool {
	titles := make(map[string]bool, len(schemas))
	for _, schema := range schemas {
		title, ok := schema["title"].(string)
		if !ok {
			continue
		}
		titles[title] = true
	}
	return titles
}

func schemaProperties(t *testing.T, schema map[string]any) map[string]any {
	t.Helper()
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema missing properties map: %v", schema)
	}
	return props
}

func schemaProperty(t *testing.T, props map[string]any, flag string) map[string]any {
	t.Helper()
	value, ok := props[flag]
	if !ok {
		t.Fatalf("schema properties missing flag %q", flag)
	}
	flagSchema, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("flag %q schema is not an object: %v", flag, value)
	}
	return flagSchema
}

func assertStringSet(t *testing.T, got, want []string) {
	t.Helper()
	gotCopy := slices.Clone(got)
	wantCopy := slices.Clone(want)
	slices.Sort(gotCopy)
	slices.Sort(wantCopy)
	if !slices.Equal(gotCopy, wantCopy) {
		t.Fatalf("values = %v, want %v", gotCopy, wantCopy)
	}
}
