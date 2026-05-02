package common

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestAddDateRangeFlags(t *testing.T) {
	cmd := &cobra.Command{}
	AddDateRangeFlags(cmd)

	assertFlag(t, cmd, "start-date", "string", "")
	assertFlag(t, cmd, "end-date", "string", "")
	assertFlag(t, cmd, "days", "int", "0")
}

func TestAddOptionalDateRangeFlags(t *testing.T) {
	cmd := &cobra.Command{}
	AddOptionalDateRangeFlags(cmd)

	assertFlag(t, cmd, "start-date", "string", "")
	assertFlag(t, cmd, "end-date", "string", "")
	assertFlag(t, cmd, "days", "int", "0")
}

func TestAddVolumeRangeFlags(t *testing.T) {
	cmd := &cobra.Command{}
	AddVolumeRangeFlags(cmd)

	assertFlag(t, cmd, "min-volume", "int", "0")
	assertFlag(t, cmd, "max-volume", "int", "2000000000")
}

func TestAddPriceRangeFlags(t *testing.T) {
	cmd := &cobra.Command{}
	AddPriceRangeFlags(cmd)

	assertFlag(t, cmd, "min-price", "float64", "0")
	assertFlag(t, cmd, "max-price", "float64", "100000")
}

func TestAddDollarRangeFlags(t *testing.T) {
	cmd := &cobra.Command{}
	AddDollarRangeFlags(cmd, 250000)

	assertFlag(t, cmd, "min-dollars", "float64", "250000")
	assertFlag(t, cmd, "max-dollars", "float64", "30000000000")
}

func TestAddPaginationFlags(t *testing.T) {
	cmd := &cobra.Command{}
	AddPaginationFlags(cmd, 100, 1, "asc")

	assertFlag(t, cmd, "start", "int", "0")
	assertFlag(t, cmd, "length", "int", "100")
	assertFlag(t, cmd, "order-col", "int", "1")
	assertFlag(t, cmd, "order-dir", "string", "asc")
}

func TestAddOutputFormatFlags(t *testing.T) {
	cmd := &cobra.Command{}
	AddOutputFormatFlags(cmd)

	assertFlag(t, cmd, "format", "string", "json")
}

func TestAddTickersFlag(t *testing.T) {
	cmd := &cobra.Command{}
	AddTickersFlag(cmd)

	assertFlag(t, cmd, "tickers", "string", "")
	assertFlagMissing(t, cmd, "symbol")
	assertFlagMissing(t, cmd, "symbols")
}

func TestAddTickerFlag(t *testing.T) {
	cmd := &cobra.Command{}
	AddTickerFlag(cmd)

	assertFlag(t, cmd, "ticker", "string", "")
	assertFlagMissing(t, cmd, "symbol")
	assertFlagMissing(t, cmd, "symbols")
}

func assertFlag(t *testing.T, cmd *cobra.Command, name, flagType, defaultValue string) {
	t.Helper()

	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("expected flag %q to be registered", name)
	}
	if got := flag.Value.Type(); got != flagType {
		t.Fatalf("expected flag %q type %q, got %q", name, flagType, got)
	}
	if flag.DefValue != defaultValue {
		t.Fatalf("expected flag %q default %q, got %q", name, defaultValue, flag.DefValue)
	}
}

func assertFlagMissing(t *testing.T, cmd *cobra.Command, name string) {
	t.Helper()

	if flag := cmd.Flags().Lookup(name); flag != nil {
		t.Fatalf("expected flag %q to be absent", name)
	}
}
