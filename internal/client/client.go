// Package client provides an authenticated HTTP client for the VolumeLeaders
// API. It handles browser cookie extraction, XSRF token probing, and
// builds session material plus an HTTP client consumed by volumeleaders-go.
package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/major/volumeleaders-agent/internal/auth"
	vlgo "github.com/major/volumeleaders-go/volumeleaders"
	"resty.dev/v3"
)

// BaseURL is the VolumeLeaders web application origin.
const BaseURL = "https://www.volumeleaders.com"

// Client wraps authenticated VolumeLeaders HTTP access.
type Client struct {
	client           *resty.Client
	noRedirectClient *resty.Client
	baseURL          string
	cookies          map[string]string
	xsrfToken        string
	closeOnce        sync.Once
	closeErr         error
}

// NewForTesting creates a Client for test use, bypassing browser-based
// authentication. Callers provide their own http.Client (typically backed
// by httptest.Server) and base URL.
func NewForTesting(httpClient *http.Client, baseURL string) *Client {
	testCookies := map[string]string{
		"ASP.NET_SessionId": "test-session",
		".ASPXAUTH":         "test-auth",
	}
	restyClient := resty.NewWithClient(httpClient)
	configureClient(restyClient, baseURL, "test-token")
	restyClient.SetCookies(buildCookies(testCookies))

	noRedirectClient := resty.NewWithClient(httpClient)
	configureClient(noRedirectClient, baseURL, "test-token")
	noRedirectClient.SetRedirectPolicy(resty.NoRedirectPolicy())
	noRedirectClient.SetCookies(buildCookies(testCookies))

	return &Client{
		client:           restyClient,
		noRedirectClient: noRedirectClient,
		baseURL:          baseURL,
		cookies:          testCookies,
		xsrfToken:        "test-token",
	}
}

// New creates an authenticated VolumeLeaders client from browser cookies.
//
// Cached cookies are tried first. If the XSRF token fetch detects a
// session expiry, the cache is invalidated and fresh cookies are
// extracted from browser stores before failing.
func New(ctx context.Context) (*Client, error) {
	cookies, xsrfToken, err := authenticate(ctx)
	if err != nil {
		return nil, err
	}

	restyClient := resty.New()
	restyClient.SetTimeout(60 * time.Second)
	configureClient(restyClient, BaseURL, xsrfToken)
	restyClient.SetCookies(buildCookies(cookies))

	noRedirectClient := resty.New()
	noRedirectClient.SetTimeout(60 * time.Second)
	configureClient(noRedirectClient, BaseURL, xsrfToken)
	noRedirectClient.SetRedirectPolicy(resty.NoRedirectPolicy())
	noRedirectClient.SetCookies(buildCookies(cookies))

	return &Client{
		client:           restyClient,
		noRedirectClient: noRedirectClient,
		baseURL:          BaseURL,
		cookies:          cookies,
		xsrfToken:        xsrfToken,
	}, nil
}

// authenticate extracts cookies and fetches an XSRF token. When the
// first attempt uses cached cookies that turn out to be stale, the
// cache is cleared and a second attempt reads fresh cookies directly
// from browser stores.
func authenticate(ctx context.Context) (cookies map[string]string, xsrfToken string, err error) {
	cookies, err = auth.ExtractCookies(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("extract cookies: %w", err)
	}

	xsrfToken, err = probeXSRFToken(ctx, cookies)
	if err != nil && auth.IsSessionExpired(err) {
		// Cached cookies may be stale. Clear cache and retry with
		// fresh browser cookies.
		_ = auth.InvalidateCache()

		cookies, err = auth.ExtractCookies(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("extract fresh cookies: %w", err)
		}

		xsrfToken, err = probeXSRFToken(ctx, cookies)
	}
	if err != nil {
		return nil, "", fmt.Errorf("fetch XSRF token: %w", err)
	}

	return cookies, xsrfToken, nil
}

// probeXSRFToken fetches the XSRF token using the volumeleaders-go
// library. The library handles the ExecutiveSummary page request,
// login-redirect detection, and token parsing.
func probeXSRFToken(ctx context.Context, cookies map[string]string) (string, error) {
	session := vlgo.SessionFromCookies(buildCookies(cookies), "", nil)
	return vlgo.FetchXSRFToken(ctx, session)
}

// Close releases resources held by both resty clients.
// It is safe to call Close more than once.
func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		c.closeErr = errors.Join(c.client.Close(), c.noRedirectClient.Close())
	})
	return c.closeErr
}

// configureClient applies the base URL and request middleware shared by all
// resty clients. Cookies are set separately by each constructor to avoid
// duplicate appends (resty's SetCookies appends rather than replaces).
func configureClient(c *resty.Client, baseURL, xsrfToken string) {
	c.SetBaseURL(baseURL)
	c.AddRequestMiddleware(buildRequestMiddleware(xsrfToken))
}

func buildRequestMiddleware(xsrfToken string) resty.RequestMiddleware {
	return func(_ *resty.Client, req *resty.Request) error {
		// Set shared browser headers
		for k, v := range auth.BrowserHeaders {
			req.SetHeader(k, v)
		}
		// Set API-specific headers
		req.SetHeader("x-xsrf-token", xsrfToken)
		req.SetHeader("x-requested-with", "XMLHttpRequest")
		req.SetHeader("Accept", "application/json, text/javascript, */*; q=0.01")
		// Only set default Content-Type if the caller hasn't already specified one.
		if req.Header.Get("Content-Type") == "" {
			req.SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
		}
		return nil
	}
}

func buildCookies(cookies map[string]string) []*http.Cookie {
	result := make([]*http.Cookie, 0, len(cookies))
	for name, value := range cookies {
		result = append(result, &http.Cookie{
			Name:     name,
			Value:    value,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	return result
}


