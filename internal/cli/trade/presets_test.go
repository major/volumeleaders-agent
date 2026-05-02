package trade

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/major/volumeleaders-agent/internal/cli/testutil"
	"github.com/major/volumeleaders-agent/internal/models"
)

func TestApplyExplicitFlagsPreservesPresetValuesUnlessChanged(t *testing.T) {
	t.Parallel()

	preset, err := findPreset("Top-100 Rank")
	if err != nil {
		t.Fatalf("find preset: %v", err)
	}
	filters := mapsClone(preset.filters)
	cmd := newTradeListCommand()
	if err := cmd.Flags().Set("dark-pools", "1"); err != nil {
		t.Fatalf("set dark-pools: %v", err)
	}

	applyExplicitFlags(cmd, filters)

	if got := filters["DarkPools"]; got != "1" {
		t.Fatalf("DarkPools = %q, want 1", got)
	}
	if got := filters["TradeRank"]; got != "100" {
		t.Fatalf("TradeRank = %q, want preset value 100", got)
	}
	if got := filters["MaxDollars"]; got != "100000000000" {
		t.Fatalf("MaxDollars = %q, want preset value 100000000000", got)
	}
}

func TestChangedTrueWhenFlagSetToDefaultValue(t *testing.T) {
	t.Parallel()

	cmd := newTradeSentimentCommand()
	if err := cmd.Flags().Set("vcd", "97"); err != nil {
		t.Fatalf("set vcd: %v", err)
	}
	if !cmd.Flags().Changed("vcd") {
		t.Fatal("Changed(vcd) = false, want true for explicit default value")
	}
}

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
			presets: []tradePreset{{name: "Ticker Preset", group: "Test Group", filters: map[string]string{
				"Tickers":        " AAPL, NVDA,,AAPL,MSFT ",
				"SectorIndustry": "Technology",
			}}},
			want: models.PresetTickersInfo{Preset: "Ticker Preset", Group: "Test Group", Type: "tickers", Tickers: []string{"AAPL", "NVDA", "MSFT"}},
		},
		{
			name:      "sector preset returns sector filter",
			presetArg: "Sector Preset",
			presets: []tradePreset{{name: "Sector Preset", group: "Test Group", filters: map[string]string{
				"SectorIndustry": "Healthcare",
			}}},
			want: models.PresetTickersInfo{Preset: "Sector Preset", Group: "Test Group", Type: "sector-filter", SectorIndustry: "Healthcare"},
		},
		{
			name:      "unfiltered preset returns unfiltered type",
			presetArg: "All Trades",
			presets:   []tradePreset{{name: "All Trades", group: "Test Group", filters: map[string]string{}}},
			want:      models.PresetTickersInfo{Preset: "All Trades", Group: "Test Group", Type: "unfiltered"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tradePresets = tt.presets
			cmd := newTradePresetTickersCommand()
			stdout, _, err := testutil.ExecuteCommand(t, cmd, t.Context(), "--preset", tt.presetArg)
			testutil.AssertErrContains(t, err, tt.wantErrSubstr)
			if tt.wantErrSubstr != "" {
				return
			}

			var got models.PresetTickersInfo
			if err := json.Unmarshal([]byte(stdout), &got); err != nil {
				t.Fatalf("unmarshal preset tickers output: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("preset tickers mismatch\nexpected: %#v\ngot:      %#v", tt.want, got)
			}
		})
	}
}

func mapsClone(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
