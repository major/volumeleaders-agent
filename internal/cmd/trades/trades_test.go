package trades

import (
	"bytes"
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
	"github.com/spf13/cobra"
)

func TestTradesCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertGetTradesRequest(t, r, "2026-04-30", "")
		fmt.Fprint(w, `{"draw":1,"recordsTotal":1492,"recordsFiltered":1492,"data":[{"Ticker":"KRE","Dollars":17501965.25,"RelativeSize":5}]}`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	cmd, err := NewCommand()
	if err != nil {
		t.Fatalf("NewCommand() error = %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--date", "2026-04-30"})

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
	if got.Date != "2026-04-30" {
		t.Fatalf("Date = %q, want 2026-04-30", got.Date)
	}
	if got.RecordsTotal != 1492 || got.RecordsFiltered != 1492 {
		t.Fatalf("record counts = %d/%d, want 1492/1492", got.RecordsTotal, got.RecordsFiltered)
	}
	if len(got.Trades) != 1 {
		t.Fatalf("len(Trades) = %d, want 1", len(got.Trades))
	}
	if !bytes.Contains(got.Trades[0], []byte(`"Ticker": "KRE"`)) {
		t.Fatalf("trade payload = %s, want KRE ticker", string(got.Trades[0]))
	}
}

func TestRankedTradesCommands(t *testing.T) {
	tests := []struct {
		name        string
		newCommand  func() (*cobra.Command, error)
		args        []string
		wantRank    int
		wantLength  int
		wantPreset  string
		wantTickers string
	}{
		{
			name:       "top 10 ranked trades",
			newCommand: NewTop10Command,
			args:       []string{"--date", "2026-04-30"},
			wantRank:   10,
			wantLength: 10,
			wantPreset: "623",
		},
		{
			name:        "top 100 ranked trades with tickers",
			newCommand:  NewTop100Command,
			args:        []string{"--date", "2026-04-30", "--ticker", "aapl,msft"},
			wantRank:    100,
			wantLength:  100,
			wantPreset:  "568",
			wantTickers: "AAPL,MSFT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset := &rankedPreset{rank: tt.wantRank, length: tt.wantLength, presetID: tt.wantPreset}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				options := rankedGetTradesRequestOptions(preset)
				assertGetTradesRequestWithOptions(t, r, "2026-04-30", tt.wantTickers, &options)
				fmt.Fprint(w, `{"draw":1,"recordsTotal":76,"recordsFiltered":76,"data":[{"Ticker":"SNDQ","TradeRank":1}]}`)
			}))
			t.Cleanup(server.Close)

			withCommandDependencies(t, server.Client(), server.URL, nil, nil)

			cmd, err := tt.newCommand()
			if err != nil {
				t.Fatalf("new command error = %v", err)
			}

			var stdout bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			var got RankedResult
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
			}
			if got.Status != "ok" {
				t.Fatalf("Status = %q, want ok", got.Status)
			}
			if got.RankLimit != tt.wantRank {
				t.Fatalf("RankLimit = %d, want %d", got.RankLimit, tt.wantRank)
			}
			if len(got.Trades) != 1 {
				t.Fatalf("len(Trades) = %d, want 1", len(got.Trades))
			}
		})
	}
}

func TestSignalTradesCommands(t *testing.T) {
	tests := []struct {
		name           string
		newCommand     func() (*cobra.Command, error)
		args           []string
		wantPreset     *signalPreset
		wantTickers    string
		wantTradeField string
	}{
		{
			name:       "phantom trades",
			newCommand: NewPhantomCommand,
			args:       []string{"--date", "2026-04-30"},
			wantPreset: &signalPreset{
				phantom:    "1",
				offsetting: "0",
				darkPools:  "1",
				presetID:   "857",
			},
			wantTradeField: "PhantomPrint",
		},
		{
			name:       "offsetting trades with tickers",
			newCommand: NewOffsettingCommand,
			args:       []string{"--date", "2026-04-30", "--ticker", "pltr"},
			wantPreset: &signalPreset{
				phantom:    "0",
				offsetting: "1",
				darkPools:  "-1",
				presetID:   "858",
			},
			wantTickers:    "PLTR",
			wantTradeField: "OffsettingTradeDate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				options := signalGetTradesRequestOptions(tt.wantPreset)
				assertGetTradesRequestWithOptions(t, r, "2026-04-30", tt.wantTickers, &options)
				fmt.Fprintf(w, `{"draw":1,"recordsTotal":2,"recordsFiltered":2,"data":[{"Ticker":"PLTR","%s":1}]}`, tt.wantTradeField)
			}))
			t.Cleanup(server.Close)

			withCommandDependencies(t, server.Client(), server.URL, nil, nil)

			cmd, err := tt.newCommand()
			if err != nil {
				t.Fatalf("new command error = %v", err)
			}

			var stdout bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(tt.args)

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
			if got.Date != "2026-04-30" {
				t.Fatalf("Date = %q, want 2026-04-30", got.Date)
			}
			if got.RecordsTotal != 2 || got.RecordsFiltered != 2 {
				t.Fatalf("record counts = %d/%d, want 2/2", got.RecordsTotal, got.RecordsFiltered)
			}
			if len(got.Trades) != 1 {
				t.Fatalf("len(Trades) = %d, want 1", len(got.Trades))
			}
			if !bytes.Contains(got.Trades[0], []byte(tt.wantTradeField)) {
				t.Fatalf("trade payload = %s, want %s field", string(got.Trades[0]), tt.wantTradeField)
			}
		})
	}
}

func TestLeverageTradesCommands(t *testing.T) {
	tests := []struct {
		name           string
		newCommand     func() (*cobra.Command, error)
		args           []string
		wantPreset     *leveragePreset
		wantTickers    string
		responseSector string
		wantTradeField string
	}{
		{
			name:       "bull leverage trades",
			newCommand: NewBullLeverageCommand,
			args:       []string{"--date", "2026-04-30"},
			wantPreset: &leveragePreset{
				sectorIndustry: "X Bull",
				presetID:       "5",
			},
			responseSector: "3x Bull Nasdaq",
			wantTradeField: "3x Bull Nasdaq",
		},
		{
			name:       "bear leverage trades with tickers",
			newCommand: NewBearLeverageCommand,
			args:       []string{"--date", "2026-04-30", "--ticker", "spxu"},
			wantPreset: &leveragePreset{
				sectorIndustry: "X Bear",
				presetID:       "6",
			},
			wantTickers:    "SPXU",
			responseSector: "3x Bear S&P 500",
			wantTradeField: `3x Bear S\u0026P 500`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				options := leverageGetTradesRequestOptions(tt.wantPreset)
				assertGetTradesRequestWithOptions(t, r, "2026-04-30", tt.wantTickers, &options)
				fmt.Fprintf(w, `{"draw":1,"recordsTotal":8,"recordsFiltered":8,"data":[{"Ticker":"TQQQ","Sector":"%s"}]}`, tt.responseSector)
			}))
			t.Cleanup(server.Close)

			withCommandDependencies(t, server.Client(), server.URL, nil, nil)

			cmd, err := tt.newCommand()
			if err != nil {
				t.Fatalf("new command error = %v", err)
			}

			var stdout bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(tt.args)

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
			if got.Date != "2026-04-30" {
				t.Fatalf("Date = %q, want 2026-04-30", got.Date)
			}
			if got.RecordsTotal != 8 || got.RecordsFiltered != 8 {
				t.Fatalf("record counts = %d/%d, want 8/8", got.RecordsTotal, got.RecordsFiltered)
			}
			if len(got.Trades) != 1 {
				t.Fatalf("len(Trades) = %d, want 1", len(got.Trades))
			}
			if !bytes.Contains(got.Trades[0], []byte(tt.wantTradeField)) {
				t.Fatalf("trade payload = %s, want %s field", string(got.Trades[0]), tt.wantTradeField)
			}
		})
	}
}

func TestTradesCommandTickerFilters(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantTickers string
	}{
		{
			name:        "no ticker",
			args:        []string{"--date", "2026-04-30"},
			wantTickers: "",
		},
		{
			name:        "single ticker",
			args:        []string{"--date", "2026-04-30", "--tickers", "AAPL"},
			wantTickers: "AAPL",
		},
		{
			name:        "multiple tickers",
			args:        []string{"--date", "2026-04-30", "--tickers", "AAPL,IONQ"},
			wantTickers: "AAPL,IONQ",
		},
		{
			name:        "ticker alias accepts comma list",
			args:        []string{"--date", "2026-04-30", "--ticker", "AAPL,IONQ"},
			wantTickers: "AAPL,IONQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assertGetTradesRequest(t, r, "2026-04-30", tt.wantTickers)
				fmt.Fprint(w, `{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":[]}`)
			}))
			t.Cleanup(server.Close)

			withCommandDependencies(t, server.Client(), server.URL, nil, nil)

			cmd, err := NewCommand()
			if err != nil {
				t.Fatalf("NewCommand() error = %v", err)
			}
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
		})
	}
}

func TestTradesCommandValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing date fails",
			args:    []string{},
			wantErr: "date",
		},
		{
			name:    "invalid date fails",
			args:    []string{"--date", "04/30/2026"},
			wantErr: "use YYYY-MM-DD",
		},
		{
			name:    "empty ticker element fails",
			args:    []string{"--date", "2026-04-30", "--tickers", "AAPL,,IONQ"},
			wantErr: "empty ticker",
		},
		{
			name:    "invalid ticker fails",
			args:    []string{"--date", "2026-04-30", "--tickers", "AAPL,$BAD"},
			wantErr: "invalid ticker",
		},
		{
			name:    "ticker spaces fail",
			args:    []string{"--date", "2026-04-30", "--tickers", "AAPL, IONQ, AMZN"},
			wantErr: "without spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd, err := NewCommand()
			if err != nil {
				t.Fatalf("NewCommand() error = %v", err)
			}

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(tt.args)

			err = cmd.Execute()
			if err == nil {
				t.Fatalf("expected error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestFetchDisproportionatelyLargeTradesHandlesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"error":"bad filter"}`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	_, err := fetchDisproportionatelyLargeTrades(t.Context(), "2026-04-30", "")
	if err == nil {
		t.Fatalf("expected API error")
	}
	if !strings.Contains(err.Error(), "bad filter") {
		t.Fatalf("expected bad filter error, got %v", err)
	}
}

func TestFetchDisproportionatelyLargeTradesHandlesAuthStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	_, err := fetchDisproportionatelyLargeTrades(t.Context(), "2026-04-30", "")
	if err == nil {
		t.Fatalf("expected auth error")
	}
	if !strings.Contains(err.Error(), "Authentication required") {
		t.Fatalf("expected authentication remediation, got %v", err)
	}
	if !auth.IsSessionExpired(err) {
		t.Fatalf("expected auth.IsSessionExpired to match %v", err)
	}
}

func TestFetchDisproportionatelyLargeTradesHandlesLoginRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Login" {
			fmt.Fprint(w, `<html>login</html>`)
			return
		}
		http.Redirect(w, r, "/Login", http.StatusFound)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	_, err := fetchDisproportionatelyLargeTrades(t.Context(), "2026-04-30", "")
	if err == nil {
		t.Fatalf("expected auth error")
	}
	if !auth.IsSessionExpired(err) {
		t.Fatalf("expected auth.IsSessionExpired to match %v", err)
	}
}

func TestFetchDisproportionatelyLargeTradesPropagatesCancellation(t *testing.T) {
	canceledCtx, cancel := context.WithCancel(t.Context())
	cancel()

	withCommandDependencies(t, http.DefaultClient, getTradesPath, func(ctx context.Context) (map[string]string, error) {
		return nil, ctx.Err()
	}, nil)

	_, err := fetchDisproportionatelyLargeTrades(canceledCtx, "2026-04-30", "")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestFetchDisproportionatelyLargeTradesDoesNotLeakSecrets(t *testing.T) {
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

	_, err := fetchDisproportionatelyLargeTrades(t.Context(), "2026-04-30", "")
	if err == nil {
		t.Fatalf("expected auth error")
	}
	for _, secret := range []string{secretCookie, secretToken} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("error leaked secret %q: %v", secret, err)
		}
	}
}

func assertGetTradesRequest(t *testing.T, r *http.Request, tradeDate, tickers string) {
	t.Helper()
	options := defaultGetTradesRequestOptions()
	assertGetTradesRequestWithOptions(t, r, tradeDate, tickers, &options)
}

func assertGetTradesRequestWithOptions(t *testing.T, r *http.Request, tradeDate, tickers string, options *getTradesRequestOptions) {
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
	if got := r.Header.Get("Sec-Fetch-Dest"); got != "empty" {
		t.Fatalf("Sec-Fetch-Dest = %q", got)
	}
	if got := r.Header.Get("Sec-Fetch-Mode"); got != "cors" {
		t.Fatalf("Sec-Fetch-Mode = %q", got)
	}
	if got := r.Header.Get("Sec-Fetch-Site"); got != "same-origin" {
		t.Fatalf("Sec-Fetch-Site = %q", got)
	}
	if got := r.Header.Get("Referer"); !strings.Contains(got, "PresetSearchTemplateID="+options.presetSearchTemplateID) || !strings.Contains(got, "StartDate="+url.QueryEscape(tradeDate)) || !strings.Contains(got, "Tickers="+url.QueryEscape(tickers)) {
		t.Fatalf("Referer = %q, want preset and single date", got)
	}
	assertCookie(t, r, "ASP.NET_SessionId", "session-cookie")
	assertCookie(t, r, ".ASPXAUTH", "auth-cookie")
	assertCookie(t, r, "__RequestVerificationToken", "cookie-token")

	if err := r.ParseForm(); err != nil {
		t.Fatalf("ParseForm() error = %v", err)
	}
	assertFormValue(t, r.Form, "StartDate", tradeDate)
	assertFormValue(t, r.Form, "EndDate", tradeDate)
	assertFormValue(t, r.Form, "DarkPools", options.darkPools)
	assertFormValue(t, r.Form, "IncludePhantom", options.includePhantom)
	assertFormValue(t, r.Form, "IncludeOffsetting", options.includeOffsetting)
	assertFormValue(t, r.Form, "VCD", options.vcd)
	assertFormValue(t, r.Form, "SectorIndustry", options.sectorIndustry)
	if got := r.Form.Get("PresetSearchTemplateID"); got != "" {
		t.Fatalf("form[PresetSearchTemplateID] = %q, want empty because preset is carried in Referer", got)
	}
	wantForm := getTradesForm(tradeDate, tickers, options)
	if r.Form.Encode() != wantForm.Encode() {
		t.Fatalf("form mismatch\ngot:  %s\nwant: %s", r.Form.Encode(), wantForm.Encode())
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

func withCommandDependencies(
	t *testing.T,
	client *http.Client,
	endpoint string,
	extract func(context.Context) (map[string]string, error),
	fetch func(context.Context, *http.Client, map[string]string) (string, error),
) {
	t.Helper()

	oldClient := getTradesHTTPClient
	oldEndpoint := getTradesEndpoint
	oldExtract := extractCookies
	oldFetch := fetchXSRFToken
	getTradesHTTPClient = client
	getTradesEndpoint = endpoint
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
		getTradesHTTPClient = oldClient
		getTradesEndpoint = oldEndpoint
		extractCookies = oldExtract
		fetchXSRFToken = oldFetch
	})
}
