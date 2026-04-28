package models

// DailySummary is a compact cross-endpoint snapshot for one trading day.
type DailySummary struct {
	Date                          string                        `json:"date"`
	TopInstitutionalVolumeTickers []DailyInstitutionalVolumeRow `json:"top_institutional_volume_tickers"`
	TopClustersByDollars          []DailyClusterRow             `json:"top_clusters_by_dollars"`
	TopClustersByMultiplier       []DailyClusterRow             `json:"top_clusters_by_multiplier"`
	RepeatedClusterTickers        []DailyRepeatedClusterTicker  `json:"repeated_cluster_tickers"`
	SectorTotals                  []DailySectorTotal            `json:"sector_totals"`
	ClusterBombs                  []DailyClusterBombRow         `json:"cluster_bombs"`
	LevelTouches                  DailyLevelTouchSummary        `json:"level_touches"`
	LeveragedETFSentiment         DailyLeveragedETFSentiment    `json:"leveraged_etf_sentiment"`
	MarketExhaustion              ExhaustionScore               `json:"market_exhaustion"`
}

// DailyInstitutionalVolumeRow summarizes high institutional volume tickers.
type DailyInstitutionalVolumeRow struct {
	Ticker                       string  `json:"ticker"`
	Sector                       string  `json:"sector,omitempty"`
	Price                        float64 `json:"price"`
	Volume                       int     `json:"volume"`
	TotalInstitutionalDollars    float64 `json:"total_institutional_dollars"`
	TotalInstitutionalDollarRank int     `json:"total_institutional_dollar_rank"`
}

// DailyClusterRow summarizes one trade cluster in daily leaderboards.
type DailyClusterRow struct {
	Ticker                 string  `json:"ticker"`
	Sector                 string  `json:"sector,omitempty"`
	Dollars                float64 `json:"dollars"`
	DollarsMultiplier      float64 `json:"dollars_multiplier"`
	Volume                 int     `json:"volume"`
	TradeCount             int     `json:"trade_count"`
	TradeClusterRank       int     `json:"trade_cluster_rank"`
	CumulativeDistribution float64 `json:"cumulative_distribution"`
}

// DailyRepeatedClusterTicker records tickers with multiple clusters in the day.
type DailyRepeatedClusterTicker struct {
	Ticker        string  `json:"ticker"`
	Sector        string  `json:"sector,omitempty"`
	Clusters      int     `json:"clusters"`
	Dollars       float64 `json:"dollars"`
	TradeCount    int     `json:"trade_count"`
	MaxMultiplier float64 `json:"max_multiplier"`
}

// DailySectorTotal aggregates institutional activity by sector.
type DailySectorTotal struct {
	Sector     string  `json:"sector"`
	Tickers    int     `json:"tickers"`
	Trades     int     `json:"trades"`
	Dollars    float64 `json:"dollars"`
	Volume     int     `json:"volume"`
	TradeCount int     `json:"trade_count"`
}

// DailyClusterBombRow summarizes high-priority cluster bomb entries.
type DailyClusterBombRow struct {
	Ticker                 string  `json:"ticker"`
	Sector                 string  `json:"sector,omitempty"`
	Dollars                float64 `json:"dollars"`
	DollarsMultiplier      float64 `json:"dollars_multiplier"`
	Volume                 int     `json:"volume"`
	TradeCount             int     `json:"trade_count"`
	TradeClusterBombRank   int     `json:"trade_cluster_bomb_rank"`
	CumulativeDistribution float64 `json:"cumulative_distribution"`
}

// DailyLevelTouchSummary contains level-touch leaderboards using two sort keys.
type DailyLevelTouchSummary struct {
	ByRelativeSize []DailyLevelTouchRow `json:"by_relative_size"`
	ByDollars      []DailyLevelTouchRow `json:"by_dollars"`
}

// DailyLevelTouchRow summarizes one notable level-touch event.
type DailyLevelTouchRow struct {
	Ticker            string  `json:"ticker"`
	Sector            string  `json:"sector,omitempty"`
	Price             float64 `json:"price"`
	Dollars           float64 `json:"dollars"`
	RelativeSize      float64 `json:"relative_size"`
	Volume            int     `json:"volume"`
	Trades            int     `json:"trades"`
	TradeLevelRank    int     `json:"trade_level_rank"`
	TradeLevelTouches int     `json:"trade_level_touches"`
}

// DailyLeveragedETFSentiment summarizes daily leveraged ETF directional flow.
type DailyLeveragedETFSentiment struct {
	Bear   TradeSentimentSide   `json:"bear"`
	Bull   TradeSentimentSide   `json:"bull"`
	Ratio  *float64             `json:"ratio"`
	Signal TradeSentimentSignal `json:"signal"`
}
