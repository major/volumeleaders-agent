package client

import (
	"fmt"

	vlgo "github.com/major/volumeleaders-go/volumeleaders"
)

// NewVolumeLeadersClient builds a volumeleaders-go client from this client's
// existing authenticated browser session material. Keeping this bridge in the
// client package lets command packages adopt volumeleaders-go endpoint by
// endpoint without duplicating cookie or XSRF handling.
func (c *Client) NewVolumeLeadersClient() (*vlgo.Client, error) {
	session := vlgo.SessionFromCookies(buildCookies(c.cookies), c.xsrfToken, nil)
	if err := vlgo.ValidateSession(session); err != nil {
		return nil, fmt.Errorf("validate volumeleaders-go session: %w", err)
	}
	vlClient, err := vlgo.NewClient(
		session,
		vlgo.WithHTTPClient(c.client.Client()),
		vlgo.WithBaseURL(c.baseURL),
	)
	if err != nil {
		return nil, fmt.Errorf("create volumeleaders-go client: %w", err)
	}
	return vlClient, nil
}
