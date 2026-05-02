package common_test

import (
	"testing"

	"github.com/major/volumeleaders-agent/internal/cli"
	"github.com/major/volumeleaders-agent/internal/cli/testutil"
)

// flagSpec describes the expected name, type, and default value for a CLI flag
// registered via structcli.Bind struct tags.
type flagSpec struct {
	name     string
	typeName string
	defValue string
}

// assertFlags verifies that every flag in specs exists on the command found at
// cmdPath and has the expected type and default value.
func assertFlags(t *testing.T, cmdPath []string, specs []flagSpec) {
	t.Helper()
	rootCmd := cli.NewRootCmd("test")
	cmd, _, err := rootCmd.Find(cmdPath)
	if err != nil {
		t.Fatalf("find %v: %v", cmdPath, err)
	}

	for _, tt := range specs {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := cmd.Flags().Lookup(tt.name)
			if f == nil {
				t.Fatalf("flag --%s not registered", tt.name)
			}
			if f.Value.Type() != tt.typeName {
				t.Errorf("flag --%s type = %q, want %q", tt.name, f.Value.Type(), tt.typeName)
			}
			if f.DefValue != tt.defValue {
				t.Errorf("flag --%s default = %q, want %q", tt.name, f.DefValue, tt.defValue)
			}
		})
	}
}

// TestVolumeInstitutionalFlagRegistration verifies that struct-tag-based flag
// registration on the volume institutional command produces the expected flags
// with correct types and defaults from volumeOptions{Length: 100, OrderCol: 1,
// OrderDir: "asc"}.
func TestVolumeInstitutionalFlagRegistration(t *testing.T) {
	t.Parallel()
	assertFlags(t, []string{"volume", "institutional"}, []flagSpec{
		{"date", "string", ""},
		{"tickers", "string", ""},
		{"format", "string", "json"},
		{"start", "int", "0"},
		{"length", "int", "100"},
		{"order-col", "int", "1"},
		{"order-dir", "string", "asc"},
	})
}

// TestVolumeAHInstitutionalFlagRegistration verifies ah-institutional shares
// the same flag set and defaults as institutional.
func TestVolumeAHInstitutionalFlagRegistration(t *testing.T) {
	t.Parallel()
	assertFlags(t, []string{"volume", "ah-institutional"}, []flagSpec{
		{"date", "string", ""},
		{"tickers", "string", ""},
		{"format", "string", "json"},
		{"start", "int", "0"},
		{"length", "int", "100"},
		{"order-col", "int", "1"},
		{"order-dir", "string", "asc"},
	})
}

// TestVolumeTotalFlagRegistration verifies the total subcommand shares the same
// flag set and defaults.
func TestVolumeTotalFlagRegistration(t *testing.T) {
	t.Parallel()
	assertFlags(t, []string{"volume", "total"}, []flagSpec{
		{"date", "string", ""},
		{"tickers", "string", ""},
		{"format", "string", "json"},
		{"start", "int", "0"},
		{"length", "int", "100"},
		{"order-col", "int", "1"},
		{"order-dir", "string", "asc"},
	})
}

// TestMarketEarningsFlagRegistration verifies that struct-tag-based flag
// registration on market earnings produces the expected flags from
// earningsOptions{}.
func TestMarketEarningsFlagRegistration(t *testing.T) {
	t.Parallel()
	assertFlags(t, []string{"market", "earnings"}, []flagSpec{
		{"start-date", "string", ""},
		{"end-date", "string", ""},
		{"days", "int", "0"},
		{"format", "string", "json"},
		{"fields", "string", ""},
	})
}

// TestAlertCreateFlagRegistration verifies a representative sample of flags on
// alert create, including defaults set in the constructor (TickerGroup,
// TradeConditions, ClosingTradeConditions).
func TestAlertCreateFlagRegistration(t *testing.T) {
	t.Parallel()
	assertFlags(t, []string{"alert", "create"}, []flagSpec{
		{"name", "string", ""},
		{"ticker-group", "string", "AllTickers"},
		{"tickers", "string", ""},
		{"trade-conditions", "string", "0"},
		{"closing-trade-conditions", "string", "0"},
		{"dark-pool", "bool", "false"},
		{"sweep", "bool", "false"},
		{"trade-rank-lte", "int", "0"},
		{"trade-dollars-gte", "int", "0"},
	})
}

// TestEnumFlagValidation verifies structcli rejects invalid bounded values
// during flag parsing, before command handlers can perform API requests.
func TestEnumFlagValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "format",
			args:    []string{"alert", "configs", "--format", "table"},
			wantErr: `invalid value "table"`,
		},
		{
			name:    "order-dir",
			args:    []string{"volume", "institutional", "--date", "2025-01-15", "--order-dir", "sideways"},
			wantErr: `invalid value "sideways"`,
		},
		{
			name:    "group-by",
			args:    []string{"trade", "list", "--summary", "--group-by", "sector"},
			wantErr: `invalid value "sector"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := testutil.ExecuteCommand(t, cli.NewRootCmd("test"), t.Context(), tt.args...)
			testutil.AssertErrContains(t, err, tt.wantErr)
		})
	}
}
