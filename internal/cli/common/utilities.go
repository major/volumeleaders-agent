package common

import (
	"strconv"
	"strings"
)

// IntStr converts an int to its decimal string representation.
func IntStr(value int) string {
	return strconv.Itoa(value)
}

// FormatFloat converts a float64 to a string without trailing zeros.
func FormatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// BoolString returns "true" or "false" for the given bool. It is used instead
// of strconv.FormatBool to match VolumeLeaders form field values.
func BoolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// ToDateKey normalizes YYYY-MM-DD to YYYYMMDD for the API.
func ToDateKey(value string) string {
	return strings.ReplaceAll(value, "-", "")
}

// ParseSnapshotString parses the semicolon-delimited "TICKER:PRICE" response
// from GetAllSnapshots into a ticker-to-price map.
func ParseSnapshotString(raw string) map[string]float64 {
	result := make(map[string]float64)
	if raw == "" {
		return result
	}

	for pair := range strings.SplitSeq(raw, ";") {
		ticker, priceStr, found := strings.Cut(pair, ":")
		if !found {
			continue
		}
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			continue
		}
		result[ticker] = price
	}
	return result
}
