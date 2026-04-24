package models

// WatchListTicker represents a ticker row returned for a selected watch list.
type WatchListTicker struct {
	Ticker                      string     `json:"Ticker"`
	Price                       float64    `json:"Price"`
	NearestTop10TradeDate       AspNetDate `json:"NearestTop10TradeDate"`
	NearestTop10TradeClusterDate AspNetDate `json:"NearestTop10TradeClusterDate"`
	NearestTop10TradeLevel      *float64   `json:"NearestTop10TradeLevel"`
}

// WatchListConfig represents a saved watch list search template and filters.
type WatchListConfig struct {
	SearchTemplateKey        int     `json:"SearchTemplateKey"`
	UserKey                  int     `json:"UserKey"`
	SearchTemplateTypeKey    int     `json:"SearchTemplateTypeKey"`
	Name                     string  `json:"Name"`
	Tickers                  string  `json:"Tickers"`
	SortOrder                *int    `json:"SortOrder"`
	MinVolume                int     `json:"MinVolume"`
	MaxVolume                int     `json:"MaxVolume"`
	MinDollars               float64 `json:"MinDollars"`
	MaxDollars               float64 `json:"MaxDollars"`
	MinPrice                 float64 `json:"MinPrice"`
	MaxPrice                 float64 `json:"MaxPrice"`
	RSIOverboughtHourly      *int    `json:"RSIOverboughtHourly"`
	RSIOverboughtDaily       *int    `json:"RSIOverboughtDaily"`
	RSIOversoldHourly        *int    `json:"RSIOversoldHourly"`
	RSIOversoldDaily         *int    `json:"RSIOversoldDaily"`
	Conditions               string  `json:"Conditions"`
	RSIOverboughtHourlySelected *bool `json:"RSIOverboughtHourlySelected"`
	RSIOverboughtDailySelected  *bool `json:"RSIOverboughtDailySelected"`
	RSIOversoldHourlySelected   *bool `json:"RSIOversoldHourlySelected"`
	RSIOversoldDailySelected    *bool `json:"RSIOversoldDailySelected"`
	MinRelativeSize          int     `json:"MinRelativeSize"`
	MinRelativeSizeSelected  *bool   `json:"MinRelativeSizeSelected"`
	MaxTradeRank             int     `json:"MaxTradeRank"`
	SecurityTypeKey          int     `json:"SecurityTypeKey"`
	SecurityType             *string `json:"SecurityType"`
	MaxTradeRankSelected     *bool   `json:"MaxTradeRankSelected"`
	MinVCD                   float64 `json:"MinVCD"`
	NormalPrints             bool    `json:"NormalPrints"`
	SignaturePrints          bool    `json:"SignaturePrints"`
	LatePrints               bool    `json:"LatePrints"`
	TimelyPrints             bool    `json:"TimelyPrints"`
	DarkPools                bool    `json:"DarkPools"`
	LitExchanges             bool    `json:"LitExchanges"`
	Sweeps                   bool    `json:"Sweeps"`
	Blocks                   bool    `json:"Blocks"`
	PremarketTrades          bool    `json:"PremarketTrades"`
	RTHTrades                bool    `json:"RTHTrades"`
	AHTrades                 bool    `json:"AHTrades"`
	OpeningTrades            bool    `json:"OpeningTrades"`
	ClosingTrades            bool    `json:"ClosingTrades"`
	PhantomTrades            bool    `json:"PhantomTrades"`
	OffsettingTrades         bool    `json:"OffsettingTrades"`
	NormalPrintsSelected     bool    `json:"NormalPrintsSelected"`
	SignaturePrintsSelected  bool    `json:"SignaturePrintsSelected"`
	LatePrintsSelected       bool    `json:"LatePrintsSelected"`
	TimelyPrintsSelected     bool    `json:"TimelyPrintsSelected"`
	DarkPoolsSelected        bool    `json:"DarkPoolsSelected"`
	LitExchangesSelected     bool    `json:"LitExchangesSelected"`
	SweepsSelected           bool    `json:"SweepsSelected"`
	BlocksSelected           bool    `json:"BlocksSelected"`
	PremarketTradesSelected  bool    `json:"PremarketTradesSelected"`
	RTHTradesSelected        bool    `json:"RTHTradesSelected"`
	AHTradesSelected         bool    `json:"AHTradesSelected"`
	OpeningTradesSelected    bool    `json:"OpeningTradesSelected"`
	ClosingTradesSelected    bool    `json:"ClosingTradesSelected"`
	PhantomTradesSelected    bool    `json:"PhantomTradesSelected"`
	OffsettingTradesSelected bool    `json:"OffsettingTradesSelected"`
	SectorIndustry           *string `json:"SectorIndustry"`
	APIKey                   *string `json:"APIKey"`
}
