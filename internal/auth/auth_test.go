package auth

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestXSRFTokenPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		html      string
		wantToken string
		wantMatch bool
	}{
		{
			name:      "matches token input",
			html:      `<input name="__RequestVerificationToken" type="hidden" value="token-123" />`,
			wantToken: "token-123",
			wantMatch: true,
		},
		{
			name:      "matches extra whitespace",
			html:      `<input   name="__RequestVerificationToken"   type="hidden"   value="token-with-space">`,
			wantToken: "token-with-space",
			wantMatch: true,
		},
		{
			name:      "does not match different attribute order",
			html:      `<input type="hidden" name="__RequestVerificationToken" value="token-123" />`,
			wantMatch: false,
		},
		{
			name:      "does not match missing token input",
			html:      `<input name="other" type="hidden" value="token-123" />`,
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matches := xsrfTokenPattern.FindStringSubmatch(tt.html)
			if tt.wantMatch && matches == nil {
				t.Fatalf("expected token match")
			}
			if !tt.wantMatch && matches != nil {
				t.Fatalf("expected no token match, got %q", matches[1])
			}
			if tt.wantMatch && matches[1] != tt.wantToken {
				t.Errorf("expected token %q, got %q", tt.wantToken, matches[1])
			}
		})
	}
}

func TestFetchXSRFToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		handler   http.HandlerFunc
		wantToken string
		wantErr   string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assertBrowserHeaders(t, r)
				assertRequestCookies(t, r)
				fmt.Fprint(w, `<input name="__RequestVerificationToken" type="hidden" value="token-123" />`)
			},
			wantToken: "token-123",
		},
		{
			name: "session expired redirect",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/ExecutiveSummary" {
					http.Redirect(w, r, "/Login", http.StatusFound)
					return
				}
				fmt.Fprint(w, "login")
			},
			wantErr: "session expired",
		},
		{
			name: "non 200 status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "forbidden", http.StatusForbidden)
			},
			wantErr: "status 403",
		},
		{
			name: "gzip response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Encoding", "gzip")
				gz := gzip.NewWriter(w)
				defer gz.Close()
				fmt.Fprint(gz, `<input name="__RequestVerificationToken" type="hidden" value="gzip-token" />`)
			},
			wantToken: "gzip-token",
		},
		{
			name: "missing token",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprint(w, `<html><body>No token here</body></html>`)
			},
			wantErr: "XSRF token not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(tt.handler)
			t.Cleanup(server.Close)

			client := server.Client()
			client.Transport = rewriteHostTransport{base: client.Transport, target: server.URL}

			token, err := FetchXSRFToken(t.Context(), client, map[string]string{
				"ASP.NET_SessionId": "session-cookie",
				".ASPXAUTH":         "auth-cookie",
			})
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token != tt.wantToken {
				t.Errorf("expected token %q, got %q", tt.wantToken, token)
			}
		})
	}
}

func TestFetchXSRFToken_CanceledContext(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<input name="__RequestVerificationToken" type="hidden" value="token-123" />`)
	}))
	t.Cleanup(server.Close)

	client := server.Client()
	client.Transport = rewriteHostTransport{base: client.Transport, target: server.URL}

	ctx, cancel := context.WithCancel(t.Context())
	cancel() // cancel immediately

	_, err := FetchXSRFToken(ctx, client, map[string]string{
		"ASP.NET_SessionId": "session-cookie",
		".ASPXAUTH":         "auth-cookie",
	})
	if err == nil {
		t.Fatal("expected error from canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

type rewriteHostTransport struct {
	base   http.RoundTripper
	target string
}

func (t rewriteHostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	targetReq := req.Clone(req.Context())
	targetURL := *req.URL
	serverURL := strings.TrimPrefix(t.target, "http://")
	targetURL.Scheme = "http"
	targetURL.Host = serverURL
	targetReq.URL = &targetURL
	targetReq.Host = req.URL.Host
	return t.base.RoundTrip(targetReq)
}

func assertBrowserHeaders(t *testing.T, r *http.Request) {
	t.Helper()

	checks := map[string]string{
		"User-Agent":        UserAgent,
		"Sec-Ch-Ua":         `"Chromium";v="147", "Not A(Brand";v="24", "Google Chrome";v="147"`,
		"Sec-Ch-Ua-Mobile":  "?0",
		"Sec-Ch-Ua-Platform": `"Windows"`,
		"Sec-Fetch-Dest":    "empty",
		"Sec-Fetch-Mode":    "cors",
		"Sec-Fetch-Site":    "same-origin",
		"Accept-Language":   "en-US,en;q=0.9",
		"Accept-Encoding":   "gzip, deflate, br",
	}
	for key, expected := range checks {
		if got := r.Header.Get(key); got != expected {
			t.Errorf("%s: expected %q, got %q", key, expected, got)
		}
	}
}

func assertRequestCookies(t *testing.T, r *http.Request) {
	t.Helper()

	checks := map[string]string{
		"ASP.NET_SessionId": "session-cookie",
		".ASPXAUTH":         "auth-cookie",
	}
	for name, expected := range checks {
		cookie, err := r.Cookie(name)
		if err != nil {
			t.Errorf("missing cookie %s: %v", name, err)
			continue
		}
		if cookie.Value != expected {
			t.Errorf("cookie %s: expected %q, got %q", name, expected, cookie.Value)
		}
	}
}
