package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/major/volumeleaders-agent/internal/client"
	"github.com/major/volumeleaders-agent/internal/datatables"
	cli "github.com/urfave/cli/v3"
)

type contextKey int

const prettyJSONKey contextKey = iota

type dataTableOptions struct {
	start, length, orderCol int
	orderDir                string
	filters                 map[string]string
}

// runDataTablesCommand is the shared handler for all DataTables-backed
// commands. It creates the client, builds and posts the request, and
// prints the result as JSON. The label is used in error messages.
func runDataTablesCommand[T any](ctx context.Context, path string, columns []string, opts dataTableOptions, label string) error {
	vlClient, err := newCommandClient(ctx)
	if err != nil {
		return err
	}

	request := newDataTablesRequest(columns, opts)
	var result []T
	if err := vlClient.PostDataTables(ctx, path, request.Encode(), &result); err != nil {
		slog.Error("failed to "+label, "error", err)
		return fmt.Errorf("%s: %w", label, err)
	}
	return printJSON(ctx, result)
}

// newCommandClient centralizes authenticated client creation so command
// handlers keep identical error text while avoiding repeated boilerplate.
func newCommandClient(ctx context.Context) (*client.Client, error) {
	vlClient, err := client.New(ctx)
	if err != nil {
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

// dateRangeFlags returns required --start-date and --end-date flags.
func dateRangeFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "start-date", Required: true, Usage: "Start date YYYY-MM-DD"},
		&cli.StringFlag{Name: "end-date", Required: true, Usage: "End date YYYY-MM-DD"},
	}
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
