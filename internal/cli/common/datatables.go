package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/major/volumeleaders-agent/internal/client"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/spf13/cobra"
)

// RunDataTablesCommand is the shared handler for DataTables-backed commands.
func RunDataTablesCommand[T any](cmd *cobra.Command, path string, columns []string, opts DataTableOptions, formatValue OutputFormat, label string) error {
	ctx, vlClient, format, err := newDataTablesSetup(cmd, formatValue)
	if err != nil {
		return err
	}
	if opts.Length < 0 {
		return RunPaginatedCommand[T](ctx, vlClient, cmd.OutOrStdout(), path, columns, opts, format, label)
	}
	request := NewDataTablesRequest(columns, opts)
	var result []T
	if err := vlClient.PostDataTables(ctx, path, request.Encode(), &result); err != nil {
		slog.Error("failed to "+label, "error", err)
		return fmt.Errorf("%s: %w", label, err)
	}
	return PrintDataTablesResult(cmd.OutOrStdout(), ctx, result, opts.Fields, format)
}

// RunDataTablesSingleRequestCommand sends exactly one DataTables request, even
// when opts.Length is -1.
func RunDataTablesSingleRequestCommand[T any](cmd *cobra.Command, path string, columns []string, opts DataTableOptions, formatValue OutputFormat, label string) error {
	ctx, vlClient, format, err := newDataTablesSetup(cmd, formatValue)
	if err != nil {
		return err
	}
	request := NewDataTablesRequest(columns, opts)
	var result []T
	if err := vlClient.PostDataTables(ctx, path, request.Encode(), &result); err != nil {
		slog.Error("failed to"+label, "error", err)
		return fmt.Errorf("%s: %w", label, err)
	}
	return PrintDataTablesResult(cmd.OutOrStdout(), ctx, result, opts.Fields, format)
}

// newDataTablesSetup extracts the common setup for DataTables commands:
// parsing output format, getting context, and creating an authenticated client.
func newDataTablesSetup(cmd *cobra.Command, formatValue OutputFormat) (context.Context, *client.Client, OutputFormat, error) {
	format, err := ParseOutputFormat(formatValue)
	if err != nil {
		return nil, nil, "", err
	}
	ctx := cmd.Context()
	vlClient, err := NewCommandClient(ctx)
	if err != nil {
		return nil, nil, "", err
	}
	return ctx, vlClient, format, nil
}

// RunPaginatedCommand fetches all records by paginating through the DataTables
// endpoint in pages of PaginationPageSize.
func RunPaginatedCommand[T any](ctx context.Context, vlClient *client.Client, w io.Writer, path string, columns []string, opts DataTableOptions, format OutputFormat, label string) error {
	opts.Length = PaginationPageSize
	all := make([]T, 0)
	for {
		request := NewDataTablesRequest(columns, opts)
		resp, err := vlClient.PostDataTablesPage(ctx, path, request.Encode())
		if err != nil {
			slog.Error("failed to "+label, "error", err)
			return fmt.Errorf("%s: %w", label, err)
		}
		var page []T
		if err := json.Unmarshal(resp.Data, &page); err != nil {
			return fmt.Errorf("unmarshal %s page: %w", label, err)
		}
		if len(page) == 0 {
			break
		}
		all = append(all, page...)
		if resp.RecordsFiltered > 0 && len(all) >= resp.RecordsFiltered {
			break
		}
		if len(page) < PaginationPageSize {
			break
		}
		opts.Start += len(page)
	}
	return PrintDataTablesResult(w, ctx, all, opts.Fields, format)
}

// NewDataTablesRequest builds the common DataTables request shape.
func NewDataTablesRequest(columns []string, opts DataTableOptions) datatables.Request {
	return datatables.Request{Columns: columns, Start: opts.Start, Length: opts.Length, OrderColumnIndex: opts.OrderCol, OrderDirection: string(opts.OrderDir), CustomFilters: opts.Filters, Draw: 1}
}
