package models

// Earnings represents an upcoming or recent earnings row with related trade activity counts.
type Earnings struct {
	Ticker                string     `json:"Ticker"`
	Name                  string     `json:"Name"`
	Sector                *string    `json:"Sector"`
	Industry              *string    `json:"Industry"`
	EarningsDate          AspNetDate `json:"EarningsDate"`
	AfterMarketClose      bool       `json:"AfterMarketClose"`
	TradeCount            int        `json:"TradeCount"`
	TradeClusterCount     int        `json:"TradeClusterCount"`
	TradeClusterBombCount int        `json:"TradeClusterBombCount"`
}
