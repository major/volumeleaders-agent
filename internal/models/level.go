package models

// TradeLevel represents a significant price level aggregate for a ticker.
type TradeLevel struct {
	Ticker                 *string    `json:"Ticker"`
	Name                   *string    `json:"Name"`
	Price                  float64    `json:"Price"`
	Dollars                float64    `json:"Dollars"`
	Volume                 int        `json:"Volume"`
	Trades                 int        `json:"Trades"`
	RelativeSize           float64    `json:"RelativeSize"`
	CumulativeDistribution float64    `json:"CumulativeDistribution"`
	TradeLevelRank         int        `json:"TradeLevelRank"`
	MinDate                AspNetDate `json:"MinDate"`
	MaxDate                AspNetDate `json:"MaxDate"`
	Dates                  string     `json:"Dates"`
}

// TradeLevelTouch represents a trade event at a notable price level.
type TradeLevelTouch struct {
	Ticker                 string     `json:"Ticker"`
	Sector                 *string    `json:"Sector"`
	Industry               *string    `json:"Industry"`
	Name                   string     `json:"Name"`
	Date                   AspNetDate `json:"Date"`
	MinDate                AspNetDate `json:"MinDate"`
	MaxDate                AspNetDate `json:"MaxDate"`
	FullDateTime           string     `json:"FullDateTime"`
	FullTimeString24       *string    `json:"FullTimeString24"`
	Dates                  string     `json:"Dates"`
	Price                  float64    `json:"Price"`
	Dollars                float64    `json:"Dollars"`
	Volume                 int        `json:"Volume"`
	Trades                 int        `json:"Trades"`
	CumulativeDistribution float64    `json:"CumulativeDistribution"`
	TradeLevelRank         int        `json:"TradeLevelRank"`
	TotalRows              int        `json:"TotalRows"`
	TradeLevelTouches      int        `json:"TradeLevelTouches"`
	RelativeSize           float64    `json:"RelativeSize"`
}
