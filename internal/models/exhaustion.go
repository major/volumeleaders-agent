package models

// ExhaustionScore represents exhaustion score rank values for a specific trading day.
type ExhaustionScore struct {
	DateKey                      int `json:"DateKey"`
	ExhaustionScoreRank          int `json:"ExhaustionScoreRank"`
	ExhaustionScoreRank30Day     int `json:"ExhaustionScoreRank30Day"`
	ExhaustionScoreRank90Day     int `json:"ExhaustionScoreRank90Day"`
	ExhaustionScoreRank365Day    int `json:"ExhaustionScoreRank365Day"`
}
