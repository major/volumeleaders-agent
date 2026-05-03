package trade

import "fmt"

// validateRange checks that value is between 1 and max (inclusive).
// flagName is the CLI flag name (e.g., "length"), context describes the operation.
func validateRange(value, max int, flagName, context string) error {
	if value < 1 || value > max {
		return fmt.Errorf("--%s must be between 1 and %d for %s", flagName, max, context)
	}
	return nil
}
