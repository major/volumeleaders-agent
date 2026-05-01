// Package watchlists contains commands for VolumeLeaders watchlist workflows.
package watchlists

import (
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
	defaultFieldPreset     = "core"
	defaultOutputShape     = "array"
	fullFieldPreset        = "full"
	objectOutputShape      = "objects"
	getWatchListsPath      = "https://www.volumeleaders.com/WatchListConfigs/GetWatchLists"
	watchListsPage         = "https://www.volumeleaders.com/WatchListConfigs"
	includedTradeTypesName = "IncludedTradeTypes"
	watchListsGuide        = "LLM field guide: VolumeLeaders watchlists are saved trade and cluster filters, not necessarily long ticker lists. Name identifies the saved filter. Tickers is empty when the watchlist applies to all symbols. MinVolume, MaxVolume, MinDollars, MaxDollars, MinPrice, MaxPrice, MinRelativeSize, MaxTradeRank, RSI fields, Conditions, SecurityTypeKey, SectorIndustry, and included trade-type booleans describe which trades appear on the watchlist. MaxTradeRank=-1 means no rank limit; otherwise trades must rank at that level or better, so 10 means ranks 1 through 10. SecurityTypeKey=-1 means all security types, 1 means stocks only, 26 means ETFs only, and 4 means REITs only. Conditions entries such as IgnoreOBD and IgnoreOBH mean the corresponding daily or hourly RSI condition is ignored. IncludedTradeTypes is a compact derived field listing enabled print, venue, session, auction, phantom, and offsetting categories. Internal-only upstream fields such as SearchTemplateTypeKey, SortOrder, MinVCD, and APIKey are excluded from expanded output; use --preset-fields full only when raw upstream payloads are needed for debugging."
)

var (
	extractCookies          = auth.ExtractCookies
	fetchXSRFToken          = auth.FetchXSRFToken
	getWatchListsHTTPClient = http.DefaultClient
	getWatchListsEndpoint   = getWatchListsPath
	nullJSON                = json.RawMessage("null")
)

var watchListFieldPresets = map[string][]string{
	"core": {
		"SearchTemplateKey",
		"Name",
		"Tickers",
		"MinVolume",
		"MaxVolume",
		"MinDollars",
		"MaxDollars",
		"MinPrice",
		"MaxPrice",
		"MinRelativeSize",
		"MaxTradeRank",
		"Conditions",
		includedTradeTypesName,
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
	Fields       string `flag:"fields" flagdescr:"Comma-separated watchlist fields to include. Overrides --preset-fields. Use upstream field names such as Name,Tickers,MaxTradeRank." flagenv:"true" flaggroup:"Output" mod:"trim"`
	PresetFields string `flag:"preset-fields" flagdescr:"Field preset to include: core, expanded, or full. Defaults to core for token-efficient output. Expanded includes annotated non-internal watchlist filter fields; full returns raw upstream payloads." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Shape        string `flag:"shape" flagdescr:"Watchlist row shape: array or objects. Array is the default and is most token-efficient." flagenv:"true" flaggroup:"Output" mod:"trim"`
	Pretty       bool   `flag:"pretty" flagdescr:"Pretty-print JSON output. Compact JSON is the default for token-efficient LLM and MCP use." flagenv:"true" flaggroup:"Output"`
}

// Result is the stable response shape for configured watchlists.
type Result struct {
	Status     string              `json:"status"`
	Count      int                 `json:"count"`
	Fields     []string            `json:"fields,omitempty"`
	Rows       [][]json.RawMessage `json:"rows,omitempty"`
	WatchLists []json.RawMessage   `json:"watchlists,omitempty"`
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
		Long:    "List the watchlists configured in the authenticated VolumeLeaders account. These watchlists are saved trade and cluster filter definitions that can include ticker, dollar, price, relative-size, rank, RSI, and trade-type criteria.\n\n" + watchListsGuide,
		Example: "volumeleaders-agent watchlists\nvolumeleaders-agent watchlists --preset-fields expanded --pretty",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), cmd, opts)
		},
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind watchlists options: %w", err)
	}

	return cmd, nil
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

	result := Result{Status: "ok", Count: len(apiResponse.Data)}
	if err := applyWatchListOutput(&result, apiResponse.Data, fields, shape); err != nil {
		return err
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
		return nil, "", fmt.Errorf("invalid preset-fields %q: use core, expanded, or full", rawPreset)
	}

	return fields, shape, nil
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

func encodeResult(w io.Writer, result *Result, pretty bool) error {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("encode watchlists response: %w", err)
	}

	return nil
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
