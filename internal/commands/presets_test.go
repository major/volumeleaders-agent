package commands

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/major/volumeleaders-agent/internal/models"
	cli "github.com/urfave/cli/v3"
)

func TestRunTradePresetTickers(t *testing.T) {
	originalPresets := tradePresets
	t.Cleanup(func() { tradePresets = originalPresets })

	tests := []struct {
		name          string
		presets       []tradePreset
		presetArg     string
		want          models.PresetTickersInfo
		wantErrSubstr string
	}{
		{
			name:          "missing preset returns error",
			presets:       []tradePreset{},
			presetArg:     "Missing",
			wantErrSubstr: "preset \"Missing\" not found",
		},
		{
			name:      "ticker preset returns normalized tickers",
			presetArg: "Ticker Preset",
			presets: []tradePreset{
				{
					name:  "Ticker Preset",
					group: "Test Group",
					filters: map[string]string{
						"Tickers":        " AAPL, NVDA,,AAPL,MSFT ",
						"SectorIndustry": "Technology",
					},
				},
			},
			want: models.PresetTickersInfo{
				Preset:  "Ticker Preset",
				Group:   "Test Group",
				Type:    "tickers",
				Tickers: []string{"AAPL", "NVDA", "MSFT"},
			},
		},
		{
			name:      "sector preset returns sector filter",
			presetArg: "Sector Preset",
			presets: []tradePreset{
				{
					name:  "Sector Preset",
					group: "Test Group",
					filters: map[string]string{
						"SectorIndustry": "Healthcare",
					},
				},
			},
			want: models.PresetTickersInfo{
				Preset:         "Sector Preset",
				Group:          "Test Group",
				Type:           "sector-filter",
				SectorIndustry: "Healthcare",
			},
		},
		{
			name:      "unfiltered preset returns unfiltered type",
			presetArg: "All Trades",
			presets: []tradePreset{
				{
					name:    "All Trades",
					group:   "Test Group",
					filters: map[string]string{},
				},
			},
			want: models.PresetTickersInfo{
				Preset: "All Trades",
				Group:  "Test Group",
				Type:   "unfiltered",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tradePresets = tt.presets

			root := &cli.Command{Commands: []*cli.Command{newTradePresetTickersCommand()}}
			var runErr error
			output := captureStdout(t, func() {
				runErr = root.Run(context.Background(), []string{"app", "preset-tickers", "--preset", tt.presetArg})
			})

			assertErrContains(t, runErr, tt.wantErrSubstr)
			if tt.wantErrSubstr != "" {
				return
			}

			var got models.PresetTickersInfo
			if err := json.Unmarshal([]byte(output), &got); err != nil {
				t.Fatalf("unmarshal preset tickers output: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("preset tickers mismatch\nexpected: %#v\ngot:      %#v", tt.want, got)
			}
		})
	}
}
