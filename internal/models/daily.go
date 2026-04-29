package models

// DailySummary is a compact cross-endpoint snapshot for one trading day.
type DailySummary struct {
	Date                  string                     `json:"date"`
	InstitutionalVolume   []DailyInstitutionalVolume `json:"institutional_volume"`
	Clusters              DailyClusterSummary        `json:"clusters"`
	ClusterBombs          []DailyClusterBomb         `json:"cluster_bombs"`
	LevelTouches          []DailyLevelTouch          `json:"level_touches"`
	LeveragedETFSentiment DailyLeveragedETFSentiment `json:"leveraged_etf_sentiment"`
	MarketExhaustion      DailyMarketExhaustion      `json:"market_exhaustion"`
}

// DailyInstitutionalVolume summarizes high institutional volume tickers.
type DailyInstitutionalVolume struct {
	Ticker               string  `json:"ticker"`
	Sector               string  `json:"sector,omitempty"`
	Price                float64 `json:"price"`
	InstitutionalDollars float64 `json:"institutional_dollars"`
	Rank                 int     `json:"rank"`
}

// DailyClusterSummary contains compact cluster signals for the day.
type DailyClusterSummary struct {
	Top             []DailyCluster               `json:"top"`
	RepeatedTickers []DailyRepeatedClusterTicker `json:"repeated_tickers"`
}

// DailyCluster summarizes one notable trade cluster.
type DailyCluster struct {
	Ticker                 string   `json:"ticker"`
	Sector                 string   `json:"sector,omitempty"`
	TopBy                  []string `json:"top_by"`
	Dollars                float64  `json:"dollars"`
	DollarsMultiplier      float64  `json:"dollars_multiplier"`
	TradeCount             int      `json:"trade_count"`
	Rank                   int      `json:"rank"`
	CumulativeDistribution float64  `json:"cumulative_distribution"`
}

// DailyRepeatedClusterTicker records tickers with multiple clusters in the day.
type DailyRepeatedClusterTicker struct {
	Ticker        string  `json:"ticker"`
	Sector        string  `json:"sector,omitempty"`
	Clusters      int     `json:"clusters"`
	Dollars       float64 `json:"dollars"`
	TradeCount    int     `json:"trade_count"`
	MaxMultiplier float64 `json:"max_multiplier"`
	BestRank      int     `json:"best_rank"`
}

// DailyClusterBomb summarizes high-priority cluster bomb entries.
type DailyClusterBomb struct {
	Ticker                 string  `json:"ticker"`
	Sector                 string  `json:"sector,omitempty"`
	Dollars                float64 `json:"dollars"`
	DollarsMultiplier      float64 `json:"dollars_multiplier"`
	TradeCount             int     `json:"trade_count"`
	Rank                   int     `json:"rank"`
	CumulativeDistribution float64 `json:"cumulative_distribution"`
}

// DailyLevelTouch summarizes one notable level-touch event.
type DailyLevelTouch struct {
	Ticker                 string   `json:"ticker"`
	Sector                 string   `json:"sector,omitempty"`
	TopBy                  []string `json:"top_by"`
	Price                  float64  `json:"price"`
	Dollars                float64  `json:"dollars"`
	RelativeSize           float64  `json:"relative_size"`
	Trades                 int      `json:"trades"`
	Rank                   int      `json:"rank"`
	Touches                int      `json:"touches"`
	CumulativeDistribution float64  `json:"cumulative_distribution"`
}

// DailyLeveragedETFSentiment summarizes daily leveraged ETF directional proxy flow.
type DailyLeveragedETFSentiment struct {
	Signal      TradeSentimentSignal `json:"signal"`
	Ratio       *float64             `json:"ratio"`
	BullDollars float64              `json:"bull_dollars"`
	BearDollars float64              `json:"bear_dollars"`
	BullTrades  int                  `json:"bull_trades"`
	BearTrades  int                  `json:"bear_trades"`
	BullTickers []string             `json:"bull_tickers"`
	BearTickers []string             `json:"bear_tickers"`
}

// DailyMarketExhaustion keeps only the daily exhaustion ranks useful for decisions.
type DailyMarketExhaustion struct {
	Rank     int `json:"rank"`
	Rank30D  int `json:"rank_30d"`
	Rank90D  int `json:"rank_90d"`
	Rank365D int `json:"rank_365d"`
}
