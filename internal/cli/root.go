package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/major/volumeleaders-agent/internal/cli/alert"
	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/market"
	"github.com/major/volumeleaders-agent/internal/cli/report"
	"github.com/major/volumeleaders-agent/internal/cli/trade"
	updatecmd "github.com/major/volumeleaders-agent/internal/cli/update"
	"github.com/major/volumeleaders-agent/internal/cli/volume"
	"github.com/major/volumeleaders-agent/internal/cli/watchlist"
	updater "github.com/major/volumeleaders-agent/internal/update"
)

// rootOptions holds flags bound to the root command.
// The flags are registered explicitly via BoolVarP and validated before PersistentPreRunE fires.
type rootOptions struct {
	Pretty bool
}

// NewRootCmd returns the root cobra command for volumeleaders-agent.
func NewRootCmd(version string) *cobra.Command {
	opts := &rootOptions{}
	cmd := &cobra.Command{
		Use:   "volumeleaders-agent",
		Short: "CLI tool for querying VolumeLeaders institutional trade data",
		Long: `volumeleaders-agent queries institutional trade data from VolumeLeaders. Use it for trades, volume leaderboards, market data, alerts, and watchlists.

Auth: reads browser cookies automatically. If auth fails with exit code 2 and "Authentication required: VolumeLeaders session has expired.", log in at https://www.volumeleaders.com in your browser, then retry.

Output: compact JSON to stdout by default. Use --pretty before the command group for indented JSON. Use --jsonschema on any command for machine-readable input JSON Schema output, --jsonschema=tree on the root for the full CLI tree, outputschema for machine-readable stdout contracts, or --mcp on the root to serve leaf commands as MCP tools over stdio. Errors and logs go to stderr.

COMMAND CHOOSER

Goal                                          Start with                              Notes
--------------------------------------------  --------------------------------------  -----------------------------------------------
Run safe preset trade scans                   report list                             Prefer reports before raw trade filters
Find ranked institutional prints              report top-100-rank                     Vetted browser preset, timeout-aware defaults
Find strongest ranked prints                  report top-10-rank                      Narrower ranked-trade preset
Find dark pool sweep activity                 report dark-pool-sweeps                 Vetted dark-pool sweep preset
Find unusually large prints                   report disproportionately-large          5x relative size browser preset
Get comprehensive ticker overview            trade dashboard X --days N              First stop for any single-ticker work
Find individual institutional prints          trade list X --days N                   Advanced path after dashboard or reports
Compare leveraged ETF bull/bear flow          trade sentiment --days N                Fixed leveraged ETF universe, not buy/sell flow
Find converging price-level activity          trade clusters --days N                 Cluster conviction around similar prices
Find sudden aggressive bursts                 trade cluster-bombs --days N            Burst detection, different defaults than clusters
Inspect trade or cluster alerts               trade alerts --date D                   System-generated alerts
Find support/resistance levels                trade levels X --days N                 Level-only drilldown after dashboard
Find revisits to institutional levels         trade level-touches X --days N          Level retests, capped pagination
See institutional volume leaders              volume institutional --date D            Same trade model, volume-ranked
See after-hours institutional leaders         volume ah-institutional --date D        After-hours institutional flow
See total volume leaders                      volume total --date D                   Total market volume across trade types
Find earnings with prior institutional flow   market earnings --days N                CSV/TSV supported
Check exhaustion/reversal signals             market exhaustion                       Optional --date, lower rank is stronger
Manage alert configs                          alert configs/create/edit/delete        Edit replaces unspecified values with defaults
Manage watchlists                             watchlist configs/create/edit/delete    Edit replaces unspecified values with defaults
Get watchlist tickers                         watchlist tickers --watchlist-key K     Key comes from watchlist configs
Check or install CLI updates                  update check or update                  Notices use stderr and update verifies checksums

ANALYSIS WORKFLOW

1. report list to choose a vetted preset report before raw filters.
2. report top-100-rank or report disproportionately-large for the broad scan.
3. trade dashboard X --days N for any single-ticker question before deeper drilling.
4. trade list --preset NAME only when report commands are not specific enough.
5. trade levels X --days N only when the dashboard level section needs level-only output, CSV/TSV, or fields.
6. trade clusters X --days N when dashboard or trade output shows concentration around a price.
7. market earnings --days N and market exhaustion for event and reversal context.

GLOBAL CONVENTIONS

Dates: YYYY-MM-DD. Commands with date ranges accept either --start-date D --end-date D or --days N. --days counts backward from today unless --end-date is also set, and cannot be combined with --start-date.

Pagination: --start offset, --length count, --length -1 means all rows unless a capped endpoint rejects it. trade list does not expose --length; multi-day lookups whose effective filters include tickers return the top 10 long-period trades with VolumeLeaders' lightweight chart query shape, while trade list --summary, single-day trade scans, all-market trade scans, sector-only presets, trade clusters, and trade cluster-bombs fetch all rows internally in browser-sized 100-row pages. trade level-touches only allows 1 to 50 rows. trade dashboard count and trade levels/level-touches level counts only allow values of 5, 10, 20, or 50.

Toggle filters: -1 means all/unfiltered, 0 means exclude, 1 means include/only.

Tickers: --tickers is comma-separated, --ticker is single-symbol. Commands that take tickers generally accept positional tickers too, for example: trade list XLE XLK. Trade and volume ticker filters also accept --symbol and --symbols aliases.

Output formats: list-style commands may support --format json/csv/tsv. CSV/TSV include headers, booleans render as true/false, null or missing values render as empty cells. Nested summaries and single-object commands are JSON-only unless the input schema shows a format flag. Use outputschema to inspect the success stdout shape for each command.

Updates: when release binaries run interactively, they check GitHub releases at most once per day and write update notifications to stderr only. Use update config --check-notifications=false to disable notifications, update check to inspect release status, and update to download a verified release archive and replace the binary.

Performance: use report commands first. Start with one vetted report, one day, and tickers when possible, then expand. VolumeLeaders endpoints can be expensive; broad custom trade list filters are easy to overdo. report commands reject broad multi-day scans without tickers, trade list uses a bounded chart-style request for multi-day ticker lookups, and full-result retrieval keeps the browser's 100-row page size.

RECOVERY PLAYBOOK

Authentication failed or exit code 2: log in at https://www.volumeleaders.com in the same browser profile, confirm the site loads, then retry the exact command. Do not paste cookies or session values into commands.

Date validation failed: use YYYY-MM-DD. For required ranges, provide either --start-date D --end-date D or --days N. Do not combine --days with --start-date.

Pagination validation failed: reduce --length to the documented cap. trade level-touches accepts 1 to 50 rows per request. Do not add --length to trade list, trade clusters, or trade cluster-bombs because they page internally at 100 rows per request.

Unknown flag or enum value: run the same command with --help or --jsonschema to inspect supported flags, defaults, allowed values, and required fields before retrying.

Empty or too broad output: use report list to pick a vetted preset report first, then add tickers or explicit dates. For a single ticker, run trade dashboard TICKER before raw trade, cluster, or level drilldowns. If JSON is too verbose, use --fields where supported or --format csv for list-style commands. Avoid hand-building raw filters unless report commands and trade list --preset cannot answer the question.

COMMAND SEQUENCES

Broad scan: report top-100-rank, then report disproportionately-large, then trade dashboard TICKER --days N, then targeted drilldowns only for the sections that need more detail.

Preset workflow: report list, then report NAME for safe defaults, then trade list --preset NAME only if advanced customization is needed.

Ticker drilldown: trade dashboard TICKER --days N first, then trade list, trade levels, trade clusters, or trade cluster-bombs only for the dashboard sections that need deeper pagination, CSV/TSV, or field selection.

Event context: market earnings --days N, then trade list TICKER --start-date D --end-date D, then market exhaustion with optional --date.

Watchlist workflow: watchlist configs to find keys and names, watchlist tickers --watchlist-key K to inspect symbols, then trade list --watchlist NAME --days N.`,
		Version:          version,
		SilenceErrors:    true,
		SilenceUsage:     true,
		TraverseChildren: true,
		Args:             cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
			cmd.SetContext(context.WithValue(cmd.Context(), common.PrettyJSONKey, opts.Pretty))
			updater.NotifyIfDue(cmd.Context(), version, cmd.CommandPath())
			return nil
		},
	}
	cmd.AddGroup(
		&cobra.Group{ID: "trading", Title: "Trading Commands:"},
		&cobra.Group{ID: "volume", Title: "Volume Commands:"},
		&cobra.Group{ID: "market", Title: "Market Commands:"},
		&cobra.Group{ID: "alerts", Title: "Alert Commands:"},
		&cobra.Group{ID: "watchlists", Title: "Watchlist Commands:"},
		&cobra.Group{ID: "system", Title: "System Commands:"},
		&cobra.Group{ID: "reference", Title: "Reference Commands:"},
	)
	updateCommand := updatecmd.NewCmd(version)
	updateCommand.GroupID = "system"
	cmd.AddCommand(
		report.NewCmd(),
		trade.NewCmd(),
		volume.NewVolumeCommand(),
		market.NewMarketCommand(),
		alert.NewAlertCommand(),
		watchlist.NewCmd(),
		updateCommand,
		newOutputSchemaCmd(),
	)
	cmd.Flags().BoolVarP(&opts.Pretty, "pretty", "p", false, "Pretty-print JSON output with indentation")
	common.AnnotateFlagGroup(cmd, "pretty", "Output")
	common.WrapValidation(cmd, opts)
	return cmd
}

// SetupCLI configures root-level discovery features on the command tree. It is
// separated from NewRootCmd so tests can build raw command trees without adding
// process-oriented schema and MCP behavior.
func SetupCLI(cmd *cobra.Command) {
	var jsonSchemaMode string
	cmd.Flags().StringVar(&jsonSchemaMode, "jsonschema", "", "Output JSON Schema for commands (use 'tree' for full CLI tree)")
	cmd.Flags().Lookup("jsonschema").NoOptDefVal = "command"
	addSubcommandJSONSchemaFlags(cmd, &jsonSchemaMode)

	var mcp bool
	cmd.Flags().BoolVar(&mcp, "mcp", false, "Serve leaf commands as MCP tools over stdio")

	installGroupedHelp(cmd)
	wrapDiscoveryHandlers(cmd, &jsonSchemaMode, &mcp)
}

func wrapDiscoveryHandlers(cmd *cobra.Command, jsonSchemaMode *string, mcp *bool) {
	originalRunE := cmd.RunE
	originalRun := cmd.Run
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if rootJSONSchemaRequested(cmd) {
			return runJSONSchema(cmd, *jsonSchemaMode)
		}
		if mcp != nil && *mcp {
			return runMCP(cmd.Root(), cmd.InOrStdin(), cmd.OutOrStdout())
		}
		if originalRunE != nil {
			return originalRunE(cmd, args)
		}
		if originalRun != nil {
			originalRun(cmd, args)
		}
		return nil
	}
	cmd.Run = nil
	for _, sub := range cmd.Commands() {
		wrapDiscoveryHandlers(sub, jsonSchemaMode, mcp)
	}
}

func rootJSONSchemaRequested(cmd *cobra.Command) bool {
	flag := cmd.Root().Flags().Lookup("jsonschema")
	if flag != nil && flag.Changed {
		return true
	}
	flag = cmd.Flags().Lookup("jsonschema")
	return flag != nil && flag.Changed
}

func addSubcommandJSONSchemaFlags(cmd *cobra.Command, mode *string) {
	for _, sub := range cmd.Commands() {
		sub.Flags().StringVar(mode, "jsonschema", "", "Output JSON Schema for commands (use 'tree' for full CLI tree)")
		sub.Flags().Lookup("jsonschema").NoOptDefVal = "command"
		sub.Flags().Lookup("jsonschema").Hidden = true
		addSubcommandJSONSchemaFlags(sub, mode)
	}
}

func runJSONSchema(cmd *cobra.Command, mode string) error {
	if mode == "" || mode == "command" {
		return common.PrintJSON(cmd.OutOrStdout(), context.WithValue(cmd.Context(), common.PrettyJSONKey, prettyFromCommand(cmd)), commandSchema(cmd))
	}
	if mode == "tree" {
		var schemas []map[string]any
		walkCommandTree(cmd.Root(), func(c *cobra.Command) {
			schemas = append(schemas, commandSchema(c))
		})
		return common.PrintJSON(cmd.OutOrStdout(), context.WithValue(cmd.Context(), common.PrettyJSONKey, prettyFromCommand(cmd)), schemas)
	}
	target, err := findCommandByPath(cmd.Root(), strings.Fields(mode))
	if err != nil {
		return err
	}
	return common.PrintJSON(cmd.OutOrStdout(), context.WithValue(cmd.Context(), common.PrettyJSONKey, prettyFromCommand(cmd)), commandSchema(target))
}

func walkCommandTree(cmd *cobra.Command, visit func(*cobra.Command)) {
	visit(cmd)
	for _, sub := range cmd.Commands() {
		walkCommandTree(sub, visit)
	}
}

func findCommandByPath(root *cobra.Command, path []string) (*cobra.Command, error) {
	current := root
	for _, part := range path {
		var next *cobra.Command
		for _, candidate := range current.Commands() {
			if candidate.Name() == part || slices.Contains(candidate.Aliases, part) {
				next = candidate
				break
			}
		}
		if next == nil {
			return nil, fmt.Errorf("unknown command path %q", strings.Join(path, " "))
		}
		current = next
	}
	return current, nil
}

func commandSchema(cmd *cobra.Command) map[string]any {
	schema := map[string]any{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"type":        "object",
		"title":       cmd.CommandPath(),
		"description": cmd.Short,
		"properties":  flagSchemas(cmd),
		"required":    requiredFlags(cmd),
	}
	if subcommands := commandSubcommands(cmd); len(subcommands) > 0 {
		schema["x-flag-subcommands"] = subcommands
	}
	if groups := flagGroups(cmd); len(groups) > 0 {
		schema["x-flag-groups"] = groups
	}
	return schema
}

func commandSubcommands(cmd *cobra.Command) []string {
	var names []string
	for _, sub := range cmd.Commands() {
		if sub.Hidden {
			continue
		}
		names = append(names, sub.Name())
	}
	slices.Sort(names)
	return names
}

func flagSchemas(cmd *cobra.Command) map[string]any {
	props := make(map[string]any)
	visitCommandFlags(cmd, func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		item := map[string]any{
			"type":        jsonTypeForFlag(flag),
			"description": flag.Usage,
			"default":     defaultValueForFlag(flag),
		}
		if group := common.FlagGroup(flag); group != "" {
			item["x-flag-group"] = group
		}
		if flag.Shorthand != "" {
			item["x-flag-shorthand"] = flag.Shorthand
		}
		if values := common.FlagEnumValues(flag); len(values) > 0 {
			item["enum"] = values
			item["x-flag-enum"] = values
		}
		props[flag.Name] = item
	})
	return props
}

func visitCommandFlags(cmd *cobra.Command, visit func(*pflag.Flag)) {
	seen := make(map[string]struct{})
	for _, flags := range []*pflag.FlagSet{cmd.InheritedFlags(), cmd.NonInheritedFlags()} {
		flags.VisitAll(func(flag *pflag.Flag) {
			if _, ok := seen[flag.Name]; ok {
				return
			}
			seen[flag.Name] = struct{}{}
			visit(flag)
		})
	}
}

func jsonTypeForFlag(flag *pflag.Flag) string {
	switch flag.Value.Type() {
	case "bool":
		return "boolean"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	default:
		return "string"
	}
}

func defaultValueForFlag(flag *pflag.Flag) any {
	switch jsonTypeForFlag(flag) {
	case "boolean":
		value, err := strconv.ParseBool(flag.DefValue)
		if err == nil {
			return value
		}
	case "integer":
		value, err := strconv.ParseInt(flag.DefValue, 10, 64)
		if err == nil {
			return value
		}
	case "number":
		value, err := strconv.ParseFloat(flag.DefValue, 64)
		if err == nil {
			return value
		}
	}
	return flag.DefValue
}

func requiredFlags(cmd *cobra.Command) []string {
	var required []string
	visitCommandFlags(cmd, func(flag *pflag.Flag) {
		if common.IsFlagRequired(flag) {
			required = append(required, flag.Name)
		}
	})
	slices.Sort(required)
	return required
}

func flagGroups(cmd *cobra.Command) map[string]any {
	groups := make(map[string]any)
	visitCommandFlags(cmd, func(flag *pflag.Flag) {
		if group := common.FlagGroup(flag); group != "" {
			groups[group] = map[string]any{}
		}
	})
	return groups
}

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
	Path        []string       `json:"-"`
}

func runMCP(root *cobra.Command, input io.Reader, output io.Writer) error {
	tools := mcpTools(root)
	toolsByName := make(map[string]mcpTool, len(tools))
	for _, tool := range tools {
		toolsByName[tool.Name] = tool
	}

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var request mcpRequest
		if err := json.Unmarshal(line, &request); err != nil {
			writeMCPError(output, nil, -32700, err.Error())
			continue
		}
		handleMCPRequest(output, root.Version, tools, toolsByName, request)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read MCP request: %w", err)
	}
	return nil
}

func handleMCPRequest(output io.Writer, version string, tools []mcpTool, toolsByName map[string]mcpTool, request mcpRequest) {
	switch request.Method {
	case "initialize":
		writeMCPResult(output, request.ID, map[string]any{"protocolVersion": "2024-11-05", "serverInfo": map[string]any{"name": "volumeleaders-agent", "version": version}, "capabilities": map[string]any{"tools": map[string]any{}}})
	case "tools/list":
		writeMCPResult(output, request.ID, map[string]any{"tools": tools})
	case "tools/call":
		result := callMCPTool(version, toolsByName, request.Params)
		writeMCPResult(output, request.ID, result)
	default:
		writeMCPError(output, request.ID, -32601, "method not found")
	}
}

func mcpTools(root *cobra.Command) []mcpTool {
	var tools []mcpTool
	walkCommandTree(root, func(cmd *cobra.Command) {
		if !mcpLeafCommand(cmd) {
			return
		}
		path := strings.Fields(strings.TrimPrefix(cmd.CommandPath(), root.CommandPath()+" "))
		name := strings.ReplaceAll(strings.Join(path, "-"), "_", "-")
		tools = append(tools, mcpTool{Name: name, Description: cmd.Short, InputSchema: commandSchema(cmd), Path: path})
	})
	slices.SortFunc(tools, func(a, b mcpTool) int { return strings.Compare(a.Name, b.Name) })
	return tools
}

func mcpLeafCommand(cmd *cobra.Command) bool {
	if !cmd.Runnable() || len(cmd.Commands()) > 0 || cmd.Hidden {
		return false
	}
	switch cmd.Name() {
	case "help", "completion", "bash", "fish", "powershell", "zsh", "outputschema":
		return false
	default:
		return true
	}
}

func callMCPTool(version string, toolsByName map[string]mcpTool, params json.RawMessage) map[string]any {
	var request struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return mcpTextResult(err.Error(), true)
	}
	tool, ok := toolsByName[request.Name]
	if !ok {
		return mcpTextResult(fmt.Sprintf("unknown tool %q", request.Name), true)
	}
	args := append([]string{}, tool.Path...)
	args = append(args, cliArgsFromToolArguments(request.Arguments)...)

	freshRoot := NewRootCmd(version)
	SetupCLI(freshRoot)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	freshRoot.SetOut(&stdout)
	freshRoot.SetErr(&stderr)
	freshRoot.SetArgs(args)
	err := freshRoot.Execute()
	if err != nil {
		text := strings.TrimSpace(stderr.String())
		if text == "" {
			text = err.Error()
		}
		return mcpTextResult(text, true)
	}
	return mcpTextResult(strings.TrimSpace(stdout.String()), false)
}

func cliArgsFromToolArguments(arguments map[string]any) []string {
	keys := make([]string, 0, len(arguments))
	for key := range arguments {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	var args []string
	for _, key := range keys {
		value := arguments[key]
		switch typed := value.(type) {
		case bool:
			if typed {
				args = append(args, "--"+key)
			} else {
				args = append(args, "--"+key+"=false")
			}
		case []any:
			parts := make([]string, 0, len(typed))
			for _, item := range typed {
				parts = append(parts, fmt.Sprint(item))
			}
			args = append(args, "--"+key, strings.Join(parts, ","))
		default:
			args = append(args, "--"+key, fmt.Sprint(value))
		}
	}
	return args
}

func mcpTextResult(text string, isError bool) map[string]any {
	return map[string]any{"content": []map[string]string{{"type": "text", "text": text}}, "isError": isError}
}

func writeMCPResult(output io.Writer, id any, result any) {
	_ = json.NewEncoder(output).Encode(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
}

func writeMCPError(output io.Writer, id any, code int, message string) {
	_ = json.NewEncoder(output).Encode(map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": message}})
}

func installGroupedHelp(root *cobra.Command) {
	walkCommandTree(root, func(cmd *cobra.Command) {
		cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
			fmt.Fprint(c.OutOrStdout(), c.UsageString())
			writeGroupedFlagHelp(c.OutOrStdout(), c)
		})
	})
}

func writeGroupedFlagHelp(output io.Writer, cmd *cobra.Command) {
	groups := make(map[string][]*pflag.Flag)
	visitCommandFlags(cmd, func(flag *pflag.Flag) {
		if group := common.FlagGroup(flag); group != "" {
			groups[group] = append(groups[group], flag)
		}
	})
	if len(groups) == 0 {
		return
	}
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		fmt.Fprintf(output, "\n%s Flags:\n", name)
		for _, flag := range groups[name] {
			fmt.Fprintf(output, "  --%s\t%s\n", flag.Name, flag.Usage)
		}
	}
}
