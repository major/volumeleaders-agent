package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	ctx = addPrettyJSON(ctx)
	output := captureStdout(t, func() {
		err := runDataTablesCommand[models.Trade](ctx, "/Test/GetData",
			datatables.TradeColumns,
			dataTableOptions{start: 0, length: 100, orderCol: 1, orderDir: "desc"},
			"test query")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "\n") {
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
		"test query")
	assertErrContains(t, err, "test query")
}
