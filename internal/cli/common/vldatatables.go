package common

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"

	vlgo "github.com/major/volumeleaders-go/volumeleaders"

	"github.com/spf13/cobra"
)

// VLFetcher calls a typed vlgo endpoint with the given DataTables request and
// filters. Each command package defines a thin wrapper that delegates to the
// appropriate vlgo.Client method (e.g. GetTrades, GetTradeClusters).
type VLFetcher[T any] func(ctx context.Context, c *vlgo.Client, dt *vlgo.DataTablesRequest, filters url.Values) (*vlgo.DataTablesResponse[T], error)

// RunVLDataTablesCommand is the shared handler for DataTables-backed commands
// that have migrated to volumeleaders-go. It mirrors RunDataTablesCommand but
// fetches via typed vlgo methods and maps results to internal model types.
func RunVLDataTablesCommand[Src, Dst any](
	cmd *cobra.Command,
	opts DataTableOptions,
	format OutputFormat,
	label string,
	fetch VLFetcher[Src],
	mapFn func(*Src) Dst,
) error {
	return RunVLDataTablesCommandWithPageSize(cmd, opts, format, PaginationPageSize, label, fetch, mapFn)
}

// RunVLDataTablesCommandWithPageSize handles endpoint-specific page sizes when
// opts.Length is -1.
func RunVLDataTablesCommandWithPageSize[Src, Dst any](
	cmd *cobra.Command,
	opts DataTableOptions,
	format OutputFormat,
	pageSize int,
	label string,
	fetch VLFetcher[Src],
	mapFn func(*Src) Dst,
) error {
	ctx, vlClient, parsedFormat, err := newVLSetup(cmd, format)
	if err != nil {
		return err
	}
	if opts.Length < 0 {
		return runVLPaginated(ctx, vlClient, cmd.OutOrStdout(), opts, parsedFormat, pageSize, label, fetch, mapFn)
	}

	dt := NewVLDataTablesRequest(opts)
	filters := FiltersToValues(opts.Filters)
	resp, err := fetch(ctx, vlClient, &dt, filters)
	if err != nil {
		slog.Error("failed to "+label, "error", err)
		return fmt.Errorf("%s: %w", label, err)
	}
	return PrintDataTablesResult(cmd.OutOrStdout(), ctx, MapSlice(resp.Data, mapFn), opts.Fields, parsedFormat)
}

// RunVLSingleRequestCommand sends exactly one DataTables request via vlgo,
// even when opts.Length is -1.
func RunVLSingleRequestCommand[Src, Dst any](
	cmd *cobra.Command,
	opts DataTableOptions,
	format OutputFormat,
	label string,
	fetch VLFetcher[Src],
	mapFn func(*Src) Dst,
) error {
	ctx, vlClient, parsedFormat, err := newVLSetup(cmd, format)
	if err != nil {
		return err
	}

	dt := NewVLDataTablesRequest(opts)
	filters := FiltersToValues(opts.Filters)
	resp, err := fetch(ctx, vlClient, &dt, filters)
	if err != nil {
		slog.Error("failed to "+label, "error", err)
		return fmt.Errorf("%s: %w", label, err)
	}
	return PrintDataTablesResult(cmd.OutOrStdout(), ctx, MapSlice(resp.Data, mapFn), opts.Fields, parsedFormat)
}

// newVLSetup extracts the common setup for vlgo-backed commands: parsing output
// format, getting context, and creating an authenticated vlgo client.
func newVLSetup(cmd *cobra.Command, formatValue OutputFormat) (context.Context, *vlgo.Client, OutputFormat, error) {
	parsedFormat, err := ParseOutputFormat(formatValue)
	if err != nil {
		return nil, nil, "", err
	}
	ctx := cmd.Context()
	vlClient, err := NewVLClient(ctx)
	if err != nil {
		return nil, nil, "", err
	}
	return ctx, vlClient, parsedFormat, nil
}

// runVLPaginated fetches all records by paginating through a vlgo endpoint.
func runVLPaginated[Src, Dst any](
	ctx context.Context,
	vlClient *vlgo.Client,
	w io.Writer,
	opts DataTableOptions,
	format OutputFormat,
	pageSize int,
	label string,
	fetch VLFetcher[Src],
	mapFn func(*Src) Dst,
) error {
	all, err := FetchVLPages(ctx, vlClient, opts, pageSize, label, fetch)
	if err != nil {
		return err
	}
	return PrintDataTablesResult(w, ctx, MapSlice(all, mapFn), opts.Fields, format)
}

// FetchVLPages pages through a vlgo endpoint until all records are fetched.
//
// Workaround: volumeleaders-go only provides GetTradesLimit for trades today.
// This generic helper adds pagination for all other endpoints until upstream
// adds *Limit helpers. See https://github.com/major/volumeleaders-go/issues/8
func FetchVLPages[T any](
	ctx context.Context,
	vlClient *vlgo.Client,
	opts DataTableOptions,
	pageSize int,
	label string,
	fetch VLFetcher[T],
) ([]T, error) {
	if pageSize == 0 {
		pageSize = PaginationPageSize
	}

	pageState := newDataTablePageState(opts, pageSize)
	filters := FiltersToValues(opts.Filters)
	all := make([]T, 0)

	for {
		dt := NewVLDataTablesRequest(pageState.options())
		resp, err := fetch(ctx, vlClient, &dt, filters)
		if err != nil {
			slog.Error("failed to "+label, "error", err)
			return nil, fmt.Errorf("%s: %w", label, err)
		}
		if len(resp.Data) == 0 {
			break
		}
		all = append(all, resp.Data...)
		if pageState.fetchedAllRecords(len(all), resp.RecordsFiltered) {
			break
		}
		if pageState.isShortPage(len(resp.Data)) {
			break
		}
		pageState.advance(len(resp.Data))
	}
	return all, nil
}

// NewVLDataTablesRequest converts CLI options to a vlgo DataTablesRequest.
func NewVLDataTablesRequest(opts DataTableOptions) vlgo.DataTablesRequest {
	return vlgo.DataTablesRequest{
		Draw:   1,
		Start:  opts.Start,
		Length: opts.Length,
		Order: []vlgo.DataTablesOrder{{
			Column: opts.OrderCol,
			Dir:    string(opts.OrderDir),
		}},
	}
}

// FiltersToValues converts a string map to url.Values for vlgo request filters.
func FiltersToValues(filters map[string]string) url.Values {
	if len(filters) == 0 {
		return nil
	}
	values := make(url.Values, len(filters))
	for k, v := range filters {
		values.Set(k, v)
	}
	return values
}

// MapSlice applies mapFn to each element of src and returns the results.
// The map function receives a pointer to avoid copying large structs.
func MapSlice[Src, Dst any](src []Src, mapFn func(*Src) Dst) []Dst {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Dst, len(src))
	for i := range src {
		dst[i] = mapFn(&src[i])
	}
	return dst
}
