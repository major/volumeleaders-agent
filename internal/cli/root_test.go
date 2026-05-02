package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/cli/testutil"
)

func TestRootPersistentPreRunStoresPrettyFlagInContext(t *testing.T) {
	t.Parallel()
	var got bool
	rootCmd := NewRootCmd("test")
	rootCmd.AddCommand(&cobra.Command{
		Use: "child",
		RunE: func(cmd *cobra.Command, _ []string) error {
			got, _ = cmd.Context().Value(common.PrettyJSONKey).(bool)
			return nil
		},
	})

	_, _, err := testutil.ExecuteCommand(t, rootCmd, context.Background(), "--pretty", "child")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatal("expected pretty flag to be stored as true in command context")
	}
}

func TestRootSilenceErrorsPreventsCobraErrorOutput(t *testing.T) {
	t.Parallel()
	rootCmd := NewRootCmd("test")
	rootCmd.AddCommand(&cobra.Command{
		Use: "fail",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("boom")
		},
	})

	_, stderr, err := testutil.ExecuteCommand(t, rootCmd, context.Background(), "fail")
	if err == nil {
		t.Fatal("expected command error")
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
}

func TestRootSilenceUsagePreventsUsageOutputOnError(t *testing.T) {
	t.Parallel()
	rootCmd := NewRootCmd("test")
	rootCmd.AddCommand(&cobra.Command{
		Use: "fail",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("boom")
		},
	})

	stdout, stderr, err := testutil.ExecuteCommand(t, rootCmd, context.Background(), "fail")
	if err == nil {
		t.Fatal("expected command error")
	}
	combinedOutput := stdout + stderr
	if strings.Contains(combinedOutput, "Usage:") {
		t.Fatalf("expected no usage output, got stdout=%q stderr=%q", stdout, stderr)
	}
}
