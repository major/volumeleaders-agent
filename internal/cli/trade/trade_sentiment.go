package trade

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/client"
	"github.com/major/volumeleaders-agent/internal/datatables"
	"github.com/major/volumeleaders-agent/internal/models"
)

func tradesOptionsFromSentimentOptions(opts *tradeSentimentOptions, startDate, endDate string) *tradesOptions {
	return &tradesOptions{startDate: startDate, endDate: endDate, minVolume: opts.MinVolume, maxVolume: opts.MaxVolume, minPrice: opts.MinPrice, maxPrice: opts.MaxPrice, minDollars: opts.MinDollars, maxDollars: opts.MaxDollars, conditions: opts.Conditions, vcd: opts.VCD, securityType: opts.SecurityType, relativeSize: opts.RelativeSize, darkPools: opts.DarkPools.Int(), sweeps: opts.Sweeps.Int(), latePrints: opts.LatePrints.Int(), sigPrints: opts.SigPrints.Int(), evenShared: opts.EvenShared.Int(), tradeRank: opts.TradeRank, rankSnapshot: opts.RankSnapshot, marketCap: opts.MarketCap, premarket: opts.Premarket.Int(), rth: opts.RTH.Int(), ah: opts.AH.Int(), opening: opts.Opening.Int(), closing: opts.Closing.Int(), phantom: opts.Phantom.Int(), offsetting: opts.Offsetting.Int()}
}

func runTradeSentiment(cmd *cobra.Command, opts *tradeSentimentOptions) error {
	format, err := common.ParseOutputFormat(opts.Format)
	if err != nil {
		return err
	}
	startDate, endDate, err := common.ResolveDateRange(cmd, 0, true)
	if err != nil {
		return err
	}
	tradeOpts := tradesOptionsFromSentimentOptions(opts, startDate, endDate)
	tradeOpts.sector = "X B"
	filters := buildTradeFilters(tradeOpts)
	vlClient, err := common.NewCommandClient(cmd.Context())
	if err != nil {
		return err
	}
	dtOpts := common.NewDataTableOptions(common.DataTableRequestConfig{Start: 0, Length: maxTradeRequestLength, OrderCol: 1, OrderDir: common.OrderDirectionDESC, Filters: filters})
	trades, err := fetchTradeSentimentTrades(cmd, vlClient, dtOpts)
	if err != nil {
		return err
	}
	sentiment := summarizeTradeSentiment(trades, startDate, endDate)
	if format == common.OutputFormatJSON {
		return common.PrintJSON(cmd.OutOrStdout(), cmd.Context(), sentiment)
	}
	return common.PrintDataTablesResult(cmd.OutOrStdout(), cmd.Context(), flattenTradeSentiment(&sentiment), nil, format)
}

func flattenTradeSentiment(summary *models.TradeSentiment) []models.TradeSentimentRow {
	rows := make([]models.TradeSentimentRow, 0, len(summary.Daily)+1)
	for i := range summary.Daily {
		day := &summary.Daily[i]
		rows = append(rows, tradeSentimentDayRow(day))
	}
	rows = append(rows, tradeSentimentTotalsRow(&summary.Totals))
	return rows
}

func tradeSentimentDayRow(day *models.TradeSentimentDay) models.TradeSentimentRow {
	return models.TradeSentimentRow{Date: day.Date, BearTrades: day.Bear.Trades, BearDollars: day.Bear.Dollars, BearTopTickers: strings.Join(day.Bear.TopTickers, ";"), BullTrades: day.Bull.Trades, BullDollars: day.Bull.Dollars, BullTopTickers: strings.Join(day.Bull.TopTickers, ";"), Ratio: day.Ratio, Signal: day.Signal}
}

func tradeSentimentTotalsRow(totals *models.TradeSentimentTotals) models.TradeSentimentRow {
	return models.TradeSentimentRow{Date: "total", BearTrades: totals.Bear.Trades, BearDollars: totals.Bear.Dollars, BearTopTickers: strings.Join(totals.Bear.TopTickers, ";"), BullTrades: totals.Bull.Trades, BullDollars: totals.Bull.Dollars, BullTopTickers: strings.Join(totals.Bull.TopTickers, ";"), Ratio: totals.Ratio, Signal: totals.Signal}
}

func fetchTradeSentimentTrades(cmd *cobra.Command, vlClient *client.Client, opts common.DataTableOptions) ([]models.Trade, error) {
	request := common.NewDataTablesRequest(datatables.TradeColumns, opts)
	var result []models.Trade
	if err := vlClient.PostDataTables(cmd.Context(), "/Trades/GetTrades", request.Encode(), &result); err != nil {
		slog.Error("failed to query trade sentiment", "error", err)
		return nil, fmt.Errorf("query trade sentiment: %w", err)
	}
	return result, nil
}

type tradeSentimentAccumulator struct {
	trades        int
	dollars       float64
	tickerDollars map[string]float64
}

type tradeSentimentDayAccumulator struct{ bear, bull tradeSentimentAccumulator }

func summarizeTradeSentiment(trades []models.Trade, startDate, endDate string) models.TradeSentiment {
	byDay := make(map[string]*tradeSentimentDayAccumulator)
	var totals tradeSentimentDayAccumulator
	for i := range trades {
		trade := &trades[i]
		if !trade.Date.Valid {
			continue
		}
		side := classifyTradeSentimentSide(trade)
		if side == "" {
			continue
		}
		day := trade.Date.Format("2006-01-02")
		acc, ok := byDay[day]
		if !ok {
			acc = &tradeSentimentDayAccumulator{}
			byDay[day] = acc
		}
		acc.add(side, trade)
		totals.add(side, trade)
	}
	days := make([]string, 0, len(byDay))
	for day := range byDay {
		days = append(days, day)
	}
	sort.Strings(days)
	daily := make([]models.TradeSentimentDay, 0, len(days))
	for _, day := range days {
		daily = append(daily, byDay[day].summary(day))
	}
	return models.TradeSentiment{DateRange: models.TradeSentimentDateRange{Start: startDate, End: endDate}, Daily: daily, Totals: totals.summaryTotals()}
}

func (a *tradeSentimentDayAccumulator) add(side string, trade *models.Trade) {
	switch side {
	case "bear":
		a.bear.add(trade)
	case "bull":
		a.bull.add(trade)
	}
}

func (a *tradeSentimentAccumulator) add(trade *models.Trade) {
	if a.tickerDollars == nil {
		a.tickerDollars = make(map[string]float64)
	}
	a.trades++
	a.dollars += trade.Dollars
	a.tickerDollars[trade.Ticker] += trade.Dollars
}

func (a tradeSentimentDayAccumulator) summary(day string) models.TradeSentimentDay {
	ratio := tradeSentimentRatio(a.bull.dollars, a.bear.dollars)
	return models.TradeSentimentDay{Date: day, Bear: a.bear.summary(), Bull: a.bull.summary(), Ratio: ratio, Signal: tradeSentimentSignal(ratio, a.bull.dollars, a.bear.dollars)}
}

func (a tradeSentimentDayAccumulator) summaryTotals() models.TradeSentimentTotals {
	ratio := tradeSentimentRatio(a.bull.dollars, a.bear.dollars)
	return models.TradeSentimentTotals{Bear: a.bear.summary(), Bull: a.bull.summary(), Ratio: ratio, Signal: tradeSentimentSignal(ratio, a.bull.dollars, a.bear.dollars)}
}

func (a tradeSentimentAccumulator) summary() models.TradeSentimentSide {
	return models.TradeSentimentSide{Trades: a.trades, Dollars: a.dollars, TopTickers: topTradeSentimentTickers(a.tickerDollars, 3)}
}

func tradeSentimentRatio(bullDollars, bearDollars float64) *float64 {
	if bearDollars == 0 {
		return nil
	}
	ratio := bullDollars / bearDollars
	return &ratio
}

func tradeSentimentSignal(ratio *float64, bullDollars, bearDollars float64) models.TradeSentimentSignal {
	if ratio == nil {
		switch {
		case bullDollars > 0:
			return models.TradeSentimentExtremeBull
		case bearDollars > 0:
			return models.TradeSentimentExtremeBear
		default:
			return models.TradeSentimentNeutral
		}
	}
	switch {
	case *ratio < 0.2:
		return models.TradeSentimentExtremeBear
	case *ratio < 0.5:
		return models.TradeSentimentModerateBear
	case *ratio <= 2.0:
		return models.TradeSentimentNeutral
	case *ratio <= 5.0:
		return models.TradeSentimentModerateBull
	default:
		return models.TradeSentimentExtremeBull
	}
}

type tradeSentimentTickerTotal struct {
	ticker  string
	dollars float64
}

func topTradeSentimentTickers(tickerDollars map[string]float64, limit int) []string {
	if len(tickerDollars) == 0 {
		return []string{}
	}
	totals := make([]tradeSentimentTickerTotal, 0, len(tickerDollars))
	for ticker, dollars := range tickerDollars {
		totals = append(totals, tradeSentimentTickerTotal{ticker: ticker, dollars: dollars})
	}
	sort.Slice(totals, func(i, j int) bool {
		if totals[i].dollars == totals[j].dollars {
			return totals[i].ticker < totals[j].ticker
		}
		return totals[i].dollars > totals[j].dollars
	})
	if len(totals) < limit {
		limit = len(totals)
	}
	tickers := make([]string, 0, limit)
	for _, total := range totals[:limit] {
		tickers = append(tickers, total.ticker)
	}
	return tickers
}

func classifyTradeSentimentSide(trade *models.Trade) string {
	fields := []string{trade.Sector, trade.Name}
	if trade.Industry != nil {
		fields = append(fields, *trade.Industry)
	}
	for _, field := range fields {
		field = strings.ToLower(field)
		switch {
		case strings.Contains(field, "bear"):
			return "bear"
		case strings.Contains(field, "bull"):
			return "bull"
		}
	}
	return leveragedETFDirection(trade.Ticker)
}

func leveragedETFDirection(ticker string) string {
	switch strings.ToUpper(strings.TrimSpace(ticker)) {
	case "AAPD", "AMDD", "BERZ", "BITI", "BNKD", "BZQ", "DUST", "EDZ", "ERY", "FAZ", "HIBS", "KOLD", "LABD", "MEXZ", "MYY", "NVDD", "QID", "REK", "REW", "RXD", "SARK", "SCO", "SDD", "SDOW", "SDS", "SEF", "SH", "SMDD", "SOXS", "SPDN", "SPXU", "SPXS", "SQQQ", "SRS", "SSG", "SVIX", "TSDD", "TSLQ", "TSLS", "TZA", "UVIX", "WEBS", "YANG", "YCS", "ZSL":
		return "bear"
	case "AAPU", "AMDL", "BITU", "BOIL", "BRZU", "CURE", "CWEB", "DFEN", "DIG", "DPST", "DRN", "EDC", "ERX", "FAS", "FNGU", "GUSH", "HIBL", "LABU", "MIDU", "NAIL", "NVDL", "QLD", "ROM", "SOXL", "SPXL", "SSO", "TECL", "TMF", "TNA", "TQQQ", "TSLL", "TURB", "UDOW", "UMDD", "UPRO", "URTY", "USD", "UWM", "WEBL", "YINN":
		return "bull"
	default:
		return ""
	}
}
