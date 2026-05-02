package common

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestResolveDateRangeBranches(t *testing.T) {
	frozen := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	origTimeNow := TimeNow
	TimeNow = func() time.Time { return frozen }
	t.Cleanup(func() { TimeNow = origTimeNow })
	tests := []struct {
		name         string
		flags        map[string]string
		lookbackDays int
		required     bool
		wantStart    string
		wantEnd      string
		wantErr      string
	}{
		{name: "days and end-date uses end-date as base", flags: map[string]string{"end-date": "2025-03-01", "days": "5"}, required: true, wantStart: "2025-02-24", wantEnd: "2025-03-01"},
		{name: "days without end-date uses today as base", flags: map[string]string{"days": "7"}, required: true, wantStart: "2025-06-08", wantEnd: "2025-06-15"},
		{name: "explicit start and end range", flags: map[string]string{"start-date": "2025-01-01", "end-date": "2025-01-31"}, required: true, wantStart: "2025-01-01", wantEnd: "2025-01-31"},
		{name: "required range missing returns error", required: true, wantErr: "required unless --days is set"},
		{name: "optional range missing uses lookback before today", lookbackDays: 90, wantStart: "2025-03-17", wantEnd: "2025-06-15"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd, err := ResolveDateRange(dateTestCommand(t, tt.flags), tt.lookbackDays, tt.required)
			assertErrContains(t, err, tt.wantErr)
			if tt.wantErr != "" {
				return
			}
			if gotStart != tt.wantStart || gotEnd != tt.wantEnd {
				t.Fatalf("ResolveDateRange() = %q, %q; want %q, %q", gotStart, gotEnd, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestDefaultDates(t *testing.T) {
	frozen := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	origTimeNow := TimeNow
	TimeNow = func() time.Time { return frozen }
	t.Cleanup(func() { TimeNow = origTimeNow })
	tests := []struct {
		name         string
		flags        map[string]string
		lookbackDays int
		wantStart    string
		wantEnd      string
	}{
		{name: "both dates explicit", flags: map[string]string{"start-date": "2025-01-01", "end-date": "2025-03-01"}, lookbackDays: 90, wantStart: "2025-01-01", wantEnd: "2025-03-01"},
		{name: "today-only default", lookbackDays: 0, wantStart: "2025-06-15", wantEnd: "2025-06-15"},
		{name: "only start-date explicit", flags: map[string]string{"start-date": "2025-01-01"}, lookbackDays: 90, wantStart: "2025-01-01", wantEnd: "2025-06-15"},
		{name: "only end-date explicit", flags: map[string]string{"end-date": "2025-03-01"}, lookbackDays: 90, wantStart: "2025-03-17", wantEnd: "2025-03-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := DefaultDates(dateTestCommand(t, tt.flags), tt.lookbackDays)
			if gotStart != tt.wantStart || gotEnd != tt.wantEnd {
				t.Fatalf("DefaultDates() = %q, %q; want %q, %q", gotStart, gotEnd, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestResolveDateRangeErrors(t *testing.T) {
	tests := []struct {
		name    string
		flags   map[string]string
		wantErr string
	}{
		{name: "negative days", flags: map[string]string{"days": "-1"}, wantErr: "--days must be greater than or equal to 0"},
		{name: "days conflicts with start-date", flags: map[string]string{"start-date": "2025-06-01", "days": "7"}, wantErr: "--days cannot be used with --start-date"},
		{name: "days parses explicit end-date", flags: map[string]string{"end-date": "bad-date", "days": "7"}, wantErr: "parse --end-date for --days"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := RequiredDateRange(dateTestCommand(t, tt.flags))
			assertErrContains(t, err, tt.wantErr)
		})
	}
}

func dateTestCommand(t *testing.T, values map[string]string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("start-date", "", "")
	cmd.Flags().String("end-date", "", "")
	cmd.Flags().Int("days", 0, "")
	for name, value := range values {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set %s: %v", name, err)
		}
	}
	return cmd
}

func assertErrContains(t *testing.T, err error, want string) {
	t.Helper()
	if want == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got: %v", want, err)
	}
}
