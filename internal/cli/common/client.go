package common

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/major/volumeleaders-agent/internal/auth"
	"github.com/major/volumeleaders-agent/internal/client"
)

// NewCommandClient centralizes authenticated client creation for command
// handlers while allowing tests to inject a pre-built client through context.
func NewCommandClient(ctx context.Context) (*client.Client, error) {
	if c, ok := ctx.Value(TestClientKey).(*client.Client); ok {
		return c, nil
	}
	vlClient, err := client.New(ctx)
	if err != nil {
		if auth.IsSessionExpired(err) {
			var detail interface{ Detail() string }
			if errors.As(err, &detail) {
				slog.Debug("VolumeLeaders session expired", "detail", detail.Detail())
			}
			return nil, fmt.Errorf("%s: %w", auth.SessionExpiredMessage, err)
		}
		slog.Error("failed to create client", "error", err)
		return nil, fmt.Errorf("create client: %w", err)
	}
	return vlClient, nil
}
