package models

// PriceBar represents a one-minute OHLCV bar with trade metadata overlays.
type PriceBar struct {
	DateKey                  int        `json:"DateKey"`
	TimeKey                  int        `json:"TimeKey"`
	SecurityKey              int        `json:"SecurityKey"`
	TradeID                  int        `json:"TradeID"`
	Date                     AspNetDate `json:"Date"`
	DateString               *string    `json:"DateString"`
	FullDateTime             string     `json:"FullDateTime"`
	MinFullDateTime          *string    `json:"MinFullDateTime"`
	MaxFullDateTime          *string    `json:"MaxFullDateTime"`
	FullDateTimePlotted      *string    `json:"FullDateTimePlotted"`
	FullTimeString24         string     `json:"FullTimeString24"`
	Ticker                   *string    `json:"Ticker"`
	Volume                   int        `json:"Volume"`
	Dollars                  float64    `json:"Dollars"`
	OpenPrice                float64    `json:"OpenPrice"`
	ClosePrice               float64    `json:"ClosePrice"`
	HighPrice                float64    `json:"HighPrice"`
	LowPrice                 float64    `json:"LowPrice"`
	Price                    *float64   `json:"Price"`
	Trades                   int        `json:"Trades"`
	BlockSizeNTile           *float64   `json:"BlockSizeNTile"`
	CumulativeDistribution   float64    `json:"CumulativeDistribution"`
	TradeConditions          *string    `json:"TradeConditions"`
	DarkPoolTrade            bool       `json:"DarkPoolTrade"`
	LatePrint                bool       `json:"LatePrint"`
	OpeningTrade             bool       `json:"OpeningTrade"`
	ClosingTrade             bool       `json:"ClosingTrade"`
	SignaturePrint           bool       `json:"SignaturePrint"`
	PhantomPrint             bool       `json:"PhantomPrint"`
	Sweep                    bool       `json:"Sweep"`
	Dates                    *string    `json:"Dates"`
	TradeRank                int        `json:"TradeRank"`
	TradeRankSnapshot        int        `json:"TradeRankSnapshot"`
	TradeLevelRank           int        `json:"TradeLevelRank"`
	LastComparibleTradeDate  AspNetDate `json:"LastComparibleTradeDate"`
	IPODate                  AspNetDate `json:"IPODate"`
	DollarsMultiplier        float64    `json:"DollarsMultiplier"`
	SharesMultiplier         float64    `json:"SharesMultiplier"`
	RelativeSize             float64    `json:"RelativeSize"`
	TotalTrades              int        `json:"TotalTrades"`
	FrequencyLast30TD        int        `json:"FrequencyLast30TD"`
	FrequencyLast90TD        int        `json:"FrequencyLast90TD"`
	FrequencyLast1CY         int        `json:"FrequencyLast1CY"`
}

// Company holds reference metadata for a company or ETF.
type Company struct {
	SecurityKey                      int        `json:"SecurityKey"`
	SecurityTypeKey                  int        `json:"SecurityTypeKey"`
	Active                          bool       `json:"Active"`
	Valid                           bool       `json:"Valid"`
	Name                            string     `json:"Name"`
	Ticker                          string     `json:"Ticker"`
	Sector                          *string    `json:"Sector"`
	Industry                        *string    `json:"Industry"`
	Description                     *string    `json:"Description"`
	HomePageURL                     *string    `json:"HomePageURL"`
	IPODate                         AspNetDate `json:"IPODate"`
	AverageBlockSizeDollars         float64    `json:"AverageBlockSizeDollars"`
	AverageBlockSizeDollars30Days   float64    `json:"AverageBlockSizeDollars30Days"`
	AverageBlockSizeDollars90Days   float64    `json:"AverageBlockSizeDollars90Days"`
	AverageBlockSizeShares          int        `json:"AverageBlockSizeShares"`
	AverageDailyVolume              int        `json:"AverageDailyVolume"`
	AverageDailyVolume30Days        int        `json:"AverageDailyVolume30Days"`
	AverageDailyVolume90Days        int        `json:"AverageDailyVolume90Days"`
	AverageTradeShares              int        `json:"AverageTradeShares"`
	AverageTradeShares30Days        int        `json:"AverageTradeShares30Days"`
	AverageTradeShares90Days        int        `json:"AverageTradeShares90Days"`
	AverageDailyRange               float64    `json:"AverageDailyRange"`
	AverageDailyRange30Days         float64    `json:"AverageDailyRange30Days"`
	AverageDailyRange90Days         float64    `json:"AverageDailyRange90Days"`
	AverageDailyRangePct            float64    `json:"AverageDailyRangePct"`
	AverageDailyRangePct30Days      float64    `json:"AverageDailyRangePct30Days"`
	AverageDailyRangePct90Days      float64    `json:"AverageDailyRangePct90Days"`
	AverageQualifyingBlockTrades    int        `json:"AverageQualifyingBlockTrades"`
	AverageClosingTradeShares       int        `json:"AverageClosingTradeShares"`
	AverageClosingTradeDollars      float64    `json:"AverageClosingTradeDollars"`
	AverageClosingTradeDollars30Days float64   `json:"AverageClosingTradeDollars30Days"`
	AverageClosingTradeDollars90Days float64   `json:"AverageClosingTradeDollars90Days"`
	AverageClusterSizeDollars       float64    `json:"AverageClusterSizeDollars"`
	AverageLevelSizeDollars         float64    `json:"AverageLevelSizeDollars"`
	Priority                        int        `json:"Priority"`
	Free                            bool       `json:"Free"`
	TotalTrades                     int        `json:"TotalTrades"`
	FirstTradeDate                  AspNetDate `json:"FirstTradeDate"`
	PreviousTicker                  *string    `json:"PreviousTicker"`
	PreviousTickerExpirationDate    AspNetDate `json:"PreviousTickerExpirationDate"`
	News                            *string    `json:"News"`
	Financials                      *string    `json:"Financials"`
	CurrentPrice                    *float64   `json:"CurrentPrice"`
	Splits                          *string    `json:"Splits"`
	MarketCap                       float64    `json:"MarketCap"`
	MaxDate                         AspNetDate `json:"MaxDate"`
	OptionsEnabled                  bool       `json:"OptionsEnabled"`
}

// Quote represents a bid/ask quote in a snapshot response.
type Quote struct {
	Timestamp int     `json:"t"`
	BidPrice  float64 `json:"p"`
	BidSize   float64 `json:"s"`
	AskPrice  float64 `json:"P"`
	AskSize   float64 `json:"S"`
}

// LastTrade represents the most recent trade in a snapshot response.
type LastTrade struct {
	Price float64 `json:"p"`
}

// Snapshot combines quote and last-trade context for a ticker.
type Snapshot struct {
	LastQuote        Quote     `json:"lastQuote"`
	LastTrade        LastTrade `json:"lastTrade"`
	Ticker           string    `json:"ticker"`
	TodaysChange     float64   `json:"todaysChange"`
	TodaysChangePerc float64   `json:"todaysChangePerc"`
}

// SnapshotResponse is the envelope returned by the GetSnapshot endpoint.
type SnapshotResponse struct {
	Snapshot Snapshot `json:"ticker"`
	Status   string   `json:"status"`
}
