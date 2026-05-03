package trade

import "fmt"

// validateRange checks that value is between 1 and maxValue (inclusive).
// flagName is the CLI flag name (e.g., "length"), context describes the operation.
//
//nolint:unparam // maxValue is generic; callers pass different constants
func validateRange(value, maxValue int, flagName, context string) error {
	if value < 1 || value > maxValue {
		return fmt.Errorf("--%s must be between 1 and %d for %s", flagName, maxValue, context)
	}
	return nil
}
