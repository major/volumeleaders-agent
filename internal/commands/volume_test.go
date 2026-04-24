package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/major/volumeleaders-agent/internal/datatables"
	cli "github.com/urfave/cli/v3"
)

func TestRunVolume(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/InstitutionalVolume/GetInstitutionalVolume" {
			t.Errorf("expected path /InstitutionalVolume/GetInstitutionalVolume, got %s", r.URL.Path)
		}
		fmt.Fprint(w, dataTablesJSON(`[{}]`))
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	captureStdout(t, func() {
		opts := volumeOptions{
			date:     "2025-01-15",
			tickers:  "AAPL",
			start:    0,
			length:   100,
			orderCol: 1,
			orderDir: "asc",
		}
		if err := runVolume(ctx, opts, "/InstitutionalVolume/GetInstitutionalVolume", datatables.InstitutionalVolumeColumns); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunVolumeServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	ctx := contextWithTestClient(server.URL)
	opts := volumeOptions{date: "2025-01-15", start: 0, length: 100, orderCol: 1, orderDir: "asc"}
	err := runVolume(ctx, opts, "/InstitutionalVolume/GetInstitutionalVolume", datatables.InstitutionalVolumeColumns)
	assertErrContains(t, err, "query volume data")
}

func TestVolumeSubcommands(t *testing.T) {
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
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("expected path %s, got %s", tt.path, r.URL.Path)
				}
				fmt.Fprint(w, dataTablesJSON(`[{}]`))
			}))
			t.Cleanup(server.Close)

			ctx := contextWithTestClient(server.URL)
			captureStdout(t, func() {
				root := &cli.Command{Commands: []*cli.Command{NewVolumeCommand()}}
				if err := root.Run(ctx, []string{"app", "volume", tt.name, "--date", "2025-01-15"}); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})
		})
	}
}
