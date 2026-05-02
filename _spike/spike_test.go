package spike

import (
	"bytes"
	"context"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/leodido/structcli"
	"github.com/spf13/cobra"
)

type changedOptions struct {
	StartDate string `flag:"start-date" flagdescr:"Start date"`
}

type defaultOptions struct {
	Length int `flag:"length" flagdescr:"Number of rows"`
}

type emptyDefaultOptions struct {
	Length int `flag:"length" flagdescr:"Number of rows" default:""`
}

type explicitDefaultOptions struct {
	Length int `flag:"length" flagdescr:"Number of rows" default:"25"`
}

type embeddedDateOptions struct {
	StartDate string `flag:"start-date" flagdescr:"Start date"`
}

type parentOptions struct {
	embeddedDateOptions
}

func TestA1ChangedReportsExplicitFlagsAfterBindAndExecuteC(t *testing.T) {
	for _, tc := range []struct {
		name        string
		args        []string
		wantChanged bool
	}{
		{name: "unset flag remains unchanged", wantChanged: false},
		{name: "set flag is changed", args: []string{"--start-date", "2026-05-02"}, wantChanged: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			structcli.Reset()
			opts := &changedOptions{}
			var gotChanged bool
			cmd := &cobra.Command{
				Use:  "spike",
				Args: cobra.NoArgs,
				RunE: func(cmd *cobra.Command, _ []string) error {
					gotChanged = cmd.Flags().Changed("start-date")
					return nil
				},
			}
			mustBind(t, cmd, opts)
			cmd.SetArgs(tc.args)

			_, err := structcli.ExecuteC(cmd)
			if err != nil {
				t.Fatalf("ExecuteC failed: %v", err)
			}
			if gotChanged != tc.wantChanged {
				t.Fatalf("Changed returned %v, want %v", gotChanged, tc.wantChanged)
			}
		})
	}
}

func TestA2StructcliVersionCompilesInGoModule(t *testing.T) {
	structcli.Reset()
	if info, ok := debug.ReadBuildInfo(); ok {
		t.Logf("go version: %s", runtime.Version())
		t.Logf("module go version: %s", info.GoVersion)
		for _, dep := range info.Deps {
			if dep.Path == "github.com/leodido/structcli" || dep.Path == "github.com/spf13/pflag" {
				t.Logf("dependency %s %s", dep.Path, dep.Version)
			}
		}
	}
	opts := &changedOptions{}
	cmd := &cobra.Command{Use: "spike", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
	mustBind(t, cmd, opts)
	cmd.SetArgs(nil)

	if _, err := structcli.ExecuteC(cmd); err != nil {
		t.Fatalf("structcli v0.17.0 did not compile and execute in the spike module: %v", err)
	}
}

func TestA3FlagTagOverridesEmbeddedStructFieldPath(t *testing.T) {
	structcli.Reset()
	opts := &parentOptions{}
	cmd := &cobra.Command{Use: "spike", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
	mustBind(t, cmd, opts)

	if cmd.Flags().Lookup("start-date") == nil {
		t.Fatalf("expected embedded field flag start-date to exist")
	}
	if cmd.Flags().Lookup("embeddeddates.start-date") != nil || cmd.Flags().Lookup("embeddedDateOptions.start-date") != nil {
		t.Fatalf("embedded flag tag was unexpectedly prefixed")
	}
}

func TestA4PresetFieldValuesBecomePflagDefaults(t *testing.T) {
	for _, tc := range []struct {
		name    string
		opts    any
		wantDef string
	}{
		{name: "no default tag uses preset field", opts: &defaultOptions{Length: 100}, wantDef: "100"},
		{name: "empty default tag uses preset field", opts: &emptyDefaultOptions{Length: 100}, wantDef: "100"},
		{name: "nonempty default tag wins for DefValue", opts: &explicitDefaultOptions{Length: 100}, wantDef: "25"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			structcli.Reset()
			cmd := &cobra.Command{Use: "spike", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
			mustBind(t, cmd, tc.opts)

			flag := cmd.Flags().Lookup("length")
			if flag == nil {
				t.Fatalf("expected length flag to exist")
			}
			if flag.DefValue != tc.wantDef {
				t.Fatalf("DefValue = %q, want %q", flag.DefValue, tc.wantDef)
			}
		})
	}
}

func TestA5PersistentPreRunEFiresForChildCommands(t *testing.T) {
	structcli.Reset()
	called := false
	root := &cobra.Command{
		Use:              "spike",
		TraverseChildren: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			called = true
			return nil
		},
	}
	child := &cobra.Command{Use: "child", RunE: func(_ *cobra.Command, _ []string) error { return nil }}
	root.AddCommand(child)
	mustBind(t, child, &defaultOptions{})
	root.SetArgs([]string{"child"})

	if _, err := structcli.ExecuteC(root); err != nil {
		t.Fatalf("ExecuteC failed: %v", err)
	}
	if !called {
		t.Fatalf("root PersistentPreRunE was not called for child command")
	}
}

func TestA6SilenceErrorsWithFlagErrorsDoesNotDoublePrintUnknownFlag(t *testing.T) {
	structcli.Reset()
	var stderr bytes.Buffer
	root := &cobra.Command{
		Use:           "spike",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          func(_ *cobra.Command, _ []string) error { return nil },
	}
	root.SetErr(&stderr)
	if err := structcli.Setup(root, structcli.WithFlagErrors()); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	root.SetArgs([]string{"--unknown"})

	_, err := structcli.ExecuteC(root)
	if err == nil {
		t.Fatalf("expected unknown flag error")
	}
	if count := strings.Count(stderr.String(), "unknown flag"); count > 1 {
		t.Fatalf("unknown flag error printed %d times to stderr: %q", count, stderr.String())
	}
}

func TestA7ExecuteCReplacesCobraExecuteForBoundSubcommands(t *testing.T) {
	structcli.Reset()
	type tradeOptions struct {
		Ticker string `flag:"ticker" flagdescr:"Ticker symbol"`
		Length int    `flag:"length" flagdescr:"Number of rows"`
	}
	opts := &tradeOptions{Length: 10}
	called := false
	root := &cobra.Command{Use: "spike"}
	child := &cobra.Command{
		Use:  "trade",
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			called = true
			if opts.Ticker != "NVDA" {
				t.Fatalf("Ticker = %q, want NVDA", opts.Ticker)
			}
			if opts.Length != 25 {
				t.Fatalf("Length = %d, want 25", opts.Length)
			}
			return nil
		},
	}
	root.AddCommand(child)
	mustBind(t, child, opts)
	root.SetArgs([]string{"trade", "--ticker", "NVDA", "--length", "25"})

	executed, err := structcli.ExecuteC(root)
	if err != nil {
		t.Fatalf("ExecuteC failed: %v", err)
	}
	if !called {
		t.Fatalf("subcommand RunE was not called")
	}
	if executed != child {
		t.Fatalf("executed command = %q, want trade", executed.Name())
	}
}

func mustBind(t *testing.T, cmd *cobra.Command, opts any) {
	t.Helper()
	if err := structcli.Bind(cmd, opts); err != nil {
		t.Fatalf("Bind failed: %v", err)
	}
	cmd.SetContext(context.Background())
}
