package commands

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

func TestNewCommandClientWithTestClient(t *testing.T) {
	t.Parallel()
	ctx := contextWithTestClient("http://example.test")
	c, err := newCommandClient(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
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
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		err := runDataTablesCommand[models.Trade](ctx, "/Test/GetData",
			datatables.TradeColumns,
			dataTableOptions{start: 0, length: 100, orderCol: 1, orderDir: "desc"},
			"",
			"test query")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "[") {
		t.Errorf("expected JSON array output, got: %s", output)
	}
}

func TestRunDataTablesCommandPrettyJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSON(`[{"A":1},{"B":2}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	ctx = addPrettyJSON(ctx)
	output := captureStdout(t, func() {
		err := runDataTablesCommand[models.Trade](ctx, "/Test/GetData",
			datatables.TradeColumns,
			dataTableOptions{start: 0, length: 100, orderCol: 1, orderDir: "desc"},
			"",
			"test query")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// MarshalIndent uses two-space indentation; compact output never contains "  ".
	if !strings.Contains(output, "\n  ") {
		t.Errorf("expected indented output, got: %s", output)
	}
}

func TestRunDataTablesCommandServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runDataTablesCommand[models.Trade](ctx, "/Test/GetData",
		datatables.TradeColumns,
		dataTableOptions{start: 0, length: 100, orderCol: 1, orderDir: "desc"},
		"",
		"test query")
	assertErrContains(t, err, "test query")
}

func TestRunDataTablesCommandInvalidFormatDoesNotQueryAPI(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runDataTablesCommand[models.Trade](ctx, "/Test/GetData",
		datatables.TradeColumns,
		dataTableOptions{start: 0, length: 100, orderCol: 1, orderDir: "desc"},
		"table",
		"test query")
	assertErrContains(t, err, "valid formats: json,csv,tsv")
	if requestCount != 0 {
		t.Errorf("expected invalid format to fail before API query, got %d requests", requestCount)
	}
}

func TestPaginatedCommandSinglePage(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		fmt.Fprint(w, dataTablesJSONPage(`[{"Ticker":"AAPL"},{"Ticker":"MSFT"}]`, 2))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		err := runDataTablesCommand[models.Trade](ctx, "/Test/GetData",
			datatables.TradeColumns,
			dataTableOptions{start: 0, length: -1, orderCol: 1, orderDir: "desc"},
			"",
			"test paginated")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if requestCount != 1 {
		t.Errorf("expected 1 request, got %d", requestCount)
	}
	if !strings.Contains(output, "AAPL") || !strings.Contains(output, "MSFT") {
		t.Errorf("expected both tickers in output, got: %s", output)
	}
}

func TestPaginatedCommandMultiPage(t *testing.T) {
	totalRecords := paginationPageSize + 3
	var requestCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		body, _ := io.ReadAll(r.Body)
		params, _ := url.ParseQuery(string(body))
		start := params.Get("start")

		if start == "0" {
			// First page: return exactly paginationPageSize items.
			items := make([]string, paginationPageSize)
			for i := range items {
				items[i] = fmt.Sprintf(`{"Ticker":"P1_%d"}`, i)
			}
			fmt.Fprint(w, dataTablesJSONPage("["+strings.Join(items, ",")+"]", totalRecords))
		} else {
			// Second page: return remaining items.
			fmt.Fprint(w, dataTablesJSONPage(
				`[{"Ticker":"P2_0"},{"Ticker":"P2_1"},{"Ticker":"P2_2"}]`, totalRecords))
		}
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		err := runDataTablesCommand[models.Trade](ctx, "/Test/GetData",
			datatables.TradeColumns,
			dataTableOptions{start: 0, length: -1, orderCol: 1, orderDir: "desc"},
			"",
			"test paginated multi")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if requestCount != 2 {
		t.Errorf("expected 2 requests, got %d", requestCount)
	}
	// Verify first page items are present.
	if !strings.Contains(output, "P1_0") {
		t.Errorf("expected first page items in output, got: %s", output[:min(200, len(output))])
	}
	// Verify second page items are present.
	if !strings.Contains(output, "P2_2") {
		t.Errorf("expected second page items in output, got: %s", output[max(0, len(output)-200):])
	}
}

func TestPaginatedCommandEmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSONPage(`[]`, 0))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	output := captureStdout(t, func() {
		err := runDataTablesCommand[models.Trade](ctx, "/Test/GetData",
			datatables.TradeColumns,
			dataTableOptions{start: 0, length: -1, orderCol: 1, orderDir: "desc"},
			"",
			"test paginated empty")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Empty pagination should output an empty JSON array, not null.
	trimmed := strings.TrimSpace(output)
	if trimmed != "[]" {
		t.Errorf("expected empty array [], got: %s", trimmed)
	}
}

func TestPaginatedCommandServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	err := runDataTablesCommand[models.Trade](ctx, "/Test/GetData",
		datatables.TradeColumns,
		dataTableOptions{start: 0, length: -1, orderCol: 1, orderDir: "desc"},
		"",
		"test paginated error")
	assertErrContains(t, err, "test paginated error")
}
