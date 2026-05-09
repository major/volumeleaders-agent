package models

import (
	"encoding/json"
	"testing"
)

func TestNewTradeDashboardHoistsMetadataAndStripsRowIdentity(t *testing.T) {
	industry := "Software"
	dashboard := NewTradeDashboard(
		"IGV",
		TradeDashboardDateRange{Start: "2025-05-04", End: "2026-05-04"},
		10,
		[]string{"trades", "clusters", "clusterBombs"},
		[]string{"trades", "clusters", "clusterBombs"},
		[]Trade{{Ticker: "IGV", Name: "iShares Expanded Tech-Software Sector ETF", Sector: "Technology", Industry: &industry, Dollars: 1000, DollarsMultiplier: 1.234, Volume: 10}},
		[]TradeCluster{{Ticker: "IGV", Name: "iShares Expanded Tech-Software Sector ETF", Sector: "Technology", Industry: &industry, Dollars: 2000, DollarsMultiplier: 2.345, Volume: 20, TradeCount: 2}},
		nil,
		[]TradeClusterBomb{{Ticker: "IGV", Name: "iShares Expanded Tech-Software Sector ETF", Sector: "Technology", Industry: &industry, Dollars: 3000, DollarsMultiplier: 3.456, Volume: 30, TradeCount: 3}},
	)

	if dashboard.Name != "iShares Expanded Tech-Software Sector ETF" || dashboard.Sector != "Technology" || dashboard.Industry == nil || *dashboard.Industry != "Software" {
		t.Fatalf("NewTradeDashboard metadata = %#v, want hoisted IGV metadata", dashboard)
	}
	if dashboard.Trades[0].DollarsMultiplier != 1.23 || dashboard.Clusters[0].DollarsMultiplier != 2.35 || dashboard.ClusterBombs[0].DollarsMultiplier != 3.46 {
		t.Fatalf("NewTradeDashboard dollars multipliers = (%v, %v, %v), want (1.23, 2.35, 3.46)", dashboard.Trades[0].DollarsMultiplier, dashboard.Clusters[0].DollarsMultiplier, dashboard.ClusterBombs[0].DollarsMultiplier)
	}

	encoded, err := json.Marshal(dashboard)
	if err != nil {
		t.Fatalf("json.Marshal(NewTradeDashboard) error = %v", err)
	}
	assertDashboardRowsOmitIdentityFields(t, encoded, "trades")
	assertDashboardRowsOmitIdentityFields(t, encoded, "clusters")
	assertDashboardRowsOmitIdentityFields(t, encoded, "clusterBombs")
}

func TestTradeDashboardOmitsEmptyClusterBombs(t *testing.T) {
	dashboard := NewTradeDashboard("IGV", TradeDashboardDateRange{Start: "2025-05-04", End: "2026-05-04"}, 10, []string{"trades"}, []string{"trades"}, nil, nil, nil, nil)
	encoded, err := json.Marshal(dashboard)
	if err != nil {
		t.Fatalf("json.Marshal(NewTradeDashboard) error = %v", err)
	}

	var output map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &output); err != nil {
		t.Fatalf("json.Unmarshal(NewTradeDashboard) error = %v", err)
	}
	if _, ok := output["clusterBombs"]; ok {
		t.Fatalf("NewTradeDashboard JSON keys include clusterBombs for empty cluster bombs: %s", string(encoded))
	}
}

func assertDashboardRowsOmitIdentityFields(t *testing.T, encoded []byte, section string) {
	t.Helper()

	var dashboard map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &dashboard); err != nil {
		t.Fatalf("json.Unmarshal(NewTradeDashboard) error = %v", err)
	}
	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(dashboard[section], &rows); err != nil {
		t.Fatalf("json.Unmarshal(NewTradeDashboard.%s) error = %v", section, err)
	}
	if len(rows) != 1 {
		t.Fatalf("NewTradeDashboard.%s row count = %d, want 1", section, len(rows))
	}
	for _, field := range []string{"Ticker", "Name", "Sector", "Industry"} {
		if _, ok := rows[0][field]; ok {
			t.Fatalf("NewTradeDashboard.%s includes identity field %q", section, field)
		}
	}
}
