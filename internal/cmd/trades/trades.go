// Package trades contains commands for VolumeLeaders trade workflows.
package trades

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/leodido/structcli"
	"github.com/major/volumeleaders-agent/internal/auth"
	"github.com/spf13/cobra"
)

const (
	dateLayout    = "2006-01-02"
	getTradesPath = "https://www.volumeleaders.com/Trades/GetTrades"
	tradesPage    = "https://www.volumeleaders.com/Trades"
)

var (
	extractCookies      = auth.ExtractCookies
	fetchXSRFToken      = auth.FetchXSRFToken
	getTradesHTTPClient = http.DefaultClient
	getTradesEndpoint   = getTradesPath
	tickerPattern       = regexp.MustCompile(`^[A-Z0-9.-]+$`)
)

// Options defines the LLM-readable contract for fetching unusual trades.
type Options struct {
	Date    string `flag:"date" flagshort:"d" flagdescr:"Single trading date to query, formatted as YYYY-MM-DD. The disproportionately large trades preset is intentionally limited to one day." flagenv:"true" flagrequired:"true" flaggroup:"Query" validate:"required" mod:"trim"`
	Tickers string `flag:"tickers" flagdescr:"Optional ticker filter. Use one symbol or a comma-delimited list without spaces, for example AAPL or AAPL,MSFT." flagenv:"true" flaggroup:"Query" mod:"trim"`
}

// Result is the stable response shape for the unusual trades command.
type Result struct {
	Status          string            `json:"status"`
	Date            string            `json:"date"`
	RecordsTotal    int               `json:"recordsTotal"`
	RecordsFiltered int               `json:"recordsFiltered"`
	Trades          []json.RawMessage `json:"trades"`
}

type getTradesResponse struct {
	Draw            int               `json:"draw"`
	RecordsTotal    int               `json:"recordsTotal"`
	RecordsFiltered int               `json:"recordsFiltered"`
	Data            []json.RawMessage `json:"data"`
	Error           string            `json:"error"`
}

type tradeColumn struct {
	data      string
	name      string
	orderable string
}

var getTradesColumns = []tradeColumn{
	{data: "FullTimeString24", name: "", orderable: "false"},
	{data: "FullTimeString24", name: "FullTimeString24", orderable: "true"},
	{data: "Ticker", name: "Ticker", orderable: "true"},
	{data: "Current", name: "Current", orderable: "false"},
	{data: "Trade", name: "Trade", orderable: "false"},
	{data: "Sector", name: "Sector", orderable: "true"},
	{data: "Industry", name: "Industry", orderable: "true"},
	{data: "Volume", name: "Sh", orderable: "true"},
	{data: "Dollars", name: "$$", orderable: "true"},
	{data: "DollarsMultiplier", name: "RS", orderable: "true"},
	{data: "CumulativeDistribution", name: "PCT", orderable: "true"},
	{data: "TradeRank", name: "Rank", orderable: "true"},
	{data: "LastComparibleTradeDate", name: "Last Traded", orderable: "true"},
	{data: "LastComparibleTradeDate", name: "Charts", orderable: "false"},
}

// NewCommand builds the large unusual trades command.
func NewCommand() (*cobra.Command, error) {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:     "trades",
		Aliases: []string{"large-trades", "unusual-trades"},
		Short:   "Fetch disproportionately large trades for one date",
		Long:    "Fetch VolumeLeaders disproportionately large trades for a single trading day. This reproduces the default Disproportionately large trades GetTrades request and intentionally does not allow multi-day ranges.",
		Example: "volumeleaders-agent trades --date 2026-04-30\nvolumeleaders-agent trades --date 2026-04-30 --tickers AAPL,MSFT",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), cmd, opts)
		},
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind trades options: %w", err)
	}
	cmd.Flags().StringVar(&opts.Tickers, "ticker", "", "Optional ticker filter. Alias for --tickers.")

	return cmd, nil
}

func run(ctx context.Context, cmd *cobra.Command, opts *Options) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("run trades command: %w", ctx.Err())
	default:
	}

	tradeDate, err := time.Parse(dateLayout, opts.Date)
	if err != nil {
		return fmt.Errorf("invalid date %q: use YYYY-MM-DD: %w", opts.Date, err)
	}
	formattedDate := tradeDate.Format(dateLayout)
	tickers, err := normalizeTickers(opts.Tickers)
	if err != nil {
		return err
	}

	apiResponse, err := fetchDisproportionatelyLargeTrades(ctx, formattedDate, tickers)
	if err != nil {
		return err
	}

	result := Result{
		Status:          "ok",
		Date:            formattedDate,
		RecordsTotal:    apiResponse.RecordsTotal,
		RecordsFiltered: apiResponse.RecordsFiltered,
		Trades:          apiResponse.Data,
	}
	if result.Trades == nil {
		result.Trades = []json.RawMessage{}
	}

	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("encode trades response: %w", err)
	}

	return nil
}

func fetchDisproportionatelyLargeTrades(ctx context.Context, tradeDate, tickers string) (getTradesResponse, error) {
	cookies, err := extractCookies(ctx)
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("extract VolumeLeaders browser cookies: %w", err)
	}

	token, err := fetchXSRFToken(ctx, getTradesHTTPClient, cookies)
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("fetch VolumeLeaders XSRF token: %w", err)
	}

	form := getTradesForm(tradeDate, tickers)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, getTradesEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("create GetTrades request: %w", err)
	}
	setGetTradesHeaders(req, token, tradeDate, tickers)
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	resp, err := getTradesHTTPClient.Do(req)
	if err != nil {
		return getTradesResponse{}, fmt.Errorf("post GetTrades request: %w", err)
	}
	defer resp.Body.Close()

	if sessionExpiredResponse(resp) {
		return getTradesResponse{}, sessionExpiredCommandError()
	}
	if resp.StatusCode != http.StatusOK {
		return getTradesResponse{}, fmt.Errorf("GetTrades request returned status %d", resp.StatusCode)
	}

	bodyReader, closeReader, err := responseBodyReader(resp)
	if err != nil {
		return getTradesResponse{}, err
	}
	defer closeReader()

	var apiResponse getTradesResponse
	if err := json.NewDecoder(bodyReader).Decode(&apiResponse); err != nil {
		return getTradesResponse{}, fmt.Errorf("decode GetTrades response: %w", err)
	}
	if apiResponse.Error != "" {
		return getTradesResponse{}, fmt.Errorf("GetTrades response error: %s", apiResponse.Error)
	}
	if apiResponse.Data == nil {
		apiResponse.Data = []json.RawMessage{}
	}

	return apiResponse, nil
}

func setGetTradesHeaders(req *http.Request, token, tradeDate, tickers string) {
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-XSRF-Token", token)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://www.volumeleaders.com")
	req.Header.Set("Referer", tradesReferer(tradeDate, tickers))
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
}

func tradesReferer(tradeDate, tickers string) string {
	query := url.Values{}
	query.Set("Tickers", tickers)
	query.Set("StartDate", tradeDate)
	query.Set("EndDate", tradeDate)
	query.Set("MinVolume", "0")
	query.Set("MaxVolume", "2000000000")
	query.Set("Conditions", "-1")
	query.Set("VCD", "0")
	query.Set("RelativeSize", "5")
	query.Set("DarkPools", "-1")
	query.Set("Sweeps", "-1")
	query.Set("LatePrints", "-1")
	query.Set("SignaturePrints", "-1")
	query.Set("EvenShared", "-1")
	query.Set("SecurityTypeKey", "-1")
	query.Set("MinPrice", "0")
	query.Set("MaxPrice", "100000")
	query.Set("MinDollars", "500000")
	query.Set("MaxDollars", "30000000000")
	query.Set("TradeRank", "-1")
	query.Set("TradeRankSnapshot", "-1")
	query.Set("MarketCap", "0")
	query.Set("IncludePremarket", "1")
	query.Set("IncludeRTH", "1")
	query.Set("IncludeAH", "1")
	query.Set("IncludeOpening", "1")
	query.Set("IncludeClosing", "1")
	query.Set("IncludePhantom", "1")
	query.Set("IncludeOffsetting", "1")
	query.Set("SectorIndustry", "")
	query.Set("PresetSearchTemplateID", "87")
	query.Set("ViewMode", "Automatic")
	return tradesPage + "?" + query.Encode()
}

func getTradesForm(tradeDate, tickers string) url.Values {
	form := url.Values{}
	form.Set("draw", "1")
	for i, column := range getTradesColumns {
		prefix := fmt.Sprintf("columns[%d]", i)
		form.Set(prefix+"[data]", column.data)
		form.Set(prefix+"[name]", column.name)
		form.Set(prefix+"[searchable]", "true")
		form.Set(prefix+"[orderable]", column.orderable)
		form.Set(prefix+"[search][value]", "")
		form.Set(prefix+"[search][regex]", "false")
	}
	form.Set("order[0][column]", "1")
	form.Set("order[0][dir]", "DESC")
	form.Set("order[0][name]", "FullTimeString24")
	form.Set("start", "0")
	form.Set("length", "100")
	form.Set("search[value]", "")
	form.Set("search[regex]", "false")
	form.Set("Tickers", tickers)
	form.Set("StartDate", tradeDate)
	form.Set("EndDate", tradeDate)
	form.Set("MinVolume", "0")
	form.Set("MaxVolume", "2000000000")
	form.Set("MinPrice", "0")
	form.Set("MaxPrice", "100000")
	form.Set("MinDollars", "500000")
	form.Set("MaxDollars", "30000000000")
	form.Set("Conditions", "-1")
	form.Set("VCD", "0")
	form.Set("SecurityTypeKey", "-1")
	form.Set("RelativeSize", "5")
	form.Set("DarkPools", "-1")
	form.Set("Sweeps", "-1")
	form.Set("LatePrints", "-1")
	form.Set("SignaturePrints", "-1")
	form.Set("EvenShared", "-1")
	form.Set("TradeRank", "-1")
	form.Set("TradeRankSnapshot", "-1")
	form.Set("MarketCap", "0")
	form.Set("IncludePremarket", "1")
	form.Set("IncludeRTH", "1")
	form.Set("IncludeAH", "1")
	form.Set("IncludeOpening", "1")
	form.Set("IncludeClosing", "1")
	form.Set("IncludePhantom", "1")
	form.Set("IncludeOffsetting", "1")
	form.Set("SectorIndustry", "")
	return form
}

func normalizeTickers(rawTickers string) (string, error) {
	if strings.TrimSpace(rawTickers) == "" {
		return "", nil
	}
	if strings.ContainsAny(rawTickers, " \t\n\r") {
		return "", fmt.Errorf("invalid tickers %q: use a comma-delimited list without spaces", rawTickers)
	}

	parts := strings.Split(rawTickers, ",")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		ticker := strings.ToUpper(part)
		if ticker == "" {
			return "", fmt.Errorf("invalid tickers %q: comma-delimited list contains an empty ticker", rawTickers)
		}
		if !tickerPattern.MatchString(ticker) {
			return "", fmt.Errorf("invalid ticker %q: use letters, numbers, dots, or hyphens", ticker)
		}
		normalized = append(normalized, ticker)
	}

	return strings.Join(normalized, ","), nil
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

func responseBodyReader(resp *http.Response) (io.Reader, func(), error) {
	if resp.Header.Get("Content-Encoding") != "gzip" {
		return resp.Body, func() {}, nil
	}
	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, func() {}, fmt.Errorf("decompress GetTrades response: %w", err)
	}
	return gr, func() { _ = gr.Close() }, nil
}
