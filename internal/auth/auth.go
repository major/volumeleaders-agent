// Package auth extracts browser cookies and XSRF tokens needed to
// authenticate with the VolumeLeaders web application.
package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/major/volumeleaders-go/volumeleaders"
	"github.com/major/volumeleaders-go/volumeleaders/browserauth"
)

// UserAgent mimics Chrome 147 on Windows for authenticated requests.
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"

// ErrSessionExpired marks auth failures caused by an expired browser session.
var ErrSessionExpired = errors.New("session expired")

// SessionExpiredMessage is the user-facing remediation for expired sessions.
const SessionExpiredMessage = "Authentication required: VolumeLeaders session has expired. Log in at https://www.volumeleaders.com in your browser, then retry."

//nolint:gochecknoglobals // Test seam for avoiding real browser stores in auth package tests.
var findBrowserSession = browserauth.FindSession

// BrowserHeaders contains the 9 browser-fingerprint headers that mimic Chrome 147 on Windows.
var BrowserHeaders = map[string]string{
	"User-Agent":         UserAgent,
	"Sec-Ch-Ua":          `"Chromium";v="147", "Not A(Brand";v="24", "Google Chrome";v="147"`,
	"Sec-Ch-Ua-Mobile":   "?0",
	"Sec-Ch-Ua-Platform": `"Windows"`,
	"Sec-Fetch-Dest":     "empty",
	"Sec-Fetch-Mode":     "cors",
	"Sec-Fetch-Site":     "same-origin",
	"Accept-Language":    "en-US,en;q=0.9",
	"Accept-Encoding":    "gzip, deflate, br",
}

// IsSessionExpired reports whether err indicates an expired VolumeLeaders session.
func IsSessionExpired(err error) bool {
	return errors.Is(err, ErrSessionExpired) || errors.Is(err, volumeleaders.ErrSessionExpired)
}

// ExtractCookies reads required VolumeLeaders cookies, checking a local
// cache first and falling back to browser cookie stores on cache miss.
// Successfully extracted cookies are cached so subsequent invocations avoid
// the cost of reading browser stores. Call InvalidateCache when the server
// reports a session expiry so the next call re-reads from browser stores.
func ExtractCookies(ctx context.Context) (map[string]string, error) {
	if cached, err := loadCache(); err == nil {
		return cached, nil
	}

	session, err := findBrowserSession(ctx, browserauth.WithoutValidation())
	if err != nil {
		return nil, fmt.Errorf("discover browser session: %w", err)
	}

	found := make(map[string]string, len(session.Cookies))
	for _, c := range session.Cookies {
		if c != nil {
			found[c.Name] = c.Value
		}
	}

	// Cache for subsequent runs (best-effort, failures are not fatal).
	_ = saveCache(found)

	return found, nil
}
