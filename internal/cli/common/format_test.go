package common

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

type fieldTestRow struct {
	Ticker string  `json:"Ticker"`
	Name   string  `json:"Name,omitempty"`
	Price  float64 `json:"Price"`
	Hidden string  `json:"-"`
	NoTag  string
}

func TestParseOutputFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    OutputFormat
		expectError bool
	}{
		{name: "empty defaults to json", input: "", expected: OutputFormatJSON},
		{name: "json", input: "json", expected: OutputFormatJSON},
		{name: "csv", input: "csv", expected: OutputFormatCSV},
		{name: "tsv with whitespace and case", input: " TSV ", expected: OutputFormatTSV},
		{name: "invalid", input: "table", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseOutputFormat(tt.input)
			if tt.expectError {
				requireErrContains(t, err, "valid formats: json,csv,tsv")
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ParseOutputFormat(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSelectJSONFields(t *testing.T) {
	t.Parallel()

	rows := []fieldTestRow{{Ticker: "AAPL", Name: "Apple", Price: 123.45}}
	selected, err := SelectJSONFields(rows, []string{"Ticker", "Price", "Missing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 1 {
		t.Fatalf("expected one row, got %d", len(selected))
	}
	if string(selected[0]["Ticker"]) != `"AAPL"` {
		t.Errorf("Ticker = %s, want AAPL JSON string", selected[0]["Ticker"])
	}
	if string(selected[0]["Price"]) != `123.45` {
		t.Errorf("Price = %s, want 123.45", selected[0]["Price"])
	}
	if selected[0]["Missing"] != nil {
		t.Errorf("Missing = %s, want nil RawMessage", selected[0]["Missing"])
	}
}

func TestSelectJSONFieldsMarshalError(t *testing.T) {
	t.Parallel()

	_, err := SelectJSONFields([]chan int{make(chan int)}, []string{"Ticker"})
	requireErrContains(t, err, "marshal item for field selection")
}

func TestParseJSONFieldList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr string
	}{
		{name: "empty", input: "", want: nil},
		{name: "whitespace", input: "   ", want: nil},
		{name: "single field", input: "Ticker", want: []string{"Ticker"}},
		{name: "trims and dedupes", input: " Ticker, Price, Ticker ", want: []string{"Ticker", "Price"}},
		{name: "skips empty segments", input: "Ticker,,Price", want: []string{"Ticker", "Price"}},
		{name: "invalid", input: "Ticker,BadField", wantErr: `invalid field "BadField"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseJSONFieldList[fieldTestRow](tt.input)
			requireErrContains(t, err, tt.wantErr)
			if tt.wantErr != "" {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseJSONFieldList(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestOutputFields(t *testing.T) {
	t.Parallel()

	defaults := []string{"Ticker"}
	got, err := OutputFields[fieldTestRow]("", defaults)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, defaults) {
		t.Errorf("default OutputFields = %v, want %v", got, defaults)
	}

	got, err = OutputFields[fieldTestRow]("all", defaults)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"Ticker", "Name", "Price"}) {
		t.Errorf("all OutputFields = %v, want [Ticker Name Price]", got)
	}

	got, err = OutputFields[fieldTestRow]("Name,Price", defaults)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"Name", "Price"}) {
		t.Errorf("selected OutputFields = %v, want [Name Price]", got)
	}
}

func TestJSONFieldNames(t *testing.T) {
	t.Parallel()

	got := JSONFieldNames[fieldTestRow]()
	want := []string{"Name", "Price", "Ticker"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("JSONFieldNames = %v, want %v", got, want)
	}
	if !slices.IsSorted(got) {
		t.Errorf("JSONFieldNames not sorted: %v", got)
	}
}

func TestJSONFieldNamesInOrder(t *testing.T) {
	t.Parallel()

	got := JSONFieldNamesInOrder[*fieldTestRow]()
	want := []string{"Ticker", "Name", "Price"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("JSONFieldNamesInOrder = %v, want %v", got, want)
	}
}

func TestSelectJSONFieldsDecodeError(t *testing.T) {
	t.Parallel()

	items := [][]int{{1, 2, 3}}
	_, err := SelectJSONFields(items, []string{"Ticker"})
	requireErrContains(t, err, "decode item for field selection")
}

func requireErrContains(t *testing.T, err error, want string) {
	t.Helper()
	if want == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected error containing %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}
