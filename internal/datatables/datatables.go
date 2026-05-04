// Package datatables encodes server-side DataTables protocol requests
// and defines column layouts for each VolumeLeaders endpoint.
package datatables

import (
	"cmp"
	"net/url"
	"strconv"
	"strings"
)

// TradeColumns contains the DataTables column names used by the trades endpoint.
var TradeColumns = []string{
	"FullTimeString24", "FullTimeString24", "Ticker", "Current", "Trade",
	"Sector", "Industry", "Volume", "Dollars", "DollarsMultiplier",
	"CumulativeDistribution", "TradeRank", "RelativeSize",
	"LastComparibleTradeDate", "LastComparibleTradeDate",
}

// TradeChartColumns contains the compact chart DataTables layout VolumeLeaders
// uses for long-period ticker trade lookups. This shape avoids asking the
// backend to scan and return every matching trade before the CLI can show the
// high-value rows users usually need from longer ticker windows.
var TradeChartColumns = []Column{
	{Data: "FullTimeString24", Name: "FullTimeString24", Searchable: true, Orderable: false},
	{Data: "Volume", Name: "Sh", Searchable: true, Orderable: false},
	{Data: "Price", Name: "Price", Searchable: true, Orderable: false},
	{Data: "Dollars", Name: "$$", Searchable: true, Orderable: false},
	{Data: "DollarsMultiplier", Name: "RS", Searchable: true, Orderable: false},
	{Data: "TradeRank", Name: "R", Searchable: true, Orderable: false},
	{Data: "LastComparibleTradeDate", Name: "Last Comp", Searchable: true, Orderable: false},
}

// TradeClusterColumns contains the DataTables column names used by the trade clusters endpoint.
var TradeClusterColumns = []string{
	"MinFullTimeString24", "MinFullTimeString24", "Ticker", "TradeCount",
	"Current", "Cluster", "Sector", "Industry", "Volume", "Dollars",
	"DollarsMultiplier", "CumulativeDistribution", "TradeClusterRank",
	"LastComparibleTradeClusterDate", "LastComparibleTradeClusterDate",
}

// TradeClusterBombColumns contains the DataTables column names used by the trade cluster bombs endpoint.
var TradeClusterBombColumns = []string{
	"MinFullTimeString24", "MinFullTimeString24", "Ticker", "TradeCount",
	"Sector", "Industry", "Volume", "Dollars", "DollarsMultiplier",
	"CumulativeDistribution", "TradeClusterBombRank",
	"LastComparableTradeClusterBombDate", "LastComparableTradeClusterBombDate",
}

// InstitutionalVolumeColumns contains the DataTables column names used by the institutional volume endpoint.
var InstitutionalVolumeColumns = []string{
	"Ticker", "Ticker", "Price", "Sector", "Industry", "Volume",
	"TotalInstitutionalDollars", "TotalInstitutionalDollarsRank",
	"LastComparibleTradeDate", "LastComparibleTradeDate",
}

// TotalVolumeColumns contains the DataTables column names used by the total volume endpoint.
var TotalVolumeColumns = []string{
	"Ticker", "Ticker", "Price", "Sector", "Industry", "Volume",
	"TotalDollars", "TotalDollarsRank",
	"LastComparibleTradeDate", "LastComparibleTradeDate",
}

// TradeLevelColumns contains the DataTables column names used by the trade levels endpoint.
var TradeLevelColumns = []string{
	"Price", "Dollars", "Volume", "Trades", "RelativeSize",
	"CumulativeDistribution", "TradeLevelRank", "Level Date Range",
}

// TradeLevelTouchColumns contains the DataTables column names used by the trade level touches endpoint.
var TradeLevelTouchColumns = []string{
	"FullDateTime", "Ticker", "Sector", "Industry", "Dollars",
	"Volume", "Trades", "Price", "RelativeSize",
	"CumulativeDistribution", "TradeLevelRank", "TradeLevelTouches", "Dates",
}

// AlertConfigColumns contains the DataTables column names used by the alert configs endpoint.
var AlertConfigColumns = []string{
	"Name", "Name", "Tickers", "Conditions",
}

// EarningsColumns contains the DataTables column names used by the earnings endpoint.
var EarningsColumns = []string{
	"Date", "Ticker", "Current", "Sector", "Industry",
	"TradeCount", "TradeClusterCount", "TradeClusterBombCount", "Ticker",
}

// WatchlistTickerColumns contains the DataTables column names used by the watchlist tickers endpoint.
var WatchlistTickerColumns = []string{
	"Ticker", "Price", "NearestTop10TradeDate",
	"NearestTop10TradeClusterDate", "NearestTop10TradeLevel", "Charts",
}

// WatchlistConfigColumns contains the DataTables column names used by the watchlist configs endpoint.
var WatchlistConfigColumns = []string{
	"Name", "Name", "Tickers", "Criteria",
}

// Column describes one DataTables column definition.
type Column struct {
	Data       string
	Name       string
	Searchable bool
	Orderable  bool
}

// Request describes a server-side DataTables form request.
type Request struct {
	Columns          []string
	ColumnDefs       []Column
	Start            int
	Length           int
	OrderColumnIndex int
	OrderDirection   string
	OrderName        string
	IncludeSearch    bool
	SearchValue      string
	SearchRegex      bool
	CustomFilters    map[string]string
	Draw             int
}

type pair struct {
	key   string
	value string
}

// Encode returns the URL-encoded form body in DataTables key order.
func (r *Request) Encode() string {
	draw := cmp.Or(r.Draw, 1)
	orderDirection := cmp.Or(r.OrderDirection, "desc")

	pairs := []pair{{"draw", strconv.Itoa(draw)}, {"start", strconv.Itoa(r.Start)}, {"length", strconv.Itoa(r.Length)}, {"order[0][column]", strconv.Itoa(r.OrderColumnIndex)}, {"order[0][dir]", orderDirection}}
	if r.OrderName != "" {
		pairs = append(pairs, pair{"order[0][name]", r.OrderName})
	}
	if r.IncludeSearch {
		pairs = append(pairs, pair{"search[value]", r.SearchValue}, pair{"search[regex]", strconv.FormatBool(r.SearchRegex)})
	}
	for index, column := range r.columnDefinitions() {
		prefix := "columns[" + strconv.Itoa(index) + "]"
		pairs = append(pairs, pair{prefix + "[data]", column.Data}, pair{prefix + "[name]", column.Name}, pair{prefix + "[searchable]", strconv.FormatBool(column.Searchable)}, pair{prefix + "[orderable]", strconv.FormatBool(column.Orderable)}, pair{prefix + "[search][value]", ""}, pair{prefix + "[search][regex]", "false"})
	}
	for key, value := range r.CustomFilters {
		pairs = append(pairs, pair{key, value})
	}

	encoded := make([]string, 0, len(pairs))
	for _, item := range pairs {
		encoded = append(encoded, url.QueryEscape(item.key)+"="+url.QueryEscape(item.value))
	}
	return strings.Join(encoded, "&")
}

func (r *Request) columnDefinitions() []Column {
	if len(r.ColumnDefs) > 0 {
		return r.ColumnDefs
	}
	columns := make([]Column, 0, len(r.Columns))
	for _, column := range r.Columns {
		columns = append(columns, Column{Data: column, Name: column, Searchable: true, Orderable: true})
	}
	return columns
}
