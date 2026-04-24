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
