package watchlists

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/major/volumeleaders-agent/internal/auth"
)

func TestWatchListsCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertGetWatchListsRequest(t, r)
		fmt.Fprint(w, `{"data":[{"SearchTemplateKey":4952,"Name":"BigOnes","Tickers":"","MinVolume":0,"MaxVolume":2000000000,"MinDollars":10000000.00,"MaxDollars":300000000000.00,"MinPrice":5.00,"MaxPrice":100000.00,"MinRelativeSize":5,"MaxTradeRank":10,"Conditions":"IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH","NormalPrints":true,"DarkPools":true,"Sweeps":true,"AHTrades":false,"OffsettingTrades":true}]}`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	cmd, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand() error = %v", err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if got.Status != "ok" {
		t.Fatalf("Status = %q, want ok", got.Status)
	}
	if got.Count != 1 {
		t.Fatalf("Count = %d, want 1", got.Count)
	}
	if strings.Join(got.Fields, ",") != strings.Join(watchListFieldPresets["summary"], ",") {
		t.Fatalf("Fields = %v, want summary fields", got.Fields)
	}
	if len(got.Rows) != 1 {
		t.Fatalf("len(Rows) = %d, want 1", len(got.Rows))
	}
	if len(got.WatchLists) != 0 {
		t.Fatalf("len(WatchLists) = %d, want 0 for default array shape", len(got.WatchLists))
	}
	if string(got.Rows[0][1]) != `"BigOnes"` {
		t.Fatalf("watchlist name = %s, want BigOnes", string(got.Rows[0][1]))
	}
	if len(got.Rows[0]) != 2 {
		t.Fatalf("summary row width = %d, want 2", len(got.Rows[0]))
	}
	if strings.Contains(stdout.String(), "\n  ") {
		t.Fatalf("default output should be compact JSON, got %q", stdout.String())
	}
}

func TestWatchListsCommandExpandedPresetIncludesConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"data":[{"SearchTemplateKey":4952,"Name":"BigOnes","Tickers":"","MinVolume":0,"MaxVolume":2000000000,"MinDollars":10000000.00,"MaxDollars":300000000000.00,"MinPrice":5.00,"MaxPrice":100000.00,"MinRelativeSize":5,"MaxTradeRank":10,"Conditions":"IgnoreOBD,IgnoreOBH,IgnoreOSD,IgnoreOSH","NormalPrints":true,"DarkPools":true,"Sweeps":true,"AHTrades":false,"OffsettingTrades":true}]}`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	cmd, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand() error = %v", err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--preset-fields", "expanded"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if strings.Join(got.Fields, ",") != strings.Join(watchListFieldPresets["expanded"], ",") {
		t.Fatalf("Fields = %v, want expanded fields", got.Fields)
	}
	tradeTypesIndex := fieldIndex(t, got.Fields, includedTradeTypesName)
	if string(got.Rows[0][tradeTypesIndex]) != `["NormalPrints","DarkPools","Sweeps","OffsettingTrades"]` {
		t.Fatalf("included trade types = %s", string(got.Rows[0][tradeTypesIndex]))
	}
}

func TestWatchListsCommandFiltersBySearchTemplateKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"data":[{"SearchTemplateKey":4952,"Name":"BigOnes"},{"SearchTemplateKey":6277,"Name":"Testing Testing"}]}`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	cmd, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand() error = %v", err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--search-template-key", "6277"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if got.Count != 1 {
		t.Fatalf("Count = %d, want 1", got.Count)
	}
	if string(got.Rows[0][0]) != "6277" || string(got.Rows[0][1]) != `"Testing Testing"` {
		t.Fatalf("filtered row = %s, want Testing Testing", got.Rows[0])
	}
}

func TestWatchListsCommandExpandedObjectsPretty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"data":[{"SearchTemplateKey":7,"SearchTemplateTypeKey":0,"Name":"DP Sweep over 18M","Tickers":"AAPL,MSFT","SortOrder":null,"MaxTradeRank":-1,"MinVCD":0,"DarkPools":true,"Sweeps":true,"Blocks":false,"APIKey":null}]}`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	cmd, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand() error = %v", err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--preset-fields", "expanded", "--shape", "objects", "--pretty"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "\n  ") {
		t.Fatalf("pretty output should be indented, got %q", stdout.String())
	}

	var got Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if len(got.Rows) != 0 {
		t.Fatalf("len(Rows) = %d, want 0 for object shape", len(got.Rows))
	}
	if len(got.WatchLists) != 1 {
		t.Fatalf("len(WatchLists) = %d, want 1", len(got.WatchLists))
	}
	var watchList map[string]json.RawMessage
	if err := json.Unmarshal(got.WatchLists[0], &watchList); err != nil {
		t.Fatalf("unmarshal watchlist: %v", err)
	}
	var includedTradeTypes []string
	if err := json.Unmarshal(watchList[includedTradeTypesName], &includedTradeTypes); err != nil {
		t.Fatalf("unmarshal included trade types: %v", err)
	}
	if strings.Join(includedTradeTypes, ",") != "DarkPools,Sweeps" {
		t.Fatalf("included trade types = %v", includedTradeTypes)
	}
	for _, internalField := range []string{"SearchTemplateTypeKey", "SortOrder", "MinVCD", "APIKey"} {
		if _, ok := watchList[internalField]; ok {
			t.Fatalf("expanded output included internal field %q: %s", internalField, got.WatchLists[0])
		}
	}
}

func TestWatchListsCommandFullPreset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"data":[{"SearchTemplateKey":7,"Name":"Raw","APIKey":null}]}`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	cmd, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand() error = %v", err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--preset-fields", "full"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if len(got.Fields) != 0 || len(got.Rows) != 0 {
		t.Fatalf("full preset should not project fields, got fields=%v rows=%v", got.Fields, got.Rows)
	}
	if len(got.WatchLists) != 1 || !strings.Contains(string(got.WatchLists[0]), "APIKey") {
		t.Fatalf("full preset did not return raw watchlist: %s", got.WatchLists)
	}
}

func TestSaveWatchListCommandCreatesNewWatchList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertSaveWatchListRequest(t, r, 0)
		fmt.Fprint(w, `<html><body>saved</body></html>`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL+"/WatchListConfig", nil, nil)

	cmd, err := NewSaveCommand()
	if err != nil {
		t.Fatalf("NewSaveCommand() error = %v", err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"--name", "Testing Testing",
		"--tickers", "AAPL,MSFT",
		"--min-dollars", "10000000",
		"--min-relative-size", "5",
		"--max-trade-rank", "10",
		"--dark-pools=false",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got SaveResult
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if got.Status != "ok" || got.Action != "created" {
		t.Fatalf("status/action = %q/%q, want ok/created", got.Status, got.Action)
	}
	if got.SearchTemplateKey != 0 {
		t.Fatalf("SearchTemplateKey = %d, want 0", got.SearchTemplateKey)
	}
	if got.Name != "Testing Testing" {
		t.Fatalf("Name = %q, want Testing Testing", got.Name)
	}
}

func TestSaveWatchListCommandUpdatesExistingWatchList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertSaveWatchListRequest(t, r, 4952)
		fmt.Fprint(w, `<html><body>saved</body></html>`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL+"/WatchListConfig", nil, nil)

	cmd, err := NewSaveCommand()
	if err != nil {
		t.Fatalf("NewSaveCommand() error = %v", err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--search-template-key", "4952", "--name", "BigOnes"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got SaveResult
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if got.Action != "updated" {
		t.Fatalf("Action = %q, want updated", got.Action)
	}
	if got.SearchTemplateKey != 4952 {
		t.Fatalf("SearchTemplateKey = %d, want 4952", got.SearchTemplateKey)
	}
}

func TestSaveWatchListCommandRejectsNegativeSearchTemplateKey(t *testing.T) {
	withCommandDependencies(t, http.DefaultClient, saveWatchListPath, nil, nil)

	cmd, err := NewSaveCommand()
	if err != nil {
		t.Fatalf("NewSaveCommand() error = %v", err)
	}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--search-template-key", "-1", "--name", "BigOnes"})

	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected negative search-template-key error")
	}
	if !strings.Contains(err.Error(), "search-template-key must be 0") {
		t.Fatalf("error = %v, want search-template-key validation", err)
	}
}

func TestDeleteWatchListCommandDeletesWatchList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertDeleteWatchListRequest(t, r, 4952)
		fmt.Fprint(w, `{"status":"ok"}`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL+"/WatchListConfigs/DeleteWatchList", nil, nil)

	cmd, err := NewDeleteCommand()
	if err != nil {
		t.Fatalf("NewDeleteCommand() error = %v", err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--search-template-key", "4952"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got DeleteResult
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if got.Status != "ok" || got.Action != "deleted" {
		t.Fatalf("status/action = %q/%q, want ok/deleted", got.Status, got.Action)
	}
	if got.SearchTemplateKey != 4952 {
		t.Fatalf("SearchTemplateKey = %d, want 4952", got.SearchTemplateKey)
	}
}

func TestDeleteWatchListCommandRejectsInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertDeleteWatchListRequest(t, r, 4952)
		fmt.Fprint(w, `<html><body>not json</body></html>`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL+"/WatchListConfigs/DeleteWatchList", nil, nil)

	cmd, err := NewDeleteCommand()
	if err != nil {
		t.Fatalf("NewDeleteCommand() error = %v", err)
	}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--search-template-key", "4952"})

	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected invalid JSON response error")
	}
	if !strings.Contains(err.Error(), "decode DeleteWatchList response JSON") {
		t.Fatalf("error = %v, want decode DeleteWatchList response JSON", err)
	}
}

func TestDeleteWatchListCommandRejectsApplicationErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertDeleteWatchListRequest(t, r, 4952)
		fmt.Fprint(w, `{"success":false,"message":"cannot delete"}`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL+"/WatchListConfigs/DeleteWatchList", nil, nil)

	cmd, err := NewDeleteCommand()
	if err != nil {
		t.Fatalf("NewDeleteCommand() error = %v", err)
	}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--search-template-key", "4952"})

	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected application error response")
	}
	if !strings.Contains(err.Error(), "cannot delete") {
		t.Fatalf("error = %v, want application error message", err)
	}
}

func TestDeleteWatchListCommandRequiresSearchTemplateKey(t *testing.T) {
	withCommandDependencies(t, http.DefaultClient, deleteWatchListPath, nil, nil)

	cmd, err := NewDeleteCommand()
	if err != nil {
		t.Fatalf("NewDeleteCommand() error = %v", err)
	}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected missing search-template-key error")
	}
	if !strings.Contains(err.Error(), "search-template-key") {
		t.Fatalf("error = %v, want search-template-key", err)
	}
}

func TestWatchListsCommandRejectsInvalidOutputOptions(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "invalid preset", args: []string{"--preset-fields", "tiny"}, wantErr: "invalid preset-fields"},
		{name: "invalid shape", args: []string{"--shape", "yaml"}, wantErr: "invalid shape"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withCommandDependencies(t, http.DefaultClient, getWatchListsPath, nil, nil)

			cmd, err := NewCommand()
			if err != nil {
				t.Fatalf("NewCommand() error = %v", err)
			}
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(tt.args)

			err = cmd.Execute()
			if err == nil {
				t.Fatalf("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestFetchWatchListsHandlesGzip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		_, _ = gz.Write([]byte(`{"data":[{"SearchTemplateKey":42,"Name":"Gzip"}]}`))
		if err := gz.Close(); err != nil {
			t.Fatalf("close gzip writer: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	got, err := fetchWatchLists(t.Context())
	if err != nil {
		t.Fatalf("fetchWatchLists() error = %v", err)
	}
	if len(got.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(got.Data))
	}
}

func TestFetchWatchListsPropagatesCancellation(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(t.Context())
	cancel()

	withCommandDependencies(t, http.DefaultClient, getWatchListsPath, func(ctx context.Context) (map[string]string, error) {
		return nil, ctx.Err()
	}, nil)

	_, err := fetchWatchLists(canceledCtx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestFetchWatchListsDoesNotLeakSecrets(t *testing.T) {
	secretCookie := "secret-session-cookie"
	secretToken := "secret-xsrf-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, func(context.Context) (map[string]string, error) {
		return map[string]string{
			"ASP.NET_SessionId":          secretCookie,
			".ASPXAUTH":                  "secret-auth-cookie",
			"__RequestVerificationToken": "secret-cookie-token",
		}, nil
	}, func(context.Context, *http.Client, map[string]string) (string, error) {
		return secretToken, nil
	})

	_, err := fetchWatchLists(t.Context())
	if err == nil {
		t.Fatalf("expected auth error")
	}
	for _, secret := range []string{secretCookie, secretToken} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("error leaked secret %q: %v", secret, err)
		}
	}
}

func TestDeleteWatchListDoesNotLeakSecrets(t *testing.T) {
	secretCookie := "secret-session-cookie"
	secretToken := "secret-xsrf-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, func(context.Context) (map[string]string, error) {
		return map[string]string{
			"ASP.NET_SessionId":          secretCookie,
			".ASPXAUTH":                  "secret-auth-cookie",
			"__RequestVerificationToken": "secret-cookie-token",
		}, nil
	}, func(context.Context, *http.Client, map[string]string) (string, error) {
		return secretToken, nil
	})

	err := deleteWatchList(t.Context(), 4952)
	if err == nil {
		t.Fatalf("expected auth error")
	}
	for _, secret := range []string{secretCookie, secretToken} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("error leaked secret %q: %v", secret, err)
		}
	}
}

func assertGetWatchListsRequest(t *testing.T, r *http.Request) {
	t.Helper()

	if r.Method != http.MethodPost {
		t.Fatalf("method = %s, want POST", r.Method)
	}
	if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded; charset=UTF-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := r.Header.Get("Accept"); got != "application/json, text/javascript, */*; q=0.01" {
		t.Fatalf("Accept = %q", got)
	}
	if got := r.Header.Get("Accept-Encoding"); got != "gzip" {
		t.Fatalf("Accept-Encoding = %q", got)
	}
	if got := r.Header.Get("User-Agent"); got != auth.UserAgent {
		t.Fatalf("User-Agent = %q", got)
	}
	if got := r.Header.Get("X-XSRF-Token"); got != "xsrf-token" {
		t.Fatalf("X-XSRF-Token = %q", got)
	}
	if got := r.Header.Get("X-Requested-With"); got != "XMLHttpRequest" {
		t.Fatalf("X-Requested-With = %q", got)
	}
	if got := r.Header.Get("Origin"); got != "https://www.volumeleaders.com" {
		t.Fatalf("Origin = %q", got)
	}
	if got := r.Header.Get("Referer"); got != watchListsPage {
		t.Fatalf("Referer = %q", got)
	}
	assertCookie(t, r, "ASP.NET_SessionId", "session-cookie")
	assertCookie(t, r, ".ASPXAUTH", "auth-cookie")
	assertCookie(t, r, "__RequestVerificationToken", "cookie-token")

	if err := r.ParseForm(); err != nil {
		t.Fatalf("ParseForm() error = %v", err)
	}
	assertFormValue(t, r.Form, "draw", "1")
	assertFormValue(t, r.Form, "columns[0][data]", "Name")
	assertFormValue(t, r.Form, "columns[1][data]", "Tickers")
	assertFormValue(t, r.Form, "columns[2][data]", "Criteria")
	assertFormValue(t, r.Form, "order[0][column]", "0")
	assertFormValue(t, r.Form, "order[0][dir]", "asc")
	assertFormValue(t, r.Form, "start", "0")
	assertFormValue(t, r.Form, "length", "-1")
	assertFormValue(t, r.Form, "search[value]", "")
	assertFormValue(t, r.Form, "search[regex]", "false")
	if r.Form.Encode() != getWatchListsForm().Encode() {
		t.Fatalf("form mismatch\ngot:  %s\nwant: %s", r.Form.Encode(), getWatchListsForm().Encode())
	}
}

func assertSaveWatchListRequest(t *testing.T, r *http.Request, searchTemplateKey int) {
	t.Helper()

	if r.Method != http.MethodPost {
		t.Fatalf("method = %s, want POST", r.Method)
	}
	if r.URL.Path != "/WatchListConfig" {
		t.Fatalf("path = %s, want /WatchListConfig", r.URL.Path)
	}
	if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := r.Header.Get("Accept"); got != "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8" {
		t.Fatalf("Accept = %q", got)
	}
	if got := r.Header.Get("Accept-Encoding"); got != "gzip" {
		t.Fatalf("Accept-Encoding = %q", got)
	}
	if got := r.Header.Get("User-Agent"); got != auth.UserAgent {
		t.Fatalf("User-Agent = %q", got)
	}
	if got := r.Header.Get("X-XSRF-Token"); got != "xsrf-token" {
		t.Fatalf("X-XSRF-Token = %q", got)
	}
	if got := r.Header.Get("Origin"); got != "https://www.volumeleaders.com" {
		t.Fatalf("Origin = %q", got)
	}
	wantReferer := watchListConfigPage
	if searchTemplateKey != 0 {
		wantReferer = fmt.Sprintf("%s?SearchTemplateKey=%d", watchListConfigPage, searchTemplateKey)
	}
	if got := r.Header.Get("Referer"); got != wantReferer {
		t.Fatalf("Referer = %q, want %q", got, wantReferer)
	}
	assertCookie(t, r, "ASP.NET_SessionId", "session-cookie")
	assertCookie(t, r, ".ASPXAUTH", "auth-cookie")
	assertCookie(t, r, "__RequestVerificationToken", "cookie-token")

	if err := r.ParseForm(); err != nil {
		t.Fatalf("ParseForm() error = %v", err)
	}
	assertFormValue(t, r.Form, "__RequestVerificationToken", "xsrf-token")
	assertFormValue(t, r.Form, "SearchTemplateKey", fmt.Sprintf("%d", searchTemplateKey))
	if searchTemplateKey == 0 {
		assertFormValue(t, r.Form, "Name", "Testing Testing")
		assertFormValue(t, r.Form, "Tickers", "AAPL,MSFT")
		assertFormValue(t, r.Form, "MinDollars", "10000000")
		assertFormValue(t, r.Form, "MinRelativeSizeSelected", "5")
		assertFormValue(t, r.Form, "MaxTradeRankSelected", "10")
		assertFormValues(t, r.Form, "DarkPoolsSelected", []string{"false"})
	} else {
		assertFormValue(t, r.Form, "Name", "BigOnes")
		assertFormValue(t, r.Form, "MaxTradeRankSelected", "-1")
		assertFormValues(t, r.Form, "DarkPoolsSelected", []string{"true", "false"})
	}
	assertFormValue(t, r.Form, "SecurityTypeKey", "-1")
	assertFormValue(t, r.Form, "RSIOverboughtDailySelected", "-1")
	assertFormValue(t, r.Form, "RSIOverboughtHourlySelected", "-1")
	assertFormValue(t, r.Form, "RSIOversoldDailySelected", "-1")
	assertFormValue(t, r.Form, "RSIOversoldHourlySelected", "-1")
}

func assertDeleteWatchListRequest(t *testing.T, r *http.Request, searchTemplateKey int) {
	t.Helper()

	if r.Method != http.MethodPost {
		t.Fatalf("method = %s, want POST", r.Method)
	}
	if r.URL.Path != "/WatchListConfigs/DeleteWatchList" {
		t.Fatalf("path = %s, want /WatchListConfigs/DeleteWatchList", r.URL.Path)
	}
	if got := r.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := r.Header.Get("Accept"); got != "application/json, text/javascript, */*; q=0.01" {
		t.Fatalf("Accept = %q", got)
	}
	if got := r.Header.Get("Accept-Encoding"); got != "gzip" {
		t.Fatalf("Accept-Encoding = %q", got)
	}
	if got := r.Header.Get("User-Agent"); got != auth.UserAgent {
		t.Fatalf("User-Agent = %q", got)
	}
	if got := r.Header.Get("X-XSRF-Token"); got != "xsrf-token" {
		t.Fatalf("X-XSRF-Token = %q", got)
	}
	if got := r.Header.Get("X-Requested-With"); got != "XMLHttpRequest" {
		t.Fatalf("X-Requested-With = %q", got)
	}
	if got := r.Header.Get("Origin"); got != "https://www.volumeleaders.com" {
		t.Fatalf("Origin = %q", got)
	}
	if got := r.Header.Get("Referer"); got != watchListsPage {
		t.Fatalf("Referer = %q", got)
	}
	assertCookie(t, r, "ASP.NET_SessionId", "session-cookie")
	assertCookie(t, r, ".ASPXAUTH", "auth-cookie")
	assertCookie(t, r, "__RequestVerificationToken", "cookie-token")

	var body map[string]int
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if got := body["WatchListKey"]; got != searchTemplateKey {
		t.Fatalf("WatchListKey = %d, want %d", got, searchTemplateKey)
	}
}

func assertCookie(t *testing.T, r *http.Request, name, want string) {
	t.Helper()

	cookie, err := r.Cookie(name)
	if err != nil {
		t.Fatalf("missing cookie %s: %v", name, err)
	}
	if cookie.Value != want {
		t.Fatalf("cookie %s = %q, want %q", name, cookie.Value, want)
	}
}

func assertFormValue(t *testing.T, form url.Values, name, want string) {
	t.Helper()

	if got := form.Get(name); got != want {
		t.Fatalf("form[%s] = %q, want %q", name, got, want)
	}
}

func assertFormValues(t *testing.T, form url.Values, name string, want []string) {
	t.Helper()

	got := form[name]
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("form[%s] = %v, want %v", name, got, want)
	}
}

func fieldIndex(t *testing.T, fields []string, want string) int {
	t.Helper()

	for i, field := range fields {
		if field == want {
			return i
		}
	}
	t.Fatalf("field %q not found in %v", want, fields)
	return -1
}

func withCommandDependencies(
	t *testing.T,
	client *http.Client,
	endpoint string,
	extract func(context.Context) (map[string]string, error),
	fetch func(context.Context, *http.Client, map[string]string) (string, error),
) {
	t.Helper()

	oldClient := getWatchListsHTTPClient
	oldEndpoint := getWatchListsEndpoint
	oldSaveEndpoint := saveWatchListEndpoint
	oldDeleteEndpoint := deleteWatchListEndpoint
	oldExtract := extractCookies
	oldFetch := fetchXSRFToken
	getWatchListsHTTPClient = client
	getWatchListsEndpoint = endpoint
	saveWatchListEndpoint = endpoint
	deleteWatchListEndpoint = endpoint
	if extract == nil {
		extract = func(context.Context) (map[string]string, error) {
			return map[string]string{
				"ASP.NET_SessionId":          "session-cookie",
				".ASPXAUTH":                  "auth-cookie",
				"__RequestVerificationToken": "cookie-token",
			}, nil
		}
	}
	if fetch == nil {
		fetch = func(context.Context, *http.Client, map[string]string) (string, error) {
			return "xsrf-token", nil
		}
	}
	extractCookies = extract
	fetchXSRFToken = fetch
	t.Cleanup(func() {
		getWatchListsHTTPClient = oldClient
		getWatchListsEndpoint = oldEndpoint
		saveWatchListEndpoint = oldSaveEndpoint
		deleteWatchListEndpoint = oldDeleteEndpoint
		extractCookies = oldExtract
		fetchXSRFToken = oldFetch
	})
}
