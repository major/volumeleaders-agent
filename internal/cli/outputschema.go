package cli

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	reportcmd "github.com/major/volumeleaders-agent/internal/cli/report"
	"github.com/major/volumeleaders-agent/internal/models"
)

// outputContract describes the stdout contract for one executable command.
// It intentionally lives outside structcli because structcli v0.17.0 only
// exposes input schemas. Keeping output contracts separate avoids changing
// command execution paths while still giving LLM clients machine-readable
// success shapes.
type outputContract struct {
	Command        string          `json:"command"`
	Description    string          `json:"description"`
	Formats        []string        `json:"formats"`
	DefaultFormat  string          `json:"default_format"`
	Schema         outputSchema    `json:"schema"`
	FieldSelection *fieldSelection `json:"field_selection,omitempty"`
	Variants       []outputVariant `json:"variants,omitempty"`
	Notes          []string        `json:"notes,omitempty"`
}

// outputVariant captures conditional output shapes selected by flags such as
// --summary or --fields. A command may have one default schema plus variants.
type outputVariant struct {
	When           string          `json:"when"`
	Formats        []string        `json:"formats"`
	Schema         outputSchema    `json:"schema"`
	FieldSelection *fieldSelection `json:"field_selection,omitempty"`
	Notes          []string        `json:"notes,omitempty"`
}

// fieldSelection documents commands where --fields narrows JSON object keys or
// CSV and TSV columns.
type fieldSelection struct {
	Flag          string   `json:"flag"`
	DefaultFields []string `json:"default_fields,omitempty"`
	AllFields     []string `json:"all_fields"`
	AllValue      string   `json:"all_value,omitempty"`
}

// outputSchema is a compact JSON Schema-like description of command stdout.
// It is intentionally small and stable so agents can consume it without needing
// a full JSON Schema validator.
type outputSchema struct {
	Type                 string                  `json:"type"`
	Model                string                  `json:"model,omitempty"`
	Format               string                  `json:"format,omitempty"`
	Nullable             bool                    `json:"nullable,omitempty"`
	Items                *outputSchema           `json:"items,omitempty"`
	Properties           map[string]outputSchema `json:"properties,omitempty"`
	AdditionalProperties *outputSchema           `json:"additional_properties,omitempty"`
	Any                  bool                    `json:"any,omitempty"`
}

// newOutputSchemaCmd returns a reference command that prints machine-readable
// stdout contracts for leaf commands.
func newOutputSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "outputschema [command...]",
		Short:   "Print command output contracts",
		GroupID: "reference",
		Args:    cobra.ArbitraryArgs,
		Long:    "Print machine-readable stdout contracts for executable commands. With no arguments it returns every contract as a JSON array. Pass a command path such as trade list to return one contract. This describes success output only; structured errors are documented by structcli flag errors.",
		Example: `volumeleaders-agent outputschema
volumeleaders-agent outputschema trade list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			contracts := allOutputContracts()
			if len(args) == 0 {
				return common.PrintJSON(cmd.OutOrStdout(), context.WithValue(cmd.Context(), common.PrettyJSONKey, prettyFromCommand(cmd)), contracts)
			}

			commandPath := strings.Join(args, " ")
			contract, ok := outputContractByCommand(contracts, commandPath)
			if !ok {
				return fmt.Errorf("unknown output contract %q; run outputschema with no arguments for all contracts", commandPath)
			}
			return common.PrintJSON(cmd.OutOrStdout(), context.WithValue(cmd.Context(), common.PrettyJSONKey, prettyFromCommand(cmd)), contract)
		},
	}
	return cmd
}

func prettyFromCommand(cmd *cobra.Command) bool {
	pretty, _ := cmd.Root().PersistentFlags().GetBool("pretty")
	return pretty
}

func outputContractByCommand(contracts []outputContract, commandPath string) (outputContract, bool) {
	for index := range contracts {
		contract := contracts[index]
		if contract.Command == commandPath {
			return contract, true
		}
	}
	return outputContract{}, false
}

func allOutputContracts() []outputContract {
	contracts := []outputContract{
		arrayOutputContract[reportcmd.ReportInfo]("report list", "List curated report commands and their fixed browser-preset filters.", outputFormats(), nil, nil),
		reportTradeOutputContract("report top-100-rank", "Run the site-vetted top 100 ranked trades report."),
		reportTradeOutputContract("report top-10-rank", "Run the site-vetted top 10 ranked trades report."),
		reportTradeOutputContract("report dark-pool-sweeps", "Run the site-vetted dark pool sweeps report."),
		reportTradeOutputContract("report disproportionately-large", "Run the site-vetted 5x relative size report."),

		objectOutputContract[models.TradeDashboard]("trade dashboard", "Return a fast ticker dashboard with trades, clusters, levels, and cluster bombs.", []string{"json"}, []string{"Defaults to a 365-day lookback and 10 rows per section."}),
		arrayOutputContract[models.TradeListRow]("trade list", "List individual institutional trades using a compact default row shape.", outputFormats(), nil, nil,
			outputVariant{When: "--fields is set or --format is csv or tsv", Formats: outputFormats(), Schema: arraySchema[models.Trade](), FieldSelection: allFieldsSelection[models.Trade](nil), Notes: []string{"CSV and TSV include a header row matching the selected fields."}},
			outputVariant{When: "--summary is set", Formats: []string{"json"}, Schema: objectSchema[models.TradeSummary](), Notes: []string{"--summary cannot be combined with --fields or non-JSON formats."}},
		),
		objectOutputContract[models.TradeSentiment]("trade sentiment", "Summarize leveraged ETF bull and bear flow.", outputFormats(), []string{"CSV and TSV flatten one row per day."},
			outputVariant{When: "--format is csv or tsv", Formats: []string{"csv", "tsv"}, Schema: arraySchema[models.TradeSentimentRow]()},
		),
		arrayOutputContract[models.TradeCluster]("trade clusters", "List institutional trade clusters around similar prices.", outputFormats(), tradeClusterDefaultFields(), allFieldsSelection[models.TradeCluster](tradeClusterDefaultFields())),
		arrayOutputContract[models.TradeClusterBomb]("trade cluster-bombs", "List bursty trade cluster bomb activity.", outputFormats(), nil, nil),
		arrayOutputContract[models.TradeAlert]("trade alerts", "List system-generated trade alerts for one date.", outputFormats(), nil, nil),
		arrayOutputContract[models.TradeClusterAlert]("trade cluster-alerts", "List system-generated trade cluster alerts for one date.", outputFormats(), nil, nil),
		arrayOutputContract[models.TradeLevelRow]("trade levels", "List support and resistance style institutional price levels.", []string{"json"}, nil, nil,
			outputVariant{When: "--fields is set", Formats: []string{"json"}, Schema: arraySchema[models.TradeLevel](), FieldSelection: allFieldsSelection[models.TradeLevel](nil)},
		),
		arrayOutputContract[models.TradeLevelTouch]("trade level-touches", "List revisits to institutional price levels.", outputFormats(), nil, nil),
		objectOutputContract[models.PresetTickersInfo]("trade preset-tickers", "List ticker symbols used by one built-in trade preset.", []string{"json"}, nil),
		arrayOutputContract[models.PresetInfo]("trade presets", "List built-in trade filter presets.", outputFormats(), nil, nil),

		arrayOutputContract[models.Trade]("volume institutional", "List regular-hours institutional volume leaders.", outputFormats(), nil, nil),
		arrayOutputContract[models.Trade]("volume ah-institutional", "List after-hours institutional volume leaders.", outputFormats(), nil, nil),
		arrayOutputContract[models.Trade]("volume total", "List total volume leaders across trade types.", outputFormats(), nil, nil),

		objectOutputContract[map[string]any]("market snapshots", "Return current price snapshots keyed by ticker.", []string{"json"}, []string{"Snapshot objects come from the VolumeLeaders API and can vary by symbol."}),
		arrayOutputContract[models.Earnings]("market earnings", "List earnings events with related institutional activity counts.", outputFormats(), marketEarningsDefaultFields(), allFieldsSelection[models.Earnings](marketEarningsDefaultFields())),
		objectOutputContract[models.MarketExhaustion]("market exhaustion", "Return market exhaustion rank metrics for one trading day.", []string{"json"}, nil),

		arrayOutputContract[models.AlertConfig]("alert configs", "List saved alert configurations.", outputFormats(), alertConfigDefaultFields(), allFieldsSelection[models.AlertConfig](alertConfigDefaultFields())),
		unknownJSONContract("alert delete", "Delete an alert configuration and return the API response."),
		mutationContract("alert create", "Create an alert configuration and return a success object."),
		mutationContract("alert edit", "Edit an alert configuration and return a success object."),

		arrayOutputContract[models.WatchListConfig]("watchlist configs", "List saved watchlist configurations.", outputFormats(), nil, nil),
		arrayOutputContract[models.WatchListTicker]("watchlist tickers", "List tickers and nearest level metadata for one watchlist.", outputFormats(), nil, nil),
		unknownJSONContract("watchlist delete", "Delete a watchlist configuration and return the API response."),
		unknownJSONContract("watchlist add-ticker", "Add a ticker to a watchlist and return the API response."),
		mutationContract("watchlist create", "Create a watchlist configuration and return a success object."),
		mutationContract("watchlist edit", "Edit a watchlist configuration and return a success object."),
	}
	slices.SortFunc(contracts, func(a, b outputContract) int {
		return strings.Compare(a.Command, b.Command)
	})
	return contracts
}

func reportTradeOutputContract(command, description string) outputContract {
	return arrayOutputContract[models.TradeListRow](command, description, outputFormats(), nil, nil,
		outputVariant{When: "--fields is set or --format is csv or tsv", Formats: outputFormats(), Schema: arraySchema[models.Trade](), FieldSelection: allFieldsSelection[models.Trade](nil), Notes: []string{"CSV and TSV include a header row matching the selected fields."}},
		outputVariant{When: "--summary is set", Formats: []string{"json"}, Schema: objectSchema[models.TradeSummary](), Notes: []string{"--summary cannot be combined with --fields or non-JSON formats."}},
	)
}

func outputFormats() []string {
	return []string{"json", "csv", "tsv"}
}

func arrayOutputContract[T any](command, description string, formats, defaultFields []string, selection *fieldSelection, variants ...outputVariant) outputContract {
	contract := outputContract{
		Command:       command,
		Description:   description,
		Formats:       formats,
		DefaultFormat: "json",
		Schema:        arraySchema[T](),
		Variants:      variants,
	}
	if selection != nil {
		contract.FieldSelection = selection
	} else if len(defaultFields) > 0 {
		contract.FieldSelection = allFieldsSelection[T](defaultFields)
	}
	return contract
}

func objectOutputContract[T any](command, description string, formats, notes []string, variants ...outputVariant) outputContract {
	return outputContract{
		Command:       command,
		Description:   description,
		Formats:       formats,
		DefaultFormat: "json",
		Schema:        schemaForType(reflect.TypeFor[T](), make(map[reflect.Type]bool)),
		Variants:      variants,
		Notes:         notes,
	}
}

func mutationContract(command, description string) outputContract {
	return outputContract{
		Command:       command,
		Description:   description,
		Formats:       []string{"json"},
		DefaultFormat: "json",
		Schema: outputSchema{
			Type:  "object",
			Model: "MutationResult",
			Properties: map[string]outputSchema{
				"success": {Type: "boolean"},
				"action":  {Type: "string"},
				"key":     {Type: "integer"},
			},
		},
	}
}

func unknownJSONContract(command, description string) outputContract {
	return outputContract{
		Command:       command,
		Description:   description,
		Formats:       []string{"json"},
		DefaultFormat: "json",
		Schema:        outputSchema{Type: "object", AdditionalProperties: &outputSchema{Any: true}},
		Notes:         []string{"Response shape is passed through from the VolumeLeaders API."},
	}
}

func arraySchema[T any]() outputSchema {
	item := schemaForType(reflect.TypeFor[T](), make(map[reflect.Type]bool))
	return outputSchema{Type: "array", Items: &item}
}

func objectSchema[T any]() outputSchema {
	return schemaForType(reflect.TypeFor[T](), make(map[reflect.Type]bool))
}

func allFieldsSelection[T any](defaultFields []string) *fieldSelection {
	return &fieldSelection{
		Flag:          "fields",
		DefaultFields: slices.Clone(defaultFields),
		AllFields:     common.JSONFieldNamesInOrder[T](),
		AllValue:      "all",
	}
}

func schemaForType(t reflect.Type, seen map[reflect.Type]bool) outputSchema {
	for t.Kind() == reflect.Pointer {
		base := schemaForType(t.Elem(), seen)
		base.Nullable = true
		return base
	}

	if t == reflect.TypeFor[models.AspNetDate]() || t == reflect.TypeFor[time.Time]() {
		return outputSchema{Type: "string", Format: "date-time", Nullable: true}
	}
	if t == reflect.TypeFor[models.FlexBool]() {
		return outputSchema{Type: "boolean"}
	}

	switch t.Kind() {
	case reflect.Bool:
		return outputSchema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return outputSchema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return outputSchema{Type: "number"}
	case reflect.String:
		return outputSchema{Type: "string"}
	case reflect.Slice, reflect.Array:
		item := schemaForType(t.Elem(), seen)
		return outputSchema{Type: "array", Items: &item}
	case reflect.Map:
		value := schemaForType(t.Elem(), seen)
		return outputSchema{Type: "object", Model: t.String(), AdditionalProperties: &value}
	case reflect.Interface:
		return outputSchema{Any: true}
	case reflect.Struct:
		if seen[t] {
			return outputSchema{Type: "object", Model: t.Name()}
		}
		seen[t] = true
		properties := make(map[string]outputSchema)
		for field := range t.Fields() {
			if !field.IsExported() {
				continue
			}
			name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
			if name == "" || name == "-" {
				continue
			}
			properties[name] = schemaForType(field.Type, seen)
		}
		delete(seen, t)
		return outputSchema{Type: "object", Model: t.Name(), Properties: properties}
	default:
		return outputSchema{Type: "string", Model: t.String()}
	}
}

func tradeClusterDefaultFields() []string {
	return []string{
		"Date",
		"Ticker",
		"Price",
		"Dollars",
		"Volume",
		"TradeCount",
		"DollarsMultiplier",
		"CumulativeDistribution",
		"TradeClusterRank",
		"MinFullDateTime",
		"MaxFullDateTime",
	}
}

func marketEarningsDefaultFields() []string {
	return []string{
		"Ticker",
		"EarningsDate",
		"AfterMarketClose",
		"TradeCount",
		"TradeClusterCount",
		"TradeClusterBombCount",
	}
}

func alertConfigDefaultFields() []string {
	return []string{
		"AlertConfigKey",
		"Name",
		"Tickers",
		"TradeConditions",
		"ClosingTradeConditions",
		"DarkPool",
		"Sweep",
		"OffsettingPrint",
		"PhantomPrint",
	}
}
