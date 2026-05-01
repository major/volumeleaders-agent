package trades

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestTradesCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
		want    Result
	}{
		{
			name: "valid date emits no-op scaffold",
			args: []string{"--date", "2026-04-30"},
			want: Result{
				Status: "not_implemented",
				Date:   "2026-04-30",
				Note:   "No API request has been wired yet. This scaffold validates inputs and exposes structcli JSON schema and MCP support.",
				Trades: []Trade{},
			},
		},
		{
			name:    "missing date fails",
			args:    []string{},
			wantErr: "date",
		},
		{
			name:    "invalid date fails",
			args:    []string{"--date", "04/30/2026"},
			wantErr: "use YYYY-MM-DD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd, err := NewCommand()
			if err != nil {
				t.Fatalf("NewCommand() error = %v", err)
			}

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(tt.args)

			err = cmd.Execute()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			var got Result
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal output: %v\noutput: %s", err, stdout.String())
			}
			if got.Status != tt.want.Status || got.Date != tt.want.Date || got.Note != tt.want.Note || len(got.Trades) != 0 {
				t.Fatalf("Result = %+v, want %+v", got, tt.want)
			}
		})
	}
}
