package models

// TradeDashboardDateRange describes the inclusive date window used by a ticker
// dashboard query.
type TradeDashboardDateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// TradeDashboard collects the fastest chart-style institutional context for a
// single ticker. It mirrors the browser dashboard sections while adding cluster
// bombs, which the VolumeLeaders page does not show in the same view.
type TradeDashboard struct {
	Ticker            string                         `json:"ticker"`
	Name              string                         `json:"name,omitempty"`
	Sector            string                         `json:"sector,omitempty"`
	Industry          *string                        `json:"industry,omitempty"`
	DateRange         TradeDashboardDateRange        `json:"dateRange"`
	Count             int                            `json:"count"`
	RequestedSections []string                       `json:"requestedSections"`
	ReturnedSections  []string                       `json:"returnedSections"`
	Trades            []TradeDashboardTradeRow       `json:"trades"`
	Clusters          []TradeDashboardClusterRow     `json:"clusters"`
	Levels            []TradeLevelRow                `json:"levels"`
	ClusterBombs      []TradeDashboardClusterBombRow `json:"clusterBombs,omitempty"`
}

// TradeDashboardSummary is the compact first-pass dashboard shape for agents.
type TradeDashboardSummary struct {
	Ticker            string                       `json:"ticker"`
	Name              string                       `json:"name,omitempty"`
	Sector            string                       `json:"sector,omitempty"`
	Industry          *string                      `json:"industry,omitempty"`
	DateRange         TradeDashboardDateRange      `json:"dateRange"`
	Count             int                          `json:"count"`
	RequestedSections []string                     `json:"requestedSections"`
	ReturnedSections  []string                     `json:"returnedSections"`
	Trades            []TradeDashboardTradeSummary `json:"trades"`
	Clusters          []TradeDashboardCluster      `json:"clusters"`
	Levels            []TradeLevelRow              `json:"levels"`
	ClusterBombs      []TradeDashboardClusterBomb  `json:"clusterBombs"`
}

// TradeDashboardTradeRow is the dashboard-specific trade projection.
// Dashboard output already scopes results to one ticker and hoists company
// metadata onto the envelope, so rows keep only section-specific analysis data.
type TradeDashboardTradeRow struct {
	Date                   AspNetDate `json:"Date"`
	FullDateTime           *string    `json:"FullDateTime"`
	Price                  float64    `json:"Price"`
	Volume                 int        `json:"Volume"`
	Dollars                float64    `json:"Dollars"`
	DollarsMultiplier      float64    `json:"DollarsMultiplier"`
	PercentDailyVolume     float64    `json:"PercentDailyVolume"`
	RelativeSize           float64    `json:"RelativeSize"`
	CumulativeDistribution float64    `json:"CumulativeDistribution"`
	TradeRank              int        `json:"TradeRank"`
	DarkPool               FlexBool   `json:"DarkPool"`
	Sweep                  FlexBool   `json:"Sweep"`
	LatePrint              FlexBool   `json:"LatePrint"`
	SignaturePrint         FlexBool   `json:"SignaturePrint"`
	OpeningTrade           FlexBool   `json:"OpeningTrade"`
	ClosingTrade           FlexBool   `json:"ClosingTrade"`
	PhantomPrint           FlexBool   `json:"PhantomPrint"`
}

// TradeDashboardClusterRow is the dashboard-specific cluster projection.
type TradeDashboardClusterRow struct {
	Date                    AspNetDate `json:"Date"`
	MinFullDateTime         string     `json:"MinFullDateTime"`
	MaxFullDateTime         string     `json:"MaxFullDateTime"`
	ClosePrice              float64    `json:"ClosePrice"`
	Price                   float64    `json:"Price"`
	Dollars                 float64    `json:"Dollars"`
	AverageBlockSizeShares  int        `json:"AverageBlockSizeShares"`
	AverageBlockSizeDollars float64    `json:"AverageBlockSizeDollars"`
	Volume                  int        `json:"Volume"`
	TradeCount              int        `json:"TradeCount"`
	DollarsMultiplier       float64    `json:"DollarsMultiplier"`
	CumulativeDistribution  float64    `json:"CumulativeDistribution"`
	AverageDailyVolume      int        `json:"AverageDailyVolume"`
	TradeClusterRank        int        `json:"TradeClusterRank"`
}

// TradeDashboardClusterBombRow is the dashboard-specific cluster-bomb projection.
type TradeDashboardClusterBombRow struct {
	Date                    AspNetDate `json:"Date"`
	MinFullDateTime         string     `json:"MinFullDateTime"`
	MaxFullDateTime         string     `json:"MaxFullDateTime"`
	ClosePrice              float64    `json:"ClosePrice"`
	Dollars                 float64    `json:"Dollars"`
	AverageBlockSizeShares  int        `json:"AverageBlockSizeShares"`
	AverageBlockSizeDollars float64    `json:"AverageBlockSizeDollars"`
	Volume                  int        `json:"Volume"`
	TradeCount              int        `json:"TradeCount"`
	DollarsMultiplier       float64    `json:"DollarsMultiplier"`
	CumulativeDistribution  float64    `json:"CumulativeDistribution"`
	AverageDailyVolume      int        `json:"AverageDailyVolume"`
	TradeClusterBombRank    int        `json:"TradeClusterBombRank"`
}

// TradeDashboardTradeSummary is a compact top trade row for dashboard summary output.
type TradeDashboardTradeSummary struct {
	Date              AspNetDate `json:"Date"`
	FullDateTime      *string    `json:"FullDateTime"`
	FullTimeString24  *string    `json:"FullTimeString24"`
	Price             float64    `json:"Price"`
	Volume            int        `json:"Volume"`
	Dollars           float64    `json:"Dollars"`
	DollarsMultiplier float64    `json:"DollarsMultiplier"`
	TradeRank         int        `json:"TradeRank"`
	DarkPool          FlexBool   `json:"DarkPool"`
	Sweep             FlexBool   `json:"Sweep"`
	ClosingTrade      FlexBool   `json:"ClosingTrade"`
	TradeConditions   *string    `json:"TradeConditions"`
}

// TradeDashboardCluster is a compact top cluster row for dashboard summary output.
type TradeDashboardCluster struct {
	MinFullDateTime        string  `json:"MinFullDateTime"`
	MaxFullDateTime        string  `json:"MaxFullDateTime"`
	MinFullTimeString24    string  `json:"MinFullTimeString24"`
	MaxFullTimeString24    string  `json:"MaxFullTimeString24"`
	Price                  float64 `json:"Price"`
	Dollars                float64 `json:"Dollars"`
	Volume                 int     `json:"Volume"`
	TradeCount             int     `json:"TradeCount"`
	DollarsMultiplier      float64 `json:"DollarsMultiplier"`
	CumulativeDistribution float64 `json:"CumulativeDistribution"`
	TradeClusterRank       int     `json:"TradeClusterRank"`
}

// TradeDashboardClusterBomb is a compact top cluster-bomb row for dashboard summary output.
type TradeDashboardClusterBomb struct {
	MinFullDateTime        string  `json:"MinFullDateTime"`
	MaxFullDateTime        string  `json:"MaxFullDateTime"`
	MinFullTimeString24    string  `json:"MinFullTimeString24"`
	MaxFullTimeString24    string  `json:"MaxFullTimeString24"`
	Dollars                float64 `json:"Dollars"`
	Volume                 int     `json:"Volume"`
	TradeCount             int     `json:"TradeCount"`
	DollarsMultiplier      float64 `json:"DollarsMultiplier"`
	CumulativeDistribution float64 `json:"CumulativeDistribution"`
	TradeClusterBombRank   int     `json:"TradeClusterBombRank"`
}

// NewTradeDashboard builds the compact dashboard response for one ticker.
func NewTradeDashboard(
	ticker string,
	dateRange TradeDashboardDateRange,
	count int,
	requestedSections, returnedSections []string,
	trades []Trade,
	clusters []TradeCluster,
	levels []TradeLevel,
	clusterBombs []TradeClusterBomb,
) TradeDashboard {
	dashboard := TradeDashboard{
		Ticker:            ticker,
		DateRange:         dateRange,
		Count:             count,
		RequestedSections: requestedSections,
		ReturnedSections:  returnedSections,
		Trades:            NewTradeDashboardTradeRows(trades),
		Clusters:          NewTradeDashboardClusterRows(clusters),
		Levels:            NewTradeLevelRows(levels),
		ClusterBombs:      NewTradeDashboardClusterBombRows(clusterBombs),
	}
	populateTradeDashboardMetadata(&dashboard, trades, clusters, levels, clusterBombs)
	return dashboard
}

// NewTradeDashboardSummary builds the compact first-pass dashboard response for one ticker.
func NewTradeDashboardSummary(
	ticker string,
	dateRange TradeDashboardDateRange,
	count int,
	requestedSections, returnedSections []string,
	trades []Trade,
	clusters []TradeCluster,
	levels []TradeLevel,
	clusterBombs []TradeClusterBomb,
) TradeDashboardSummary {
	summary := TradeDashboardSummary{
		Ticker:            ticker,
		DateRange:         dateRange,
		Count:             count,
		RequestedSections: requestedSections,
		ReturnedSections:  returnedSections,
		Trades:            NewTradeDashboardTradeSummaries(trades),
		Clusters:          NewTradeDashboardClusters(clusters),
		Levels:            NewTradeLevelRows(levels),
		ClusterBombs:      NewTradeDashboardClusterBombs(clusterBombs),
	}
	populateTradeDashboardSummaryMetadata(&summary, trades, clusters, levels, clusterBombs)
	return summary
}

// NewTradeDashboardTradeRows projects full API trade rows into dashboard rows.
func NewTradeDashboardTradeRows(trades []Trade) []TradeDashboardTradeRow {
	rows := make([]TradeDashboardTradeRow, 0, len(trades))
	for i := range trades {
		rows = append(rows, NewTradeDashboardTradeRow(&trades[i]))
	}

	return rows
}

// NewTradeDashboardTradeRow projects one full API trade row into a dashboard row.
func NewTradeDashboardTradeRow(trade *Trade) TradeDashboardTradeRow {
	return TradeDashboardTradeRow{
		Date:                   trade.Date,
		FullDateTime:           trade.FullDateTime,
		Price:                  trade.Price,
		Volume:                 trade.Volume,
		Dollars:                trade.Dollars,
		DollarsMultiplier:      roundDollarsMultiplier(trade.DollarsMultiplier),
		PercentDailyVolume:     trade.PercentDailyVolume,
		RelativeSize:           trade.RelativeSize,
		CumulativeDistribution: trade.CumulativeDistribution,
		TradeRank:              trade.TradeRank,
		DarkPool:               trade.DarkPool,
		Sweep:                  trade.Sweep,
		LatePrint:              trade.LatePrint,
		SignaturePrint:         trade.SignaturePrint,
		OpeningTrade:           trade.OpeningTrade,
		ClosingTrade:           trade.ClosingTrade,
		PhantomPrint:           trade.PhantomPrint,
	}
}

// NewTradeDashboardClusterRows projects full API cluster rows into dashboard rows.
func NewTradeDashboardClusterRows(clusters []TradeCluster) []TradeDashboardClusterRow {
	rows := make([]TradeDashboardClusterRow, 0, len(clusters))
	for i := range clusters {
		rows = append(rows, NewTradeDashboardClusterRow(&clusters[i]))
	}

	return rows
}

// NewTradeDashboardClusterRow projects one full API cluster row into a dashboard row.
func NewTradeDashboardClusterRow(cluster *TradeCluster) TradeDashboardClusterRow {
	return TradeDashboardClusterRow{
		Date:                    cluster.Date,
		MinFullDateTime:         cluster.MinFullDateTime,
		MaxFullDateTime:         cluster.MaxFullDateTime,
		ClosePrice:              cluster.ClosePrice,
		Price:                   cluster.Price,
		Dollars:                 cluster.Dollars,
		AverageBlockSizeShares:  cluster.AverageBlockSizeShares,
		AverageBlockSizeDollars: cluster.AverageBlockSizeDollars,
		Volume:                  cluster.Volume,
		TradeCount:              cluster.TradeCount,
		DollarsMultiplier:       roundDollarsMultiplier(cluster.DollarsMultiplier),
		CumulativeDistribution:  cluster.CumulativeDistribution,
		AverageDailyVolume:      cluster.AverageDailyVolume,
		TradeClusterRank:        cluster.TradeClusterRank,
	}
}

// NewTradeDashboardClusterBombRows projects full API cluster-bomb rows into dashboard rows.
func NewTradeDashboardClusterBombRows(bombs []TradeClusterBomb) []TradeDashboardClusterBombRow {
	rows := make([]TradeDashboardClusterBombRow, 0, len(bombs))
	for i := range bombs {
		rows = append(rows, NewTradeDashboardClusterBombRow(&bombs[i]))
	}

	return rows
}

// NewTradeDashboardClusterBombRow projects one full API cluster-bomb row into a dashboard row.
func NewTradeDashboardClusterBombRow(bomb *TradeClusterBomb) TradeDashboardClusterBombRow {
	return TradeDashboardClusterBombRow{
		Date:                    bomb.Date,
		MinFullDateTime:         bomb.MinFullDateTime,
		MaxFullDateTime:         bomb.MaxFullDateTime,
		ClosePrice:              bomb.ClosePrice,
		Dollars:                 bomb.Dollars,
		AverageBlockSizeShares:  bomb.AverageBlockSizeShares,
		AverageBlockSizeDollars: bomb.AverageBlockSizeDollars,
		Volume:                  bomb.Volume,
		TradeCount:              bomb.TradeCount,
		DollarsMultiplier:       roundDollarsMultiplier(bomb.DollarsMultiplier),
		CumulativeDistribution:  bomb.CumulativeDistribution,
		AverageDailyVolume:      bomb.AverageDailyVolume,
		TradeClusterBombRank:    bomb.TradeClusterBombRank,
	}
}

// NewTradeDashboardTradeSummaries projects full trade rows into compact dashboard summary rows.
func NewTradeDashboardTradeSummaries(trades []Trade) []TradeDashboardTradeSummary {
	rows := make([]TradeDashboardTradeSummary, 0, len(trades))
	for i := range trades {
		trade := &trades[i]
		rows = append(rows, TradeDashboardTradeSummary{
			Date:              trade.Date,
			FullDateTime:      trade.FullDateTime,
			FullTimeString24:  trade.FullTimeString24,
			Price:             trade.Price,
			Volume:            trade.Volume,
			Dollars:           trade.Dollars,
			DollarsMultiplier: trade.DollarsMultiplier,
			TradeRank:         trade.TradeRank,
			DarkPool:          trade.DarkPool,
			Sweep:             trade.Sweep,
			ClosingTrade:      trade.ClosingTrade,
			TradeConditions:   trade.TradeConditions,
		})
	}
	return rows
}

// NewTradeDashboardClusters projects full cluster rows into compact dashboard summary rows.
func NewTradeDashboardClusters(clusters []TradeCluster) []TradeDashboardCluster {
	rows := make([]TradeDashboardCluster, 0, len(clusters))
	for i := range clusters {
		cluster := &clusters[i]
		rows = append(rows, TradeDashboardCluster{
			MinFullDateTime:        cluster.MinFullDateTime,
			MaxFullDateTime:        cluster.MaxFullDateTime,
			MinFullTimeString24:    cluster.MinFullTimeString24,
			MaxFullTimeString24:    cluster.MaxFullTimeString24,
			Price:                  cluster.Price,
			Dollars:                cluster.Dollars,
			Volume:                 cluster.Volume,
			TradeCount:             cluster.TradeCount,
			DollarsMultiplier:      cluster.DollarsMultiplier,
			CumulativeDistribution: cluster.CumulativeDistribution,
			TradeClusterRank:       cluster.TradeClusterRank,
		})
	}
	return rows
}

// NewTradeDashboardClusterBombs projects full cluster-bomb rows into compact dashboard summary rows.
func NewTradeDashboardClusterBombs(clusterBombs []TradeClusterBomb) []TradeDashboardClusterBomb {
	rows := make([]TradeDashboardClusterBomb, 0, len(clusterBombs))
	for i := range clusterBombs {
		clusterBomb := &clusterBombs[i]
		rows = append(rows, TradeDashboardClusterBomb{
			MinFullDateTime:        clusterBomb.MinFullDateTime,
			MaxFullDateTime:        clusterBomb.MaxFullDateTime,
			MinFullTimeString24:    clusterBomb.MinFullTimeString24,
			MaxFullTimeString24:    clusterBomb.MaxFullTimeString24,
			Dollars:                clusterBomb.Dollars,
			Volume:                 clusterBomb.Volume,
			TradeCount:             clusterBomb.TradeCount,
			DollarsMultiplier:      clusterBomb.DollarsMultiplier,
			CumulativeDistribution: clusterBomb.CumulativeDistribution,
			TradeClusterBombRank:   clusterBomb.TradeClusterBombRank,
		})
	}
	return rows
}

func populateTradeDashboardMetadata(
	dashboard *TradeDashboard,
	trades []Trade,
	clusters []TradeCluster,
	levels []TradeLevel,
	clusterBombs []TradeClusterBomb,
) {
	mergeTradeDashboardMetadata(&dashboard.Name, &dashboard.Sector, &dashboard.Industry, trades, clusters, levels, clusterBombs)
}

func populateTradeDashboardSummaryMetadata(
	summary *TradeDashboardSummary,
	trades []Trade,
	clusters []TradeCluster,
	levels []TradeLevel,
	clusterBombs []TradeClusterBomb,
) {
	mergeTradeDashboardMetadata(&summary.Name, &summary.Sector, &summary.Industry, trades, clusters, levels, clusterBombs)
}

func mergeTradeDashboardMetadata(
	name, sector *string,
	industry **string,
	trades []Trade,
	clusters []TradeCluster,
	levels []TradeLevel,
	clusterBombs []TradeClusterBomb,
) {
	for i := range trades {
		mergeTradeDashboardMetadataValues(name, sector, industry, trades[i].Name, trades[i].Sector, trades[i].Industry)
	}
	for i := range clusters {
		mergeTradeDashboardMetadataValues(name, sector, industry, clusters[i].Name, clusters[i].Sector, clusters[i].Industry)
	}
	for i := range clusterBombs {
		mergeTradeDashboardMetadataValues(name, sector, industry, clusterBombs[i].Name, clusterBombs[i].Sector, clusterBombs[i].Industry)
	}
	for i := range levels {
		if levels[i].Name != nil {
			mergeTradeDashboardMetadataValues(name, sector, industry, *levels[i].Name, "", nil)
		}
	}
}

func mergeTradeDashboardMetadataValues(
	name, sector *string,
	industry **string,
	nextName, nextSector string,
	nextIndustry *string,
) {
	if *name == "" {
		*name = nextName
	}
	if *sector == "" {
		*sector = nextSector
	}
	if *industry == nil && nextIndustry != nil && *nextIndustry != "" {
		*industry = cloneStringPointer(nextIndustry)
	}
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}
