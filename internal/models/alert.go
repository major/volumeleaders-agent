package models

// AlertConfig represents a saved alert configuration for the authenticated user.
type AlertConfig struct {
	AlertConfigKey         int      `json:"AlertConfigKey"`
	UserKey                int      `json:"UserKey"`
	Name                   string   `json:"Name"`
	Tickers                string   `json:"Tickers"`
	TradeRankLTE           *int     `json:"TradeRankLTE"`
	TradeVCDGTE            *float64 `json:"TradeVCDGTE"`
	TradeMultGTE           *float64 `json:"TradeMultGTE"`
	TradeVolumeGTE         *int     `json:"TradeVolumeGTE"`
	TradeDollarsGTE        *float64 `json:"TradeDollarsGTE"`
	TradeConditions        *string  `json:"TradeConditions"`
	TradeClusterRankLTE    *int     `json:"TradeClusterRankLTE"`
	TradeClusterVCDGTE     *float64 `json:"TradeClusterVCDGTE"`
	TradeClusterMultGTE    *float64 `json:"TradeClusterMultGTE"`
	TradeClusterVolumeGTE  *int     `json:"TradeClusterVolumeGTE"`
	TradeClusterDollarsGTE *float64 `json:"TradeClusterDollarsGTE"`
	TotalRankLTE           *int     `json:"TotalRankLTE"`
	TotalVolumeGTE         *int     `json:"TotalVolumeGTE"`
	TotalDollarsGTE        *float64 `json:"TotalDollarsGTE"`
	AHRankLTE              *int     `json:"AHRankLTE"`
	AHVolumeGTE            *int     `json:"AHVolumeGTE"`
	AHDollarsGTE           *float64 `json:"AHDollarsGTE"`
	ClosingTradeRankLTE    *int     `json:"ClosingTradeRankLTE"`
	ClosingTradeVCDGTE     *float64 `json:"ClosingTradeVCDGTE"`
	ClosingTradeMultGTE    *float64 `json:"ClosingTradeMultGTE"`
	ClosingTradeVolumeGTE  *int     `json:"ClosingTradeVolumeGTE"`
	ClosingTradeDollarsGTE *float64 `json:"ClosingTradeDollarsGTE"`
	ClosingTradeConditions *string  `json:"ClosingTradeConditions"`
	OffsettingPrint        bool     `json:"OffsettingPrint"`
	PhantomPrint           bool     `json:"PhantomPrint"`
	Sweep                  bool     `json:"Sweep"`
	DarkPool               bool     `json:"DarkPool"`
}

// TradeAlert represents a trade alert row with trade details and alert metadata.
type TradeAlert struct {
	Date                         AspNetDate `json:"Date"`
	StartDate                    AspNetDate `json:"StartDate"`
	EndDate                      AspNetDate `json:"EndDate"`
	FullTimeString24             string     `json:"FullTimeString24"`
	DateKey                      int        `json:"DateKey"`
	SecurityKey                  int        `json:"SecurityKey"`
	TimeKey                      int        `json:"TimeKey"`
	TradeID                      int        `json:"TradeID"`
	SequenceNumber               int        `json:"SequenceNumber"`
	UserKey                      int        `json:"UserKey"`
	UserKeys                     *string    `json:"UserKeys"`
	Sent                         bool       `json:"Sent"`
	Email                        *string    `json:"Email"`
	Emails                       *string    `json:"Emails"`
	Ticker                       string     `json:"Ticker"`
	Sector                       *string    `json:"Sector"`
	Industry                     *string    `json:"Industry"`
	Name                         string     `json:"Name"`
	AlertType                    string     `json:"AlertType"`
	Price                        float64    `json:"Price"`
	TradeRank                    int        `json:"TradeRank"`
	VolumeCumulativeDistribution float64    `json:"VolumeCumulativeDistribution"`
	DollarsMultiplier            float64    `json:"DollarsMultiplier"`
	Volume                       int        `json:"Volume"`
	Dollars                      float64    `json:"Dollars"`
	LastComparibleTradeDateKey   int        `json:"LastComparibleTradeDateKey"`
	LastComparibleTradeDate      AspNetDate `json:"LastComparibleTradeDate"`
	OffsettingTradeDate          AspNetDate `json:"OffsettingTradeDate"`
	PhantomPrintFulfillmentDate  AspNetDate `json:"PhantomPrintFulfillmentDate"`
	FullDateTime                 string     `json:"FullDateTime"`
	IPODate                      AspNetDate `json:"IPODate"`
	RSIHour                      float64    `json:"RSIHour"`
	RSIDay                       float64    `json:"RSIDay"`
	InProcess                    bool       `json:"InProcess"`
	Complete                     bool       `json:"Complete"`
	Sweep                        FlexBool   `json:"Sweep"`
	DarkPool                     FlexBool   `json:"DarkPool"`
	LatePrint                    FlexBool   `json:"LatePrint"`
	ClosingTrade                 FlexBool   `json:"ClosingTrade"`
	SignaturePrint               FlexBool   `json:"SignaturePrint"`
	PhantomPrint                 FlexBool   `json:"PhantomPrint"`
}

// TradeAlertRow is the compact default JSON shape for trade alert output.
type TradeAlertRow struct {
	Date                         AspNetDate `json:"Date"`
	StartDate                    AspNetDate `json:"StartDate"`
	EndDate                      AspNetDate `json:"EndDate"`
	DateKey                      int        `json:"DateKey"`
	SecurityKey                  int        `json:"SecurityKey"`
	TimeKey                      int        `json:"TimeKey"`
	TradeID                      int        `json:"TradeID"`
	SequenceNumber               int        `json:"SequenceNumber"`
	UserKey                      int        `json:"UserKey"`
	UserKeys                     *string    `json:"UserKeys"`
	Sent                         bool       `json:"Sent"`
	Email                        *string    `json:"Email"`
	Emails                       *string    `json:"Emails"`
	Ticker                       string     `json:"Ticker"`
	Sector                       *string    `json:"Sector"`
	Industry                     *string    `json:"Industry"`
	Name                         string     `json:"Name"`
	AlertType                    string     `json:"AlertType"`
	Price                        float64    `json:"Price"`
	TradeRank                    int        `json:"TradeRank"`
	VolumeCumulativeDistribution float64    `json:"VolumeCumulativeDistribution"`
	DollarsMultiplier            float64    `json:"DollarsMultiplier"`
	Volume                       int        `json:"Volume"`
	Dollars                      float64    `json:"Dollars"`
	LastComparibleTradeDateKey   int        `json:"LastComparibleTradeDateKey"`
	LastComparibleTradeDate      AspNetDate `json:"LastComparibleTradeDate"`
	OffsettingTradeDate          AspNetDate `json:"OffsettingTradeDate"`
	PhantomPrintFulfillmentDate  AspNetDate `json:"PhantomPrintFulfillmentDate"`
	FullDateTime                 string     `json:"FullDateTime"`
	IPODate                      AspNetDate `json:"IPODate"`
	InProcess                    bool       `json:"InProcess"`
	Complete                     bool       `json:"Complete"`
	Sweep                        FlexBool   `json:"Sweep"`
	DarkPool                     FlexBool   `json:"DarkPool"`
	LatePrint                    FlexBool   `json:"LatePrint"`
	ClosingTrade                 FlexBool   `json:"ClosingTrade"`
	SignaturePrint               FlexBool   `json:"SignaturePrint"`
	PhantomPrint                 FlexBool   `json:"PhantomPrint"`
}

// NewTradeAlertRows projects full API alert rows into compact output rows.
func NewTradeAlertRows(alerts []TradeAlert) []TradeAlertRow {
	rows := make([]TradeAlertRow, 0, len(alerts))
	for i := range alerts {
		rows = append(rows, NewTradeAlertRow(&alerts[i]))
	}

	return rows
}

// NewTradeAlertRow projects one full API alert row into a compact output row.
func NewTradeAlertRow(alert *TradeAlert) TradeAlertRow {
	return TradeAlertRow{
		Date:                         alert.Date,
		StartDate:                    alert.StartDate,
		EndDate:                      alert.EndDate,
		DateKey:                      alert.DateKey,
		SecurityKey:                  alert.SecurityKey,
		TimeKey:                      alert.TimeKey,
		TradeID:                      alert.TradeID,
		SequenceNumber:               alert.SequenceNumber,
		UserKey:                      alert.UserKey,
		UserKeys:                     alert.UserKeys,
		Sent:                         alert.Sent,
		Email:                        alert.Email,
		Emails:                       alert.Emails,
		Ticker:                       alert.Ticker,
		Sector:                       alert.Sector,
		Industry:                     alert.Industry,
		Name:                         alert.Name,
		AlertType:                    alert.AlertType,
		Price:                        alert.Price,
		TradeRank:                    alert.TradeRank,
		VolumeCumulativeDistribution: alert.VolumeCumulativeDistribution,
		DollarsMultiplier:            roundDollarsMultiplier(alert.DollarsMultiplier),
		Volume:                       alert.Volume,
		Dollars:                      alert.Dollars,
		LastComparibleTradeDateKey:   alert.LastComparibleTradeDateKey,
		LastComparibleTradeDate:      alert.LastComparibleTradeDate,
		OffsettingTradeDate:          alert.OffsettingTradeDate,
		PhantomPrintFulfillmentDate:  alert.PhantomPrintFulfillmentDate,
		FullDateTime:                 alert.FullDateTime,
		IPODate:                      alert.IPODate,
		InProcess:                    alert.InProcess,
		Complete:                     alert.Complete,
		Sweep:                        alert.Sweep,
		DarkPool:                     alert.DarkPool,
		LatePrint:                    alert.LatePrint,
		ClosingTrade:                 alert.ClosingTrade,
		SignaturePrint:               alert.SignaturePrint,
		PhantomPrint:                 alert.PhantomPrint,
	}
}

// TradeClusterAlert is an alias for TradeCluster; the API returns the same shape.
type TradeClusterAlert = TradeCluster
