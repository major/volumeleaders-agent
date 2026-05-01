// Package client provides an authenticated HTTP client for the VolumeLeaders
// API, supporting DataTables, JSON, form, and multipart request formats.
package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/major/volumeleaders-agent/internal/auth"
	"github.com/major/volumeleaders-agent/internal/models"
	"resty.dev/v3"
)

// BaseURL is the VolumeLeaders web application origin.
const BaseURL = "https://www.volumeleaders.com"

// Client wraps authenticated VolumeLeaders HTTP access.
type Client struct {
	http             *http.Client
	client           *resty.Client
	noRedirectClient *resty.Client
	baseURL          string
	cookies          map[string]string
	xsrfToken        string
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
	restyClient.SetBaseURL(baseURL)
	restyClient.AddRequestMiddleware(buildRequestMiddleware("test-token"))
	restyClient.SetCookies(buildCookies(testCookies))

	noRedirectClient := resty.NewWithClient(httpClient)
	noRedirectClient.SetBaseURL(baseURL)
	noRedirectClient.SetRedirectPolicy(resty.NoRedirectPolicy())
	noRedirectClient.AddRequestMiddleware(buildRequestMiddleware("test-token"))
	noRedirectClient.SetCookies(buildCookies(testCookies))

	return &Client{
		http:             httpClient,
		client:           restyClient,
		noRedirectClient: noRedirectClient,
		baseURL:          baseURL,
		cookies:          testCookies,
		xsrfToken:        "test-token",
	}
}

// New creates an authenticated VolumeLeaders client from browser cookies.
func New(ctx context.Context) (*Client, error) {
	cookies, err := auth.ExtractCookies(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract cookies: %w", err)
	}

	restyClient := resty.New()
	restyClient.SetTimeout(60 * time.Second)

	// auth.FetchXSRFToken still expects the stdlib client during the migration.
	httpClient := restyClient.Client()
	xsrfToken, err := auth.FetchXSRFToken(ctx, httpClient, cookies)
	if err != nil {
		return nil, fmt.Errorf("fetch XSRF token: %w", err)
	}

	restyClient.SetBaseURL(BaseURL)
	restyClient.AddRequestMiddleware(buildRequestMiddleware(xsrfToken))
	restyClient.SetCookies(buildCookies(cookies))

	noRedirectClient := resty.New()
	noRedirectClient.SetTimeout(60 * time.Second)
	noRedirectClient.SetBaseURL(BaseURL)
	noRedirectClient.SetRedirectPolicy(resty.NoRedirectPolicy())
	noRedirectClient.AddRequestMiddleware(buildRequestMiddleware(xsrfToken))
	noRedirectClient.SetCookies(buildCookies(cookies))

	return &Client{
		http:             httpClient,
		client:           restyClient,
		noRedirectClient: noRedirectClient,
		baseURL:          BaseURL,
		cookies:          cookies,
		xsrfToken:        xsrfToken,
	}, nil
}

func buildRequestMiddleware(xsrfToken string) resty.RequestMiddleware {
	return func(_ *resty.Client, req *resty.Request) error {
		req.SetHeaders(map[string]string{
			"User-Agent":         auth.UserAgent,
			"x-xsrf-token":       xsrfToken,
			"x-requested-with":   "XMLHttpRequest",
			"Accept":             "application/json, text/javascript, */*; q=0.01",
			"Sec-Ch-Ua":          `"Chromium";v="147", "Not A(Brand";v="24", "Google Chrome";v="147"`,
			"Sec-Ch-Ua-Mobile":   "?0",
			"Sec-Ch-Ua-Platform": `"Windows"`,
			"Sec-Fetch-Dest":     "empty",
			"Sec-Fetch-Mode":     "cors",
			"Sec-Fetch-Site":     "same-origin",
			"Accept-Language":    "en-US,en;q=0.9",
			"Accept-Encoding":    "gzip, deflate, br",
		})
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
		result = append(result, &http.Cookie{Name: name, Value: value})
	}
	return result
}

// PostDataTablesPage posts a form-encoded DataTables request and returns the
// full response envelope, including RecordsFiltered for pagination decisions.
func (c *Client) PostDataTablesPage(ctx context.Context, path, body string) (*models.DataTablesResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create DataTables request: %w", err)
	}

	responseBody, err := c.doRequest(req, "DataTables")
	if err != nil {
		return nil, err
	}

	var wrapper models.DataTablesResponse
	if err := json.Unmarshal(responseBody, &wrapper); err != nil {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create JSON request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	responseBody, err := c.doRequest(req, "JSON")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(responseBody, result); err != nil {
		return fmt.Errorf("decode JSON response: %w", err)
	}
	return nil
}

// PostForm sends a form-encoded POST and decodes the JSON response directly
// (no DataTables envelope).
func (c *Client) PostForm(ctx context.Context, path string, values url.Values, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("create form request: %w", err)
	}

	responseBody, err := c.doRequest(req, "form")
	if err != nil {
		return err
	}

	if result != nil {
		if err := json.Unmarshal(responseBody, result); err != nil {
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, &buf)
	if err != nil {
		return fmt.Errorf("create multipart request: %w", err)
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.setCookies(req)

	noRedirect := &http.Client{
		Transport: c.http.Transport,
		Timeout:   c.http.Timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := noRedirect.Do(req)
	if err != nil {
		return fmt.Errorf("post multipart request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		responseBody, _ := readResponseBody(resp)
		return fmt.Errorf("post multipart request: status %d: %s", resp.StatusCode, string(responseBody))
	}
	return nil
}

// doRequest executes req with standard auth headers/cookies, reads the response
// body, and checks for non-200 status codes. Callers that need a non-default
// Content-Type (e.g. application/json) should set it on req before calling.
func (c *Client) doRequest(req *http.Request, label string) ([]byte, error) {
	c.setHeaders(req)
	c.setCookies(req)

	resp, err := c.http.Do(req) //nolint:gosec // baseURL is the hardcoded BaseURL constant, not user input
	if err != nil {
		return nil, fmt.Errorf("post %s request: %w", label, err)
	}
	defer resp.Body.Close()

	body, err := readResponseBody(resp)
	if err != nil {
		return nil, fmt.Errorf("read %s response: %w", label, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("post %s request: status %d: %s", label, resp.StatusCode, string(body))
	}
	return body, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("x-xsrf-token", c.xsrfToken)
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	// Only set default Content-Type if the caller hasn't already specified one.
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	}
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="147", "Not A(Brand";v="24", "Google Chrome";v="147"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
}

func (c *Client) setCookies(req *http.Request) {
	for name, value := range c.cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("open gzip response body: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	return body, nil
}
