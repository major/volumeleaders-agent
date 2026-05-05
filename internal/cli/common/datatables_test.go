package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/major/volumeleaders-agent/internal/client"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
	"github.com/spf13/cobra"
)

func TestNewDataTablesRequest(t *testing.T) {
	columns := []string{"Col1", "Col2"}
	opts := DataTableOptions{Start: 10, Length: 25, OrderCol: 1, OrderDir: "asc", Filters: map[string]string{"Ticker": "AAPL"}}
	got := NewDataTablesRequest(columns, opts)
	if got.Draw != 1 || got.Start != 10 || got.Length != 25 || got.OrderColumnIndex != 1 || got.OrderDirection != "asc" {
		t.Fatalf("unexpected request: %+v", got)
	}
	if got.CustomFilters["Ticker"] != "AAPL" {
		t.Fatalf("expected Ticker filter, got %+v", got.CustomFilters)
	}
}

func TestNewDataTableOptions(t *testing.T) {
	t.Parallel()

	filters := map[string]string{"Ticker": "AAPL"}
	fields := []string{"Ticker", "Price"}
	got := NewDataTableOptions(DataTableRequestConfig{Start: 5, Length: 25, OrderCol: 2, OrderDir: OrderDirectionDESC, Filters: filters, Fields: fields})

	if got.Start != 5 || got.Length != 25 || got.OrderCol != 2 || got.OrderDir != OrderDirectionDESC {
		t.Fatalf("unexpected options: %+v", got)
	}
	if got.Filters["Ticker"] != "AAPL" || len(got.Fields) != 2 {
		t.Fatalf("expected filters and fields to be preserved, got %+v", got)
	}
}

func TestRunDataTablesCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/Test/GetData" {
			t.Errorf("expected path /Test/GetData, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{"Ticker":"AAPL"}]`))
	}))
	t.Cleanup(server.Close)
	cmd, output := dataTablesCommand(t, server.URL)
	err := RunDataTablesCommand[models.Trade](cmd, "/Test/GetData", datatables.TradeColumns, DataTableOptions{Start: 0, Length: 100, OrderCol: 1, OrderDir: "desc"}, "", "test query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output.String(), "AAPL") {
		t.Fatalf("expected JSON output, got: %s", output.String())
	}
}

func TestRunDataTablesCommandPrettyJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, dataTablesJSON(`[{"Ticker":"AAPL"}]`)) }))
	t.Cleanup(server.Close)
	cmd, output := dataTablesCommand(t, server.URL)
	cmd.SetContext(context.WithValue(cmd.Context(), PrettyJSONKey, true))
	err := RunDataTablesCommand[models.Trade](cmd, "/Test/GetData", datatables.TradeColumns, DataTableOptions{Start: 0, Length: 100, OrderCol: 1, OrderDir: "desc"}, "", "test query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output.String(), "\n  ") {
		t.Fatalf("expected indented output, got: %s", output.String())
	}
}

func TestRunDataTablesCommandServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	cmd, _ := dataTablesCommand(t, server.URL)
	err := RunDataTablesCommand[models.Trade](cmd, "/Test/GetData", datatables.TradeColumns, DataTableOptions{Start: 0, Length: 100, OrderCol: 1, OrderDir: "desc"}, "", "test query")
	assertErrContains(t, err, "test query")
}

func TestRunDataTablesCommandInvalidFormatDoesNotQueryAPI(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { requestCount++; fmt.Fprint(w, dataTablesJSON(`[{}]`)) }))
	t.Cleanup(server.Close)
	cmd, _ := dataTablesCommand(t, server.URL)
	err := RunDataTablesCommand[models.Trade](cmd, "/Test/GetData", datatables.TradeColumns, DataTableOptions{Start: 0, Length: 100, OrderCol: 1, OrderDir: "desc"}, "table", "test query")
	assertErrContains(t, err, "valid formats: json,csv,tsv")
	if requestCount != 0 {
		t.Fatalf("expected invalid format to fail before API query, got %d requests", requestCount)
	}
}

func TestRunDataTablesSingleRequestCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		params, _ := url.ParseQuery(string(body))
		if params.Get("length") != "-1" {
			t.Errorf("expected length=-1, got %q", params.Get("length"))
		}
		fmt.Fprint(w, dataTablesJSON(`[{"Ticker":"AAPL"}]`))
	}))
	t.Cleanup(server.Close)
	cmd, output := dataTablesCommand(t, server.URL)
	err := RunDataTablesSingleRequestCommand[models.Trade](cmd, "/Test/GetData", datatables.TradeColumns, DataTableOptions{Start: 0, Length: -1, OrderCol: 1, OrderDir: "desc"}, "", "single request")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output.String(), "AAPL") {
		t.Fatalf("expected response output, got %s", output.String())
	}
}

func TestPaginatedCommandMultiPage(t *testing.T) {
	totalRecords := PaginationPageSize + 3
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		body, _ := io.ReadAll(r.Body)
		params, _ := url.ParseQuery(string(body))
		if params.Get("start") == "0" {
			items := make([]string, PaginationPageSize)
			for i := range items {
				items[i] = fmt.Sprintf(`{"Ticker":"P1_%d"}`, i)
			}
			fmt.Fprint(w, dataTablesJSONPage("["+strings.Join(items, ",")+"]", totalRecords))
			return
		}
		fmt.Fprint(w, dataTablesJSONPage(`[{"Ticker":"P2_0"},{"Ticker":"P2_1"},{"Ticker":"P2_2"}]`, totalRecords))
	}))
	t.Cleanup(server.Close)
	cmd, output := dataTablesCommand(t, server.URL)
	err := RunDataTablesCommand[models.Trade](cmd, "/Test/GetData", datatables.TradeColumns, DataTableOptions{Start: 0, Length: -1, OrderCol: 1, OrderDir: "desc"}, "", "test paginated multi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("expected 2 requests, got %d", requestCount)
	}
	if !strings.Contains(output.String(), "P1_0") || !strings.Contains(output.String(), "P2_2") {
		t.Fatalf("expected both pages in output, got: %s", output.String())
	}
}

func TestPaginatedCommandEmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, dataTablesJSONPage(`[]`, 0)) }))
	t.Cleanup(server.Close)
	cmd, output := dataTablesCommand(t, server.URL)
	err := RunDataTablesCommand[models.Trade](cmd, "/Test/GetData", datatables.TradeColumns, DataTableOptions{Start: 0, Length: -1, OrderCol: 1, OrderDir: "desc"}, "", "test paginated empty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(output.String()) != "[]" {
		t.Fatalf("expected empty array [], got: %s", strings.TrimSpace(output.String()))
	}
}

func dataTablesCommand(t *testing.T, baseURL string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	vlClient := client.NewForTesting(&http.Client{Timeout: 5 * time.Second}, baseURL)
	cmd := &cobra.Command{Use: "test"}
	cmd.SetContext(context.WithValue(t.Context(), TestClientKey, vlClient))
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	return cmd, output
}

func dataTablesJSON(data string) string {
	return `{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":` + data + `}`
}

func dataTablesJSONPage(data string, recordsFiltered int) string {
	return fmt.Sprintf(`{"draw":1,"recordsTotal":%d,"recordsFiltered":%d,"data":%s}`, recordsFiltered, recordsFiltered, data)
}
