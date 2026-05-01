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
	if strings.Join(got.Fields, ",") != strings.Join(watchListFieldPresets["core"], ",") {
		t.Fatalf("Fields = %v, want core fields", got.Fields)
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
	tradeTypesIndex := fieldIndex(t, got.Fields, includedTradeTypesName)
	if string(got.Rows[0][tradeTypesIndex]) != `["NormalPrints","DarkPools","Sweeps","OffsettingTrades"]` {
		t.Fatalf("included trade types = %s", string(got.Rows[0][tradeTypesIndex]))
	}
	if strings.Contains(stdout.String(), "\n  ") {
		t.Fatalf("default output should be compact JSON, got %q", stdout.String())
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
	oldExtract := extractCookies
	oldFetch := fetchXSRFToken
	getWatchListsHTTPClient = client
	getWatchListsEndpoint = endpoint
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
		extractCookies = oldExtract
		fetchXSRFToken = oldFetch
	})
}
