package common

import "testing"

func TestContextKeyValues(t *testing.T) {
	t.Parallel()

	if PrettyJSONKey != 1 {
		t.Errorf("PrettyJSONKey = %d, want 1", PrettyJSONKey)
	}
	if TestClientKey != 2 {
		t.Errorf("TestClientKey = %d, want 2", TestClientKey)
	}
}

func TestOutputFormatConstants(t *testing.T) {
	t.Parallel()

	if OutputFormatJSON != "json" {
		t.Errorf("OutputFormatJSON = %q, want json", OutputFormatJSON)
	}
	if OutputFormatCSV != "csv" {
		t.Errorf("OutputFormatCSV = %q, want csv", OutputFormatCSV)
	}
	if OutputFormatTSV != "tsv" {
		t.Errorf("OutputFormatTSV = %q, want tsv", OutputFormatTSV)
	}
}

func TestDataTableOptionsFields(t *testing.T) {
	t.Parallel()

	opts := DataTableOptions{
		Start:    10,
		Length:   25,
		OrderCol: 1,
		OrderDir: "asc",
		Filters:  map[string]string{"Ticker": "AAPL"},
		Fields:   []string{"Ticker", "Name"},
	}

	if opts.Start != 10 || opts.Length != 25 || opts.OrderCol != 1 || opts.OrderDir != "asc" {
		t.Fatalf("unexpected DataTableOptions scalar fields: %+v", opts)
	}
	if opts.Filters["Ticker"] != "AAPL" {
		t.Errorf("Filters[Ticker] = %q, want AAPL", opts.Filters["Ticker"])
	}
	if len(opts.Fields) != 2 || opts.Fields[0] != "Ticker" || opts.Fields[1] != "Name" {
		t.Errorf("Fields = %v, want [Ticker Name]", opts.Fields)
	}
}

func TestPaginationPageSize(t *testing.T) {
	t.Parallel()

	if PaginationPageSize != 1000 {
		t.Errorf("PaginationPageSize = %d, want 1000", PaginationPageSize)
	}
}
