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
	Ticker            string                  `json:"ticker"`
	DateRange         TradeDashboardDateRange `json:"dateRange"`
	Count             int                     `json:"count"`
	RequestedSections []string                `json:"requestedSections"`
	ReturnedSections  []string                `json:"returnedSections"`
	Trades            []TradeListRow          `json:"trades"`
	Clusters          []TradeClusterRow       `json:"clusters"`
	Levels            []TradeLevelRow         `json:"levels"`
	ClusterBombs      []TradeClusterBombRow   `json:"clusterBombs"`
}

// TradeDashboardSummary is the compact first-pass dashboard shape for agents.
type TradeDashboardSummary struct {
	Ticker            string                       `json:"ticker"`
	DateRange         TradeDashboardDateRange      `json:"dateRange"`
	Count             int                          `json:"count"`
	RequestedSections []string                     `json:"requestedSections"`
	ReturnedSections  []string                     `json:"returnedSections"`
	Trades            []TradeDashboardTradeSummary `json:"trades"`
	Clusters          []TradeDashboardCluster      `json:"clusters"`
	Levels            []TradeLevelRow              `json:"levels"`
	ClusterBombs      []TradeDashboardClusterBomb  `json:"clusterBombs"`
}

// TradeDashboardTradeSummary is a compact top trade row for dashboard summary output.
type TradeDashboardTradeSummary struct {
	Date              AspNetDate `json:"Date"`
	FullDateTime      *string    `json:"FullDateTime"`
	FullTimeString24  *string    `json:"FullTimeString24"`
	Price             float64    `json:"Price"`
	Volume            int        `json:"Volume"`
	Dollars           float64    `json:"Dollars"`
	DollarsMultiplier float64    `json:"DollarsMultiplier"`
	TradeRank         int        `json:"TradeRank"`
	DarkPool          FlexBool   `json:"DarkPool"`
	Sweep             FlexBool   `json:"Sweep"`
	ClosingTrade      FlexBool   `json:"ClosingTrade"`
	TradeConditions   *string    `json:"TradeConditions"`
}

// TradeDashboardCluster is a compact top cluster row for dashboard summary output.
type TradeDashboardCluster struct {
	MinFullDateTime        string  `json:"MinFullDateTime"`
	MaxFullDateTime        string  `json:"MaxFullDateTime"`
	MinFullTimeString24    string  `json:"MinFullTimeString24"`
	MaxFullTimeString24    string  `json:"MaxFullTimeString24"`
	Price                  float64 `json:"Price"`
	Dollars                float64 `json:"Dollars"`
	Volume                 int     `json:"Volume"`
	TradeCount             int     `json:"TradeCount"`
	DollarsMultiplier      float64 `json:"DollarsMultiplier"`
	CumulativeDistribution float64 `json:"CumulativeDistribution"`
	TradeClusterRank       int     `json:"TradeClusterRank"`
}

// TradeDashboardClusterBomb is a compact top cluster-bomb row for dashboard summary output.
type TradeDashboardClusterBomb struct {
	MinFullDateTime        string  `json:"MinFullDateTime"`
	MaxFullDateTime        string  `json:"MaxFullDateTime"`
	MinFullTimeString24    string  `json:"MinFullTimeString24"`
	MaxFullTimeString24    string  `json:"MaxFullTimeString24"`
	Dollars                float64 `json:"Dollars"`
	Volume                 int     `json:"Volume"`
	TradeCount             int     `json:"TradeCount"`
	DollarsMultiplier      float64 `json:"DollarsMultiplier"`
	CumulativeDistribution float64 `json:"CumulativeDistribution"`
	TradeClusterBombRank   int     `json:"TradeClusterBombRank"`
}

// NewTradeDashboardTradeSummaries projects full trade rows into compact dashboard summary rows.
func NewTradeDashboardTradeSummaries(trades []Trade) []TradeDashboardTradeSummary {
	rows := make([]TradeDashboardTradeSummary, 0, len(trades))
	for i := range trades {
		trade := &trades[i]
		rows = append(rows, TradeDashboardTradeSummary{
			Date:              trade.Date,
			FullDateTime:      trade.FullDateTime,
			FullTimeString24:  trade.FullTimeString24,
			Price:             trade.Price,
			Volume:            trade.Volume,
			Dollars:           trade.Dollars,
			DollarsMultiplier: trade.DollarsMultiplier,
			TradeRank:         trade.TradeRank,
			DarkPool:          trade.DarkPool,
			Sweep:             trade.Sweep,
			ClosingTrade:      trade.ClosingTrade,
			TradeConditions:   trade.TradeConditions,
		})
	}
	return rows
}

// NewTradeDashboardClusters projects full cluster rows into compact dashboard summary rows.
func NewTradeDashboardClusters(clusters []TradeCluster) []TradeDashboardCluster {
	rows := make([]TradeDashboardCluster, 0, len(clusters))
	for i := range clusters {
		cluster := &clusters[i]
		rows = append(rows, TradeDashboardCluster{
			MinFullDateTime:        cluster.MinFullDateTime,
			MaxFullDateTime:        cluster.MaxFullDateTime,
			MinFullTimeString24:    cluster.MinFullTimeString24,
			MaxFullTimeString24:    cluster.MaxFullTimeString24,
			Price:                  cluster.Price,
			Dollars:                cluster.Dollars,
			Volume:                 cluster.Volume,
			TradeCount:             cluster.TradeCount,
			DollarsMultiplier:      cluster.DollarsMultiplier,
			CumulativeDistribution: cluster.CumulativeDistribution,
			TradeClusterRank:       cluster.TradeClusterRank,
		})
	}
	return rows
}

// NewTradeDashboardClusterBombs projects full cluster-bomb rows into compact dashboard summary rows.
func NewTradeDashboardClusterBombs(clusterBombs []TradeClusterBomb) []TradeDashboardClusterBomb {
	rows := make([]TradeDashboardClusterBomb, 0, len(clusterBombs))
	for i := range clusterBombs {
		clusterBomb := &clusterBombs[i]
		rows = append(rows, TradeDashboardClusterBomb{
			MinFullDateTime:        clusterBomb.MinFullDateTime,
			MaxFullDateTime:        clusterBomb.MaxFullDateTime,
			MinFullTimeString24:    clusterBomb.MinFullTimeString24,
			MaxFullTimeString24:    clusterBomb.MaxFullTimeString24,
			Dollars:                clusterBomb.Dollars,
			Volume:                 clusterBomb.Volume,
			TradeCount:             clusterBomb.TradeCount,
			DollarsMultiplier:      clusterBomb.DollarsMultiplier,
			CumulativeDistribution: clusterBomb.CumulativeDistribution,
			TradeClusterBombRank:   clusterBomb.TradeClusterBombRank,
		})
	}
	return rows
}
