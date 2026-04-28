// Package models defines the response types returned by the VolumeLeaders API,
// including trades, charts, earnings, alerts, and watchlists.
package models

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const (
	dateTimeMinEpochMillis = -62135596800000
	date1900EpochMillis    = -2208988800000
)

var aspNetDatePattern = regexp.MustCompile(`^/Date\((-?\d+)\)/$`)

// AspNetDate is a nullable time that parses ASP.NET /Date(epoch_ms)/ format.
type AspNetDate struct {
	time.Time
	Valid bool
}

// UnmarshalJSON parses ASP.NET JSON date strings into UTC times.
func (d *AspNetDate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*d = AspNetDate{}
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("unmarshal ASP.NET date string: %w", err)
	}
	if value == "" {
		*d = AspNetDate{}
		return nil
	}

	matches := aspNetDatePattern.FindStringSubmatch(value)
	if matches == nil {
		return fmt.Errorf("invalid ASP.NET date format: %q", value)
	}

	epochMillis, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return fmt.Errorf("parse ASP.NET date epoch milliseconds: %w", err)
	}
	if epochMillis == dateTimeMinEpochMillis || epochMillis == date1900EpochMillis {
		*d = AspNetDate{}
		return nil
	}

	*d = AspNetDate{Time: time.UnixMilli(epochMillis).UTC(), Valid: true}
	return nil
}

// MarshalJSON serializes valid dates as RFC3339 strings and invalid dates as null.
func (d AspNetDate) MarshalJSON() ([]byte, error) {
	if !d.Valid {
		return []byte("null"), nil
	}

	data, err := json.Marshal(d.Time.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("marshal ASP.NET date: %w", err)
	}
	return data, nil
}

// FlexBool handles JSON booleans that arrive as 0/1 integers or true/false.
// The VolumeLeaders API returns boolean fields as numeric 0 and 1.
type FlexBool bool

// UnmarshalJSON accepts JSON true, false, 0, and 1.
func (b *FlexBool) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "true", "1":
		*b = true
	case "false", "0", "null":
		*b = false
	default:
		return fmt.Errorf("cannot unmarshal %s into bool", string(data))
	}
	return nil
}

// MarshalJSON writes a standard JSON boolean.
func (b FlexBool) MarshalJSON() ([]byte, error) {
	if b {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

// Trade represents a VolumeLeaders institutional trade row.
type Trade struct {
	Date                          AspNetDate `json:"Date"`
	StartDate                     AspNetDate `json:"StartDate"`
	EndDate                       AspNetDate `json:"EndDate"`
	TD30                          AspNetDate `json:"TD30"`
	TD90                          AspNetDate `json:"TD90"`
	TD1CY                         AspNetDate `json:"TD1CY"`
	DateKey                       int        `json:"DateKey"`
	TimeKey                       int        `json:"TimeKey"`
	SecurityKey                   int        `json:"SecurityKey"`
	TradeID                       int        `json:"TradeID"`
	SequenceNumber                int        `json:"SequenceNumber"`
	EOM                           FlexBool   `json:"EOM"`
	EOQ                           FlexBool   `json:"EOQ"`
	EOY                           FlexBool   `json:"EOY"`
	OPEX                          FlexBool   `json:"OPEX"`
	VOLEX                         FlexBool   `json:"VOLEX"`
	Ticker                        string     `json:"Ticker"`
	Sector                        string     `json:"Sector"`
	Industry                      *string    `json:"Industry"`
	Name                          string     `json:"Name"`
	FullDateTime                  *string    `json:"FullDateTime"`
	FullTimeString24              *string    `json:"FullTimeString24"`
	Price                         float64    `json:"Price"`
	Bid                           float64    `json:"Bid"`
	Ask                           float64    `json:"Ask"`
	Dollars                       float64    `json:"Dollars"`
	AverageBlockSizeDollars       float64    `json:"AverageBlockSizeDollars"`
	AverageBlockSizeShares        int        `json:"AverageBlockSizeShares"`
	DollarsMultiplier             float64    `json:"DollarsMultiplier"`
	Volume                        int        `json:"Volume"`
	AverageDailyVolume            int        `json:"AverageDailyVolume"`
	PercentDailyVolume            float64    `json:"PercentDailyVolume"`
	RelativeSize                  float64    `json:"RelativeSize"`
	LastComparibleTradeDate       AspNetDate `json:"LastComparibleTradeDate"`
	IPODate                       AspNetDate `json:"IPODate"`
	OffsettingTradeDate           AspNetDate `json:"OffsettingTradeDate"`
	PhantomPrintFulfillmentDate   AspNetDate `json:"PhantomPrintFulfillmentDate"`
	PhantomPrintFulfillmentDays   *int       `json:"PhantomPrintFulfillmentDays"`
	TradeCount                    int        `json:"TradeCount"`
	CumulativeDistribution        float64    `json:"CumulativeDistribution"`
	TradeRank                     int        `json:"TradeRank"`
	TradeRankSnapshot             int        `json:"TradeRankSnapshot"`
	LatePrint                     FlexBool   `json:"LatePrint"`
	Sweep                         FlexBool   `json:"Sweep"`
	DarkPool                      FlexBool   `json:"DarkPool"`
	OpeningTrade                  FlexBool   `json:"OpeningTrade"`
	ClosingTrade                  FlexBool   `json:"ClosingTrade"`
	PhantomPrint                  FlexBool   `json:"PhantomPrint"`
	InsideBar                     FlexBool   `json:"InsideBar"`
	DoubleInsideBar               FlexBool   `json:"DoubleInsideBar"`
	SignaturePrint                FlexBool   `json:"SignaturePrint"`
	NewPosition                   FlexBool   `json:"NewPosition"`
	AHInstitutionalDollars        float64    `json:"AHInstitutionalDollars"`
	AHInstitutionalDollarsRank    int        `json:"AHInstitutionalDollarsRank"`
	AHInstitutionalVolume         int        `json:"AHInstitutionalVolume"`
	TotalInstitutionalDollars     float64    `json:"TotalInstitutionalDollars"`
	TotalInstitutionalDollarsRank int        `json:"TotalInstitutionalDollarsRank"`
	TotalInstitutionalVolume      int        `json:"TotalInstitutionalVolume"`
	ClosingTradeDollars           float64    `json:"ClosingTradeDollars"`
	ClosingTradeDollarsRank       int        `json:"ClosingTradeDollarsRank"`
	ClosingTradeVolume            int        `json:"ClosingTradeVolume"`
	TotalDollars                  float64    `json:"TotalDollars"`
	TotalDollarsRank              int        `json:"TotalDollarsRank"`
	TotalVolume                   int        `json:"TotalVolume"`
	ClosePrice                    float64    `json:"ClosePrice"`
	RSIHour                       float64    `json:"RSIHour"`
	RSIDay                        float64    `json:"RSIDay"`
	TotalRows                     int        `json:"TotalRows"`
	TradeConditions               *string    `json:"TradeConditions"`
	FrequencyLast30TD             int        `json:"FrequencyLast30TD"`
	FrequencyLast90TD             int        `json:"FrequencyLast90TD"`
	FrequencyLast1CY              int        `json:"FrequencyLast1CY"`
	Cancelled                     FlexBool   `json:"Cancelled"` //nolint:misspell // VolumeLeaders API spells this field "Cancelled".
	TotalTrades                   int        `json:"TotalTrades"`
	ExternalFeed                  FlexBool   `json:"ExternalFeed"`
}

// TradeSummary represents aggregate trade metrics for a trade list query.
type TradeSummary struct {
	TotalTrades  int                          `json:"totalTrades"`
	TotalDollars float64                      `json:"totalDollars"`
	DateRange    TradeSummaryDateRange        `json:"dateRange"`
	ByTicker     map[string]TradeGroupSummary `json:"byTicker,omitempty"`
	ByDay        map[string]TradeGroupSummary `json:"byDay,omitempty"`
	ByTickerDay  map[string]TradeGroupSummary `json:"byTickerDay,omitempty"`
}

// TradeSummaryDateRange records the CLI date range used for a summary query.
type TradeSummaryDateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// TradeGroupSummary represents aggregate metrics for one summary group.
type TradeGroupSummary struct {
	Trades                    int     `json:"trades"`
	Dollars                   float64 `json:"dollars"`
	AvgDollarsMultiplier      float64 `json:"avgDollarsMultiplier"`
	PctDarkPool               float64 `json:"pctDarkPool"`
	PctSweep                  float64 `json:"pctSweep"`
	AvgCumulativeDistribution float64 `json:"avgCumulativeDistribution"`
}

// TradeSentiment summarizes bull/bear leveraged ETF flow over a date range.
type TradeSentiment struct {
	DateRange TradeSentimentDateRange `json:"dateRange"`
	Daily     []TradeSentimentDay     `json:"daily"`
	Totals    TradeSentimentTotals    `json:"totals"`
}

// TradeSentimentDateRange is the CLI date range used for sentiment analysis.
type TradeSentimentDateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// TradeSentimentDay summarizes leveraged ETF flow for one trading day.
type TradeSentimentDay struct {
	Date   string               `json:"date"`
	Bear   TradeSentimentSide   `json:"bear"`
	Bull   TradeSentimentSide   `json:"bull"`
	Ratio  *float64             `json:"ratio"`
	Signal TradeSentimentSignal `json:"signal"`
}

// TradeSentimentSide summarizes one side of leveraged ETF sentiment.
type TradeSentimentSide struct {
	Trades     int      `json:"trades"`
	Dollars    float64  `json:"dollars"`
	TopTickers []string `json:"topTickers"`
}

// TradeSentimentTotals summarizes leveraged ETF flow for the whole date range.
type TradeSentimentTotals struct {
	Bear   TradeSentimentSide   `json:"bear"`
	Bull   TradeSentimentSide   `json:"bull"`
	Ratio  *float64             `json:"ratio"`
	Signal TradeSentimentSignal `json:"signal"`
}

// TradeSentimentRow is a flat representation of one day's sentiment data,
// suitable for tabular (CSV/TSV) output.
type TradeSentimentRow struct {
	Date           string               `json:"date"`
	BearTrades     int                  `json:"bear_trades"`
	BearDollars    float64              `json:"bear_dollars"`
	BearTopTickers string               `json:"bear_top_tickers"`
	BullTrades     int                  `json:"bull_trades"`
	BullDollars    float64              `json:"bull_dollars"`
	BullTopTickers string               `json:"bull_top_tickers"`
	Ratio          *float64             `json:"ratio"`
	Signal         TradeSentimentSignal `json:"signal"`
}

// TradeSentimentSignal labels the bull/bear flow ratio for quick interpretation.
type TradeSentimentSignal string

const (
	TradeSentimentExtremeBear  TradeSentimentSignal = "extreme_bear"
	TradeSentimentModerateBear TradeSentimentSignal = "moderate_bear"
	TradeSentimentNeutral      TradeSentimentSignal = "neutral"
	TradeSentimentModerateBull TradeSentimentSignal = "moderate_bull"
	TradeSentimentExtremeBull  TradeSentimentSignal = "extreme_bull"
)

// DataTablesResponse represents the server-side DataTables JSON envelope.
type DataTablesResponse struct {
	Draw            int             `json:"draw"`
	RecordsTotal    int             `json:"recordsTotal"`
	RecordsFiltered int             `json:"recordsFiltered"`
	Data            json.RawMessage `json:"data"`
}
