package common

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	vlgo "github.com/major/volumeleaders-go/volumeleaders"

	"github.com/major/volumeleaders-agent/internal/auth"
)

// NewVLClient returns a volumeleaders-go client for the current command
// context. Tests inject a pre-built client via TestVLClientKey; production
// callers fall through to the browser-authenticated internal client bridge.
func NewVLClient(ctx context.Context) (*vlgo.Client, error) {
	if c, ok := ctx.Value(TestVLClientKey).(*vlgo.Client); ok {
		return c, nil
	}

	commandClient, err := NewCommandClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create command client: %w", err)
	}

	vlClient, err := commandClient.NewVolumeLeadersClient()
	if err != nil {
		if auth.IsSessionExpired(err) {
			var detail interface{ Detail() string }
			if errors.As(err, &detail) {
				slog.Debug("VolumeLeaders session expired during vlgo bridge", "detail", detail.Detail())
			}
			return nil, fmt.Errorf("%s: %w", auth.SessionExpiredMessage, err)
		}
		return nil, fmt.Errorf("create volumeleaders-go client: %w", err)
	}
	return vlClient, nil
}
