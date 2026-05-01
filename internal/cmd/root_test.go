package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandWiresStructCLIFeatures(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantOuts []string
	}{
		{
			name:     "json schema tree is available",
			args:     []string{"--jsonschema=tree"},
			wantOuts: []string{"trades", "phantom", "offsetting", "bull-leverage", "bear-leverage", "biotech", "bonds", "commodities", "communications-services"},
		},
		{
			name:     "env vars reference topic is available",
			args:     []string{"env-vars"},
			wantOuts: []string{"VOLUMELEADERS_AGENT_TRADES_DATE", "VOLUMELEADERS_AGENT_PHANTOM_DATE", "VOLUMELEADERS_AGENT_OFFSETTING_DATE", "VOLUMELEADERS_AGENT_BULL_LEVERAGE_DATE", "VOLUMELEADERS_AGENT_BEAR_LEVERAGE_DATE", "VOLUMELEADERS_AGENT_BIOTECH_DATE", "VOLUMELEADERS_AGENT_BONDS_DATE", "VOLUMELEADERS_AGENT_COMMODITIES_DATE", "VOLUMELEADERS_AGENT_COMMUNICATIONS_SERVICES_DATE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			for _, wantOut := range tt.wantOuts {
				if !strings.Contains(output.String(), wantOut) {
					t.Fatalf("expected output to contain %q, got %q", wantOut, output.String())
				}
			}
		})
	}
}
