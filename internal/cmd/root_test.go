package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandWiresStructCLIFeatures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantOut string
	}{
		{
			name:    "json schema tree is available",
			args:    []string{"--jsonschema=tree"},
			wantOut: "trades",
		},
		{
			name:    "env vars reference topic is available",
			args:    []string{"env-vars"},
			wantOut: "VOLUMELEADERS_AGENT_TRADES_DATE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rootCmd, err := NewRootCmd()
			if err != nil {
				t.Fatalf("NewRootCmd() error = %v", err)
			}

			var output bytes.Buffer
			rootCmd.SetOut(&output)
			rootCmd.SetErr(&output)
			rootCmd.SetArgs(tt.args)

			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v\noutput: %s", err, output.String())
			}
			if !strings.Contains(output.String(), tt.wantOut) {
				t.Fatalf("expected output to contain %q, got %q", tt.wantOut, output.String())
			}
		})
	}
}
