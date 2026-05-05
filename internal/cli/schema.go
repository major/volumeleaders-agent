package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/major/volumeleaders-agent/internal/cli/common"
)

const jsonSchemaFlagName = "jsonschema"

func configureCobraIntrospection(root *cobra.Command) {
	var schemaMode string
	root.PersistentFlags().StringVar(&schemaMode, jsonSchemaFlagName, "", "Print command input JSON Schema; use --jsonschema=tree on the root for all commands")
	root.PersistentFlags().Lookup(jsonSchemaFlagName).NoOptDefVal = "true"
	var mcpMode bool
	root.PersistentFlags().BoolVar(&mcpMode, "mcp", false, "Serve leaf commands as MCP tools over stdio")

	root.RunE = func(cmd *cobra.Command, _ []string) error {
		if mcpMode {
			return serveMCP(cmd)
		}
		if schemaMode != "" {
			return writeSchema(cmd, schemaMode)
		}
		return cmd.Help()
	}
	wrapCommandTree(root, func(cmd *cobra.Command) {
		installSchemaRunWrapper(cmd, &schemaMode)
		installFlagGroupHelp(cmd)
	})
}

func installSchemaRunWrapper(cmd *cobra.Command, schemaMode *string) {
	if cmd.Parent() == nil || cmd.RunE == nil {
		return
	}
	original := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if *schemaMode != "" {
			return writeSchema(cmd, *schemaMode)
		}
		return original(cmd, args)
	}
}

func writeSchema(cmd *cobra.Command, mode string) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	if mode == "tree" {
		return encoder.Encode(commandSchemas(cmd.Root()))
	}
	return encoder.Encode(commandSchema(cmd))
}

func commandSchemas(root *cobra.Command) []map[string]any {
	schemas := make([]map[string]any, 0)
	walkCobraCommands(root, func(cmd *cobra.Command) {
		if cmd.Runnable() && len(cmd.Commands()) == 0 && !isReferenceHelper(cmd) {
			schemas = append(schemas, commandSchema(cmd))
		}
	})
	return schemas
}

func commandSchema(cmd *cobra.Command) map[string]any {
	properties := map[string]any{}
	groups := map[string]any{}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden || flag.Name == jsonSchemaFlagName || flag.Name == "mcp" {
			return
		}
		flagSchema := map[string]any{
			"type":        schemaType(flag),
			"description": flag.Usage,
		}
		if flag.DefValue != "" {
			flagSchema["default"] = schemaDefault(flag)
		}
		if group := common.FlagGroup(flag); group != "" {
			flagSchema["x-structcli-group"] = group
			groups[group] = map[string]any{}
		}
		if flag.Shorthand != "" {
			flagSchema["x-structcli-shorthand"] = flag.Shorthand
		}
		if enumValues := common.FlagEnum(flag); len(enumValues) > 0 {
			flagSchema["enum"] = stringAnySlice(enumValues)
		}
		properties[flag.Name] = flagSchema
	})
	schema := map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"title":                cmd.CommandPath(),
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
		"x-structcli-groups":   groups,
	}
	if required := requiredFlags(cmd); len(required) > 0 {
		schema["required"] = stringAnySlice(required)
	}
	return schema
}

func schemaType(flag *pflag.Flag) string {
	switch flag.Value.Type() {
	case "bool":
		return "boolean"
	case "int", "int8", "int16", "int32", "int64":
		return "integer"
	case "float32", "float64", "number":
		return "number"
	default:
		return "string"
	}
}

func schemaDefault(flag *pflag.Flag) any {
	switch flag.Value.Type() {
	case "bool":
		value, _ := strconv.ParseBool(flag.DefValue)
		return value
	case "int", "int8", "int16", "int32", "int64":
		value, _ := strconv.ParseInt(flag.DefValue, 10, 64)
		return value
	case "float32", "float64", "number":
		value, _ := strconv.ParseFloat(flag.DefValue, 64)
		return value
	default:
		return flag.DefValue
	}
}

func requiredFlags(cmd *cobra.Command) []string {
	required := make([]string, 0)
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if common.IsFlagRequired(flag) {
			required = append(required, flag.Name)
		}
	})
	slices.Sort(required)
	return required
}

func serveMCP(root *cobra.Command) error {
	scanner := bufio.NewScanner(root.InOrStdin())
	for scanner.Scan() {
		var request struct {
			JSONRPC string `json:"jsonrpc"`
			ID      any    `json:"id"`
			Method  string `json:"method"`
			Params  struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			return err
		}

		if request.ID == nil {
			continue
		}
		switch request.Method {
		case "initialize":
			result := map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]any{"tools": map[string]any{}},
				"serverInfo":      map[string]any{"name": root.Name(), "version": root.Version},
			}
			if err := writeMCPResponse(root.OutOrStdout(), request.ID, result); err != nil {
				return err
			}
		case "tools/list":
			if err := writeMCPResponse(root.OutOrStdout(), request.ID, map[string]any{"tools": mcpTools(root)}); err != nil {
				return err
			}
		case "tools/call":
			if err := writeMCPResponse(root.OutOrStdout(), request.ID, callMCPTool(root, request.Params.Name, request.Params.Arguments)); err != nil {
				return err
			}
		default:
			if err := writeMCPResponse(root.OutOrStdout(), request.ID, map[string]any{"content": []map[string]string{{"type": "text", "text": "unsupported method"}}, "isError": true}); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func mcpTools(root *cobra.Command) []map[string]any {
	tools := make([]map[string]any, 0)
	walkCobraCommands(root, func(cmd *cobra.Command) {
		if cmd.Runnable() && len(cmd.Commands()) == 0 && !isReferenceHelper(cmd) {
			tools = append(tools, map[string]any{"name": toolName(cmd), "description": cmd.Short, "inputSchema": commandSchema(cmd)})
		}
	})
	return tools
}

func callMCPTool(root *cobra.Command, name string, arguments map[string]any) map[string]any {
	cmd := commandByToolName(root, name)
	if cmd == nil {
		return mcpTextError("unknown tool " + name)
	}
	for key, value := range arguments {
		flag := cmd.Flags().Lookup(key)
		if flag == nil || flag.Hidden {
			return mcpTextError("unknown flag: --" + key)
		}
		if err := cmd.Flags().Set(key, fmt.Sprint(value)); err != nil {
			return mcpTextError(err.Error())
		}
	}
	for _, required := range requiredFlags(cmd) {
		if _, ok := arguments[required]; !ok {
			return mcpTextError(fmt.Sprintf("required flag %q not set", required))
		}
	}

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if cmd.PreRunE != nil {
		if err := cmd.PreRunE(cmd, nil); err != nil {
			return mcpTextError(err.Error())
		}
	}
	if cmd.RunE != nil {
		if err := cmd.RunE(cmd, nil); err != nil {
			return mcpTextError(err.Error())
		}
	}
	return mcpTextResult(buf.String())
}

func mcpTextResult(text string) map[string]any {
	return map[string]any{"content": []map[string]string{{"type": "text", "text": text}}}
}

func mcpTextError(text string) map[string]any {
	return map[string]any{"content": []map[string]string{{"type": "text", "text": text}}, "isError": true}
}

func writeMCPResponse(writer io.Writer, id any, result map[string]any) error {
	return json.NewEncoder(writer).Encode(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
}

func commandByToolName(root *cobra.Command, name string) *cobra.Command {
	var found *cobra.Command
	walkCobraCommands(root, func(cmd *cobra.Command) {
		if found == nil && cmd.Runnable() && len(cmd.Commands()) == 0 && toolName(cmd) == name {
			found = cmd
		}
	})
	return found
}

func toolName(cmd *cobra.Command) string {
	parts := strings.Fields(strings.TrimPrefix(cmd.CommandPath(), cmd.Root().Name()+" "))
	return strings.Join(parts, "-")
}

func walkCobraCommands(cmd *cobra.Command, fn func(*cobra.Command)) {
	fn(cmd)
	for _, child := range cmd.Commands() {
		walkCobraCommands(child, fn)
	}
}

func wrapCommandTree(cmd *cobra.Command, fn func(*cobra.Command)) {
	for _, child := range cmd.Commands() {
		fn(child)
		wrapCommandTree(child, fn)
	}
}

func isReferenceHelper(cmd *cobra.Command) bool {
	name := cmd.Name()
	return name == "help" || name == "completion" || name == "bash" || name == "fish" || name == "powershell" || name == "zsh" || name == "outputschema" || name == "config-keys" || name == "env-vars"
}

func installFlagGroupHelp(cmd *cobra.Command) {
	groups := flagGroupNames(cmd)
	if len(groups) == 0 {
		return
	}
	defaultHelp := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelp(cmd, args)
		for _, group := range groups {
			fmt.Fprintf(cmd.OutOrStdout(), "\n%s Flags:\n", group)
			groupFlags := pflag.NewFlagSet(group, pflag.ContinueOnError)
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				if !flag.Hidden && common.FlagGroup(flag) == group {
					groupFlags.AddFlag(flag)
				}
			})
			fmt.Fprint(cmd.OutOrStdout(), groupFlags.FlagUsagesWrapped(80))
		}
	})
}

func flagGroupNames(cmd *cobra.Command) []string {
	seen := map[string]struct{}{}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if group := common.FlagGroup(flag); group != "" {
			seen[group] = struct{}{}
		}
	})
	groups := make([]string, 0, len(seen))
	for group := range seen {
		groups = append(groups, group)
	}
	slices.Sort(groups)
	return groups
}

func stringAnySlice(values []string) []any {
	result := make([]any, len(values))
	for i, value := range values {
		result[i] = value
	}
	return result
}
