package models

import (
	"encoding/json"
	"testing"
)

func TestExhaustionScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, es *ExhaustionScore)
	}{
		{
			name: "exhaustion score with all fields",
			input: `{
				"DateKey": 20250423,
				"ExhaustionScoreRank": 45,
				"ExhaustionScoreRank30Day": 120,
				"ExhaustionScoreRank90Day": 200,
				"ExhaustionScoreRank365Day": 300
			}`,
			wantErr: false,
			check: func(t *testing.T, es *ExhaustionScore) {
				if es.DateKey != 20250423 {
					t.Errorf("DateKey: expected 20250423, got %d", es.DateKey)
				}
				if es.ExhaustionScoreRank != 45 {
					t.Errorf("ExhaustionScoreRank: expected 45, got %d", es.ExhaustionScoreRank)
				}
				if es.ExhaustionScoreRank30Day != 120 {
					t.Errorf("ExhaustionScoreRank30Day: expected 120, got %d", es.ExhaustionScoreRank30Day)
				}
				if es.ExhaustionScoreRank90Day != 200 {
					t.Errorf("ExhaustionScoreRank90Day: expected 200, got %d", es.ExhaustionScoreRank90Day)
				}
				if es.ExhaustionScoreRank365Day != 300 {
					t.Errorf("ExhaustionScoreRank365Day: expected 300, got %d", es.ExhaustionScoreRank365Day)
				}
			},
		},
		{
			name: "exhaustion score with zero values",
			input: `{
				"DateKey": 20250422,
				"ExhaustionScoreRank": 0,
				"ExhaustionScoreRank30Day": 0,
				"ExhaustionScoreRank90Day": 0,
				"ExhaustionScoreRank365Day": 0
			}`,
			wantErr: false,
			check: func(t *testing.T, es *ExhaustionScore) {
				if es.DateKey != 20250422 {
					t.Errorf("DateKey: expected 20250422, got %d", es.DateKey)
				}
				if es.ExhaustionScoreRank != 0 {
					t.Errorf("ExhaustionScoreRank: expected 0, got %d", es.ExhaustionScoreRank)
				}
				if es.ExhaustionScoreRank30Day != 0 {
					t.Errorf("ExhaustionScoreRank30Day: expected 0, got %d", es.ExhaustionScoreRank30Day)
				}
			},
		},
		{
			name: "exhaustion score with high rank values",
			input: `{
				"DateKey": 20250421,
				"ExhaustionScoreRank": 500,
				"ExhaustionScoreRank30Day": 1000,
				"ExhaustionScoreRank90Day": 2000,
				"ExhaustionScoreRank365Day": 5000
			}`,
			wantErr: false,
			check: func(t *testing.T, es *ExhaustionScore) {
				if es.ExhaustionScoreRank != 500 {
					t.Errorf("ExhaustionScoreRank: expected 500, got %d", es.ExhaustionScoreRank)
				}
				if es.ExhaustionScoreRank365Day != 5000 {
					t.Errorf("ExhaustionScoreRank365Day: expected 5000, got %d", es.ExhaustionScoreRank365Day)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var es ExhaustionScore
			err := json.Unmarshal([]byte(tt.input), &es)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, &es)
		})
	}
}

func TestMarketExhaustion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, me *MarketExhaustion)
	}{
		{
			name: "market exhaustion with all fields",
			input: `{
				"date_key": 20250423,
				"rank": 50,
				"rank_30d": 150,
				"rank_90d": 250,
				"rank_365d": 350
			}`,
			wantErr: false,
			check: func(t *testing.T, me *MarketExhaustion) {
				if me.DateKey != 20250423 {
					t.Errorf("DateKey: expected 20250423, got %d", me.DateKey)
				}
				if me.Rank != 50 {
					t.Errorf("Rank: expected 50, got %d", me.Rank)
				}
				if me.Rank30D != 150 {
					t.Errorf("Rank30D: expected 150, got %d", me.Rank30D)
				}
				if me.Rank90D != 250 {
					t.Errorf("Rank90D: expected 250, got %d", me.Rank90D)
				}
				if me.Rank365D != 350 {
					t.Errorf("Rank365D: expected 350, got %d", me.Rank365D)
				}
			},
		},
		{
			name: "market exhaustion with zero ranks",
			input: `{
				"date_key": 20250422,
				"rank": 0,
				"rank_30d": 0,
				"rank_90d": 0,
				"rank_365d": 0
			}`,
			wantErr: false,
			check: func(t *testing.T, me *MarketExhaustion) {
				if me.DateKey != 20250422 {
					t.Errorf("DateKey: expected 20250422, got %d", me.DateKey)
				}
				if me.Rank != 0 {
					t.Errorf("Rank: expected 0, got %d", me.Rank)
				}
				if me.Rank30D != 0 {
					t.Errorf("Rank30D: expected 0, got %d", me.Rank30D)
				}
			},
		},
		{
			name: "market exhaustion with high ranks",
			input: `{
				"date_key": 20250421,
				"rank": 1000,
				"rank_30d": 2000,
				"rank_90d": 3000,
				"rank_365d": 4000
			}`,
			wantErr: false,
			check: func(t *testing.T, me *MarketExhaustion) {
				if me.Rank != 1000 {
					t.Errorf("Rank: expected 1000, got %d", me.Rank)
				}
				if me.Rank365D != 4000 {
					t.Errorf("Rank365D: expected 4000, got %d", me.Rank365D)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var me MarketExhaustion
			err := json.Unmarshal([]byte(tt.input), &me)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected unmarshal error: %v", err)
			}
			tt.check(t, &me)
		})
	}
}
