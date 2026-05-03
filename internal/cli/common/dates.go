package common

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// TimeNow is the clock function used by date-range defaults.
// Tests replace it to control the current time.
var TimeNow = time.Now

// ResolveDateRange applies VolumeLeaders date flag semantics.
func ResolveDateRange(cmd *cobra.Command, lookbackDays int, required bool) (startDate, endDate string, err error) {
	now := TimeNow()
	today := now.Format("2006-01-02")
	days, err := cmd.Flags().GetInt("days")
	if err != nil {
		return "", "", fmt.Errorf("get --days: %w", err)
	}
	hasDays := cmd.Flags().Changed("days")
	if hasDays && days < 0 {
		return "", "", fmt.Errorf("--days must be greater than or equal to 0")
	}
	if hasDays && cmd.Flags().Changed("start-date") {
		return "", "", fmt.Errorf("--days cannot be used with --start-date")
	}

	endDate, err = cmd.Flags().GetString("end-date")
	if err != nil {
		return "", "", fmt.Errorf("get --end-date: %w", err)
	}
	if !cmd.Flags().Changed("end-date") {
		if required && !hasDays {
			return "", "", fmt.Errorf("--start-date and --end-date are required unless --days is set")
		}
		endDate = today
	}

	startDate, err = cmd.Flags().GetString("start-date")
	if err != nil {
		return "", "", fmt.Errorf("get --start-date: %w", err)
	}
	if !cmd.Flags().Changed("start-date") {
		switch {
		case hasDays:
			base := now
			if cmd.Flags().Changed("end-date") {
				parsed, parseErr := time.Parse("2006-01-02", endDate)
				if parseErr != nil {
					return "", "", fmt.Errorf("parse --end-date for --days: %w", parseErr)
				}
				base = parsed
			}
			startDate = base.AddDate(0, 0, -days).Format("2006-01-02")
		case required:
			return "", "", fmt.Errorf("--start-date and --end-date are required unless --days is set")
		case lookbackDays > 0:
			startDate = now.AddDate(0, 0, -lookbackDays).Format("2006-01-02")
		default:
			startDate = today
		}
	}

	return startDate, endDate, nil
}
