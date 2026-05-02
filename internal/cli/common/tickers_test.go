package common

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestMultiTickerValue(t *testing.T) {
	tests := []struct {
		name string
		flag string
		args []string
		want string
	}{
		{name: "flag only", flag: "AAPL, MSFT", want: "AAPL,MSFT"},
		{name: "args only", args: []string{"AAPL", "MSFT"}, want: "AAPL,MSFT"},
		{name: "flag and args dedupe", flag: "AAPL,NVDA", args: []string{"NVDA", "MSFT"}, want: "AAPL,NVDA,MSFT"},
		{name: "empty values skipped", flag: " AAPL, ,MSFT ", args: []string{"AAPL"}, want: "AAPL,MSFT"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MultiTickerValue(tickerTestCommand(t, "tickers", tt.flag, tt.args)); got != tt.want {
				t.Fatalf("MultiTickerValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSingleTickerValue(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		args    []string
		want    string
		wantErr string
	}{
		{name: "flag ticker", flag: "AAPL", want: "AAPL"},
		{name: "positional ticker", args: []string{"MSFT"}, want: "MSFT"},
		{name: "flag and positional ticker conflict", flag: "AAPL", args: []string{"MSFT"}, wantErr: "use either --ticker or a ticker argument"},
		{name: "too many positional tickers", args: []string{"AAPL", "MSFT"}, wantErr: "expected at most one ticker argument"},
		{name: "missing ticker", wantErr: "--ticker or a ticker argument is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SingleTickerValue(tickerTestCommand(t, "ticker", tt.flag, tt.args))
			assertErrContains(t, err, tt.wantErr)
			if tt.wantErr == "" && got != tt.want {
				t.Fatalf("SingleTickerValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func tickerTestCommand(t *testing.T, flagName, flagValue string, args []string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String(flagName, "", "")
	if flagValue != "" {
		if err := cmd.Flags().Set(flagName, flagValue); err != nil {
			t.Fatalf("set %s: %v", flagName, err)
		}
	}
	cmd.SetArgs(args)
	cmd.Run = func(*cobra.Command, []string) {}
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute command: %v", err)
	}
	return cmd
}
