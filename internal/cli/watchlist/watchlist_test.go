package watchlist

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/major/volumeleaders-agent/internal/cli/testutil"
)

func TestConfigs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchListConfigs/GetWatchLists" {
			t.Errorf("expected path /WatchListConfigs/GetWatchLists, got %s", r.URL.Path)
		}
		fmt.Fprint(w, testutil.DataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "configs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout == "" {
		t.Error("expected non-empty stdout")
	}
}

func TestConfigsServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "configs")
	testutil.AssertErrContains(t, err, "query watchlist configs")
}

func TestTickers(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchLists/GetWatchListTickers" {
			t.Errorf("expected path /WatchLists/GetWatchListTickers, got %s", r.URL.Path)
		}
		fmt.Fprint(w, testutil.DataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "tickers", "--watchlist-key", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout == "" {
		t.Error("expected non-empty stdout")
	}
}

func TestTickersDefaultKey(t *testing.T) {
	t.Parallel()

	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// DataTables body is form-encoded; extract the WatchListKey filter.
		for _, pair := range strings.Split(string(body), "&") {
			// The filter is encoded as a custom parameter.
			if strings.Contains(pair, "WatchListKey") {
				parts := strings.SplitN(pair, "=", 2)
				if len(parts) == 2 {
					gotKey = parts[1]
				}
			}
		}
		fmt.Fprint(w, testutil.DataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "tickers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotKey != "-1" {
		t.Errorf("watchlist-key default = %q, want %q", gotKey, "-1")
	}
}

func TestTickersServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "tickers")
	testutil.AssertErrContains(t, err, "query watchlist tickers")
}

func TestDelete(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchListConfigs/DeleteWatchList" {
			t.Errorf("expected path /WatchListConfigs/DeleteWatchList, got %s", r.URL.Path)
		}
		// Verify the request body is JSON with WatchListKey.
		body, _ := io.ReadAll(r.Body)
		var payload map[string]int
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("expected JSON body, got: %s", body)
		}
		if payload["WatchListKey"] != 1 {
			t.Errorf("WatchListKey = %d, want 1", payload["WatchListKey"])
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "delete", "--key", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, `"ok"`) {
		t.Errorf("expected output to contain ok, got: %s", stdout)
	}
}

func TestDeleteMissingKey(t *testing.T) {
	t.Parallel()

	ctx := testutil.ContextWithTestClient(t, "http://unused")
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "delete")
	if err == nil {
		t.Fatal("expected error for missing --key flag, got nil")
	}
}

func TestDeleteServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "delete", "--key", "1")
	testutil.AssertErrContains(t, err, "delete watchlist")
}

func TestAddTicker(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Chart0/UpdateWatchList" {
			t.Errorf("expected path /Chart0/UpdateWatchList, got %s", r.URL.Path)
		}
		// Verify form-encoded body.
		if err := r.ParseForm(); err != nil {
			t.Errorf("failed to parse form: %v", err)
		}
		if got := r.PostFormValue("WatchListKey"); got != "1" {
			t.Errorf("WatchListKey = %q, want %q", got, "1")
		}
		if got := r.PostFormValue("Ticker"); got != "NVDA" {
			t.Errorf("Ticker = %q, want %q", got, "NVDA")
		}
		fmt.Fprint(w, `{"ok":true}`)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "add-ticker", "--watchlist-key", "1", "--ticker", "NVDA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, `"ok"`) {
		t.Errorf("expected output to contain ok, got: %s", stdout)
	}
}

func TestAddTickerMissingFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"missing both", []string{"add-ticker"}},
		{"missing ticker", []string{"add-ticker", "--watchlist-key", "1"}},
		{"missing watchlist-key", []string{"add-ticker", "--ticker", "NVDA"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := testutil.ContextWithTestClient(t, "http://unused")
			cmd := NewCmd()
			_, _, err := testutil.ExecuteCommand(t, cmd, ctx, tt.args...)
			if err == nil {
				t.Fatal("expected error for missing required flags, got nil")
			}
		})
	}
}

func TestAddTickerServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "add-ticker", "--watchlist-key", "1", "--ticker", "INVALID")
	testutil.AssertErrContains(t, err, "add ticker to watchlist")
}

func TestCreate(t *testing.T) {
	t.Parallel()

	var gotFields map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchListConfig" {
			t.Errorf("expected path /WatchListConfig, got %s", r.URL.Path)
		}
		gotFields = parseMultipartFields(t, r)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "create", "--name", "Test List")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, `"created"`) {
		t.Errorf("expected output to contain created, got: %s", stdout)
	}

	// Verify key fields in multipart body.
	if gotFields["SearchTemplateKey"] != "0" {
		t.Errorf("SearchTemplateKey = %q, want %q", gotFields["SearchTemplateKey"], "0")
	}
	if gotFields["Name"] != "Test List" {
		t.Errorf("Name = %q, want %q", gotFields["Name"], "Test List")
	}
}

func TestCreateMissingName(t *testing.T) {
	t.Parallel()

	ctx := testutil.ContextWithTestClient(t, "http://unused")
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "create")
	if err == nil {
		t.Fatal("expected error for missing --name flag, got nil")
	}
}

func TestCreateServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "create", "--name", "Test")
	testutil.AssertErrContains(t, err, "save watchlist config")
}

func TestEdit(t *testing.T) {
	t.Parallel()

	var gotFields map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchListConfig" {
			t.Errorf("expected path /WatchListConfig, got %s", r.URL.Path)
		}
		gotFields = parseMultipartFields(t, r)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, "edit", "--key", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, `"updated"`) {
		t.Errorf("expected output to contain updated, got: %s", stdout)
	}

	// Verify key field in multipart body.
	if gotFields["SearchTemplateKey"] != "1" {
		t.Errorf("SearchTemplateKey = %q, want %q", gotFields["SearchTemplateKey"], "1")
	}
}

func TestEditMissingKey(t *testing.T) {
	t.Parallel()

	ctx := testutil.ContextWithTestClient(t, "http://unused")
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "edit")
	if err == nil {
		t.Fatal("expected error for missing --key flag, got nil")
	}
}

func TestEditServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "edit", "--key", "1")
	testutil.AssertErrContains(t, err, "save watchlist config")
}

func TestCreateDefaultBoolFlags(t *testing.T) {
	t.Parallel()

	var gotFields map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFields = parseMultipartFields(t, r)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	// Only pass --name (required); all booleans should default to true.
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "create", "--name", "Defaults Test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	boolFields := []struct {
		field string
		want  string
	}{
		{"NormalPrintsSelected", "true"},
		{"SignaturePrintsSelected", "true"},
		{"LatePrintsSelected", "true"},
		{"TimelyPrintsSelected", "true"},
		{"DarkPoolsSelected", "true"},
		{"LitExchangesSelected", "true"},
		{"SweepsSelected", "true"},
		{"BlocksSelected", "true"},
		{"PremarketTradesSelected", "true"},
		{"RTHTradesSelected", "true"},
		{"AHTradesSelected", "true"},
		{"OpeningTradesSelected", "true"},
		{"ClosingTradesSelected", "true"},
		{"PhantomTradesSelected", "true"},
		{"OffsettingTradesSelected", "true"},
	}
	for _, bf := range boolFields {
		if gotFields[bf.field] != bf.want {
			t.Errorf("%s = %q, want %q", bf.field, gotFields[bf.field], bf.want)
		}
	}
}

func TestCreateDefaultIntFlags(t *testing.T) {
	t.Parallel()

	var gotFields map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFields = parseMultipartFields(t, r)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "create", "--name", "Defaults Test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	intFields := []struct {
		field string
		want  string
	}{
		{"SecurityTypeKey", "-1"},
		{"MaxTradeRankSelected", "-1"},
		{"RSIOverboughtDailySelected", "-1"},
		{"RSIOverboughtHourlySelected", "-1"},
		{"RSIOversoldDailySelected", "-1"},
		{"RSIOversoldHourlySelected", "-1"},
		{"MaxVolume", "2000000000"},
		{"MaxDollars", "30000000000"},
		{"MaxPrice", "100000"},
	}
	for _, f := range intFields {
		if gotFields[f.field] != f.want {
			t.Errorf("%s = %q, want %q", f.field, gotFields[f.field], f.want)
		}
	}
}

func TestCreateCustomFlags(t *testing.T) {
	t.Parallel()

	var gotFields map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFields = parseMultipartFields(t, r)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewCmd()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "create",
		"--name", "Custom",
		"--tickers", "AAPL,MSFT",
		"--dark-pools=false",
		"--sweeps=false",
		"--security-type", "1",
		"--min-price", "10.5",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []struct {
		field string
		want  string
	}{
		{"Name", "Custom"},
		{"Tickers", "AAPL,MSFT"},
		{"DarkPoolsSelected", "false"},
		{"SweepsSelected", "false"},
		{"SecurityTypeKey", "1"},
		{"MinPrice", "10.5"},
		// Unchanged bools should still be true.
		{"NormalPrintsSelected", "true"},
		{"BlocksSelected", "true"},
	}
	for _, c := range checks {
		if gotFields[c.field] != c.want {
			t.Errorf("%s = %q, want %q", c.field, gotFields[c.field], c.want)
		}
	}
}

func TestBuildWatchlistConfigFields(t *testing.T) {
	t.Parallel()

	// Verify all expected form field names are present.
	cmd := newCreateCmd()
	cmd.SetArgs([]string{"--name", "Test"})
	// Parse flags so they have values.
	if err := cmd.ParseFlags([]string{"--name", "Test"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	fields := buildWatchlistConfigFields(cmd, 42)

	expectedKeys := []string{
		"SearchTemplateKey", "Name", "Tickers",
		"MinVolume", "MaxVolume", "MinDollars", "MaxDollars",
		"MinPrice", "MaxPrice", "MinVCD",
		"SectorIndustry", "SecurityTypeKey",
		"MinRelativeSizeSelected", "MaxTradeRankSelected",
		"NormalPrintsSelected", "SignaturePrintsSelected",
		"LatePrintsSelected", "TimelyPrintsSelected",
		"DarkPoolsSelected", "LitExchangesSelected",
		"SweepsSelected", "BlocksSelected",
		"PremarketTradesSelected", "RTHTradesSelected",
		"AHTradesSelected", "OpeningTradesSelected",
		"ClosingTradesSelected", "PhantomTradesSelected",
		"OffsettingTradesSelected",
		"RSIOverboughtDailySelected", "RSIOverboughtHourlySelected",
		"RSIOversoldDailySelected", "RSIOversoldHourlySelected",
	}

	for _, key := range expectedKeys {
		if _, ok := fields[key]; !ok {
			t.Errorf("expected form field %q to be present", key)
		}
	}

	if fields["SearchTemplateKey"] != "42" {
		t.Errorf("SearchTemplateKey = %q, want %q", fields["SearchTemplateKey"], "42")
	}
	if fields["Name"] != "Test" {
		t.Errorf("Name = %q, want %q", fields["Name"], "Test")
	}
}

// parseMultipartFields reads all multipart form fields from a request.
func parseMultipartFields(t *testing.T, r *http.Request) map[string]string {
	t.Helper()

	contentType := r.Header.Get("Content-Type")
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("parse content type: %v", err)
	}

	reader := multipart.NewReader(r.Body, params["boundary"])
	fields := make(map[string]string)
	for {
		part, err := reader.NextPart()
		if err != nil {
			break
		}
		value, _ := io.ReadAll(part)
		fields[part.FormName()] = string(value)
		part.Close()
	}
	return fields
}
