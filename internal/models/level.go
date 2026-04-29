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

// TradeLevelRow is the compact default JSON shape for trade levels output.
// The command already scopes results to one ticker, so per-row ticker and
// company metadata stay available through explicit --fields requests instead
// of repeating on every default row.
type TradeLevelRow struct {
	Price                  float64    `json:"Price"`
	Dollars                float64    `json:"Dollars"`
	Volume                 int        `json:"Volume"`
	Trades                 int        `json:"Trades"`
	RelativeSize           float64    `json:"RelativeSize"`
	CumulativeDistribution float64    `json:"CumulativeDistribution"`
	TradeLevelRank         int        `json:"TradeLevelRank"`
	MinDate                AspNetDate `json:"MinDate"`
	MaxDate                AspNetDate `json:"MaxDate"`
}

// NewTradeLevelRows projects full API level rows into the compact default
// shape used by trade levels JSON output.
func NewTradeLevelRows(levels []TradeLevel) []TradeLevelRow {
	rows := make([]TradeLevelRow, 0, len(levels))
	for i := range levels {
		level := &levels[i]
		rows = append(rows, TradeLevelRow{
			Price:                  level.Price,
			Dollars:                level.Dollars,
			Volume:                 level.Volume,
			Trades:                 level.Trades,
			RelativeSize:           level.RelativeSize,
			CumulativeDistribution: level.CumulativeDistribution,
			TradeLevelRank:         level.TradeLevelRank,
			MinDate:                level.MinDate,
			MaxDate:                level.MaxDate,
		})
	}

	return rows
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
