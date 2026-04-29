package commands

import (
	"bytes"
	"context"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

func TestParseSnapshotString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected map[string]float64
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]float64{},
		},
		{
			name:  "single ticker",
			input: "AAPL:255.30",
			expected: map[string]float64{
				"AAPL": 255.30,
			},
		},
		{
			name:  "multiple tickers",
			input: "A:114.73;AA:70.96;AAPL:255.30",
			expected: map[string]float64{
				"A":    114.73,
				"AA":   70.96,
				"AAPL": 255.30,
			},
		},
		{
			name:  "trailing semicolon",
			input: "AAPL:255.30;MSFT:420.50;",
			expected: map[string]float64{
				"AAPL": 255.30,
				"MSFT": 420.50,
			},
		},
		{
			name:  "skips malformed entries",
			input: "AAPL:255.30;BADENTRY;MSFT:420.50",
			expected: map[string]float64{
				"AAPL": 255.30,
				"MSFT": 420.50,
			},
		},
		{
			name:  "skips non-numeric prices",
			input: "AAPL:255.30;BAD:notanumber;MSFT:420.50",
			expected: map[string]float64{
				"AAPL": 255.30,
				"MSFT": 420.50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseSnapshotString(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d entries, got %d", len(tt.expected), len(result))
			}
			for ticker, expectedPrice := range tt.expected {
				got, ok := result[ticker]
				if !ok {
					t.Errorf("missing ticker %s", ticker)
					continue
				}
				if got != expectedPrice {
					t.Errorf("ticker %s: expected %f, got %f", ticker, expectedPrice, got)
				}
			}
		})
	}
}

func TestIntStr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{42, "42"},
		{-1, "-1"},
		{2000000000, "2000000000"},
	}
	for _, tt := range tests {
		if got := intStr(tt.input); got != tt.expected {
			t.Errorf("intStr(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0"},
		{1.5, "1.5"},
		{100000, "100000"},
		{99.25, "99.25"},
		{30000000000, "30000000000"},
		{0.001, "0.001"},
	}
	for _, tt := range tests {
		if got := formatFloat(tt.input); got != tt.expected {
			t.Errorf("formatFloat(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestBoolString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    bool
		expected string
	}{
		{true, "true"},
		{false, "false"},
	}
	for _, tt := range tests {
		if got := boolString(tt.input); got != tt.expected {
			t.Errorf("boolString(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestToDateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input, expected string
	}{
		{"2025-01-15", "20250115"},
		{"2026-04-24", "20260424"},
		{"20250115", "20250115"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := toDateKey(tt.input); got != tt.expected {
			t.Errorf("toDateKey(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNewDataTablesRequest(t *testing.T) {
	t.Parallel()

	columns := []string{"Col1", "Col2"}
	opts := dataTableOptions{
		start:    10,
		length:   25,
		orderCol: 1,
		orderDir: "asc",
		filters:  map[string]string{"Ticker": "AAPL"},
	}
	got := newDataTablesRequest(columns, opts)
	expected := datatables.Request{
		Columns:          columns,
		Start:            10,
		Length:           25,
		OrderColumnIndex: 1,
		OrderDirection:   "asc",
		CustomFilters:    map[string]string{"Ticker": "AAPL"},
		Draw:             1,
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("newDataTablesRequest mismatch\nexpected: %+v\ngot:      %+v", expected, got)
	}
}

func TestNewDataTablesRequestNilFilters(t *testing.T) {
	t.Parallel()

	got := newDataTablesRequest(nil, dataTableOptions{orderDir: "desc"})
	if got.Draw != 1 {
		t.Errorf("expected Draw=1, got %d", got.Draw)
	}
	if got.OrderDirection != "desc" {
		t.Errorf("expected OrderDirection=desc, got %q", got.OrderDirection)
	}
}

func TestPrintJSON(t *testing.T) {
	// Not parallel: captures os.Stdout.
	tests := []struct {
		name     string
		pretty   bool
		input    any
		expected string
	}{
		{
			name:     "compact object",
			pretty:   false,
			input:    map[string]string{"ticker": "AAPL"},
			expected: "{\"ticker\":\"AAPL\"}\n",
		},
		{
			name:     "pretty object",
			pretty:   true,
			input:    map[string]string{"ticker": "AAPL"},
			expected: "{\n  \"ticker\": \"AAPL\"\n}\n",
		},
		{
			name:     "compact array",
			pretty:   false,
			input:    []int{1, 2, 3},
			expected: "[1,2,3]\n",
		},
		{
			name:     "null value",
			pretty:   false,
			input:    nil,
			expected: "null\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(t.Context(), prettyJSONKey, tt.pretty)

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := printJSON(ctx, tt.input)
			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var buf bytes.Buffer
			if _, copyErr := io.Copy(&buf, r); copyErr != nil {
				t.Fatalf("read pipe: %v", copyErr)
			}
			if buf.String() != tt.expected {
				t.Errorf("printJSON output:\nexpected: %q\ngot:      %q", tt.expected, buf.String())
			}
		})
	}
}

func TestPrintJSONMarshalError(t *testing.T) {
	ctx := t.Context()
	err := printJSON(ctx, make(chan int))
	if err == nil {
		t.Fatal("expected marshal error")
	}
	if !strings.Contains(err.Error(), "marshal JSON") {
		t.Errorf("expected marshal JSON error, got: %v", err)
	}
}

func TestDateRangeFlags(t *testing.T) {
	t.Parallel()

	flags := dateRangeFlags()
	if len(flags) != 3 {
		t.Fatalf("expected 3 flags, got %d", len(flags))
	}
	assertFlagName(t, flags[0], "start-date")
	assertFlagName(t, flags[1], "end-date")
	assertFlagName(t, flags[2], "days")
	assertStringFlagRequired(t, flags[0], false)
	assertStringFlagRequired(t, flags[1], false)
}

func TestOptionalDateRangeFlags(t *testing.T) {
	t.Parallel()

	flags := optionalDateRangeFlags()
	if len(flags) != 3 {
		t.Fatalf("expected 3 flags, got %d", len(flags))
	}
	assertFlagName(t, flags[0], "start-date")
	assertFlagName(t, flags[1], "end-date")
	assertFlagName(t, flags[2], "days")
	assertStringFlagRequired(t, flags[0], false)
	assertStringFlagRequired(t, flags[1], false)
}

func TestDefaultDates(t *testing.T) {
	t.Parallel()

	frozen := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	origTimeNow := timeNow
	timeNow = func() time.Time { return frozen }
	t.Cleanup(func() { timeNow = origTimeNow })

	tests := []struct {
		name         string
		args         []string
		lookbackDays int
		wantStart    string
		wantEnd      string
	}{
		{
			name:         "both dates explicit",
			args:         []string{"app", "sub", "--start-date", "2025-01-01", "--end-date", "2025-03-01"},
			lookbackDays: 90,
			wantStart:    "2025-01-01",
			wantEnd:      "2025-03-01",
		},
		{
			name:         "90-day lookback default",
			args:         []string{"app", "sub"},
			lookbackDays: 90,
			wantStart:    "2025-03-17",
			wantEnd:      "2025-06-15",
		},
		{
			name:         "today-only default",
			args:         []string{"app", "sub"},
			lookbackDays: 0,
			wantStart:    "2025-06-15",
			wantEnd:      "2025-06-15",
		},
		{
			name:         "365-day lookback default",
			args:         []string{"app", "sub"},
			lookbackDays: 365,
			wantStart:    "2024-06-15",
			wantEnd:      "2025-06-15",
		},
		{
			name:         "only start-date explicit",
			args:         []string{"app", "sub", "--start-date", "2025-01-01"},
			lookbackDays: 90,
			wantStart:    "2025-01-01",
			wantEnd:      "2025-06-15",
		},
		{
			name:         "only end-date explicit",
			args:         []string{"app", "sub", "--end-date", "2025-03-01"},
			lookbackDays: 90,
			wantStart:    "2025-03-17",
			wantEnd:      "2025-03-01",
		},
		{
			name:         "days defaults from today",
			args:         []string{"app", "sub", "--days", "5"},
			lookbackDays: 90,
			wantStart:    "2025-06-10",
			wantEnd:      "2025-06-15",
		},
		{
			name:         "days uses explicit end-date as base",
			args:         []string{"app", "sub", "--end-date", "2025-03-01", "--days", "5"},
			lookbackDays: 90,
			wantStart:    "2025-02-24",
			wantEnd:      "2025-03-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotStart, gotEnd string
			sub := &cli.Command{
				Name:  "sub",
				Flags: slices.Concat(optionalDateRangeFlags()),
				Action: func(_ context.Context, cmd *cli.Command) error {
					gotStart, gotEnd = defaultDates(cmd, tt.lookbackDays)
					return nil
				},
			}
			root := &cli.Command{Commands: []*cli.Command{sub}}
			if err := root.Run(context.Background(), tt.args); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotStart != tt.wantStart {
				t.Errorf("start date = %q, want %q", gotStart, tt.wantStart)
			}
			if gotEnd != tt.wantEnd {
				t.Errorf("end date = %q, want %q", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestRequiredDateRange(t *testing.T) {
	frozen := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	origTimeNow := timeNow
	timeNow = func() time.Time { return frozen }
	t.Cleanup(func() { timeNow = origTimeNow })

	tests := []struct {
		name      string
		args      []string
		wantStart string
		wantEnd   string
		wantErr   string
	}{
		{
			name:      "explicit range",
			args:      []string{"app", "sub", "--start-date", "2025-01-01", "--end-date", "2025-01-31"},
			wantStart: "2025-01-01",
			wantEnd:   "2025-01-31",
		},
		{
			name:      "days supplies range",
			args:      []string{"app", "sub", "--days", "7"},
			wantStart: "2025-06-08",
			wantEnd:   "2025-06-15",
		},
		{
			name:    "missing range",
			args:    []string{"app", "sub"},
			wantErr: "required unless --days is set",
		},
		{
			name:    "negative days",
			args:    []string{"app", "sub", "--days", "-1"},
			wantErr: "--days must be greater than or equal to 0",
		},
		{
			name:    "days conflicts with start-date",
			args:    []string{"app", "sub", "--start-date", "2025-06-01", "--days", "7"},
			wantErr: "--days cannot be used with --start-date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotStart, gotEnd string
			sub := &cli.Command{
				Name:  "sub",
				Flags: dateRangeFlags(),
				Action: func(_ context.Context, cmd *cli.Command) error {
					var err error
					gotStart, gotEnd, err = requiredDateRange(cmd)
					return err
				},
			}
			root := &cli.Command{Commands: []*cli.Command{sub}}
			err := root.Run(context.Background(), tt.args)
			assertErrContains(t, err, tt.wantErr)
			if tt.wantErr != "" {
				return
			}
			if gotStart != tt.wantStart {
				t.Errorf("start date = %q, want %q", gotStart, tt.wantStart)
			}
			if gotEnd != tt.wantEnd {
				t.Errorf("end date = %q, want %q", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestSingleTickerValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr string
	}{
		{
			name: "flag ticker",
			args: []string{"app", "sub", "--ticker", "AAPL"},
			want: "AAPL",
		},
		{
			name: "positional ticker",
			args: []string{"app", "sub", "MSFT"},
			want: "MSFT",
		},
		{
			name:    "flag and positional ticker conflict",
			args:    []string{"app", "sub", "MSFT", "--ticker", "AAPL"},
			wantErr: "use either --ticker or a ticker argument, not both",
		},
		{
			name:    "too many positional tickers",
			args:    []string{"app", "sub", "AAPL", "MSFT"},
			wantErr: "expected at most one ticker argument",
		},
		{
			name:    "missing ticker",
			args:    []string{"app", "sub"},
			wantErr: "--ticker or a ticker argument is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			sub := &cli.Command{
				Name: "sub",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "ticker"},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					var err error
					got, err = singleTickerValue(cmd)
					return err
				},
			}
			root := &cli.Command{Commands: []*cli.Command{sub}}
			err := root.Run(context.Background(), tt.args)
			assertErrContains(t, err, tt.wantErr)
			if tt.wantErr == "" && got != tt.want {
				t.Errorf("ticker = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVolumeRangeFlags(t *testing.T) {
	t.Parallel()

	flags := volumeRangeFlags()
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}
	assertFlagName(t, flags[0], "min-volume")
	assertFlagName(t, flags[1], "max-volume")
}

func TestPriceRangeFlags(t *testing.T) {
	t.Parallel()

	flags := priceRangeFlags()
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}
	assertFlagName(t, flags[0], "min-price")
	assertFlagName(t, flags[1], "max-price")
}

func TestDollarRangeFlags(t *testing.T) {
	t.Parallel()

	flags := dollarRangeFlags(500000)
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}
	assertFlagName(t, flags[0], "min-dollars")
	assertFlagName(t, flags[1], "max-dollars")
}

func TestPaginationFlags(t *testing.T) {
	t.Parallel()

	flags := paginationFlags(100, 1, "asc")
	if len(flags) != 4 {
		t.Fatalf("expected 4 flags, got %d", len(flags))
	}
	assertFlagName(t, flags[0], "start")
	assertFlagName(t, flags[1], "length")
	assertFlagName(t, flags[2], "order-col")
	assertFlagName(t, flags[3], "order-dir")
}

func TestOutputFormatFlags(t *testing.T) {
	t.Parallel()

	flags := outputFormatFlags()
	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	assertFlagName(t, flags[0], "format")
}

func TestParseOutputFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    outputFormat
		expectError bool
	}{
		{name: "empty defaults to json", input: "", expected: outputFormatJSON},
		{name: "json", input: "json", expected: outputFormatJSON},
		{name: "csv", input: "csv", expected: outputFormatCSV},
		{name: "tsv with whitespace and case", input: " TSV ", expected: outputFormatTSV},
		{name: "invalid", input: "table", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseOutputFormat(tt.input)
			if tt.expectError {
				assertErrContains(t, err, "valid formats: json,csv,tsv")
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("parseOutputFormat(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPrintDataTablesResultDelimited(t *testing.T) {
	// Not parallel: captures os.Stdout.
	industry := "Software, Infrastructure"
	rows := []models.Trade{
		{
			Ticker:   "AAPL",
			Name:     "Apple, Inc.",
			Industry: &industry,
			DarkPool: true,
			Dollars:  123.45,
			TradeID:  9007199254740993,
		},
		{
			Ticker:   "MSFT",
			Name:     "Microsoft Corp",
			DarkPool: false,
			Dollars:  0,
		},
	}

	tests := []struct {
		name     string
		format   outputFormat
		fields   []string
		expected string
	}{
		{
			name:     "csv quotes strings and leaves nil empty",
			format:   outputFormatCSV,
			fields:   []string{"Ticker", "Name", "Industry", "DarkPool", "Dollars", "TradeID"},
			expected: "Ticker,Name,Industry,DarkPool,Dollars,TradeID\nAAPL,\"Apple, Inc.\",\"Software, Infrastructure\",true,123.45,9007199254740993\nMSFT,Microsoft Corp,,false,0,0\n",
		},
		{
			name:     "tsv uses tabs",
			format:   outputFormatTSV,
			fields:   []string{"Ticker", "DarkPool"},
			expected: "Ticker\tDarkPool\nAAPL\ttrue\nMSFT\tfalse\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(t, func() {
				err := printDataTablesResult(context.Background(), rows, tt.fields, tt.format)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})
			if output != tt.expected {
				t.Errorf("delimited output:\nexpected: %q\ngot:      %q", tt.expected, output)
			}
		})
	}
}

func TestRequireStringFlag(t *testing.T) {
	t.Parallel()

	flags := []cli.Flag{
		&cli.StringFlag{Name: "name"},
		&cli.IntFlag{Name: "count"},
		&cli.StringFlag{Name: "other"},
	}
	requireStringFlag(flags, "name")

	nameFlag := flags[0].(*cli.StringFlag)
	if !nameFlag.Required {
		t.Error("expected 'name' flag to be required")
	}
	otherFlag := flags[2].(*cli.StringFlag)
	if otherFlag.Required {
		t.Error("expected 'other' flag to remain not required")
	}
}

func TestRequireStringFlagNoMatch(t *testing.T) {
	t.Parallel()

	flags := []cli.Flag{
		&cli.StringFlag{Name: "name"},
	}
	// Should not panic when the target flag is not found.
	requireStringFlag(flags, "nonexistent")
}

func assertFlagName(t *testing.T, flag cli.Flag, expected string) {
	t.Helper()
	names := flag.Names()
	if len(names) == 0 || names[0] != expected {
		t.Errorf("expected flag name %q, got %v", expected, names)
	}
}

func assertStringFlagRequired(t *testing.T, flag cli.Flag, expected bool) {
	t.Helper()
	sf, ok := flag.(*cli.StringFlag)
	if !ok {
		t.Errorf("expected *cli.StringFlag")
		return
	}
	if sf.Required != expected {
		t.Errorf("expected flag %q Required=%v, got %v", sf.Name, expected, sf.Required)
	}
}
