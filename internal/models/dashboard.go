package models

// TradeDashboardDateRange describes the inclusive date window used by a ticker
// dashboard query.
type TradeDashboardDateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// TradeDashboard collects the fastest chart-style institutional context for a
// single ticker. It mirrors the browser dashboard sections while adding cluster
// bombs, which the VolumeLeaders page does not show in the same view.
type TradeDashboard struct {
	Ticker       string                  `json:"ticker"`
	DateRange    TradeDashboardDateRange `json:"dateRange"`
	Count        int                     `json:"count"`
	Trades       []TradeListRow          `json:"trades"`
	Clusters     []TradeCluster          `json:"clusters"`
	Levels       []TradeLevelRow         `json:"levels"`
	ClusterBombs []TradeClusterBomb      `json:"clusterBombs"`
}
