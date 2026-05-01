package trades

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
	if len(got.Fields) != len(tradeFieldPresets["core"]) {
		t.Fatalf("len(Fields) = %d, want core fields", len(got.Fields))
	}
	if len(got.Rows) != 1 {
		t.Fatalf("len(Rows) = %d, want 1", len(got.Rows))
	}
	if len(got.Trades) != 0 {
		t.Fatalf("len(Trades) = %d, want 0 for default array shape", len(got.Trades))
	}
	if string(got.Rows[0][0]) != `"KRE"` {
		t.Fatalf("first row ticker = %s, want KRE", string(got.Rows[0][0]))
	}
	if strings.Contains(stdout.String(), "\n  ") {
		t.Fatalf("default output should be compact JSON, got %q", stdout.String())
	}
}

func TestTradeClustersCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		options := defaultGetTradeClustersRequestOptions()
		options.length = 2
		assertGetTradeClustersRequestWithOptions(t, r, "2026-04-30", "AAPL,MSFT", &options)
		fmt.Fprint(w, `{"draw":1,"data":[{"Ticker":"AAPL","MinFullTimeString24":"10:01:04","MaxFullTimeString24":"10:01:08","Dollars":1250000,"TradeCount":7,"TradeClusterRank":14,"Sector":"Technology","TotalRows":3213},{"Ticker":"MSFT","MinFullTimeString24":"10:02:00","Dollars":1800000,"TradeCount":5,"TradeClusterRank":18,"Sector":"Technology","TotalRows":3213}]}`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	cmd, err := NewTradeClustersCommand()
	if err != nil {
		t.Fatalf("NewTradeClustersCommand() error = %v", err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--date", "2026-04-30", "--tickers", "aapl,msft", "--limit", "2"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got ClusterResult
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if got.Status != "ok" {
		t.Fatalf("Status = %q, want ok", got.Status)
	}
	if got.Date != "2026-04-30" {
		t.Fatalf("Date = %q, want 2026-04-30", got.Date)
	}
	if got.RecordsTotal != 3213 || got.RecordsFiltered != 3213 {
		t.Fatalf("record counts = %d/%d, want 3213/3213", got.RecordsTotal, got.RecordsFiltered)
	}
	if strings.Join(got.Fields, ",") != strings.Join(clusterFieldPresets["core"], ",") {
		t.Fatalf("Fields = %v, want core cluster fields", got.Fields)
	}
	if len(got.Rows) != 2 {
		t.Fatalf("len(Rows) = %d, want 2", len(got.Rows))
	}
	if len(got.Clusters) != 0 {
		t.Fatalf("len(Clusters) = %d, want 0 for default array shape", len(got.Clusters))
	}
	if string(got.Rows[0][0]) != `"AAPL"` {
		t.Fatalf("first row ticker = %s, want AAPL", string(got.Rows[0][0]))
	}
}

func TestTradeClustersCommandObjectShapeAndCustomFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		options := defaultGetTradeClustersRequestOptions()
		options.length = 1
		assertGetTradeClustersRequestWithOptions(t, r, "2026-04-30", "", &options)
		fmt.Fprint(w, `{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":[{"Ticker":"IONQ","Dollars":2500000,"TradeCount":12,"Ignored":true}]}`)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	cmd, err := NewTradeClustersCommand()
	if err != nil {
		t.Fatalf("NewTradeClustersCommand() error = %v", err)
	}

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--date", "2026-04-30", "--limit", "1", "--fields", "Ticker,Dollars,TradeCount", "--shape", "objects"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got ClusterResult
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if strings.Join(got.Fields, ",") != "Ticker,Dollars,TradeCount" {
		t.Fatalf("Fields = %v, want custom cluster fields", got.Fields)
	}
	if len(got.Clusters) != 1 {
		t.Fatalf("len(Clusters) = %d, want 1", len(got.Clusters))
	}
	if !bytes.Contains(got.Clusters[0], []byte(`"TradeCount":12`)) || bytes.Contains(got.Clusters[0], []byte("Ignored")) {
		t.Fatalf("projected cluster payload = %s", string(got.Clusters[0]))
	}
}

func TestRankedTradeClustersCommands(t *testing.T) {
	tests := []struct {
		name        string
		newCommand  func() (*cobra.Command, error)
		args        []string
		wantPreset  *clusterPreset
		wantLength  int
		wantTickers string
	}{
		{
			name:       "top 10 ranked clusters",
			newCommand: NewTop10ClustersCommand,
			args:       []string{"--date", "2026-04-30"},
			wantPreset: &clusterPreset{
				tradeClusterRank: 10,
				length:           10,
				minVolume:        "10000",
				maxDollars:       "100000000000",
				presetID:         "623",
			},
			wantLength: 10,
		},
		{
			name:       "top 100 ranked clusters with tickers",
			newCommand: NewTop100ClustersCommand,
			args:       []string{"--date", "2026-04-30", "--ticker", "aapl,msft", "--limit", "25"},
			wantPreset: &clusterPreset{
				tradeClusterRank: 100,
				length:           100,
				minVolume:        "10000",
				maxDollars:       "100000000000",
				presetID:         "568",
			},
			wantLength:  25,
			wantTickers: "AAPL,MSFT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				options := clusterPresetRequestOptions(tt.wantPreset)
				options.length = tt.wantLength
				assertGetTradeClustersRequestWithOptions(t, r, "2026-04-30", tt.wantTickers, &options)
				fmt.Fprint(w, `{"draw":1,"recordsTotal":76,"recordsFiltered":76,"data":[{"Ticker":"SNDQ","TradeClusterRank":1,"TradeCount":8,"TotalRows":76}]}`)
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

			var got ClusterResult
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
			}
			if got.Status != "ok" {
				t.Fatalf("Status = %q, want ok", got.Status)
			}
			if got.RankLimit != tt.wantPreset.tradeClusterRank {
				t.Fatalf("RankLimit = %d, want %d", got.RankLimit, tt.wantPreset.tradeClusterRank)
			}
			if len(got.Fields) != len(clusterFieldPresets["core"]) {
				t.Fatalf("len(Fields) = %d, want core cluster fields", len(got.Fields))
			}
			if len(got.Rows) != 1 {
				t.Fatalf("len(Rows) = %d, want 1", len(got.Rows))
			}
		})
	}
}

func TestTradesCommandOutputOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		options := defaultGetTradesRequestOptions()
		options.length = 2
		assertGetTradesRequestWithOptions(t, r, "2026-04-30", "AAPL", &options)
		fmt.Fprint(w, `{"draw":1,"recordsTotal":7,"recordsFiltered":7,"data":[{"Ticker":"AAPL"},{"Ticker":"AAPL"}]}`)
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
	cmd.SetArgs([]string{"--date", "2026-04-30", "--tickers", "AAPL", "--limit", "2", "--pretty"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "\n  ") {
		t.Fatalf("pretty output should be indented JSON, got %q", stdout.String())
	}

	var got Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if len(got.Rows) != 2 {
		t.Fatalf("len(Rows) = %d, want 2", len(got.Rows))
	}
}

func TestTradesCommandObjectShapeAndCustomFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertGetTradesRequest(t, r, "2026-04-30", "")
		fmt.Fprint(w, `{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":[{"Ticker":"KRE","Dollars":17501965.25,"Ignored":true}]}`)
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
	cmd.SetArgs([]string{"--date", "2026-04-30", "--fields", "Ticker,Dollars", "--shape", "objects"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if strings.Join(got.Fields, ",") != "Ticker,Dollars" {
		t.Fatalf("Fields = %v, want Ticker,Dollars", got.Fields)
	}
	if len(got.Trades) != 1 {
		t.Fatalf("len(Trades) = %d, want 1", len(got.Trades))
	}
	if !bytes.Contains(got.Trades[0], []byte(`"Ticker":"KRE"`)) || bytes.Contains(got.Trades[0], []byte("Ignored")) {
		t.Fatalf("projected trade payload = %s", string(got.Trades[0]))
	}
}

func TestTradesCommandFullPresetKeepsRawTrades(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertGetTradesRequest(t, r, "2026-04-30", "")
		fmt.Fprint(w, `{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":[{"Ticker":"KRE","Ignored":true}]}`)
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
	cmd.SetArgs([]string{"--date", "2026-04-30", "--preset-fields", "full"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if len(got.Fields) != 0 || len(got.Rows) != 0 {
		t.Fatalf("full preset should omit Fields/Rows, got fields=%v rows=%v", got.Fields, got.Rows)
	}
	if len(got.Trades) != 1 || !bytes.Contains(got.Trades[0], []byte("Ignored")) {
		t.Fatalf("full preset trade payload = %v", got.Trades)
	}
}

func TestTradesCommandSignalsPresetIncludesSignalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertGetTradesRequest(t, r, "2026-04-30", "PLTR")
		fmt.Fprint(w, `{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":[{"Ticker":"PLTR","PhantomPrint":true,"InsideBar":false,"Industry":"Software"}]}`)
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
	cmd.SetArgs([]string{"--date", "2026-04-30", "--ticker", "PLTR", "--preset-fields", "signals"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var got Result
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
	}
	if strings.Join(got.Fields, ",") != strings.Join(tradeFieldPresets["signals"], ",") {
		t.Fatalf("Fields = %v, want signals preset", got.Fields)
	}
	phantomIndex := indexOf(got.Fields, "PhantomPrint")
	if phantomIndex < 0 {
		t.Fatalf("signals preset omitted PhantomPrint: %v", got.Fields)
	}
	if string(got.Rows[0][phantomIndex]) != "true" {
		t.Fatalf("PhantomPrint cell = %s, want true", string(got.Rows[0][phantomIndex]))
	}
}

func TestTradesCommandRejectsLimitAboveMaximum(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("request should not be sent for oversized limit: %s", r.URL.String())
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
	cmd.SetArgs([]string{"--date", "2026-04-30", "--limit", "250"})

	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "100 or less") {
		t.Fatalf("Execute() error = %v, want limit cap error", err)
	}
}

func TestRankedTradesCommands(t *testing.T) {
	tests := []struct {
		name        string
		newCommand  func() (*cobra.Command, error)
		args        []string
		wantRank    int
		presetLen   int
		wantLength  int
		wantPreset  string
		wantTickers string
	}{
		{
			name:       "top 10 ranked trades",
			newCommand: NewTop10Command,
			args:       []string{"--date", "2026-04-30"},
			wantRank:   10,
			presetLen:  10,
			wantLength: 10,
			wantPreset: "623",
		},
		{
			name:        "top 100 ranked trades with tickers",
			newCommand:  NewTop100Command,
			args:        []string{"--date", "2026-04-30", "--ticker", "aapl,msft", "--limit", "25"},
			wantRank:    100,
			presetLen:   100,
			wantLength:  25,
			wantPreset:  "568",
			wantTickers: "AAPL,MSFT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset := &rankedPreset{rank: tt.wantRank, length: tt.presetLen, presetID: tt.wantPreset}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				options := rankedGetTradesRequestOptions(preset)
				options.length = tt.wantLength
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
			if len(got.Fields) != len(tradeFieldPresets["core"]) {
				t.Fatalf("len(Fields) = %d, want core fields", len(got.Fields))
			}
			if len(got.Rows) != 1 {
				t.Fatalf("len(Rows) = %d, want 1", len(got.Rows))
			}
			if len(got.Trades) != 0 {
				t.Fatalf("len(Trades) = %d, want 0 for default array shape", len(got.Trades))
			}
		})
	}
}

func TestSignalTradesCommands(t *testing.T) {
	tests := []struct {
		name        string
		newCommand  func() (*cobra.Command, error)
		args        []string
		wantPreset  *signalPreset
		wantTickers string
		wantLength  int
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
			wantLength: defaultTradeLimit,
		},
		{
			name:       "offsetting trades with tickers",
			newCommand: NewOffsettingCommand,
			args:       []string{"--date", "2026-04-30", "--ticker", "pltr", "--limit", "7"},
			wantPreset: &signalPreset{
				phantom:    "0",
				offsetting: "1",
				darkPools:  "-1",
				presetID:   "858",
			},
			wantTickers: "PLTR",
			wantLength:  7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				options := signalGetTradesRequestOptions(tt.wantPreset)
				options.length = tt.wantLength
				assertGetTradesRequestWithOptions(t, r, "2026-04-30", tt.wantTickers, &options)
				fmt.Fprint(w, `{"draw":1,"recordsTotal":2,"recordsFiltered":2,"data":[{"Ticker":"PLTR","PhantomPrint":1,"OffsettingTradeDate":"/Date(-2208988800000)/"}]}`)
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
			if len(got.Fields) != len(tradeFieldPresets["core"]) {
				t.Fatalf("len(Fields) = %d, want core fields", len(got.Fields))
			}
			if len(got.Rows) != 1 {
				t.Fatalf("len(Rows) = %d, want 1", len(got.Rows))
			}
			if len(got.Trades) != 0 {
				t.Fatalf("len(Trades) = %d, want 0 for default array shape", len(got.Trades))
			}
			if string(got.Rows[0][0]) != `"PLTR"` {
				t.Fatalf("first row ticker = %s, want PLTR", string(got.Rows[0][0]))
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
		{
			name:    "negative limit fails",
			args:    []string{"--date", "2026-04-30", "--limit", "-1"},
			wantErr: "use a value of 1 or greater",
		},
		{
			name:    "zero limit fails",
			args:    []string{"--date", "2026-04-30", "--limit", "0"},
			wantErr: "use a value of 1 or greater",
		},
		{
			name:    "limit above maximum fails",
			args:    []string{"--date", "2026-04-30", "--limit", "101"},
			wantErr: "100 or less",
		},
		{
			name:    "invalid field preset fails",
			args:    []string{"--date", "2026-04-30", "--preset-fields", "everything"},
			wantErr: "invalid preset-fields",
		},
		{
			name:    "invalid shape fails",
			args:    []string{"--date", "2026-04-30", "--shape", "table"},
			wantErr: "invalid shape",
		},
		{
			name:    "empty custom field fails",
			args:    []string{"--date", "2026-04-30", "--fields", "Ticker,,Dollars"},
			wantErr: "empty field",
		},
		{
			name:    "custom field whitespace fails",
			args:    []string{"--date", "2026-04-30", "--fields", "Ticker,Bad Field"},
			wantErr: "field names cannot contain whitespace",
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

func TestParseFieldsDeduplicatesRepeatedFields(t *testing.T) {
	t.Parallel()

	fields, err := parseFields("Ticker,Dollars,Ticker")
	if err != nil {
		t.Fatalf("parseFields() error = %v", err)
	}
	if got := strings.Join(fields, ","); got != "Ticker,Dollars" {
		t.Fatalf("fields = %q, want Ticker,Dollars", got)
	}
}

func TestRunRankedValidationReturnsBeforeNetwork(t *testing.T) {
	cmd, err := NewTop10Command()
	if err != nil {
		t.Fatalf("NewTop10Command() error = %v", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("request should not be sent when ranked validation fails")
	}))
	t.Cleanup(server.Close)
	withCommandDependencies(t, server.Client(), server.URL, func(context.Context) (map[string]string, error) {
		t.Fatal("cookies should not be extracted when ranked validation fails")
		return nil, nil
	}, func(context.Context, *http.Client, map[string]string) (string, error) {
		t.Fatal("XSRF token should not be fetched when ranked validation fails")
		return "", nil
	})

	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--date", "2026-04-30", "--fields", "Ticker,,Dollars"})

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "empty field") {
		t.Fatalf("expected empty field error, got %v", err)
	}
}

func TestProjectTradeOutputBranches(t *testing.T) {
	t.Parallel()

	trades := []json.RawMessage{
		json.RawMessage(`{"Ticker":"KRE","Dollars":17501965.25}`),
	}

	tests := []struct {
		name        string
		trades      []json.RawMessage
		fields      []string
		shape       string
		wantFields  []string
		wantRows    int
		wantTrades  int
		wantRowCell string
		wantTrade   string
	}{
		{
			name:       "nil trades become empty array rows",
			trades:     nil,
			fields:     []string{"Ticker"},
			shape:      defaultOutputShape,
			wantFields: []string{"Ticker"},
			wantRows:   0,
		},
		{
			name:        "array shape fills missing fields with null",
			trades:      trades,
			fields:      []string{"Ticker", "Missing"},
			shape:       defaultOutputShape,
			wantFields:  []string{"Ticker", "Missing"},
			wantRows:    1,
			wantRowCell: "null",
		},
		{
			name:       "object shape omits missing fields",
			trades:     trades,
			fields:     []string{"Ticker", "Missing"},
			shape:      objectOutputShape,
			wantFields: []string{"Ticker", "Missing"},
			wantTrades: 1,
			wantTrade:  `"Ticker":"KRE"`,
		},
		{
			name:       "full shape keeps raw trades",
			trades:     trades,
			fields:     nil,
			shape:      defaultOutputShape,
			wantTrades: 1,
			wantTrade:  `"Dollars":17501965.25`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := projectTradeOutput(tt.trades, tt.fields, tt.shape)
			if err != nil {
				t.Fatalf("projectTradeOutput() error = %v", err)
			}
			if strings.Join(got.fields, ",") != strings.Join(tt.wantFields, ",") {
				t.Fatalf("fields = %v, want %v", got.fields, tt.wantFields)
			}
			if len(got.rows) != tt.wantRows {
				t.Fatalf("len(rows) = %d, want %d", len(got.rows), tt.wantRows)
			}
			if len(got.trades) != tt.wantTrades {
				t.Fatalf("len(trades) = %d, want %d", len(got.trades), tt.wantTrades)
			}
			if tt.wantRowCell != "" && string(got.rows[0][1]) != tt.wantRowCell {
				t.Fatalf("missing row cell = %s, want %s", string(got.rows[0][1]), tt.wantRowCell)
			}
			if tt.wantTrade != "" && !bytes.Contains(got.trades[0], []byte(tt.wantTrade)) {
				t.Fatalf("trade = %s, want substring %s", string(got.trades[0]), tt.wantTrade)
			}
			if tt.shape == objectOutputShape && bytes.Contains(got.trades[0], []byte("Missing")) {
				t.Fatalf("object projection included missing field: %s", string(got.trades[0]))
			}
		})
	}
}

func TestProjectionRejectsInvalidTradeJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		shape string
	}{
		{name: "array projection", shape: defaultOutputShape},
		{name: "object projection", shape: objectOutputShape},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := projectTradeOutput([]json.RawMessage{json.RawMessage(`not-json`)}, []string{"Ticker"}, tt.shape)
			if err == nil {
				t.Fatal("expected decode error")
			}
			if !strings.Contains(err.Error(), "decode trade row") {
				t.Fatalf("expected decode context, got %v", err)
			}
		})
	}
}

func TestEncodeResultReportsWriterErrors(t *testing.T) {
	t.Parallel()

	err := encodeResult(failingWriter{}, "trades", Result{Status: "ok"}, false)
	if err == nil {
		t.Fatal("expected writer error")
	}
	if !strings.Contains(err.Error(), "encode trades response") {
		t.Fatalf("expected encode context, got %v", err)
	}
}

func TestFetchTradesPagesValidationAndPagination(t *testing.T) {
	tests := []struct {
		name         string
		limit        int
		responses    []string
		wantErr      string
		wantRows     int
		wantStarts   []string
		wantLengths  []string
		wantDraws    []string
		wantTotal    int
		wantFiltered int
	}{
		{
			name:    "rejects zero limit",
			limit:   0,
			wantErr: "limit must be 1 or greater",
		},
		{
			name:    "rejects oversized limit",
			limit:   101,
			wantErr: "limit must be 100 or less",
		},
		{
			name:         "truncates extra API rows to limit",
			limit:        2,
			responses:    []string{`{"draw":1,"recordsTotal":3,"recordsFiltered":3,"data":[{"Ticker":"A"},{"Ticker":"B"},{"Ticker":"C"}]}`},
			wantRows:     2,
			wantStarts:   []string{"0"},
			wantLengths:  []string{"2"},
			wantDraws:    []string{"1"},
			wantTotal:    3,
			wantFiltered: 3,
		},
		{
			name:         "stops when API returns short page",
			limit:        3,
			responses:    []string{`{"draw":1,"recordsTotal":5,"recordsFiltered":5,"data":[{"Ticker":"A"}]}`},
			wantRows:     1,
			wantStarts:   []string{"0"},
			wantLengths:  []string{"3"},
			wantDraws:    []string{"1"},
			wantTotal:    5,
			wantFiltered: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var starts []string
			var lengths []string
			var draws []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					t.Fatalf("ParseForm() error = %v", err)
				}
				starts = append(starts, r.Form.Get("start"))
				lengths = append(lengths, r.Form.Get("length"))
				draws = append(draws, r.Form.Get("draw"))
				if len(starts) > len(tt.responses) {
					t.Fatalf("unexpected request %d", len(starts))
				}
				fmt.Fprint(w, tt.responses[len(starts)-1])
			}))
			t.Cleanup(server.Close)

			withCommandDependencies(t, server.Client(), server.URL, nil, nil)

			options := defaultGetTradesRequestOptions()
			got, err := fetchTradesPages(t.Context(), "2026-04-30", "", &options, tt.limit)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("fetchTradesPages() error = %v", err)
			}
			if len(got.Data) != tt.wantRows {
				t.Fatalf("len(Data) = %d, want %d", len(got.Data), tt.wantRows)
			}
			if got.RecordsTotal != tt.wantTotal || got.RecordsFiltered != tt.wantFiltered {
				t.Fatalf("record counts = %d/%d, want %d/%d", got.RecordsTotal, got.RecordsFiltered, tt.wantTotal, tt.wantFiltered)
			}
			if strings.Join(starts, ",") != strings.Join(tt.wantStarts, ",") {
				t.Fatalf("starts = %v, want %v", starts, tt.wantStarts)
			}
			if strings.Join(lengths, ",") != strings.Join(tt.wantLengths, ",") {
				t.Fatalf("lengths = %v, want %v", lengths, tt.wantLengths)
			}
			if strings.Join(draws, ",") != strings.Join(tt.wantDraws, ",") {
				t.Fatalf("draws = %v, want %v", draws, tt.wantDraws)
			}
		})
	}
}

func TestFetchTradesPagesPropagatesCancellationBeforeRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("request should not be sent after cancellation")
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	options := defaultGetTradesRequestOptions()
	_, err := fetchTradesPages(ctx, "2026-04-30", "", &options, 2)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestFetchTradesHandlesGzipAndMalformedResponses(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr string
	}{
		{
			name: "gzip response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Encoding", "gzip")
				gz := gzip.NewWriter(w)
				defer gz.Close()
				fmt.Fprint(gz, `{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":[{"Ticker":"KRE"}]}`)
			},
		},
		{
			name: "invalid gzip response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Encoding", "gzip")
				fmt.Fprint(w, "not gzip")
			},
			wantErr: "decompress GetTrades response",
		},
		{
			name: "invalid JSON response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprint(w, `not-json`)
			},
			wantErr: "decode GetTrades response",
		},
		{
			name: "non OK status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
			},
			wantErr: "status 503",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			t.Cleanup(server.Close)

			withCommandDependencies(t, server.Client(), server.URL, nil, nil)

			got, err := fetchDisproportionatelyLargeTrades(t.Context(), "2026-04-30", "", 1)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("fetchDisproportionatelyLargeTrades() error = %v", err)
			}
			if len(got.Data) != 1 {
				t.Fatalf("len(Data) = %d, want 1", len(got.Data))
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

	_, err := fetchDisproportionatelyLargeTrades(t.Context(), "2026-04-30", "", defaultTradeLimit)
	if err == nil {
		t.Fatalf("expected API error")
	}
	if !strings.Contains(err.Error(), "bad filter") {
		t.Fatalf("expected bad filter error, got %v", err)
	}
}

func TestFetchDisproportionatelyLargeTradesHandlesDependencyErrors(t *testing.T) {
	tests := []struct {
		name    string
		extract func(context.Context) (map[string]string, error)
		fetch   func(context.Context, *http.Client, map[string]string) (string, error)
		wantErr string
	}{
		{
			name: "cookie extraction error",
			extract: func(context.Context) (map[string]string, error) {
				return nil, errors.New("cookie store unavailable")
			},
			wantErr: "extract VolumeLeaders browser cookies",
		},
		{
			name: "XSRF token error",
			fetch: func(context.Context, *http.Client, map[string]string) (string, error) {
				return "", errors.New("token page unavailable")
			},
			wantErr: "fetch VolumeLeaders XSRF token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				t.Fatal("request should not be sent when dependencies fail")
			}))
			t.Cleanup(server.Close)

			withCommandDependencies(t, server.Client(), server.URL, tt.extract, tt.fetch)

			_, err := fetchDisproportionatelyLargeTrades(t.Context(), "2026-04-30", "", 1)
			if err == nil {
				t.Fatal("expected dependency error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestFetchDisproportionatelyLargeTradesHandlesAuthStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	withCommandDependencies(t, server.Client(), server.URL, nil, nil)

	_, err := fetchDisproportionatelyLargeTrades(t.Context(), "2026-04-30", "", defaultTradeLimit)
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

func TestSessionExpiredResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		resp *http.Response
		want bool
	}{
		{
			name: "unauthorized status",
			resp: &http.Response{StatusCode: http.StatusUnauthorized},
			want: true,
		},
		{
			name: "forbidden status",
			resp: &http.Response{StatusCode: http.StatusForbidden},
			want: true,
		},
		{
			name: "login redirect request path",
			resp: &http.Response{StatusCode: http.StatusOK, Request: httptest.NewRequest(http.MethodPost, "https://www.volumeleaders.com/Login", http.NoBody)},
			want: true,
		},
		{
			name: "nil request is not expired",
			resp: &http.Response{StatusCode: http.StatusOK},
			want: false,
		},
		{
			name: "nil URL is not expired",
			resp: &http.Response{StatusCode: http.StatusOK, Request: &http.Request{}},
			want: false,
		},
		{
			name: "normal OK response",
			resp: &http.Response{StatusCode: http.StatusOK, Request: httptest.NewRequest(http.MethodPost, "https://www.volumeleaders.com/Trades/GetTrades", http.NoBody)},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := sessionExpiredResponse(tt.resp); got != tt.want {
				t.Fatalf("sessionExpiredResponse() = %t, want %t", got, tt.want)
			}
		})
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

	_, err := fetchDisproportionatelyLargeTrades(t.Context(), "2026-04-30", "", defaultTradeLimit)
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

	_, err := fetchDisproportionatelyLargeTrades(canceledCtx, "2026-04-30", "", defaultTradeLimit)
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

	_, err := fetchDisproportionatelyLargeTrades(t.Context(), "2026-04-30", "", defaultTradeLimit)
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

func assertGetTradeClustersRequestWithOptions(t *testing.T, r *http.Request, tradeDate, tickers string, options *getTradeClustersRequestOptions) {
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
	if got := r.Header.Get("Referer"); !strings.Contains(got, "TradeClusters") || !strings.Contains(got, "PresetSearchTemplateID="+options.presetSearchTemplateID) || !strings.Contains(got, "StartDate="+url.QueryEscape(tradeDate)) || !strings.Contains(got, "Tickers="+url.QueryEscape(tickers)) {
		t.Fatalf("Referer = %q, want trade clusters preset and single date", got)
	}
	assertCookie(t, r, "ASP.NET_SessionId", "session-cookie")
	assertCookie(t, r, ".ASPXAUTH", "auth-cookie")
	assertCookie(t, r, "__RequestVerificationToken", "cookie-token")

	if err := r.ParseForm(); err != nil {
		t.Fatalf("ParseForm() error = %v", err)
	}
	assertFormValue(t, r.Form, "Tickers", tickers)
	assertFormValue(t, r.Form, "StartDate", tradeDate)
	assertFormValue(t, r.Form, "EndDate", tradeDate)
	assertFormValue(t, r.Form, "MinDollars", "500000")
	assertFormValue(t, r.Form, "MaxDollars", options.maxDollars)
	assertFormValue(t, r.Form, "TradeClusterRank", fmt.Sprintf("%d", options.tradeClusterRank))
	assertFormValue(t, r.Form, "order[0][name]", "MinFullTimeString24")
	if got := r.Form.Get("PresetSearchTemplateID"); got != "" {
		t.Fatalf("form[PresetSearchTemplateID] = %q, want empty because preset is carried in Referer", got)
	}
	wantForm := getTradeClustersForm(tradeDate, tickers, options)
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

func indexOf(values []string, want string) int {
	for i, value := range values {
		if value == want {
			return i
		}
	}
	return -1
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
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
	oldClusterEndpoint := getTradeClustersEndpoint
	oldExtract := extractCookies
	oldFetch := fetchXSRFToken
	getTradesHTTPClient = client
	getTradesEndpoint = endpoint
	getTradeClustersEndpoint = endpoint
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
		getTradeClustersEndpoint = oldClusterEndpoint
		extractCookies = oldExtract
		fetchXSRFToken = oldFetch
	})
}
