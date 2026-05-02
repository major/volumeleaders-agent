package alert

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/testutil"
	"github.com/major/volumeleaders-agent/internal/models"
)

const alertConfigFixture = `[{
	"AlertConfigKey": 1,
	"UserKey": 99,
	"Name": "Big sweeps",
	"Tickers": "AAPL,MSFT",
	"TradeRankLTE": 5,
	"TradeVCDGTE": null,
	"TradeMultGTE": null,
	"TradeVolumeGTE": 1000000,
	"TradeDollarsGTE": null,
	"TradeConditions": "OBH",
	"TradeClusterRankLTE": null,
	"TradeClusterVCDGTE": null,
	"TradeClusterMultGTE": null,
	"TradeClusterVolumeGTE": null,
	"TradeClusterDollarsGTE": null,
	"TotalRankLTE": null,
	"TotalVolumeGTE": null,
	"TotalDollarsGTE": null,
	"AHRankLTE": null,
	"AHVolumeGTE": null,
	"AHDollarsGTE": null,
	"ClosingTradeRankLTE": null,
	"ClosingTradeVCDGTE": null,
	"ClosingTradeMultGTE": null,
	"ClosingTradeVolumeGTE": null,
	"ClosingTradeDollarsGTE": null,
	"ClosingTradeConditions": "OSH",
	"OffsettingPrint": false,
	"PhantomPrint": false,
	"Sweep": true,
	"DarkPool": true
}]`

func TestAlertConfigs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfigs/GetAlertConfigs" {
			t.Errorf("expected path /AlertConfigs/GetAlertConfigs, got %s", r.URL.Path)
		}
		fmt.Fprint(w, testutil.DataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewAlertCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "configs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout == "" {
		t.Error("expected non-empty stdout")
	}
}

func TestAlertConfigsDefaultFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, testutil.DataTablesJSON(alertConfigFixture))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewAlertCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "configs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows := decodeAlertConfigRows(t, stdout)
	assertAlertConfigFields(t, rows[0], alertConfigDefaultFields)
	if rows[0]["AlertConfigKey"] != float64(1) {
		t.Fatalf("expected AlertConfigKey 1, got %#v", rows[0]["AlertConfigKey"])
	}
}

func TestAlertConfigsAllFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, testutil.DataTablesJSON(alertConfigFixture))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewAlertCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "configs", "--fields", "all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows := decodeAlertConfigRows(t, stdout)
	allFields := common.JSONFieldNamesInOrder[models.AlertConfig]()
	assertAlertConfigFields(t, rows[0], allFields)
	if rows[0]["UserKey"] != float64(99) {
		t.Fatalf("expected UserKey 99 in full output, got %#v", rows[0]["UserKey"])
	}
}

func TestAlertConfigsExplicitFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, testutil.DataTablesJSON(alertConfigFixture))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewAlertCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "configs", "--fields", "AlertConfigKey,Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rows := decodeAlertConfigRows(t, stdout)
	assertAlertConfigFields(t, rows[0], []string{"AlertConfigKey", "Name"})
}

func TestAlertConfigsInvalidFields(t *testing.T) {
	t.Parallel()

	ctx := testutil.ContextWithTestClient(t, "http://example.invalid")
	cmd := NewAlertCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "configs", "--fields", "BadField")
	testutil.AssertErrContains(t, err, "invalid field")
}

func TestAlertConfigsCSVDefaultFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, testutil.DataTablesJSON(alertConfigFixture))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewAlertCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "configs", "--format", "csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantHeader := strings.Join(alertConfigDefaultFields, ",") + "\n"
	if !strings.HasPrefix(stdout, wantHeader) {
		t.Fatalf("expected CSV header %q, got %q", wantHeader, stdout)
	}
}

func TestAlertConfigsServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewAlertCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "configs")
	testutil.AssertErrContains(t, err, "query alert configs")
}

func TestAlertDelete(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfigs/DeleteAlertConfig" {
			t.Errorf("expected path /AlertConfigs/DeleteAlertConfig, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewAlertCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "delete", "--key", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, `"ok"`) {
		t.Errorf("expected output to contain ok, got: %s", stdout)
	}
}

func TestAlertDeleteServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewAlertCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "delete", "--key", "42")
	testutil.AssertErrContains(t, err, "delete alert config")
}

func TestAlertDeleteMissingKey(t *testing.T) {
	t.Parallel()

	ctx := testutil.ContextWithTestClient(t, "http://unused")
	cmd := NewAlertCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "delete")
	if err == nil {
		t.Fatal("expected error for missing --key flag, got nil")
	}
}

func TestAlertCreate(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfig" {
			t.Errorf("expected path /AlertConfig, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewAlertCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "create", "--name", "Test Alert")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, `"created"`) {
		t.Errorf("expected output to contain created, got: %s", stdout)
	}
}

func TestAlertCreateMissingName(t *testing.T) {
	t.Parallel()

	ctx := testutil.ContextWithTestClient(t, "http://unused")
	cmd := NewAlertCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "create")
	if err == nil {
		t.Fatal("expected error for missing --name flag, got nil")
	}
}

func TestAlertCreateWithTickers(t *testing.T) {
	t.Parallel()

	// Verifies that buildAlertConfigFields auto-selects SelectedTickers when
	// tickers are specified but ticker-group is left at the default.
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		capturedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewAlertCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "create", "--name", "Ticker Alert", "--tickers", "AAPL,MSFT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(capturedBody), "SelectedTickers") {
		t.Errorf("expected request body to reference SelectedTickers, got: %s", capturedBody)
	}
}

func TestBuildAlertConfigFieldsAutoSelectsTickerGroup(t *testing.T) {
	t.Parallel()

	fields := buildAlertConfigFields(&alertConfigFlags{
		TickerGroup: alertTickerGroupAll,
		Tickers:     "AAPL,MSFT",
	}, 0)

	if got := fields["TickerGroup"]; got != string(alertTickerGroupSelected) {
		t.Errorf("TickerGroup = %q, want %q", got, alertTickerGroupSelected)
	}
}

func TestAlertEdit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfig" {
			t.Errorf("expected path /AlertConfig, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewAlertCommand()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "edit", "--key", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, `"updated"`) {
		t.Errorf("expected output to contain updated, got: %s", stdout)
	}
}

func TestAlertEditMissingKey(t *testing.T) {
	t.Parallel()

	ctx := testutil.ContextWithTestClient(t, "http://unused")
	cmd := NewAlertCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "edit")
	if err == nil {
		t.Fatal("expected error for missing --key flag, got nil")
	}
}

func TestAlertCreateEditServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)

	t.Run("create", func(t *testing.T) {
		t.Parallel()
		cmd := NewAlertCommand()
		_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "create", "--name", "Test")
		testutil.AssertErrContains(t, err, "save alert config")
	})

	t.Run("edit", func(t *testing.T) {
		t.Parallel()
		cmd := NewAlertCommand()
		_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "edit", "--key", "42")
		testutil.AssertErrContains(t, err, "save alert config")
	})
}

// --- test helpers ---

// decodeAlertConfigRows unmarshals JSON array of alert config objects from output.
func decodeAlertConfigRows(t *testing.T, output string) []map[string]any {
	t.Helper()

	var rows []map[string]any
	if err := json.Unmarshal([]byte(output), &rows); err != nil {
		t.Fatalf("decode alert config output: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 alert config row, got %d", len(rows))
	}
	return rows
}

// assertAlertConfigFields verifies that the row contains exactly the expected fields.
func assertAlertConfigFields(t *testing.T, row map[string]any, fields []string) {
	t.Helper()

	if len(row) != len(fields) {
		t.Fatalf("expected %d fields, got %d: %#v", len(fields), len(row), row)
	}
	for _, field := range fields {
		if _, ok := row[field]; !ok {
			t.Fatalf("expected field %q in %#v", field, row)
		}
	}
}
