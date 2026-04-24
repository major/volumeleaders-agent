package datatables

import (
	"net/url"
	"strings"
	"testing"
)

func TestRequestEncodeDefaultsAndColumnOrder(t *testing.T) {
	t.Parallel()

	request := Request{
		Columns:          []string{"Ticker", "Ticker", "Price"},
		Start:            5,
		Length:           25,
		OrderColumnIndex: 2,
	}

	encoded := request.Encode()
	expectedPrefix := strings.Join([]string{
		"draw=1",
		"start=5",
		"length=25",
		"order%5B0%5D%5Bcolumn%5D=2",
		"order%5B0%5D%5Bdir%5D=desc",
		"columns%5B0%5D%5Bdata%5D=Ticker",
		"columns%5B0%5D%5Bname%5D=Ticker",
		"columns%5B0%5D%5Bsearchable%5D=true",
		"columns%5B0%5D%5Borderable%5D=true",
		"columns%5B0%5D%5Bsearch%5D%5Bvalue%5D=",
		"columns%5B0%5D%5Bsearch%5D%5Bregex%5D=false",
		"columns%5B1%5D%5Bdata%5D=Ticker",
		"columns%5B1%5D%5Bname%5D=Ticker",
	}, "&")

	if !strings.HasPrefix(encoded, expectedPrefix) {
		t.Fatalf("encoded request prefix mismatch\nexpected prefix: %s\ngot:             %s", expectedPrefix, encoded)
	}
}

func TestRequestEncodeCustomFilters(t *testing.T) {
	t.Parallel()

	request := Request{
		Columns:          []string{"Ticker"},
		Start:            0,
		Length:           -1,
		OrderColumnIndex: 1,
		OrderDirection:   "asc",
		CustomFilters: map[string]string{
			"Date":    "2026-04-24",
			"Tickers": "AAPL,NVDA",
		},
		Draw: 7,
	}

	values, err := url.ParseQuery(request.Encode())
	if err != nil {
		t.Fatalf("parse encoded request: %v", err)
	}

	checks := map[string]string{
		"draw":                      "7",
		"start":                     "0",
		"length":                    "-1",
		"order[0][column]":          "1",
		"order[0][dir]":             "asc",
		"columns[0][data]":          "Ticker",
		"columns[0][name]":          "Ticker",
		"columns[0][searchable]":    "true",
		"columns[0][orderable]":     "true",
		"columns[0][search][value]": "",
		"columns[0][search][regex]": "false",
		"Date":                      "2026-04-24",
		"Tickers":                   "AAPL,NVDA",
	}
	for key, expected := range checks {
		if got := values.Get(key); got != expected {
			t.Errorf("%s: expected %q, got %q", key, expected, got)
		}
	}
}
