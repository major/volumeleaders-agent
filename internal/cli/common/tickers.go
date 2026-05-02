package common

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// MultiTickerValue combines comma-separated --tickers values with positional
// ticker arguments and deduplicates them while preserving first-seen order.
func MultiTickerValue(cmd *cobra.Command) string {
	flagValue, _ := cmd.Flags().GetString("tickers")
	values := splitTickerValues(flagValue)
	values = append(values, splitTickerValues(strings.Join(cmd.Flags().Args(), ","))...)
	return strings.Join(dedupeStrings(values), ",")
}

// SingleTickerValue returns the one ticker supplied by --ticker or a positional
// argument, rejecting missing or ambiguous inputs.
func SingleTickerValue(cmd *cobra.Command) (string, error) {
	flagValue, err := cmd.Flags().GetString("ticker")
	if err != nil {
		return "", fmt.Errorf("get --ticker: %w", err)
	}
	flagValue = strings.TrimSpace(flagValue)
	args := cmd.Flags().Args()
	if len(args) > 1 {
		return "", fmt.Errorf("expected at most one ticker argument, got %d", len(args))
	}
	if flagValue != "" && len(args) == 1 {
		return "", fmt.Errorf("use either --ticker or a ticker argument, not both")
	}
	if flagValue != "" {
		return flagValue, nil
	}
	if len(args) == 1 {
		return strings.TrimSpace(args[0]), nil
	}
	return "", fmt.Errorf("--ticker or a ticker argument is required")
}

func splitTickerValues(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	items := make([]string, 0, strings.Count(value, ",")+1)
	for item := range strings.SplitSeq(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
