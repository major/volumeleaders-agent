package volume

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/major/volumeleaders-agent/internal/cli/testutil"
)

func TestVolumeSubcommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{"institutional", "/InstitutionalVolume/GetInstitutionalVolume"},
		{"ah-institutional", "/AHInstitutionalVolume/GetAHInstitutionalVolume"},
		{"total", "/TotalVolume/GetTotalVolume"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("expected path %s, got %s", tt.path, r.URL.Path)
				}
				fmt.Fprint(w, testutil.DataTablesJSON(`[{}]`))
			}))
			t.Cleanup(server.Close)

			ctx := testutil.ContextWithTestClient(t, server.URL)
			cmd := NewVolumeCommand()
			stdout, _, err := testutil.ExecuteCommand(t, cmd, ctx, tt.name, "--date", "2025-01-15")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if stdout == "" {
				t.Error("expected non-empty stdout")
			}
		})
	}
}

func TestVolumeMissingDateFlag(t *testing.T) {
	t.Parallel()

	tests := []string{"institutional", "ah-institutional", "total"}

	for _, sub := range tests {
		t.Run(sub, func(t *testing.T) {
			t.Parallel()

			ctx := testutil.ContextWithTestClient(t, "http://unused")
			cmd := NewVolumeCommand()
			_, _, err := testutil.ExecuteCommand(t, cmd, ctx, sub)
			if err == nil {
				t.Fatal("expected error for missing --date flag, got nil")
			}
		})
	}
}

func TestVolumeTickersFlag(t *testing.T) {
	t.Parallel()

	var gotTickers string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		gotTickers = form.Get("Tickers")
		fmt.Fprint(w, testutil.DataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewVolumeCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "institutional", "--date", "2025-01-15", "--tickers", "AAPL,MSFT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotTickers != "AAPL,MSFT" {
		t.Errorf("Tickers = %q, want %q", gotTickers, "AAPL,MSFT")
	}
}

func TestVolumePositionalTickers(t *testing.T) {
	t.Parallel()

	var gotTickers string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		gotTickers = form.Get("Tickers")
		fmt.Fprint(w, testutil.DataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewVolumeCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "institutional", "XLE", "XLK", "--date", "2025-01-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotTickers != "XLE,XLK" {
		t.Errorf("Tickers = %q, want %q", gotTickers, "XLE,XLK")
	}
}

func TestVolumeServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	cmd := NewVolumeCommand()
	_, _, err := testutil.ExecuteCommand(t, cmd, ctx, "institutional", "--date", "2025-01-15")
	testutil.AssertErrContains(t, err, "query volume data")
}
