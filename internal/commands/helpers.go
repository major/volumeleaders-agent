package commands

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/major/volumeleaders-agent/internal/auth"
	"github.com/major/volumeleaders-agent/internal/client"
	"github.com/major/volumeleaders-agent/internal/datatables"
	cli "github.com/urfave/cli/v3"
)

type contextKey int

const (
	prettyJSONKey contextKey = iota
	testClientKey
)

type dataTableOptions struct {
	start, length, orderCol int
	orderDir                string
	filters                 map[string]string
	fields                  []string
}

type outputFormat string

const (
	outputFormatJSON outputFormat = "json"
	outputFormatCSV  outputFormat = "csv"
	outputFormatTSV  outputFormat = "tsv"
)

// paginationPageSize is the number of records fetched per page when the user
// requests all results (--length -1).
const paginationPageSize = 1000

// runDataTablesCommand is the shared handler for all DataTables-backed
// commands. It creates the client, builds and posts the request, and
// prints the result as JSON. The label is used in error messages.
// When opts.length is negative, results are fetched in pages of
// paginationPageSize until all records have been retrieved.
func runDataTablesCommand[T any](ctx context.Context, path string, columns []string, opts dataTableOptions, formatValue, label string) error {
	format, err := parseOutputFormat(formatValue)
	if err != nil {
		return err
	}

	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	if opts.length < 0 {
		return runPaginatedCommand[T](ctx, vlClient, path, columns, opts, format, label)
	}

	request := newDataTablesRequest(columns, opts)
	var result []T
	if err := vlClient.PostDataTables(ctx, path, request.Encode(), &result); err != nil {
		slog.Error("failed to "+label, "error", err)
		return fmt.Errorf("%s: %w", label, err)
	}
	return printDataTablesResult(ctx, result, opts.fields, format)
}

// runPaginatedCommand fetches all records by paginating through the DataTables
// endpoint in pages of paginationPageSize. It accumulates results across pages
// and stops when all filtered records have been retrieved or the server returns
// an empty page.
func runPaginatedCommand[T any](ctx context.Context, vlClient *client.Client, path string, columns []string, opts dataTableOptions, format outputFormat, label string) error {
	opts.length = paginationPageSize
	all := make([]T, 0)

	for {
		request := newDataTablesRequest(columns, opts)
		resp, err := vlClient.PostDataTablesPage(ctx, path, request.Encode())
		if err != nil {
			slog.Error("failed to "+label, "error", err)
			return fmt.Errorf("%s: %w", label, err)
		}

		var page []T
		if err := json.Unmarshal(resp.Data, &page); err != nil {
			break
		}
		if len(page) == 0 {
			break
		}

		all = append(all, page...)

		if resp.RecordsFiltered > 0 && len(all) >= resp.RecordsFiltered {
			break
		}
		if len(page) < paginationPageSize {
			break
		}

		opts.start += len(page)
	}

	return printDataTablesResult(ctx, all, opts.fields, format)
}

func printDataTablesResult[T any](ctx context.Context, result []T, fields []string, format outputFormat) error {
	if format == outputFormatJSON && len(fields) == 0 {
		return printJSON(ctx, result)
	}

	selectedFields := fields
	if len(selectedFields) == 0 {
		selectedFields = jsonFieldNamesInOrder[T]()
	}

	selected, err := selectJSONFields(result, selectedFields)
	if err != nil {
		return err
	}

	switch format {
	case outputFormatJSON:
		return printJSON(ctx, selected)
	case outputFormatCSV:
		return printDelimited(selected, selectedFields, ',')
	case outputFormatTSV:
		return printDelimited(selected, selectedFields, '\t')
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func selectJSONFields[T any](items []T, fields []string) ([]map[string]json.RawMessage, error) {
	selected := make([]map[string]json.RawMessage, 0, len(items))
	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("marshal item for field selection: %w", err)
		}

		var allFields map[string]json.RawMessage
		if err := json.Unmarshal(data, &allFields); err != nil {
			return nil, fmt.Errorf("decode item for field selection: %w", err)
		}

		row := make(map[string]json.RawMessage, len(fields))
		for _, field := range fields {
			row[field] = allFields[field]
		}
		selected = append(selected, row)
	}
	return selected, nil
}

func parseJSONFieldList[T any](value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	validFields := jsonFieldNames[T]()
	valid := make(map[string]struct{}, len(validFields))
	for _, field := range validFields {
		valid[field] = struct{}{}
	}

	fields := make([]string, 0, strings.Count(value, ",")+1)
	seen := make(map[string]struct{})
	for field := range strings.SplitSeq(value, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if _, ok := valid[field]; !ok {
			return nil, fmt.Errorf("invalid field %q; valid fields: %s", field, strings.Join(validFields, ","))
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	return fields, nil
}

func jsonFieldNames[T any]() []string {
	fields := jsonFieldNamesInOrder[T]()
	slices.Sort(fields)
	return fields
}

func jsonFieldNamesInOrder[T any]() []string {
	var zero T
	typeOf := reflect.TypeOf(zero)
	for typeOf.Kind() == reflect.Pointer {
		typeOf = typeOf.Elem()
	}

	fields := make([]string, 0, typeOf.NumField())
	for field := range typeOf.Fields() {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		switch name {
		case "", "-":
			continue
		default:
			fields = append(fields, name)
		}
	}
	return fields
}

func parseOutputFormat(value string) (outputFormat, error) {
	format := outputFormat(strings.ToLower(strings.TrimSpace(value)))
	if format == "" {
		return outputFormatJSON, nil
	}

	switch format {
	case outputFormatJSON, outputFormatCSV, outputFormatTSV:
		return format, nil
	default:
		return "", fmt.Errorf("invalid format %q; valid formats: json,csv,tsv", value)
	}
}

func printDelimited(rows []map[string]json.RawMessage, fields []string, comma rune) error {
	writer := csv.NewWriter(os.Stdout)
	writer.Comma = comma

	if err := writer.Write(fields); err != nil {
		return fmt.Errorf("write delimited header: %w", err)
	}
	for _, row := range rows {
		record, err := delimitedRecord(row, fields)
		if err != nil {
			return err
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write delimited row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush delimited output: %w", err)
	}
	return nil
}

func delimitedRecord(row map[string]json.RawMessage, fields []string) ([]string, error) {
	record := make([]string, 0, len(fields))
	for _, field := range fields {
		value, err := delimitedValue(row[field])
		if err != nil {
			return nil, fmt.Errorf("format field %q: %w", field, err)
		}
		record = append(record, value)
	}
	return record, nil
}

func delimitedValue(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}
	if isJSONNumber(raw) {
		return string(raw), nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}

	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("decode JSON value: %w", err)
	}
	switch typed := value.(type) {
	case bool:
		return strconv.FormatBool(typed), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return "", fmt.Errorf("marshal complex JSON value: %w", err)
		}
		return string(data), nil
	}
}

func isJSONNumber(raw json.RawMessage) bool {
	return raw[0] == '-' || raw[0] >= '0' && raw[0] <= '9'
}

// newCommandClient centralizes authenticated client creation so command
// handlers keep identical error text while avoiding repeated boilerplate.
// During tests a pre-built client can be injected via testClientKey in the
// context, bypassing browser-based authentication entirely.
func newCommandClient(ctx context.Context) (*client.Client, error) {
	if c, ok := ctx.Value(testClientKey).(*client.Client); ok {
		return c, nil
	}
	vlClient, err := client.New(ctx)
	if err != nil {
		if auth.IsSessionExpired(err) {
			var detail interface{ Detail() string }
			if errors.As(err, &detail) {
				slog.Debug("VolumeLeaders session expired", "detail", detail.Detail())
			}
			return nil, fmt.Errorf("%s: %w", auth.SessionExpiredMessage, err)
		}
		slog.Error("failed to create client", "error", err)
		return nil, fmt.Errorf("create client: %w", err)
	}
	return vlClient, nil
}

// newDataTablesRequest builds the common DataTables request shape used by the
// VolumeLeaders endpoints. Callers still provide explicit defaults through opts.
func newDataTablesRequest(columns []string, opts dataTableOptions) datatables.Request {
	return datatables.Request{
		Columns:          columns,
		Start:            opts.start,
		Length:           opts.length,
		OrderColumnIndex: opts.orderCol,
		OrderDirection:   opts.orderDir,
		CustomFilters:    opts.filters,
		Draw:             1,
	}
}

// dateRangeFlags returns --start-date, --end-date, and --days flags for
// commands that require a range. Callers must resolve the range with
// requiredDateRange because --days can satisfy the range without explicit dates.
func dateRangeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "start-date", Usage: "Start date YYYY-MM-DD (required unless --days is set)"},
		&cli.StringFlag{Name: "end-date", Usage: "End date YYYY-MM-DD (required unless --days is set)"},
		&cli.IntFlag{Name: "days", Usage: "Look back this many days from --end-date or today"},
	}
}

// optionalDateRangeFlags returns --start-date, --end-date, and --days flags for
// commands that supply sensible defaults when dates are omitted.
func optionalDateRangeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "start-date", Usage: "Start date YYYY-MM-DD (default: auto)"},
		&cli.StringFlag{Name: "end-date", Usage: "End date YYYY-MM-DD (default: today)"},
		&cli.IntFlag{Name: "days", Usage: "Look back this many days from --end-date or today"},
	}
}

// timeNow is the clock function used by date-range defaults.
// Tests replace it to control the current time.
var timeNow = time.Now

// defaultDates returns start and end dates for a query, applying defaults when
// the user did not explicitly set them. lookbackDays controls how far back the
// start date goes; pass 0 for a today-only default.
func defaultDates(cmd *cli.Command, lookbackDays int) (startDate, endDate string) {
	startDate, endDate, _ = resolveDateRange(cmd, lookbackDays, false)
	return startDate, endDate
}

func requiredDateRange(cmd *cli.Command) (startDate, endDate string, err error) {
	return resolveDateRange(cmd, 0, true)
}

func optionalDateRange(cmd *cli.Command, lookbackDays int) (startDate, endDate string, err error) {
	return resolveDateRange(cmd, lookbackDays, false)
}

func resolveDateRange(cmd *cli.Command, lookbackDays int, required bool) (startDate, endDate string, err error) {
	now := timeNow()
	today := now.Format("2006-01-02")
	days := cmd.Int("days")
	hasDays := cmd.IsSet("days")
	if hasDays && days < 0 {
		return "", "", fmt.Errorf("--days must be greater than or equal to 0")
	}
	if hasDays && cmd.IsSet("start-date") {
		return "", "", fmt.Errorf("--days cannot be used with --start-date")
	}

	endDate = cmd.String("end-date")
	if !cmd.IsSet("end-date") {
		if required && !hasDays {
			return "", "", fmt.Errorf("--start-date and --end-date are required unless --days is set")
		}
		endDate = today
	}

	startDate = cmd.String("start-date")
	if !cmd.IsSet("start-date") {
		switch {
		case hasDays:
			base := now
			if cmd.IsSet("end-date") {
				parsed, parseErr := time.Parse("2006-01-02", endDate)
				if parseErr != nil {
					return "", "", fmt.Errorf("parse --end-date for --days: %w", parseErr)
				}
				base = parsed
			}
			startDate = base.AddDate(0, 0, -days).Format("2006-01-02")
		case required:
			return "", "", fmt.Errorf("--start-date and --end-date are required unless --days is set")
		case lookbackDays > 0:
			startDate = now.AddDate(0, 0, -lookbackDays).Format("2006-01-02")
		default:
			startDate = today
		}
	}

	return startDate, endDate, nil
}

func multiTickerValue(cmd *cli.Command) string {
	values := splitTickerValues(cmd.String("tickers"))
	values = append(values, splitTickerValues(strings.Join(cmd.Args().Slice(), ","))...)
	return strings.Join(dedupeStrings(values), ",")
}

func singleTickerValue(cmd *cli.Command) (string, error) {
	flagValue := strings.TrimSpace(cmd.String("ticker"))
	args := cmd.Args().Slice()
	if len(args) > 1 {
		return "", fmt.Errorf("expected at most one ticker argument, got %d", len(args))
	}
	if flagValue != "" && len(args) == 1 {
		return "", fmt.Errorf("use either --ticker or a ticker argument, not both")
	}
	if flagValue != "" {
		return flagValue, nil
	}
	if len(args) == 1 {
		return strings.TrimSpace(args[0]), nil
	}
	return "", fmt.Errorf("--ticker or a ticker argument is required")
}

func splitTickerValues(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	items := make([]string, 0, strings.Count(value, ",")+1)
	for item := range strings.SplitSeq(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

// volumeRangeFlags returns --min-volume and --max-volume flags with
// standard defaults (0 and 2,000,000,000).
func volumeRangeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{Name: "min-volume", Value: 0, Usage: "Minimum volume"},
		&cli.IntFlag{Name: "max-volume", Value: 2000000000, Usage: "Maximum volume"},
	}
}

// priceRangeFlags returns --min-price and --max-price flags with
// standard defaults (0 and 100,000).
func priceRangeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.FloatFlag{Name: "min-price", Value: 0, Usage: "Minimum price"},
		&cli.FloatFlag{Name: "max-price", Value: 100000, Usage: "Maximum price"},
	}
}

// dollarRangeFlags returns --min-dollars and --max-dollars flags.
// minDefault sets the --min-dollars default; --max-dollars is always 30 billion.
func dollarRangeFlags(minDefault float64) []cli.Flag {
	return []cli.Flag{
		&cli.FloatFlag{Name: "min-dollars", Value: minDefault, Usage: "Minimum dollar value"},
		&cli.FloatFlag{Name: "max-dollars", Value: 30000000000, Usage: "Maximum dollar value"},
	}
}

// paginationFlags returns the standard DataTables pagination and ordering flags.
func paginationFlags(length, orderCol int, orderDir string) []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{Name: "start", Value: 0, Usage: "DataTables start offset"},
		&cli.IntFlag{Name: "length", Value: length, Usage: "Number of results"},
		&cli.IntFlag{Name: "order-col", Value: orderCol, Usage: "Order column index"},
		&cli.StringFlag{Name: "order-dir", Value: orderDir, Usage: "Order direction"},
	}
}

// outputFormatFlags returns the standard tabular output format flag.
func outputFormatFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "format", Value: string(outputFormatJSON), Usage: "Output format: json, csv, or tsv"},
	}
}

func requireStringFlag(flags []cli.Flag, name string) {
	for _, flag := range flags {
		stringFlag, ok := flag.(*cli.StringFlag)
		if ok && stringFlag.Name == name {
			stringFlag.Required = true
			return
		}
	}
}

// printJSON marshals v to JSON and writes it to stdout. When the pretty flag
// is set in the context, the output is indented for human readability.
func printJSON(ctx context.Context, v any) error {
	var (
		output []byte
		err    error
	)
	if pretty, _ := ctx.Value(prettyJSONKey).(bool); pretty {
		output, err = json.MarshalIndent(v, "", "  ")
	} else {
		output, err = json.Marshal(v)
	}
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	_, _ = os.Stdout.Write(output)
	_, _ = os.Stdout.WriteString("\n")
	return nil
}

// intStr converts an int to its decimal string representation.
func intStr(value int) string {
	return strconv.Itoa(value)
}

// formatFloat converts a float64 to a string without trailing zeros.
func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// boolString returns "true" or "false" for the given bool. It is used
// instead of strconv.FormatBool to match VolumeLeaders form field values.
func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// toDateKey normalizes YYYY-MM-DD to YYYYMMDD for the API.
func toDateKey(value string) string {
	return strings.ReplaceAll(value, "-", "")
}

// parseSnapshotString parses the semicolon-delimited "TICKER:PRICE" response
// from GetAllSnapshots into a ticker-to-price map.
func parseSnapshotString(raw string) map[string]float64 {
	result := make(map[string]float64)
	if raw == "" {
		return result
	}

	for pair := range strings.SplitSeq(raw, ";") {
		ticker, priceStr, found := strings.Cut(pair, ":")
		if !found {
			continue
		}
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			continue
		}
		result[ticker] = price
	}
	return result
}
