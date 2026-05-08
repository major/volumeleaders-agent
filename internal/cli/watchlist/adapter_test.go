package watchlist

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	vlgo "github.com/major/volumeleaders-go/volumeleaders"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/testutil"
)

func TestVolumeLeadersWatchlistServiceConfigsFetchesAllPages(t *testing.T) {
	t.Parallel()

	var starts []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/WatchListConfigs/GetWatchLists" {
			t.Errorf("expected path /WatchListConfigs/GetWatchLists, got %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse configs request form: %v", err)
		}
		starts = append(starts, r.PostFormValue("start"))
		if got := r.PostFormValue("length"); got != strconv.Itoa(common.PaginationPageSize) {
			t.Errorf("configs request length = %q, want %d", got, common.PaginationPageSize)
		}
		if r.PostFormValue("start") == "0" {
			fmt.Fprint(w, testutil.DataTablesJSONPage(configRows(1, common.PaginationPageSize), common.PaginationPageSize+1))
			return
		}
		fmt.Fprint(w, testutil.DataTablesJSONPage(configRows(common.PaginationPageSize+1, 1), common.PaginationPageSize+1))
	}))
	t.Cleanup(server.Close)

	ctx := testutil.ContextWithTestClient(t, server.URL)
	service, err := newWatchlistService(ctx)
	if err != nil {
		t.Fatalf("newWatchlistService() error = %v", err)
	}
	configs, err := service.Configs(ctx, common.DataTableOptions{Start: 0, Length: -1, OrderCol: 1, OrderDir: "asc"})
	if err != nil {
		t.Fatalf("Configs() error = %v", err)
	}
	if got, want := len(configs), common.PaginationPageSize+1; got != want {
		t.Fatalf("Configs() returned %d configs, want %d", got, want)
	}
	wantStarts := []string{"0", strconv.Itoa(common.PaginationPageSize)}
	if len(starts) != len(wantStarts) {
		t.Fatalf("Configs() request starts = %v, want %v", starts, wantStarts)
	}
	for i, want := range wantStarts {
		if got := starts[i]; got != want {
			t.Errorf("Configs() request starts[%d] = %q, want %q", i, got, want)
		}
	}
	if got, want := configs[0].Name, "Config 1"; got != want {
		t.Errorf("Configs()[0].Name = %q, want %q", got, want)
	}
	if got, want := configs[common.PaginationPageSize].SearchTemplateKey, common.PaginationPageSize+1; got != want {
		t.Errorf("Configs()[%d].SearchTemplateKey = %d, want %d", common.PaginationPageSize, got, want)
	}
}

func configRows(startKey, count int) string {
	var rows strings.Builder
	rows.WriteByte('[')
	for i := range count {
		if i > 0 {
			rows.WriteByte(',')
		}
		key := startKey + i
		fmt.Fprintf(&rows, `{"SearchTemplateKey":%d,"Name":"Config %d"}`, key, key)
	}
	rows.WriteByte(']')
	return rows.String()
}

func TestMapWatchListConfigCopiesRepresentativeFields(t *testing.T) {
	t.Parallel()

	sortOrder := 3
	rsiSelected := true
	securityType := "Stocks"
	apiKey := "watchlist-api-key"
	config := mapWatchListConfig(&vlgo.WatchListConfig{
		SearchTemplateKey:           7,
		Name:                        "Core",
		SortOrder:                   &sortOrder,
		RSIOverboughtHourlySelected: &rsiSelected,
		NormalPrintsSelected:        true,
		SecurityType:                &securityType,
		APIKey:                      &apiKey,
	})

	if got, want := config.SearchTemplateKey, 7; got != want {
		t.Errorf("mapWatchListConfig().SearchTemplateKey = %d, want %d", got, want)
	}
	if config.SortOrder == nil || *config.SortOrder != sortOrder {
		t.Errorf("mapWatchListConfig().SortOrder = %v, want %d", config.SortOrder, sortOrder)
	}
	if config.RSIOverboughtHourlySelected == nil || *config.RSIOverboughtHourlySelected != rsiSelected {
		t.Errorf("mapWatchListConfig().RSIOverboughtHourlySelected = %v, want %t", config.RSIOverboughtHourlySelected, rsiSelected)
	}
	if got, want := config.NormalPrintsSelected, true; got != want {
		t.Errorf("mapWatchListConfig().NormalPrintsSelected = %t, want %t", got, want)
	}
	if config.SecurityType == nil || *config.SecurityType != securityType {
		t.Errorf("mapWatchListConfig().SecurityType = %v, want %q", config.SecurityType, securityType)
	}
	if config.APIKey == nil || *config.APIKey != apiKey {
		t.Errorf("mapWatchListConfig().APIKey = %v, want %q", config.APIKey, apiKey)
	}
}

func TestNewVolumeLeadersDataTablesRequestUsesExplicitOptions(t *testing.T) {
	t.Parallel()

	req := newVolumeLeadersDataTablesRequest(common.DataTableOptions{Start: 5, Length: 25, OrderCol: 1, OrderDir: "asc"})

	if got, want := req.Draw, 1; got != want {
		t.Errorf("newVolumeLeadersDataTablesRequest().Draw = %d, want %d", got, want)
	}
	if got, want := req.Start, 5; got != want {
		t.Errorf("newVolumeLeadersDataTablesRequest().Start = %d, want %d", got, want)
	}
	if got, want := req.Length, 25; got != want {
		t.Errorf("newVolumeLeadersDataTablesRequest().Length = %d, want %d", got, want)
	}
	if len(req.Order) != 1 {
		t.Fatalf("newVolumeLeadersDataTablesRequest().Order length = %d, want 1", len(req.Order))
	}
	if got, want := req.Order[0].Column, 1; got != want {
		t.Errorf("newVolumeLeadersDataTablesRequest().Order[0].Column = %d, want %d", got, want)
	}
	if got, want := req.Order[0].Dir, "asc"; got != want {
		t.Errorf("newVolumeLeadersDataTablesRequest().Order[0].Dir = %q, want %q", got, want)
	}
}

var _ watchlistService = (*volumeLeadersWatchlistService)(nil)
