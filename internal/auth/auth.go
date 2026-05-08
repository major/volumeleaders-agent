// Package auth extracts browser cookies and XSRF tokens needed to
// authenticate with the VolumeLeaders web application.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/major/volumeleaders-go/volumeleaders"
	"github.com/major/volumeleaders-go/volumeleaders/browserauth"
	"resty.dev/v3"
)

// UserAgent mimics Chrome 147 on Windows for authenticated requests.
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"

const volumeLeadersDomain = "volumeleaders.com"

// ErrSessionExpired marks auth failures caused by an expired browser session.
var ErrSessionExpired = errors.New("session expired")

// SessionExpiredMessage is the user-facing remediation for expired sessions.
const SessionExpiredMessage = "Authentication required: VolumeLeaders session has expired. Log in at https://www.volumeleaders.com in your browser, then retry."

var requiredCookieNames = []string{"ASP.NET_SessionId", ".ASPXAUTH"}

var xsrfTokenPattern = regexp.MustCompile(`<input\s+name="__RequestVerificationToken"\s+type="hidden"\s+value="([^"]+)"`)

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

type sessionExpiredError struct {
	redirectPath string
}

func (e sessionExpiredError) Error() string {
	return SessionExpiredMessage
}

func (e sessionExpiredError) Unwrap() error {
	return ErrSessionExpired
}

func (e sessionExpiredError) Detail() string {
	return fmt.Sprintf("requested host www.%s redirected to %s", volumeLeadersDomain, e.redirectPath)
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
	found := sessionCookies(session)
	if missing := missingRequiredCookies(found); len(missing) > 0 {
		return nil, fmt.Errorf(
			"required browser cookies unavailable: missing %s: %w",
			strings.Join(missing, ", "),
			browserauth.ErrRequiredCookieMissing,
		)
	}

	// Cache for subsequent runs (best-effort, failures are not fatal).
	_ = saveCache(found)

	return found, nil
}

// FetchXSRFToken retrieves the hidden request verification token from ExecutiveSummary.
func FetchXSRFToken(ctx context.Context, client *resty.Client) (string, error) {
	resp, err := client.R().SetContext(ctx).SetHeaders(BrowserHeaders).Get("https://www.volumeleaders.com/ExecutiveSummary")
	if err != nil {
		return "", fmt.Errorf("fetch XSRF token page: %w", err)
	}

	redirectPath := safeRedirectPath(resp.RawResponse)
	if normalizeRedirectPath(redirectPath) == "/login" {
		return "", sessionExpiredError{redirectPath: redirectPath}
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("fetch XSRF token page: status %d", resp.StatusCode())
	}

	matches := xsrfTokenPattern.FindSubmatch(resp.Bytes())
	if matches == nil {
		return "", fmt.Errorf("XSRF token not found in HTML")
	}
	return string(matches[1]), nil
}

func sessionCookies(session volumeleaders.Session) map[string]string {
	found := make(map[string]string, 3)
	for _, cookie := range session.Cookies {
		if cookie == nil {
			continue
		}
		switch cookie.Name {
		case "ASP.NET_SessionId", ".ASPXAUTH", "__RequestVerificationToken":
			found[cookie.Name] = cookie.Value
		}
	}
	return found
}

func missingRequiredCookies(found map[string]string) []string {
	missing := make([]string, 0, len(requiredCookieNames))
	for _, name := range requiredCookieNames {
		if found[name] == "" {
			missing = append(missing, name)
		}
	}
	return missing
}

func safeRedirectPath(resp *http.Response) string {
	if resp == nil || resp.Request == nil || resp.Request.URL == nil {
		return "unknown redirect target"
	}
	escapedPath := resp.Request.URL.EscapedPath()
	if escapedPath == "" {
		return "/"
	}
	return escapedPath
}

func normalizeRedirectPath(redirectPath string) string {
	if redirectPath == "" {
		return "/"
	}
	if !strings.HasPrefix(redirectPath, "/") {
		redirectPath = "/" + redirectPath
	}
	return strings.ToLower(path.Clean(redirectPath))
}
