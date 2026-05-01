// Package trades contains commands for VolumeLeaders trade workflows.
package trades

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/leodido/structcli"
	"github.com/spf13/cobra"
)

const dateLayout = "2006-01-02"

// Options defines the LLM-readable contract for fetching unusual trades.
type Options struct {
	Date string `flag:"date" flagshort:"d" flagdescr:"Trade date to query, formatted as YYYY-MM-DD." flagenv:"true" flagrequired:"true" flaggroup:"Query" validate:"required" mod:"trim"`
}

// Result is the stable no-op response shape for the unusual trades command.
type Result struct {
	Status string  `json:"status"`
	Date   string  `json:"date"`
	Note   string  `json:"note"`
	Trades []Trade `json:"trades"`
}

// Trade is intentionally empty until the VolumeLeaders API response shape is wired.
type Trade struct{}

// NewCommand builds the no-op large unusual trades command.
func NewCommand() (*cobra.Command, error) {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:     "trades",
		Aliases: []string{"large-trades", "unusual-trades"},
		Short:   "Fetch large unusual trades for a date",
		Long:    "Fetch large unusual VolumeLeaders trades for a particular trading day. The API call is not wired yet, so this currently returns a no-op JSON scaffold.",
		Example: "volumeleaders-agent trades --date 2026-04-30",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), cmd, opts)
		},
	}

	if err := structcli.Bind(cmd, opts); err != nil {
		return nil, fmt.Errorf("bind trades options: %w", err)
	}

	return cmd, nil
}

func run(ctx context.Context, cmd *cobra.Command, opts *Options) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("run trades command: %w", ctx.Err())
	default:
	}

	tradeDate, err := time.Parse(dateLayout, opts.Date)
	if err != nil {
		return fmt.Errorf("invalid date %q: use YYYY-MM-DD: %w", opts.Date, err)
	}

	result := Result{
		Status: "not_implemented",
		Date:   tradeDate.Format(dateLayout),
		Note:   "No API request has been wired yet. This scaffold validates inputs and exposes structcli JSON schema and MCP support.",
		Trades: []Trade{},
	}
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("encode trades scaffold response: %w", err)
	}

	return nil
}
