package commands

import (
	"bytes"
	"context"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/major/volumeleaders-agent/internal/datatables"
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
			ctx := context.WithValue(context.Background(), prettyJSONKey, tt.pretty)

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
	ctx := context.Background()
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
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}
	assertFlagName(t, flags[0], "start-date")
	assertFlagName(t, flags[1], "end-date")
	assertStringFlagRequired(t, flags[0], true)
	assertStringFlagRequired(t, flags[1], true)
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
