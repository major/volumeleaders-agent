package common

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/major/volumeleaders-agent/internal/models"
)

func TestPrintJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pretty   bool
		input    any
		expected string
	}{
		{name: "compact object", pretty: false, input: map[string]string{"ticker": "AAPL"}, expected: "{\"ticker\":\"AAPL\"}\n"},
		{name: "pretty object", pretty: true, input: map[string]string{"ticker": "AAPL"}, expected: "{\n  \"ticker\": \"AAPL\"\n}\n"},
		{name: "compact array", pretty: false, input: []int{1, 2, 3}, expected: "[1,2,3]\n"},
		{name: "null value", pretty: false, input: nil, expected: "null\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.WithValue(t.Context(), PrettyJSONKey, tt.pretty)
			var buf bytes.Buffer
			err := PrintJSON(&buf, ctx, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if buf.String() != tt.expected {
				t.Errorf("PrintJSON output:\nexpected: %q\ngot:      %q", tt.expected, buf.String())
			}

			if !json.Valid(bytes.TrimSpace(buf.Bytes())) {
				t.Errorf("PrintJSON wrote invalid JSON: %q", buf.String())
			}
		})
	}
}

func TestPrintJSONMarshalError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := PrintJSON(&buf, t.Context(), make(chan int))
	requireErrContains(t, err, "marshal JSON")
}

func TestPrintJSONWriteError(t *testing.T) {
	t.Parallel()

	err := PrintJSON(errorWriter{}, t.Context(), map[string]string{"ticker": "AAPL"})
	requireErrContains(t, err, "write JSON")
}

func TestPrintDelimited(t *testing.T) {
	t.Parallel()

	rows := []map[string]json.RawMessage{
		{
			"Ticker":   json.RawMessage(`"AAPL"`),
			"Name":     json.RawMessage(`"Apple, Inc."`),
			"Industry": json.RawMessage(`"Software, Infrastructure"`),
			"DarkPool": json.RawMessage(`true`),
			"Dollars":  json.RawMessage(`123.45`),
			"TradeID":  json.RawMessage(`9007199254740993`),
		},
		{
			"Ticker":   json.RawMessage(`"MSFT"`),
			"Name":     json.RawMessage(`"Microsoft Corp"`),
			"Industry": json.RawMessage(`null`),
			"DarkPool": json.RawMessage(`false`),
			"Dollars":  json.RawMessage(`0`),
			"TradeID":  json.RawMessage(`0`),
		},
	}

	tests := []struct {
		name     string
		comma    rune
		expected string
	}{
		{name: "csv with header row and comma delimiter", comma: ',', expected: "Ticker,Name,Industry,DarkPool,Dollars,TradeID\nAAPL,\"Apple, Inc.\",\"Software, Infrastructure\",true,123.45,9007199254740993\nMSFT,Microsoft Corp,,false,0,0\n"},
		{name: "tsv with header row and tab delimiter", comma: '\t', expected: "Ticker\tName\tIndustry\tDarkPool\tDollars\tTradeID\nAAPL\tApple, Inc.\tSoftware, Infrastructure\ttrue\t123.45\t9007199254740993\nMSFT\tMicrosoft Corp\t\tfalse\t0\t0\n"},
	}

	fields := []string{"Ticker", "Name", "Industry", "DarkPool", "Dollars", "TradeID"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := PrintDelimited(&buf, rows, fields, tt.comma)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if buf.String() != tt.expected {
				t.Errorf("delimited output:\nexpected: %q\ngot:      %q", tt.expected, buf.String())
			}
		})
	}
}

func TestPrintDelimitedErrors(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := PrintDelimited(&buf, []map[string]json.RawMessage{{"bad": json.RawMessage(`{`)}}, []string{"bad"}, ',')
	requireErrContains(t, err, "format field")

	err = PrintDelimited(errorWriter{}, nil, []string{"Ticker"}, ',')
	requireErrContains(t, err, "flush delimited output")
}

func TestDelimitedValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     json.RawMessage
		want    string
		wantErr string
	}{
		{name: "missing", raw: nil, want: ""},
		{name: "null", raw: json.RawMessage(`null`), want: ""},
		{name: "integer", raw: json.RawMessage(`42`), want: "42"},
		{name: "negative", raw: json.RawMessage(`-1.25`), want: "-1.25"},
		{name: "string", raw: json.RawMessage(`"hello"`), want: "hello"},
		{name: "boolean", raw: json.RawMessage(`true`), want: "true"},
		{name: "array", raw: json.RawMessage(`["a",1]`), want: `["a",1]`},
		{name: "object", raw: json.RawMessage(`{"a":1}`), want: `{"a":1}`},
		{name: "invalid", raw: json.RawMessage(`{`), wantErr: "decode JSON value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := delimitedValue(tt.raw)
			requireErrContains(t, err, tt.wantErr)
			if tt.wantErr == "" && got != tt.want {
				t.Errorf("delimitedValue(%s) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestIsJSONNumber(t *testing.T) {
	t.Parallel()

	for _, raw := range []json.RawMessage{json.RawMessage(`0`), json.RawMessage(`123.45`), json.RawMessage(`-1`)} {
		if !isJSONNumber(raw) {
			t.Errorf("isJSONNumber(%s) = false, want true", raw)
		}
	}
	for _, raw := range []json.RawMessage{json.RawMessage(`true`), json.RawMessage(`"1"`), json.RawMessage(`null`)} {
		if isJSONNumber(raw) {
			t.Errorf("isJSONNumber(%s) = true, want false", raw)
		}
	}
}

func TestPrintDataTablesResult(t *testing.T) {
	t.Parallel()

	industry := "Software, Infrastructure"
	rows := []models.Trade{
		{Ticker: "AAPL", Name: "Apple, Inc.", Industry: &industry, DarkPool: true, Dollars: 123.45, TradeID: 9007199254740993},
		{Ticker: "MSFT", Name: "Microsoft Corp", DarkPool: false, Dollars: 0},
	}

	tests := []struct {
		name     string
		ctx      context.Context
		format   OutputFormat
		fields   []string
		expected string
	}{
		{name: "json full compact", ctx: t.Context(), format: OutputFormatJSON},
		{name: "json selected pretty", ctx: context.WithValue(t.Context(), PrettyJSONKey, true), format: OutputFormatJSON, fields: []string{"Ticker", "DarkPool"}, expected: "[\n  {\n    \"DarkPool\": true,\n    \"Ticker\": \"AAPL\"\n  },\n  {\n    \"DarkPool\": false,\n    \"Ticker\": \"MSFT\"\n  }\n]\n"},
		{name: "csv quotes strings and leaves nil empty", ctx: t.Context(), format: OutputFormatCSV, fields: []string{"Ticker", "Name", "Industry", "DarkPool", "Dollars", "TradeID"}, expected: "Ticker,Name,Industry,DarkPool,Dollars,TradeID\nAAPL,\"Apple, Inc.\",\"Software, Infrastructure\",true,123.45,9007199254740993\nMSFT,Microsoft Corp,,false,0,0\n"},
		{name: "tsv uses tabs", ctx: t.Context(), format: OutputFormatTSV, fields: []string{"Ticker", "DarkPool"}, expected: "Ticker\tDarkPool\nAAPL\ttrue\nMSFT\tfalse\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := PrintDataTablesResult(&buf, tt.ctx, rows, tt.fields, tt.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expected == "" {
				if !strings.Contains(buf.String(), `"Ticker":"AAPL"`) || !json.Valid(bytes.TrimSpace(buf.Bytes())) {
					t.Fatalf("expected valid full JSON trade output, got %s", buf.String())
				}
				return
			}
			if buf.String() != tt.expected {
				t.Errorf("PrintDataTablesResult output:\nexpected: %q\ngot:      %q", tt.expected, buf.String())
			}
		})
	}
}

func TestPrintDataTablesResultUnsupportedFormat(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := PrintDataTablesResult(&buf, t.Context(), []fieldTestRow{{Ticker: "AAPL"}}, nil, OutputFormat("table"))
	requireErrContains(t, err, `unsupported output format "table"`)
}

type errorWriter struct{}

func (errorWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}
