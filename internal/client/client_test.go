package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/major/volumeleaders-agent/internal/auth"
)

func newTestClient(baseURL string) *Client {
	httpClient := &http.Client{Timeout: 5 * time.Second}
	return NewForTesting(httpClient, baseURL)
}

func TestMiddlewareHeaders(t *testing.T) {
	t.Parallel()

	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(server.URL)
	if _, err := client.client.R().Post("/test"); err != nil {
		t.Fatalf("post through Resty client: %v", err)
	}

	checks := map[string]string{
		"User-Agent":         auth.UserAgent,
		"x-xsrf-token":       "test-token",
		"x-requested-with":   "XMLHttpRequest",
		"Accept":             "application/json, text/javascript, */*; q=0.01",
		"Content-Type":       "application/x-www-form-urlencoded; charset=UTF-8",
		"Sec-Ch-Ua":          `"Chromium";v="147", "Not A(Brand";v="24", "Google Chrome";v="147"`,
		"Sec-Ch-Ua-Mobile":   "?0",
		"Sec-Ch-Ua-Platform": `"Windows"`,
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-origin",
		"Accept-Language":    "en-US,en;q=0.9",
		"Accept-Encoding":    "gzip, deflate, br",
	}
	for key, expected := range checks {
		if got := capturedHeaders.Get(key); got != expected {
			t.Errorf("%s: expected %q, got %q", key, expected, got)
		}
	}
}

func TestMiddlewareCookies(t *testing.T) {
	t.Parallel()

	capturedCookies := make(map[string]string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, cookie := range r.Cookies() {
			capturedCookies[cookie.Name] = cookie.Value
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`)
	}))
	t.Cleanup(server.Close)

	client := newTestClient(server.URL)
	if _, err := client.client.R().Post("/test"); err != nil {
		t.Fatalf("post through Resty client: %v", err)
	}

	checks := map[string]string{
		"ASP.NET_SessionId": "test-session",
		".ASPXAUTH":         "test-auth",
	}
	for name, expected := range checks {
		if got := capturedCookies[name]; got != expected {
			t.Errorf("cookie %s: expected %q, got %q", name, expected, got)
		}
	}
}

func TestNewForTesting(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{Timeout: 5 * time.Second}
	c := NewForTesting(httpClient, "http://test.example")
	if c.baseURL != "http://test.example" {
		t.Errorf("expected baseURL http://test.example, got %s", c.baseURL)
	}
	if c.xsrfToken != "test-token" {
		t.Errorf("expected xsrfToken test-token, got %s", c.xsrfToken)
	}
	if c.client == nil {
		t.Errorf("expected Resty client to be populated")
	}
	if c.noRedirectClient == nil {
		t.Errorf("expected no-redirect Resty client to be populated")
	}
}

func TestClose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		calls   int
		wantErr bool
	}{
		{
			name:    "single close returns nil",
			calls:   1,
			wantErr: false,
		},
		{
			name:    "double close does not panic",
			calls:   2,
			wantErr: false,
		},
		{
			name:    "triple close does not panic",
			calls:   3,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient("http://localhost")

			var firstErr error
			for i := 0; i < tt.calls; i++ {
				err := client.Close()
				if i == 0 {
					firstErr = err
				} else {
					// Verify all calls return the same error value (proving closeErr caching works)
					if (err == nil && firstErr != nil) || (err != nil && firstErr == nil) {
						t.Errorf("Close() call %d: got %v, want %v", i+1, err, firstErr)
					}
					if err != nil && firstErr != nil && err.Error() != firstErr.Error() {
						t.Errorf("Close() call %d: got %v, want %v", i+1, err, firstErr)
					}
				}
			}

			if tt.wantErr && firstErr == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && firstErr != nil {
				t.Errorf("expected nil, got error: %v", firstErr)
			}
		})
	}
}
