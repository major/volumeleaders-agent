// Package commands defines the CLI command tree and action handlers for
// querying VolumeLeaders institutional trade data.
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	cli "github.com/urfave/cli/v3"
)

// SchemaOutput is the top-level schema introspection structure.
type SchemaOutput struct {
	Commands    map[string]CommandSchema `json:"commands"`
	GlobalFlags map[string]FlagSchema    `json:"global_flags"`
}

// CommandSchema describes a single CLI command in the schema.
type CommandSchema struct {
	Description string                `json:"description"`
	Flags       map[string]FlagSchema `json:"flags"`
	Args        map[string]any        `json:"args"`
	Examples    []string              `json:"examples"`
}

// FlagSchema describes a single CLI flag in the schema.
type FlagSchema struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Default     any    `json:"default"`
	Description string `json:"description"`
}

// SchemaCommand returns the CLI command for schema introspection.
// It walks the provided app's command tree and emits a JSON description of all
// commands, flags, and their types. Output is raw JSON so agents and scripts can
// consume the CLI shape without authenticating to VolumeLeaders first.
func SchemaCommand(app *cli.Command, w io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "schema",
		Usage: "Display JSON schema of all available commands",
		UsageText: `volumeleaders-agent schema
volumeleaders-agent schema --command "trade list"`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "command",
				Usage: "Filter to a single command path (e.g., \"trade list\")",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			commands := make(map[string]CommandSchema)
			walkCommands(app, "", commands)

			globalFlags := extractFlags(app.Flags)

			filter := cmd.String("command")
			if filter != "" {
				cs, ok := commands[filter]
				if !ok {
					return fmt.Errorf("command %q not found", filter)
				}
				commands = map[string]CommandSchema{filter: cs}
			}

			schema := SchemaOutput{
				Commands:    commands,
				GlobalFlags: globalFlags,
			}

			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			enc.SetEscapeHTML(false)
			return enc.Encode(schema)
		},
	}
}

// walkCommands recursively traverses the command tree and populates commands
// with space-separated command paths as keys.
func walkCommands(cmd *cli.Command, prefix string, commands map[string]CommandSchema) {
	for _, sub := range cmd.Commands {
		path := sub.Name
		if prefix != "" {
			path = prefix + " " + sub.Name
		}

		commands[path] = CommandSchema{
			Description: sub.Usage,
			Flags:       extractFlags(sub.Flags),
			Args:        map[string]any{},
			Examples:    parseExamples(sub.UsageText),
		}

		walkCommands(sub, path, commands)
	}
}

// extractFlags converts CLI flag definitions to schema flag descriptions.
func extractFlags(flags []cli.Flag) map[string]FlagSchema {
	result := make(map[string]FlagSchema)
	for _, f := range flags {
		name, schema := classifyFlag(f)
		if name != "" {
			result[flagSchemaName(name)] = schema
		}

		// Include aliases as first-class entries so the schema reflects every
		// accepted CLI spelling agents and scripts may use.
		for _, alias := range f.Names() {
			if alias != "" && alias != name {
				result[flagSchemaName(alias)] = schema
			}
		}
	}
	return result
}

// flagSchemaName preserves the usual single-dash spelling for one-character
// aliases while keeping long flag names in their double-dash form.
func flagSchemaName(name string) string {
	if len(name) == 1 {
		return "-" + name
	}
	return "--" + name
}

// classifyFlag determines the type and properties of a single CLI flag via
// type assertion on concrete urfave/cli v3 flag types.
func classifyFlag(f cli.Flag) (string, FlagSchema) {
	switch tf := f.(type) {
	case *cli.StringFlag:
		return tf.Name, FlagSchema{
			Type:        "string",
			Required:    tf.Required,
			Default:     tf.Value,
			Description: tf.Usage,
		}
	case *cli.IntFlag:
		return tf.Name, FlagSchema{
			Type:        "int",
			Required:    tf.Required,
			Default:     tf.Value,
			Description: tf.Usage,
		}
	case *cli.FloatFlag:
		return tf.Name, FlagSchema{
			Type:        "float",
			Required:    tf.Required,
			Default:     tf.Value,
			Description: tf.Usage,
		}
	case *cli.BoolFlag:
		return tf.Name, FlagSchema{
			Type:        "bool",
			Required:    tf.Required,
			Default:     tf.Value,
			Description: tf.Usage,
		}
	default:
		// Fall back to string for unknown flag types so new CLI flag kinds still
		// appear in the schema instead of disappearing from agent-visible output.
		names := f.Names()
		if len(names) == 0 {
			return "", FlagSchema{}
		}
		return names[0], FlagSchema{Type: "string"}
	}
}

// parseExamples splits a UsageText string into individual example lines,
// trimming whitespace and dropping blanks. Returns an empty slice so JSON
// output stays consistent when commands do not provide examples.
func parseExamples(usageText string) []string {
	if strings.TrimSpace(usageText) == "" {
		return []string{}
	}

	lines := strings.Split(usageText, "\n")
	examples := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			examples = append(examples, trimmed)
		}
	}

	if len(examples) == 0 {
		return []string{}
	}
	return examples
}
