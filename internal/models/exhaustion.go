package models

// ExhaustionScore represents exhaustion score rank values for a specific trading day.
type ExhaustionScore struct {
	DateKey                   int `json:"DateKey"`
	ExhaustionScoreRank       int `json:"ExhaustionScoreRank"`
	ExhaustionScoreRank30Day  int `json:"ExhaustionScoreRank30Day"`
	ExhaustionScoreRank90Day  int `json:"ExhaustionScoreRank90Day"`
	ExhaustionScoreRank365Day int `json:"ExhaustionScoreRank365Day"`
}

// MarketExhaustion is the compact CLI projection for market exhaustion ranks.
type MarketExhaustion struct {
	DateKey  int `json:"date_key"`
	Rank     int `json:"rank"`
	Rank30D  int `json:"rank_30d"`
	Rank90D  int `json:"rank_90d"`
	Rank365D int `json:"rank_365d"`
}
