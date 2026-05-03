package cli

import (
	"context"
	"encoding/json"
	"errors"
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
	if len(groups) != 5 {
		t.Fatalf("expected 5 command groups, got %d", len(groups))
	}

	expectedGroups := []struct {
		id string
	}{
		{"trading"},
		{"volume"},
		{"market"},
		{"alerts"},
		{"watchlists"},
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

	validGroups := []string{"trading", "volume", "market", "alerts", "watchlists"}
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

func TestRunnableCommandsHaveExamples(t *testing.T) {
	t.Parallel()
	cmd := NewRootCmd("test")

	builtins := []string{"help", "completion", "config-keys", "env-vars"}

	walkCommands(cmd, func(c *cobra.Command) {
		t.Run(c.CommandPath(), func(t *testing.T) {
			t.Parallel()
			if slices.Contains(builtins, c.Name()) || !c.Runnable() {
				return
			}
			if c.Example == "" {
				t.Fatalf("command %q has empty Example; runnable commands need copy-paste guidance", c.CommandPath())
			}
			if !strings.Contains(c.Example, "volumeleaders-agent ") {
				t.Fatalf("command %q Example should include the binary name, got %q", c.CommandPath(), c.Example)
			}
		})
	})
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
			t.Parallel()
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

	expectedFlags := []string{"tickers", "start-date", "end-date", "min-dollars", "format", "length"}
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
		"length":     "Pagination",
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
		"length":     "l",
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
	for _, expectedFlag := range []string{"tickers", "days", "format", "length"} {
		if _, ok := props[expectedFlag]; !ok {
			t.Fatalf("trade-list inputSchema missing flag %q", expectedFlag)
		}
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
