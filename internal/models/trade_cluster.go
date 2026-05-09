package models

// TradeCluster represents a VolumeLeaders aggregated trade cluster row.
type TradeCluster struct {
	Date                           AspNetDate `json:"Date"`
	DateKey                        int        `json:"DateKey"`
	SecurityKey                    int        `json:"SecurityKey"`
	Ticker                         string     `json:"Ticker"`
	Sector                         string     `json:"Sector"`
	Industry                       *string    `json:"Industry"`
	Name                           string     `json:"Name"`
	MinFullDateTime                string     `json:"MinFullDateTime"`
	MaxFullDateTime                string     `json:"MaxFullDateTime"`
	MinFullTimeString24            string     `json:"MinFullTimeString24"`
	MaxFullTimeString24            string     `json:"MaxFullTimeString24"`
	ClosePrice                     float64    `json:"ClosePrice"`
	Price                          float64    `json:"Price"`
	Dollars                        float64    `json:"Dollars"`
	AverageBlockSizeShares         int        `json:"AverageBlockSizeShares"`
	AverageBlockSizeDollars        float64    `json:"AverageBlockSizeDollars"`
	Volume                         int        `json:"Volume"`
	TradeCount                     int        `json:"TradeCount"`
	IPODate                        AspNetDate `json:"IPODate"`
	DollarsMultiplier              float64    `json:"DollarsMultiplier"`
	CumulativeDistribution         float64    `json:"CumulativeDistribution"`
	AverageDailyVolume             int        `json:"AverageDailyVolume"`
	EOM                            FlexBool   `json:"EOM"`
	EOQ                            FlexBool   `json:"EOQ"`
	EOY                            FlexBool   `json:"EOY"`
	OPEX                           FlexBool   `json:"OPEX"`
	VOLEX                          FlexBool   `json:"VOLEX"`
	InsideBar                      FlexBool   `json:"InsideBar"`
	DoubleInsideBar                FlexBool   `json:"DoubleInsideBar"`
	LastComparibleTradeClusterDate AspNetDate `json:"LastComparibleTradeClusterDate"`
	TradeClusterRank               int        `json:"TradeClusterRank"`
	TotalRows                      int        `json:"TotalRows"`
	ExternalFeed                   FlexBool   `json:"ExternalFeed"`
}

// TradeClusterRow is the compact default JSON shape for trade cluster output.
type TradeClusterRow struct {
	Date                   string  `json:"Date"`
	Ticker                 string  `json:"Ticker"`
	Sector                 string  `json:"Sector"`
	Industry               *string `json:"Industry"`
	Name                   string  `json:"Name"`
	MinFullDateTime        string  `json:"MinFullDateTime"`
	MaxFullDateTime        string  `json:"MaxFullDateTime"`
	Price                  float64 `json:"Price"`
	Dollars                float64 `json:"Dollars"`
	Volume                 int     `json:"Volume"`
	TradeCount             int     `json:"TradeCount"`
	DollarsMultiplier      float64 `json:"DollarsMultiplier"`
	CumulativeDistribution float64 `json:"CumulativeDistribution"`
	TradeClusterRank       int     `json:"TradeClusterRank"`
}

// NewTradeClusterRows projects full API cluster rows into compact output rows.
func NewTradeClusterRows(clusters []TradeCluster) []TradeClusterRow {
	rows := make([]TradeClusterRow, 0, len(clusters))
	for i := range clusters {
		rows = append(rows, NewTradeClusterRow(&clusters[i]))
	}

	return rows
}

// NewTradeClusterRow projects one full API cluster row into a compact output row.
func NewTradeClusterRow(cluster *TradeCluster) TradeClusterRow {
	return TradeClusterRow{
		Date:                   compactDate(cluster.Date),
		Ticker:                 cluster.Ticker,
		Sector:                 cluster.Sector,
		Industry:               cluster.Industry,
		Name:                   cluster.Name,
		MinFullDateTime:        cluster.MinFullDateTime,
		MaxFullDateTime:        cluster.MaxFullDateTime,
		Price:                  cluster.Price,
		Dollars:                cluster.Dollars,
		Volume:                 cluster.Volume,
		TradeCount:             cluster.TradeCount,
		DollarsMultiplier:      roundDollarsMultiplier(cluster.DollarsMultiplier),
		CumulativeDistribution: cluster.CumulativeDistribution,
		TradeClusterRank:       cluster.TradeClusterRank,
	}
}

// TradeClusterBomb represents a VolumeLeaders trade cluster bomb row.
type TradeClusterBomb struct {
	Date                               AspNetDate `json:"Date"`
	DateKey                            int        `json:"DateKey"`
	SecurityKey                        int        `json:"SecurityKey"`
	Ticker                             string     `json:"Ticker"`
	Sector                             string     `json:"Sector"`
	Industry                           *string    `json:"Industry"`
	Name                               string     `json:"Name"`
	MinFullDateTime                    string     `json:"MinFullDateTime"`
	MaxFullDateTime                    string     `json:"MaxFullDateTime"`
	MinFullTimeString24                string     `json:"MinFullTimeString24"`
	MaxFullTimeString24                string     `json:"MaxFullTimeString24"`
	ClosePrice                         float64    `json:"ClosePrice"`
	Dollars                            float64    `json:"Dollars"`
	AverageBlockSizeShares             int        `json:"AverageBlockSizeShares"`
	AverageBlockSizeDollars            float64    `json:"AverageBlockSizeDollars"`
	Volume                             int        `json:"Volume"`
	TradeCount                         int        `json:"TradeCount"`
	IPODate                            AspNetDate `json:"IPODate"`
	DollarsMultiplier                  float64    `json:"DollarsMultiplier"`
	CumulativeDistribution             float64    `json:"CumulativeDistribution"`
	AverageDailyVolume                 int        `json:"AverageDailyVolume"`
	EOM                                FlexBool   `json:"EOM"`
	EOQ                                FlexBool   `json:"EOQ"`
	EOY                                FlexBool   `json:"EOY"`
	OPEX                               FlexBool   `json:"OPEX"`
	VOLEX                              FlexBool   `json:"VOLEX"`
	InsideBar                          FlexBool   `json:"InsideBar"`
	DoubleInsideBar                    FlexBool   `json:"DoubleInsideBar"`
	LastComparableTradeClusterBombDate AspNetDate `json:"LastComparableTradeClusterBombDate"`
	TradeClusterBombRank               int        `json:"TradeClusterBombRank"`
	TotalRows                          int        `json:"TotalRows"`
	ExternalFeed                       FlexBool   `json:"ExternalFeed"`
}

// TradeClusterBombRow is the compact default JSON shape for cluster bomb output.
type TradeClusterBombRow struct {
	Date                   string  `json:"Date"`
	Ticker                 string  `json:"Ticker"`
	Sector                 string  `json:"Sector"`
	Industry               *string `json:"Industry"`
	Name                   string  `json:"Name"`
	MinFullDateTime        string  `json:"MinFullDateTime"`
	MaxFullDateTime        string  `json:"MaxFullDateTime"`
	Dollars                float64 `json:"Dollars"`
	Volume                 int     `json:"Volume"`
	TradeCount             int     `json:"TradeCount"`
	DollarsMultiplier      float64 `json:"DollarsMultiplier"`
	CumulativeDistribution float64 `json:"CumulativeDistribution"`
	TradeClusterBombRank   int     `json:"TradeClusterBombRank"`
}

// NewTradeClusterBombRows projects full API cluster bomb rows into compact output rows.
func NewTradeClusterBombRows(bombs []TradeClusterBomb) []TradeClusterBombRow {
	rows := make([]TradeClusterBombRow, 0, len(bombs))
	for i := range bombs {
		rows = append(rows, NewTradeClusterBombRow(&bombs[i]))
	}

	return rows
}

// NewTradeClusterBombRow projects one full API cluster bomb row into a compact output row.
func NewTradeClusterBombRow(bomb *TradeClusterBomb) TradeClusterBombRow {
	return TradeClusterBombRow{
		Date:                   compactDate(bomb.Date),
		Ticker:                 bomb.Ticker,
		Sector:                 bomb.Sector,
		Industry:               bomb.Industry,
		Name:                   bomb.Name,
		MinFullDateTime:        bomb.MinFullDateTime,
		MaxFullDateTime:        bomb.MaxFullDateTime,
		Dollars:                bomb.Dollars,
		Volume:                 bomb.Volume,
		TradeCount:             bomb.TradeCount,
		DollarsMultiplier:      roundDollarsMultiplier(bomb.DollarsMultiplier),
		CumulativeDistribution: bomb.CumulativeDistribution,
		TradeClusterBombRank:   bomb.TradeClusterBombRank,
	}
}

func compactDate(date AspNetDate) string {
	if !date.Valid {
		return ""
	}
	return date.Format("2006-01-02")
}
