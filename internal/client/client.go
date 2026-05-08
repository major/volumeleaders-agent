// Package client provides an authenticated HTTP client for the VolumeLeaders
// API, supporting DataTables, JSON, form, and multipart request formats.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/major/volumeleaders-agent/internal/auth"
	"github.com/major/volumeleaders-agent/internal/models"
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

// probeXSRFToken creates a temporary resty client to fetch the XSRF
// token without configuring the full request middleware. The throwaway
// client avoids resty's append-only cookie behavior that would cause
// duplicates on retry.
func probeXSRFToken(ctx context.Context, cookies map[string]string) (string, error) {
	probe := resty.New()
	probe.SetTimeout(60 * time.Second)
	probe.SetCookies(buildCookies(cookies))
	defer probe.Close()

	return auth.FetchXSRFToken(ctx, probe)
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

// PostDataTablesPage posts a form-encoded DataTables request and returns the
// full response envelope, including RecordsFiltered for pagination decisions.
func (c *Client) PostDataTablesPage(ctx context.Context, path, body string) (*models.DataTablesResponse, error) {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(body).
		SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8").
		Post(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("post DataTables request: %w", err)
	}
	if resp.Err != nil {
		return nil, fmt.Errorf("post DataTables request: %w", resp.Err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("post DataTables request: status %d: %s", resp.StatusCode(), resp.String())
	}

	var wrapper models.DataTablesResponse
	if err := json.Unmarshal(resp.Bytes(), &wrapper); err != nil {
		return nil, fmt.Errorf("decode DataTables response: %w", err)
	}
	if len(wrapper.Data) == 0 || bytes.Equal(wrapper.Data, []byte("null")) {
		wrapper.Data = []byte("[]")
	}
	return &wrapper, nil
}

// PostDataTables posts a form-encoded DataTables request and unmarshals its data array.
func (c *Client) PostDataTables(ctx context.Context, path, body string, result any) error {
	wrapper, err := c.PostDataTablesPage(ctx, path, body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(wrapper.Data, result); err != nil {
		return fmt.Errorf("decode DataTables data: %w", err)
	}
	return nil
}

// PostJSON posts a JSON body and decodes the JSON response.
func (c *Client) PostJSON(ctx context.Context, path string, payload, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal JSON request: %w", err)
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(body).
		SetHeader("Content-Type", "application/json").
		Post(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("post JSON request: %w", err)
	}
	if resp.Err != nil {
		return fmt.Errorf("post JSON request: %w", resp.Err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("post JSON request: status %d: %s", resp.StatusCode(), resp.String())
	}

	if err := json.Unmarshal(resp.Bytes(), result); err != nil {
		return fmt.Errorf("decode JSON response: %w", err)
	}
	return nil
}

// PostForm sends a form-encoded POST and decodes the JSON response directly
// (no DataTables envelope).
func (c *Client) PostForm(ctx context.Context, path string, values url.Values, result any) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(values.Encode()).
		SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8").
		Post(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("post form request: %w", err)
	}
	if resp.Err != nil {
		return fmt.Errorf("post form request: %w", resp.Err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("post form request: status %d: %s", resp.StatusCode(), resp.String())
	}

	if result != nil {
		if err := json.Unmarshal(resp.Bytes(), result); err != nil {
			return fmt.Errorf("decode form response: %w", err)
		}
	}
	return nil
}

// PostMultipart sends a multipart/form-data POST. ASP.NET MVC returns 302 on
// successful form submissions, so redirects are not followed and any 2xx/3xx
// status is treated as success.
func (c *Client) PostMultipart(ctx context.Context, path string, fields map[string]string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, val := range fields {
		if err := writer.WriteField(key, val); err != nil {
			return fmt.Errorf("write multipart field %s: %w", key, err)
		}
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}

	resp, err := c.noRedirectClient.R().
		SetContext(ctx).
		SetBody(buf.Bytes()).
		SetHeader("Content-Type", writer.FormDataContentType()).
		Post(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("post multipart request: %w", err)
	}
	if resp.Err != nil {
		return fmt.Errorf("post multipart request: %w", resp.Err)
	}
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("post multipart request: status %d: %s", resp.StatusCode(), resp.String())
	}
	return nil
}
