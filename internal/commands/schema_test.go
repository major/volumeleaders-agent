package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	cli "github.com/urfave/cli/v3"
)

// testSchemaApp builds a small CLI app for schema testing with nested commands
// and representative flag types. It stays independent from NewApp so failures
// point at schema behavior rather than the production command tree.
func testSchemaApp() *cli.Command {
	return &cli.Command{
		Name:  "test-app",
		Usage: "A test application",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Enable verbose output"},
			&cli.StringFlag{Name: "output", Usage: "Output format", Value: "json"},
		},
		Commands: []*cli.Command{
			{
				Name:  "account",
				Usage: "Account operations",
				Commands: []*cli.Command{
					{
						Name:      "list",
						Usage:     "List all accounts",
						UsageText: "test-app account list --format table",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "format", Usage: "Output format", Value: "table"},
							&cli.BoolFlag{Name: "all", Usage: "Show all accounts", Required: true},
						},
					},
				},
			},
			{
				Name:  "quote",
				Usage: "Get stock quotes",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "symbol", Aliases: []string{"s", "ticker"}, Usage: "Stock symbol", Required: true},
					&cli.IntFlag{Name: "count", Usage: "Number of quotes", Value: 10},
					&cli.FloatFlag{Name: "threshold", Usage: "Price threshold", Value: 0.5},
				},
			},
		},
	}
}

func decodeSchemaOutput(t *testing.T, data []byte) SchemaOutput {
	t.Helper()
	var schema SchemaOutput
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("unmarshal schema output: %v", err)
	}
	return schema
}

func TestSchemaCommandFullOutput(t *testing.T) {
	t.Parallel()

	app := testSchemaApp()
	var buf bytes.Buffer
	schemaCmd := SchemaCommand(app, &buf)

	if err := schemaCmd.Run(context.Background(), []string{"schema"}); err != nil {
		t.Fatalf("run schema command: %v", err)
	}
	schema := decodeSchemaOutput(t, buf.Bytes())

	if got, want := len(schema.Commands), 3; got != want {
		t.Fatalf("expected %d commands, got %d", want, got)
	}
	for _, path := range []string{"account", "account list", "quote"} {
		if _, ok := schema.Commands[path]; !ok {
			t.Fatalf("missing command path %q", path)
		}
	}
	if got, want := schema.Commands["account list"].Description, "List all accounts"; got != want {
		t.Fatalf("expected description %q, got %q", want, got)
	}
	if got, want := len(schema.GlobalFlags), 3; got != want {
		t.Fatalf("expected %d global flags, got %d", want, got)
	}
	if got := schema.GlobalFlags["--output"].Default; got != "json" {
		t.Fatalf("expected output default json, got %v", got)
	}
	if _, ok := schema.GlobalFlags["-v"]; !ok {
		t.Fatalf("missing short global flag alias")
	}
}

func TestSchemaCommandFlagTypes(t *testing.T) {
	t.Parallel()

	app := testSchemaApp()
	var buf bytes.Buffer
	schemaCmd := SchemaCommand(app, &buf)

	if err := schemaCmd.Run(context.Background(), []string{"schema"}); err != nil {
		t.Fatalf("run schema command: %v", err)
	}
	schema := decodeSchemaOutput(t, buf.Bytes())

	symbolFlag := schema.Commands["quote"].Flags["--symbol"]
	if symbolFlag.Type != "string" || !symbolFlag.Required || symbolFlag.Default != "" || symbolFlag.Description != "Stock symbol" {
		t.Fatalf("unexpected string flag schema: %+v", symbolFlag)
	}
	for _, alias := range []string{"-s", "--ticker"} {
		if got := schema.Commands["quote"].Flags[alias]; got != symbolFlag {
			t.Fatalf("expected alias %q to match symbol flag schema, got %+v", alias, got)
		}
	}

	countFlag := schema.Commands["quote"].Flags["--count"]
	if countFlag.Type != "int" || countFlag.Required || countFlag.Default != float64(10) || countFlag.Description != "Number of quotes" {
		t.Fatalf("unexpected int flag schema: %+v", countFlag)
	}

	thresholdFlag := schema.Commands["quote"].Flags["--threshold"]
	if thresholdFlag.Type != "float" || thresholdFlag.Required || thresholdFlag.Default != 0.5 || thresholdFlag.Description != "Price threshold" {
		t.Fatalf("unexpected float flag schema: %+v", thresholdFlag)
	}

	allFlag := schema.Commands["account list"].Flags["--all"]
	if allFlag.Type != "bool" || !allFlag.Required || allFlag.Default != false || allFlag.Description != "Show all accounts" {
		t.Fatalf("unexpected bool flag schema: %+v", allFlag)
	}
}

func TestSchemaCommandFilterByCommand(t *testing.T) {
	t.Parallel()

	app := testSchemaApp()
	var buf bytes.Buffer
	schemaCmd := SchemaCommand(app, &buf)

	if err := schemaCmd.Run(context.Background(), []string{"schema", "--command", "account list"}); err != nil {
		t.Fatalf("run schema command: %v", err)
	}
	schema := decodeSchemaOutput(t, buf.Bytes())

	if got, want := len(schema.Commands), 1; got != want {
		t.Fatalf("expected %d command, got %d", want, got)
	}
	cmd, ok := schema.Commands["account list"]
	if !ok {
		t.Fatalf("filtered command missing")
	}
	if got, want := cmd.Examples, []string{"test-app account list --format table"}; strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("expected examples %v, got %v", want, got)
	}
	if got, want := len(schema.GlobalFlags), 3; got != want {
		t.Fatalf("expected %d global flags, got %d", want, got)
	}
}

func TestSchemaCommandNewAppRegistration(t *testing.T) {
	output := captureStdout(t, func() {
		// NewApp wires the production schema command to os.Stdout, so build the
		// app after stdout capture is installed.
		app := NewApp("test")
		var schemaCmd *cli.Command
		for _, cmd := range app.Commands {
			if cmd.Name == "schema" {
				schemaCmd = cmd
				break
			}
		}
		if schemaCmd == nil {
			t.Fatalf("schema command is not registered")
		}

		if err := schemaCmd.Run(context.Background(), []string{"schema", "--command", "schema"}); err != nil {
			t.Fatalf("run schema command: %v", err)
		}
	})
	schema := decodeSchemaOutput(t, []byte(output))

	cmd, ok := schema.Commands["schema"]
	if !ok {
		t.Fatalf("schema output does not include the schema command")
	}
	if _, ok := cmd.Flags["--command"]; !ok {
		t.Fatalf("schema command output is missing --command flag")
	}
	if _, ok := schema.GlobalFlags["--pretty"]; !ok {
		t.Fatalf("schema output is missing --pretty global flag")
	}
}

func TestSchemaCommandFilterNotFound(t *testing.T) {
	t.Parallel()

	app := testSchemaApp()
	var buf bytes.Buffer
	schemaCmd := SchemaCommand(app, &buf)

	err := schemaCmd.Run(context.Background(), []string{"schema", "--command", "nonexistent"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "nonexistent") || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSchemaCommandEmptyApp(t *testing.T) {
	t.Parallel()

	app := &cli.Command{Name: "empty", Usage: "An empty application"}
	var buf bytes.Buffer
	schemaCmd := SchemaCommand(app, &buf)

	if err := schemaCmd.Run(context.Background(), []string{"schema"}); err != nil {
		t.Fatalf("run schema command: %v", err)
	}
	schema := decodeSchemaOutput(t, buf.Bytes())

	if len(schema.Commands) != 0 {
		t.Fatalf("expected no commands, got %d", len(schema.Commands))
	}
	if len(schema.GlobalFlags) != 0 {
		t.Fatalf("expected no global flags, got %d", len(schema.GlobalFlags))
	}
}

func TestSchemaCommandNestedCommandPath(t *testing.T) {
	t.Parallel()

	app := &cli.Command{
		Name: "app",
		Commands: []*cli.Command{
			{
				Name:  "order",
				Usage: "Order operations",
				Commands: []*cli.Command{
					{
						Name:  "place",
						Usage: "Place an order",
						Commands: []*cli.Command{
							{
								Name:  "equity",
								Usage: "Place an equity order",
								Flags: []cli.Flag{
									&cli.StringFlag{Name: "symbol", Usage: "Stock symbol", Required: true},
								},
							},
						},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	schemaCmd := SchemaCommand(app, &buf)

	if err := schemaCmd.Run(context.Background(), []string{"schema"}); err != nil {
		t.Fatalf("run schema command: %v", err)
	}
	schema := decodeSchemaOutput(t, buf.Bytes())

	for _, path := range []string{"order", "order place", "order place equity"} {
		if _, ok := schema.Commands[path]; !ok {
			t.Fatalf("missing command path %q", path)
		}
	}
	if !schema.Commands["order place equity"].Flags["--symbol"].Required {
		t.Fatalf("expected symbol flag to be required")
	}
}

func TestClassifyFlagUnknownTypeFallsBackToString(t *testing.T) {
	t.Parallel()

	name, schema := classifyFlag(&cli.UintFlag{Name: "retries", Usage: "retry count"})
	if name != "retries" {
		t.Fatalf("expected retries, got %q", name)
	}
	if schema.Type != "string" {
		t.Fatalf("expected fallback string type, got %q", schema.Type)
	}
}

func TestSchemaCommandRawJSONOutput(t *testing.T) {
	t.Parallel()

	app := testSchemaApp()
	var buf bytes.Buffer
	schemaCmd := SchemaCommand(app, &buf)

	if err := schemaCmd.Run(context.Background(), []string{"schema"}); err != nil {
		t.Fatalf("run schema command: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal raw schema output: %v", err)
	}
	for _, key := range []string{"commands", "global_flags"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("missing top-level key %q", key)
		}
	}
	for _, key := range []string{"data", "metadata", "error"} {
		if _, ok := raw[key]; ok {
			t.Fatalf("unexpected envelope key %q", key)
		}
	}
}
