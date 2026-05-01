// Package watchlists contains commands for VolumeLeaders watchlist workflows.
package watchlists

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/leodido/structcli"
	"github.com/major/volumeleaders-agent/internal/auth"
	"github.com/spf13/cobra"
)

const (
	defaultFieldPreset     = "summary"
	defaultOutputShape     = "array"
	fullFieldPreset        = "full"
	objectOutputShape      = "objects"
	getWatchListsPath      = "https://www.volumeleaders.com/WatchListConfigs/GetWatchLists"
	saveWatchListPath      = "https://www.volumeleaders.com/WatchListConfig"
	deleteWatchListPath    = "https://www.volumeleaders.com/WatchListConfigs/DeleteWatchList"
	watchListsPage         = "https://www.volumeleaders.com/WatchListConfigs"
	watchListConfigPage    = "https://www.volumeleaders.com/WatchListConfig"
	includedTradeTypesName = "IncludedTradeTypes"
	watchListsGuide        = "LLM field guide: VolumeLeaders watchlists are saved trade and cluster filters, not necessarily long ticker lists. Name identifies the saved filter. Tickers is empty when the watchlist applies to all symbols. MinVolume, MaxVolume, MinDollars, MaxDollars, MinPrice, MaxPrice, MinRelativeSize, MaxTradeRank, RSI fields, Conditions, SecurityTypeKey, SectorIndustry, and included trade-type booleans describe which trades appear on the watchlist. MaxTradeRank=-1 means no rank limit; otherwise trades must rank at that level or better, so 10 means ranks 1 through 10. SecurityTypeKey=-1 means all security types, 1 means stocks only, 26 means ETFs only, and 4 means REITs only. Conditions entries such as IgnoreOBD and IgnoreOBH mean the corresponding daily or hourly RSI condition is ignored. IncludedTradeTypes is a compact derived field listing enabled print, venue, session, auction, phantom, and offsetting categories. Internal-only upstream fields such as SearchTemplateTypeKey, SortOrder, MinVCD, and APIKey are excluded from expanded output; use --preset-fields full only when raw upstream payloads are needed for debugging."
)

var (
	extractCookies          = auth.ExtractCookies
	fetchXSRFToken          = auth.FetchXSRFToken
	getWatchListsHTTPClient = http.DefaultClient
	getWatchListsEndpoint   = getWatchListsPath
	saveWatchListEndpoint   = saveWatchListPath
	deleteWatchListEndpoint = deleteWatchListPath
	nullJSON                = json.RawMessage("null")
)

var watchListFieldPresets = map[string][]string{
	"summary": {
		"SearchTemplateKey",
		"Name",
	},
	"expanded": {
		"SearchTemplateKey",
		"UserKey",
		"Name",
		"Tickers",
		"MinVolume",
		"MaxVolume",
		"MinDollars",
		"MaxDollars",
		"MinPrice",
		"MaxPrice",
		"RSIOverboughtHourly",
		"RSIOverboughtDaily",
		"RSIOversoldHourly",
		"RSIOversoldDaily",
		"Conditions",
		"RSIOverboughtHourlySelected",
		"RSIOverboughtDailySelected",
		"RSIOversoldHourlySelected",
		"RSIOversoldDailySelected",
		"MinRelativeSize",
		"MinRelativeSizeSelected",
		"MaxTradeRank",
		"MaxTradeRankSelected",
		"SecurityTypeKey",
		"SecurityType",
		"SectorIndustry",
		"NormalPrints",
		"SignaturePrints",
		"LatePrints",
		"TimelyPrints",
		"DarkPools",
		"LitExchanges",
		"Sweeps",
		"Blocks",
		"PremarketTrades",
		"RTHTrades",
		"AHTrades",
		"OpeningTrades",
		"ClosingTrades",
		"PhantomTrades",
		"OffsettingTrades",
		"NormalPrintsSelected",
		"SignaturePrintsSelected",
		"LatePrintsSelected",
		"TimelyPrintsSelected",
		"DarkPoolsSelected",
		"LitExchangesSelected",
		"SweepsSelected",
		"BlocksSelected",
		"PremarketTradesSelected",
		"RTHTradesSelected",
		"AHTradesSelected",
		"OpeningTradesSelected",
		"ClosingTradesSelected",
		"PhantomTradesSelected",
		"OffsettingTradesSelected",
		includedTradeTypesName,
	},
}

var getWatchListsColumns = []dataTableColumn{
	{data: "Name", name: "Name", orderable: "true"},
	{data: "Tickers", name: "Tickers", orderable: "true"},
	{data: "Criteria", name: "Criteria", orderable: "false"},
}

var tradeTypeFields = []string{
	"NormalPrints",
	"SignaturePrints",
	"LatePrints",
	"TimelyPrints",
	"DarkPools",
	"LitExchanges",
	"Sweeps",
	"Blocks",
	"PremarketTrades",
	"RTHTrades",
	"AHTrades",
	"OpeningTrades",
	"ClosingTrades",
	"PhantomTrades",
	"OffsettingTrades",
}

// Options defines the LLM-readable contract for listing saved watchlists.
type Options struct {
	SearchTemplateKey int    `flag:"search-template-key" flagdescr:"Optional saved watchlist identifier to return. Use this after the default summary output to inspect one watchlist's criteria." flagenv:"true" flaggroup:"Filters"`
	Fields            string `flag:"fields" flagdescr:"Comma-separated watchlist fields to include. Overrides --preset-fields. Use upstream field names such as Name,Tickers,MaxTradeRank." flagenv:"true" flaggroup:"Output" mod:"trim"`
	PresetFields      string `flag:"preset-fields" flagdescr:"Field preset to include: summary, expanded, or full. Defaults to summary with only SearchTemplateKey and Name so callers can select a watchlist before requesting details. Expanded includes the saved watchlist configuration fields; full returns raw upstream payloads." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Shape             string `flag:"shape" flagdescr:"Watchlist row shape: array or objects. Array is the default and is most token-efficient." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Pretty            bool   `flag:"pretty" flagdescr:"Pretty-print JSON output. Compact JSON is the default for token-efficient LLM and MCP use." flagenv:"true" flaggroup:"Output"`
}

// SaveOptions defines the LLM-readable contract for creating or updating saved watchlists.
type SaveOptions struct {
	SearchTemplateKey           int    `flag:"search-template-key" flagdescr:"Existing watchlist identifier to replace. Leave 0 to create a new watchlist, matching VolumeLeaders' new-watchlist form. Updates post a complete WatchListConfig form, so pass every criterion that should remain on the saved filter." flagenv:"true" flaggroup:"Identity"`
	Name                        string `flag:"name" flagdescr:"Watchlist display name to save." flagenv:"true" flagrequired:"true" flaggroup:"Identity" validate:"required" mod:"trim"`
	Tickers                     string `flag:"tickers" flagdescr:"Optional comma-delimited ticker filter. Leave empty to apply the watchlist criteria to all symbols." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	MinVolume                   string `flag:"min-volume" flagdescr:"Optional minimum share volume filter. Leave empty to match the browser default." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	MaxVolume                   string `flag:"max-volume" flagdescr:"Optional maximum share volume filter. Leave empty to match the browser default." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	MinDollars                  string `flag:"min-dollars" flagdescr:"Optional minimum dollar-value filter. Leave empty to match the browser default." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	MaxDollars                  string `flag:"max-dollars" flagdescr:"Optional maximum dollar-value filter. Leave empty to match the browser default." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	MinPrice                    string `flag:"min-price" flagdescr:"Optional minimum trade price filter. Leave empty to match the browser default." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	MaxPrice                    string `flag:"max-price" flagdescr:"Optional maximum trade price filter. Leave empty to match the browser default." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	SecurityTypeKey             int    `flag:"security-type-key" flagdescr:"Security type filter from VolumeLeaders: -1 all securities, 1 stocks, 26 ETFs, or 4 REITs." flagenv:"true" flaggroup:"Filters"`
	MinRelativeSizeSelected     int    `flag:"min-relative-size" flagdescr:"Minimum relative-size selector value from the browser form. Common values are 0, 5, 10, 25, 50, and 100; 0 means any size." flagenv:"true" flaggroup:"Filters"`
	MaxTradeRankSelected        int    `flag:"max-trade-rank" flagdescr:"Maximum all-time trade rank selector. Use -1 for no rank limit, or values such as 10 or 100 to include ranks 1 through that number." flagenv:"true" flaggroup:"Filters"`
	MinVCD                      string `flag:"min-vcd" flagdescr:"Optional upstream VCD percentile filter. Leave empty to match the browser default." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	SectorIndustry              string `flag:"sector-industry" flagdescr:"Optional comma-delimited sector or industry filter copied from the VolumeLeaders form." flagenv:"true" flaggroup:"Filters" mod:"trim"`
	NormalPrintsSelected        bool   `flag:"normal-prints" flagdescr:"Include normal prints in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	SignaturePrintsSelected     bool   `flag:"signature-prints" flagdescr:"Include signature prints in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	LatePrintsSelected          bool   `flag:"late-prints" flagdescr:"Include late prints in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	TimelyPrintsSelected        bool   `flag:"timely-prints" flagdescr:"Include timely prints in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	DarkPoolsSelected           bool   `flag:"dark-pools" flagdescr:"Include dark-pool prints in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	LitExchangesSelected        bool   `flag:"lit-exchanges" flagdescr:"Include lit-exchange prints in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	SweepsSelected              bool   `flag:"sweeps" flagdescr:"Include sweeps in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	BlocksSelected              bool   `flag:"blocks" flagdescr:"Include blocks in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	PremarketTradesSelected     bool   `flag:"premarket-trades" flagdescr:"Include premarket trades in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	RTHTradesSelected           bool   `flag:"rth-trades" flagdescr:"Include regular-hours trades in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	AHTradesSelected            bool   `flag:"ah-trades" flagdescr:"Include after-hours trades in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	OpeningTradesSelected       bool   `flag:"opening-trades" flagdescr:"Include opening-auction trades in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	ClosingTradesSelected       bool   `flag:"closing-trades" flagdescr:"Include closing-auction trades in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	PhantomTradesSelected       bool   `flag:"phantom-trades" flagdescr:"Include phantom trades in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	OffsettingTradesSelected    bool   `flag:"offsetting-trades" flagdescr:"Include offsetting trades in the saved watchlist filter. Defaults to true to match the browser create form." flagenv:"true" flaggroup:"Trade Types"`
	RSIOverboughtDailySelected  int    `flag:"rsi-overbought-daily" flagdescr:"Daily overbought RSI selector. Use -1 to ignore the condition, matching the browser default." flagenv:"true" flaggroup:"RSI"`
	RSIOverboughtHourlySelected int    `flag:"rsi-overbought-hourly" flagdescr:"Hourly overbought RSI selector. Use -1 to ignore the condition, matching the browser default." flagenv:"true" flaggroup:"RSI"`
	RSIOversoldDailySelected    int    `flag:"rsi-oversold-daily" flagdescr:"Daily oversold RSI selector. Use -1 to ignore the condition, matching the browser default." flagenv:"true" flaggroup:"RSI"`
	RSIOversoldHourlySelected   int    `flag:"rsi-oversold-hourly" flagdescr:"Hourly oversold RSI selector. Use -1 to ignore the condition, matching the browser default." flagenv:"true" flaggroup:"RSI"`
	Pretty                      bool   `flag:"pretty" flagdescr:"Pretty-print JSON output. Compact JSON is the default for token-efficient LLM and MCP use." flagenv:"true" flaggroup:"Output"`
}

// DeleteOptions defines the LLM-readable contract for removing saved watchlists.
type DeleteOptions struct {
	SearchTemplateKey int  `flag:"search-template-key" flagdescr:"Existing watchlist identifier to delete. Use SearchTemplateKey from the watchlists command output." flagenv:"true" flagrequired:"true" flaggroup:"Identity"`
	Pretty            bool `flag:"pretty" flagdescr:"Pretty-print JSON output. Compact JSON is the default for token-efficient LLM and MCP use." flagenv:"true" flaggroup:"Output"`
}

// Result is the stable response shape for configured watchlists.
type Result struct {
	Status     string              `json:"status"`
	Count      int                 `json:"count"`
	Fields     []string            `json:"fields,omitempty"`
	Rows       [][]json.RawMessage `json:"rows,omitempty"`
	WatchLists []json.RawMessage   `json:"watchlists,omitempty"`
}

// SaveResult is the stable response shape for saved watchlist mutations.
type SaveResult struct {
	Status            string `json:"status"`
	Action            string `json:"action"`
	SearchTemplateKey int    `json:"searchTemplateKey"`
	Name              string `json:"name"`
	FinalURL          string `json:"finalUrl,omitempty"`
}

// DeleteResult is the stable response shape for deleted watchlist mutations.
type DeleteResult struct {
	Status            string `json:"status"`
	Action            string `json:"action"`
	SearchTemplateKey int    `json:"searchTemplateKey"`
}

type getWatchListsResponse struct {
	Data  []json.RawMessage `json:"data"`
	Error string            `json:"error"`
}

type dataTableColumn struct {
	data      string
	name      string
	orderable string
}

type watchListOutput struct {
	fields     []string
	rows       [][]json.RawMessage
	watchLists []json.RawMessage
}

// NewCommand builds the configured watchlists command.
func NewCommand() (*cobra.Command, error) {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:     "watchlists",
		Aliases: []string{"watch-lists"},
		Short:   "List configured VolumeLeaders watchlists",
		Long:    "List the watchlists configured in the authenticated VolumeLeaders account. The default summary preset returns only SearchTemplateKey and Name so humans and LLMs can pick a watchlist before asking for its filter details. These watchlists are saved trade and cluster filter definitions that can include ticker, dollar, price, relative-size, rank, RSI, and trade-type criteria.\n\n" + watchListsGuide,
		Example: "volumeleaders-agent watchlists\nvolumeleaders-agent watchlists --search-template-key 4952 --preset-fields expanded\nvolumeleaders-agent watchlists --preset-fields expanded --pretty",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), cmd, opts)
		},
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind watchlists options: %w", err)
	}

	return cmd, nil
}

// NewSaveCommand builds the create/update watchlist command.
func NewSaveCommand() (*cobra.Command, error) {
	opts := defaultSaveOptions()
	cmd := &cobra.Command{
		Use:     "save-watchlist",
		Aliases: []string{"create-watchlist"},
		Short:   "Create or update a VolumeLeaders watchlist",
		Long:    "Create a new VolumeLeaders watchlist or replace an existing one by replaying the authenticated WatchListConfig form. SearchTemplateKey 0 creates a new watchlist; a positive key replaces that saved filter with the complete criteria supplied to this command.\n\n" + watchListsGuide,
		Example: "volumeleaders-agent save-watchlist --name 'Big dark-pool sweeps' --min-dollars 10000000 --min-relative-size 5\nvolumeleaders-agent save-watchlist --search-template-key 4952 --name BigOnes --tickers AAPL,MSFT --min-dollars 10000000 --max-trade-rank 10",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSave(cmd.Context(), cmd, opts)
		},
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind save-watchlist options: %w", err)
	}

	return cmd, nil
}

// NewDeleteCommand builds the delete watchlist command.
func NewDeleteCommand() (*cobra.Command, error) {
	opts := &DeleteOptions{}
	cmd := &cobra.Command{
		Use:     "delete-watchlist",
		Aliases: []string{"remove-watchlist"},
		Short:   "Delete a VolumeLeaders watchlist",
		Long:    "Delete an existing VolumeLeaders watchlist by replaying the authenticated WatchListConfigs/DeleteWatchList browser request. The command requires a non-zero SearchTemplateKey from the watchlists command output.\n\n" + watchListsGuide,
		Example: "volumeleaders-agent delete-watchlist --search-template-key 4952",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDelete(cmd.Context(), cmd, opts)
		},
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind delete-watchlist options: %w", err)
	}

	return cmd, nil
}

func defaultSaveOptions() *SaveOptions {
	return &SaveOptions{
		SecurityTypeKey:             -1,
		MaxTradeRankSelected:        -1,
		NormalPrintsSelected:        true,
		SignaturePrintsSelected:     true,
		LatePrintsSelected:          true,
		TimelyPrintsSelected:        true,
		DarkPoolsSelected:           true,
		LitExchangesSelected:        true,
		SweepsSelected:              true,
		BlocksSelected:              true,
		PremarketTradesSelected:     true,
		RTHTradesSelected:           true,
		AHTradesSelected:            true,
		OpeningTradesSelected:       true,
		ClosingTradesSelected:       true,
		PhantomTradesSelected:       true,
		OffsettingTradesSelected:    true,
		RSIOverboughtDailySelected:  -1,
		RSIOverboughtHourlySelected: -1,
		RSIOversoldDailySelected:    -1,
		RSIOversoldHourlySelected:   -1,
	}
}

func run(ctx context.Context, cmd *cobra.Command, opts *Options) error {
	fields, shape, err := outputOptions(opts.Fields, opts.PresetFields, opts.Shape)
	if err != nil {
		return err
	}

	apiResponse, err := fetchWatchLists(ctx)
	if err != nil {
		return err
	}
	watchLists, err := filterWatchLists(apiResponse.Data, opts.SearchTemplateKey)
	if err != nil {
		return err
	}

	result := Result{Status: "ok", Count: len(watchLists)}
	if err := applyWatchListOutput(&result, watchLists, fields, shape); err != nil {
		return err
	}

	return encodeResult(cmd.OutOrStdout(), &result, opts.Pretty)
}

func runSave(ctx context.Context, cmd *cobra.Command, opts *SaveOptions) error {
	if opts.SearchTemplateKey < 0 {
		return fmt.Errorf("search-template-key must be 0 for create or greater than 0 for save-watchlist updates")
	}

	finalURL, err := saveWatchList(ctx, opts)
	if err != nil {
		return err
	}

	action := "created"
	if opts.SearchTemplateKey != 0 {
		action = "updated"
	}
	result := SaveResult{
		Status:            "ok",
		Action:            action,
		SearchTemplateKey: opts.SearchTemplateKey,
		Name:              opts.Name,
		FinalURL:          finalURL,
	}

	return encodeResult(cmd.OutOrStdout(), &result, opts.Pretty)
}

func runDelete(ctx context.Context, cmd *cobra.Command, opts *DeleteOptions) error {
	if opts.SearchTemplateKey <= 0 {
		return fmt.Errorf("search-template-key must be greater than 0 for delete-watchlist")
	}

	if err := deleteWatchList(ctx, opts.SearchTemplateKey); err != nil {
		return err
	}
	result := DeleteResult{
		Status:            "ok",
		Action:            "deleted",
		SearchTemplateKey: opts.SearchTemplateKey,
	}

	return encodeResult(cmd.OutOrStdout(), &result, opts.Pretty)
}

func outputOptions(rawFields, rawPreset, rawShape string) (fields []string, shape string, err error) {
	shape = strings.ToLower(strings.TrimSpace(rawShape))
	if shape == "" {
		shape = defaultOutputShape
	}
	if shape != defaultOutputShape && shape != objectOutputShape {
		return nil, "", fmt.Errorf("invalid shape %q: use array or objects", rawShape)
	}

	if strings.TrimSpace(rawFields) != "" {
		return splitFields(rawFields), shape, nil
	}

	preset := strings.ToLower(strings.TrimSpace(rawPreset))
	if preset == "" {
		preset = defaultFieldPreset
	}
	if preset == fullFieldPreset {
		return nil, shape, nil
	}
	fields, ok := watchListFieldPresets[preset]
	if !ok {
		return nil, "", fmt.Errorf("invalid preset-fields %q: use summary, expanded, or full", rawPreset)
	}

	return fields, shape, nil
}

func filterWatchLists(watchLists []json.RawMessage, searchTemplateKey int) ([]json.RawMessage, error) {
	if searchTemplateKey == 0 {
		return watchLists, nil
	}
	if searchTemplateKey < 0 {
		return nil, fmt.Errorf("search-template-key must be greater than or equal to 0")
	}

	filtered := make([]json.RawMessage, 0, 1)
	for _, watchList := range watchLists {
		object, err := decodeWatchListObject(watchList)
		if err != nil {
			return nil, err
		}
		if watchListKey(object) == searchTemplateKey {
			filtered = append(filtered, watchList)
		}
	}
	return filtered, nil
}

func watchListKey(object map[string]json.RawMessage) int {
	rawKey, ok := object["SearchTemplateKey"]
	if !ok {
		return 0
	}
	var key int
	if err := json.Unmarshal(rawKey, &key); err != nil {
		return 0
	}
	return key
}

func splitFields(rawFields string) []string {
	parts := strings.Split(rawFields, ",")
	fields := make([]string, 0, len(parts))
	for _, part := range parts {
		field := strings.TrimSpace(part)
		if field != "" {
			fields = append(fields, field)
		}
	}
	return fields
}

func applyWatchListOutput(result *Result, watchLists []json.RawMessage, fields []string, shape string) error {
	output, err := projectWatchListOutput(watchLists, fields, shape)
	if err != nil {
		return err
	}
	result.Fields = output.fields
	result.Rows = output.rows
	result.WatchLists = output.watchLists
	return nil
}

func projectWatchListOutput(watchLists []json.RawMessage, fields []string, shape string) (watchListOutput, error) {
	if watchLists == nil {
		watchLists = []json.RawMessage{}
	}
	if fields == nil {
		return watchListOutput{watchLists: watchLists}, nil
	}

	if shape == objectOutputShape {
		projected, err := projectWatchListObjects(watchLists, fields)
		if err != nil {
			return watchListOutput{}, err
		}
		return watchListOutput{fields: fields, watchLists: projected}, nil
	}

	rows, err := projectWatchListRows(watchLists, fields)
	if err != nil {
		return watchListOutput{}, err
	}
	return watchListOutput{fields: fields, rows: rows}, nil
}

func projectWatchListObjects(watchLists []json.RawMessage, fields []string) ([]json.RawMessage, error) {
	projected := make([]json.RawMessage, 0, len(watchLists))
	for _, watchList := range watchLists {
		object, err := decodeWatchListObject(watchList)
		if err != nil {
			return nil, err
		}
		row := make(map[string]json.RawMessage, len(fields))
		for _, field := range fields {
			if value, ok := projectedFieldValue(object, field); ok {
				row[field] = value
			}
		}
		encoded, err := json.Marshal(row)
		if err != nil {
			return nil, fmt.Errorf("encode projected watchlist row: %w", err)
		}
		projected = append(projected, encoded)
	}

	return projected, nil
}

func projectWatchListRows(watchLists []json.RawMessage, fields []string) ([][]json.RawMessage, error) {
	rows := make([][]json.RawMessage, 0, len(watchLists))
	for _, watchList := range watchLists {
		object, err := decodeWatchListObject(watchList)
		if err != nil {
			return nil, err
		}
		row := make([]json.RawMessage, 0, len(fields))
		for _, field := range fields {
			value, ok := projectedFieldValue(object, field)
			if !ok {
				value = nullJSON
			}
			row = append(row, value)
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func decodeWatchListObject(watchList json.RawMessage) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(watchList, &object); err != nil {
		return nil, fmt.Errorf("decode watchlist row: %w", err)
	}

	return object, nil
}

func projectedFieldValue(object map[string]json.RawMessage, field string) (json.RawMessage, bool) {
	if field == includedTradeTypesName {
		return includedTradeTypesValue(object)
	}

	value, ok := object[field]
	return value, ok
}

func includedTradeTypesValue(object map[string]json.RawMessage) (json.RawMessage, bool) {
	enabled := make([]string, 0, len(tradeTypeFields))
	for _, field := range tradeTypeFields {
		if rawValue, ok := object[field]; ok {
			value, valid := rawJSONFlag(rawValue)
			if valid && value {
				enabled = append(enabled, field)
			}
		}
	}
	if len(enabled) == 0 {
		return nil, false
	}

	encoded, err := json.Marshal(enabled)
	if err != nil {
		return nil, false
	}
	return encoded, true
}

func rawJSONFlag(rawValue json.RawMessage) (value, ok bool) {
	var boolValue bool
	if err := json.Unmarshal(rawValue, &boolValue); err == nil {
		return boolValue, true
	}

	var numericValue int
	if err := json.Unmarshal(rawValue, &numericValue); err != nil {
		return false, false
	}
	return numericValue != 0, true
}

func encodeResult(w io.Writer, result any, pretty bool) error {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("encode watchlist response: %w", err)
	}

	return nil
}

func saveWatchList(ctx context.Context, opts *SaveOptions) (string, error) {
	cookies, err := extractCookies(ctx)
	if err != nil {
		return "", fmt.Errorf("extract VolumeLeaders browser cookies: %w", err)
	}

	token, err := fetchXSRFToken(ctx, getWatchListsHTTPClient, cookies)
	if err != nil {
		return "", fmt.Errorf("fetch VolumeLeaders XSRF token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, saveWatchListEndpoint, strings.NewReader(saveWatchListForm(opts, token).Encode()))
	if err != nil {
		return "", fmt.Errorf("create WatchListConfig request: %w", err)
	}
	setSaveWatchListHeaders(req, token, opts.SearchTemplateKey)
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	resp, err := getWatchListsHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("post WatchListConfig request: %w", err)
	}
	defer resp.Body.Close()

	if sessionExpiredResponse(resp) {
		return "", sessionExpiredCommandError()
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("WatchListConfig request returned status %d", resp.StatusCode)
	}

	if resp.Request == nil || resp.Request.URL == nil {
		return "", nil
	}
	return resp.Request.URL.String(), nil
}

func saveWatchListForm(opts *SaveOptions, token string) url.Values {
	form := url.Values{}
	form.Set("__RequestVerificationToken", token)
	form.Set("SearchTemplateKey", fmt.Sprintf("%d", opts.SearchTemplateKey))
	form.Set("Name", opts.Name)
	form.Set("Tickers", opts.Tickers)
	form.Set("MinVolume", opts.MinVolume)
	form.Set("MaxVolume", opts.MaxVolume)
	form.Set("MinDollars", opts.MinDollars)
	form.Set("MaxDollars", opts.MaxDollars)
	form.Set("MinPrice", opts.MinPrice)
	form.Set("MaxPrice", opts.MaxPrice)
	form.Set("SecurityTypeKey", fmt.Sprintf("%d", opts.SecurityTypeKey))
	form.Set("MinRelativeSizeSelected", fmt.Sprintf("%d", opts.MinRelativeSizeSelected))
	form.Set("MaxTradeRankSelected", fmt.Sprintf("%d", opts.MaxTradeRankSelected))
	form.Set("MinVCD", opts.MinVCD)
	form.Set("SectorIndustry", opts.SectorIndustry)
	addCheckboxFormField(form, "NormalPrintsSelected", opts.NormalPrintsSelected)
	addCheckboxFormField(form, "SignaturePrintsSelected", opts.SignaturePrintsSelected)
	addCheckboxFormField(form, "LatePrintsSelected", opts.LatePrintsSelected)
	addCheckboxFormField(form, "TimelyPrintsSelected", opts.TimelyPrintsSelected)
	addCheckboxFormField(form, "DarkPoolsSelected", opts.DarkPoolsSelected)
	addCheckboxFormField(form, "LitExchangesSelected", opts.LitExchangesSelected)
	addCheckboxFormField(form, "SweepsSelected", opts.SweepsSelected)
	addCheckboxFormField(form, "BlocksSelected", opts.BlocksSelected)
	addCheckboxFormField(form, "PremarketTradesSelected", opts.PremarketTradesSelected)
	addCheckboxFormField(form, "RTHTradesSelected", opts.RTHTradesSelected)
	addCheckboxFormField(form, "AHTradesSelected", opts.AHTradesSelected)
	addCheckboxFormField(form, "OpeningTradesSelected", opts.OpeningTradesSelected)
	addCheckboxFormField(form, "ClosingTradesSelected", opts.ClosingTradesSelected)
	addCheckboxFormField(form, "PhantomTradesSelected", opts.PhantomTradesSelected)
	addCheckboxFormField(form, "OffsettingTradesSelected", opts.OffsettingTradesSelected)
	form.Set("RSIOverboughtDailySelected", fmt.Sprintf("%d", opts.RSIOverboughtDailySelected))
	form.Set("RSIOverboughtHourlySelected", fmt.Sprintf("%d", opts.RSIOverboughtHourlySelected))
	form.Set("RSIOversoldDailySelected", fmt.Sprintf("%d", opts.RSIOversoldDailySelected))
	form.Set("RSIOversoldHourlySelected", fmt.Sprintf("%d", opts.RSIOversoldHourlySelected))
	return form
}

func addCheckboxFormField(form url.Values, name string, selected bool) {
	if selected {
		form.Add(name, "true")
	}
	form.Add(name, "false")
}

func deleteWatchList(ctx context.Context, searchTemplateKey int) error {
	cookies, err := extractCookies(ctx)
	if err != nil {
		return fmt.Errorf("extract VolumeLeaders browser cookies: %w", err)
	}

	token, err := fetchXSRFToken(ctx, getWatchListsHTTPClient, cookies)
	if err != nil {
		return fmt.Errorf("fetch VolumeLeaders XSRF token: %w", err)
	}

	body, err := json.Marshal(map[string]int{"WatchListKey": searchTemplateKey})
	if err != nil {
		return fmt.Errorf("encode DeleteWatchList request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deleteWatchListEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create DeleteWatchList request: %w", err)
	}
	setDeleteWatchListHeaders(req, token)
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	resp, err := getWatchListsHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("post DeleteWatchList request: %w", err)
	}
	defer resp.Body.Close()

	if sessionExpiredResponse(resp) {
		return sessionExpiredCommandError()
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("DeleteWatchList request returned status %d", resp.StatusCode)
	}
	if err := decodeDeleteWatchListResponse(resp.Body); err != nil {
		return err
	}

	return nil
}

func decodeDeleteWatchListResponse(body io.Reader) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read DeleteWatchList response: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return fmt.Errorf("DeleteWatchList response was empty")
	}

	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("decode DeleteWatchList response JSON: %w", err)
	}
	if errorMessage, ok := deleteWatchListError(payload); ok {
		return fmt.Errorf("DeleteWatchList response returned error: %s", errorMessage)
	}

	return nil
}

func deleteWatchListError(payload any) (string, bool) {
	response, ok := payload.(map[string]any)
	if !ok {
		return "", false
	}

	for _, name := range []string{"error", "Error", "message", "Message"} {
		value, ok := response[name]
		if !ok {
			continue
		}
		message, ok := value.(string)
		if !ok || strings.TrimSpace(message) == "" {
			continue
		}
		return strings.TrimSpace(message), true
	}

	if success, ok := response["success"].(bool); ok && !success {
		return "success=false", true
	}

	return "", false
}

func fetchWatchLists(ctx context.Context) (getWatchListsResponse, error) {
	cookies, err := extractCookies(ctx)
	if err != nil {
		return getWatchListsResponse{}, fmt.Errorf("extract VolumeLeaders browser cookies: %w", err)
	}

	token, err := fetchXSRFToken(ctx, getWatchListsHTTPClient, cookies)
	if err != nil {
		return getWatchListsResponse{}, fmt.Errorf("fetch VolumeLeaders XSRF token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, getWatchListsEndpoint, strings.NewReader(getWatchListsForm().Encode()))
	if err != nil {
		return getWatchListsResponse{}, fmt.Errorf("create GetWatchLists request: %w", err)
	}
	setGetWatchListsHeaders(req, token)
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	resp, err := getWatchListsHTTPClient.Do(req)
	if err != nil {
		return getWatchListsResponse{}, fmt.Errorf("post GetWatchLists request: %w", err)
	}
	defer resp.Body.Close()

	if sessionExpiredResponse(resp) {
		return getWatchListsResponse{}, sessionExpiredCommandError()
	}
	if resp.StatusCode != http.StatusOK {
		return getWatchListsResponse{}, fmt.Errorf("GetWatchLists request returned status %d", resp.StatusCode)
	}

	bodyReader, closeReader, err := responseBodyReader(resp, "GetWatchLists")
	if err != nil {
		return getWatchListsResponse{}, err
	}
	defer closeReader()

	var apiResponse getWatchListsResponse
	if err := json.NewDecoder(bodyReader).Decode(&apiResponse); err != nil {
		return getWatchListsResponse{}, fmt.Errorf("decode GetWatchLists response: %w", err)
	}
	if apiResponse.Error != "" {
		return getWatchListsResponse{}, fmt.Errorf("GetWatchLists response error: %s", apiResponse.Error)
	}
	if apiResponse.Data == nil {
		apiResponse.Data = []json.RawMessage{}
	}

	return apiResponse, nil
}

func getWatchListsForm() url.Values {
	form := url.Values{}
	setDataTableFormFields(form, getWatchListsColumns)
	return form
}

func setDataTableFormFields(form url.Values, columns []dataTableColumn) {
	form.Set("draw", "1")
	for i, column := range columns {
		prefix := fmt.Sprintf("columns[%d]", i)
		form.Set(prefix+"[data]", column.data)
		form.Set(prefix+"[name]", column.name)
		form.Set(prefix+"[searchable]", "true")
		form.Set(prefix+"[orderable]", column.orderable)
		form.Set(prefix+"[search][value]", "")
		form.Set(prefix+"[search][regex]", "false")
	}
	form.Set("order[0][column]", "0")
	form.Set("order[0][dir]", "asc")
	form.Set("start", "0")
	form.Set("length", "-1")
	form.Set("search[value]", "")
	form.Set("search[regex]", "false")
}

func setGetWatchListsHeaders(req *http.Request, token string) {
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-XSRF-Token", token)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://www.volumeleaders.com")
	req.Header.Set("Referer", watchListsPage)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
}

func setSaveWatchListHeaders(req *http.Request, token string, searchTemplateKey int) {
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-XSRF-Token", token)
	req.Header.Set("Origin", "https://www.volumeleaders.com")
	req.Header.Set("Referer", saveWatchListReferer(searchTemplateKey))
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
}

func setDeleteWatchListHeaders(req *http.Request, token string) {
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-XSRF-Token", token)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://www.volumeleaders.com")
	req.Header.Set("Referer", watchListsPage)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
}

func saveWatchListReferer(searchTemplateKey int) string {
	if searchTemplateKey == 0 {
		return watchListConfigPage
	}
	return fmt.Sprintf("%s?SearchTemplateKey=%d", watchListConfigPage, searchTemplateKey)
}

func sessionExpiredResponse(resp *http.Response) bool {
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return true
	}
	if resp.Request == nil || resp.Request.URL == nil {
		return false
	}
	return strings.EqualFold(resp.Request.URL.EscapedPath(), "/Login")
}

func sessionExpiredCommandError() error {
	return fmt.Errorf("%s: %w", auth.SessionExpiredMessage, auth.ErrSessionExpired)
}

func responseBodyReader(resp *http.Response, operation string) (io.Reader, func(), error) {
	if resp.Header.Get("Content-Encoding") != "gzip" {
		return resp.Body, func() {}, nil
	}
	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, func() {}, fmt.Errorf("decompress %s response: %w", operation, err)
	}
	return gr, func() { _ = gr.Close() }, nil
}
