package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
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

func TestRunAlertConfigs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfigs/GetAlertConfigs" {
			t.Errorf("expected path /AlertConfigs/GetAlertConfigs, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		if err := runAlertConfigs(ctx, "", "json"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunAlertConfigsDefaultFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSON(alertConfigFixture))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		if err := runAlertConfigs(ctx, "", "json"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	rows := decodeAlertConfigRows(t, output)
	assertAlertConfigFields(t, rows[0], alertConfigDefaultFields)
	if rows[0]["AlertConfigKey"] != float64(1) {
		t.Fatalf("expected AlertConfigKey 1, got %#v", rows[0]["AlertConfigKey"])
	}
}

func TestRunAlertConfigsAllFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSON(alertConfigFixture))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		if err := runAlertConfigs(ctx, "all", "json"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	rows := decodeAlertConfigRows(t, output)
	allFields := jsonFieldNamesInOrder[models.AlertConfig]()
	assertAlertConfigFields(t, rows[0], allFields)
	if rows[0]["UserKey"] != float64(99) {
		t.Fatalf("expected UserKey 99 in full output, got %#v", rows[0]["UserKey"])
	}
}

func TestRunAlertConfigsExplicitFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSON(alertConfigFixture))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		if err := runAlertConfigs(ctx, "AlertConfigKey,Name", "json"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	rows := decodeAlertConfigRows(t, output)
	assertAlertConfigFields(t, rows[0], []string{"AlertConfigKey", "Name"})
}

func TestRunAlertConfigsInvalidFields(t *testing.T) {
	ctx := contextWithTestClient(t, "http://example.invalid")
	err := runAlertConfigs(ctx, "BadField", "json")
	assertErrContains(t, err, "invalid field")
}

func TestRunAlertConfigsCSVDefaultFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSON(alertConfigFixture))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		if err := runAlertConfigs(ctx, "", "csv"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	wantHeader := strings.Join(alertConfigDefaultFields, ",") + "\n"
	if !strings.HasPrefix(output, wantHeader) {
		t.Fatalf("expected CSV header %q, got %q", wantHeader, output)
	}
}

func TestRunAlertConfigsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	err := runAlertConfigs(ctx, "", "json")
	assertErrContains(t, err, "query alert configs")
}

func TestRunAlertDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfigs/DeleteAlertConfig" {
			t.Errorf("expected path /AlertConfigs/DeleteAlertConfig, got %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		if err := runAlertDelete(ctx, 42); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

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

func TestRunAlertDeleteServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	err := runAlertDelete(ctx, 42)
	assertErrContains(t, err, "delete alert config")
}

func TestRunAlertCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfig" {
			t.Errorf("expected path /AlertConfig, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		if err := root.Run(ctx, []string{"app", "alert", "create", "--name", "Test Alert"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, `"created"`) {
		t.Errorf("expected output to contain created, got: %s", output)
	}
}

func TestRunAlertEdit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/AlertConfig" {
			t.Errorf("expected path /AlertConfig, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	output := captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		if err := root.Run(ctx, []string{"app", "alert", "edit", "--key", "42"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, `"updated"`) {
		t.Errorf("expected output to contain updated, got: %s", output)
	}
}

func TestRunAlertCreateWithTickers(t *testing.T) {
	// Verifies that buildAlertConfigFields auto-selects SelectedTickers when
	// tickers are specified but ticker-group is left at the default.
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		if err := root.Run(ctx, []string{"app", "alert", "create", "--name", "Ticker Alert", "--tickers", "AAPL,MSFT"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(string(capturedBody), "SelectedTickers") {
		t.Errorf("expected request body to reference SelectedTickers, got: %s", capturedBody)
	}
}

func TestAlertConfigsCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		if err := root.Run(ctx, []string{"app", "alert", "configs"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAlertDeleteCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)
	captureStdout(t, func() {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		if err := root.Run(ctx, []string{"app", "alert", "delete", "--key", "42"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunAlertCreateEditServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(t, server.URL)

	t.Run("create", func(t *testing.T) {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		err := root.Run(ctx, []string{"app", "alert", "create", "--name", "Test"})
		assertErrContains(t, err, "save alert config")
	})

	t.Run("edit", func(t *testing.T) {
		root := &cli.Command{Commands: []*cli.Command{NewAlertCommand()}}
		err := root.Run(ctx, []string{"app", "alert", "edit", "--key", "42"})
		assertErrContains(t, err, "save alert config")
	})
}
