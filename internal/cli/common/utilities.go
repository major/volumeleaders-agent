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
