package auth

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/major/volumeleaders-go/volumeleaders"
	"github.com/major/volumeleaders-go/volumeleaders/browserauth"
	"resty.dev/v3"
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
		name               string
		handler            http.HandlerFunc
		wantToken          string
		wantErr            string
		wantSessionExpired bool
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
			wantErr:            SessionExpiredMessage,
			wantSessionExpired: true,
		},
		{
			name: "non login redirect does not mark session expired",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/ExecutiveSummary" {
					http.Redirect(w, r, "/NotLogin", http.StatusFound)
					return
				}
				fmt.Fprint(w, "not login")
			},
			wantErr: "XSRF token not found",
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
			restyClient := resty.NewWithClient(client)
			restyClient.SetCookies([]*http.Cookie{
				{Name: "ASP.NET_SessionId", Value: "session-cookie"},
				{Name: ".ASPXAUTH", Value: "auth-cookie"},
			})

			token, err := FetchXSRFToken(t.Context(), restyClient)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				if got := IsSessionExpired(err); got != tt.wantSessionExpired {
					t.Fatalf("IsSessionExpired() = %t, want %t", got, tt.wantSessionExpired)
				}
				if tt.wantSessionExpired && strings.Contains(err.Error(), "/Login") {
					t.Fatalf("session expired error exposed redirect detail: %v", err)
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

func TestExtractCookiesUsesCachedCookies(t *testing.T) {
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)

	want := map[string]string{
		"ASP.NET_SessionId": "cached-session",
		".ASPXAUTH":         "cached-auth",
	}
	if err := saveCacheFile(filepath.Join(cacheDir, cacheSubdir, cacheFileName), want); err != nil {
		t.Fatalf("saveCacheFile() error = %v", err)
	}
	stubFindBrowserSession(t, func(context.Context, ...browserauth.Option) (volumeleaders.Session, error) {
		t.Fatal("ExtractCookies() called browserauth despite valid cache")
		return volumeleaders.Session{}, nil
	})

	got, err := ExtractCookies(t.Context())
	if err != nil {
		t.Fatalf("ExtractCookies() error = %v", err)
	}
	for name, wantValue := range want {
		if got[name] != wantValue {
			t.Errorf("ExtractCookies()[%s] = %q, want %q", name, got[name], wantValue)
		}
	}
}

func TestExtractCookiesFindsSessionAndCachesCookies(t *testing.T) {
	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)
	stubFindBrowserSession(t, func(_ context.Context, opts ...browserauth.Option) (volumeleaders.Session, error) {
		if len(opts) != 1 {
			t.Fatalf("ExtractCookies() browserauth options = %d, want 1", len(opts))
		}
		return volumeleaders.SessionFromCookies([]*http.Cookie{
			{Name: "ASP.NET_SessionId", Value: "session-value"},
			{Name: ".ASPXAUTH", Value: "auth-value"},
			{Name: "__RequestVerificationToken", Value: "xsrf-cookie"},
		}, "xsrf-token", nil), nil
	})

	got, err := ExtractCookies(t.Context())
	if err != nil {
		t.Fatalf("ExtractCookies() error = %v", err)
	}
	want := map[string]string{
		"ASP.NET_SessionId":          "session-value",
		".ASPXAUTH":                  "auth-value",
		"__RequestVerificationToken": "xsrf-cookie",
	}
	for name, wantValue := range want {
		if got[name] != wantValue {
			t.Errorf("ExtractCookies()[%s] = %q, want %q", name, got[name], wantValue)
		}
	}

	cached, err := loadCacheFile(filepath.Join(cacheDir, cacheSubdir, cacheFileName))
	if err != nil {
		t.Fatalf("loadCacheFile() error = %v", err)
	}
	for name, wantValue := range want {
		if cached[name] != wantValue {
			t.Errorf("cached cookie %s = %q, want %q", name, cached[name], wantValue)
		}
	}
}

func TestExtractCookiesPreservesBrowserAuthErrorSemantics(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	stubFindBrowserSession(t, func(context.Context, ...browserauth.Option) (volumeleaders.Session, error) {
		return volumeleaders.Session{}, errors.Join(
			browserauth.ErrRequiredCookieMissing,
			volumeleaders.ErrBrowserCookiesUnavailable,
		)
	})

	_, err := ExtractCookies(t.Context())
	if err == nil {
		t.Fatal("ExtractCookies() error = nil, want browserauth failure")
	}
	if !errors.Is(err, browserauth.ErrRequiredCookieMissing) {
		t.Errorf("ExtractCookies() error = %v, want ErrRequiredCookieMissing", err)
	}
	if !errors.Is(err, volumeleaders.ErrBrowserCookiesUnavailable) {
		t.Errorf("ExtractCookies() error = %v, want ErrBrowserCookiesUnavailable", err)
	}
}

func TestIsSessionExpiredRecognizesLibraryError(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("browserauth validation: %w", volumeleaders.ErrSessionExpired)
	if !IsSessionExpired(err) {
		t.Fatalf("IsSessionExpired(%v) = false, want true", err)
	}
}

func TestSafeRedirectPath(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "https://www.volumeleaders.com/Login?returnUrl=%2FAccount", http.NoBody)
	got := safeRedirectPath(&http.Response{Request: req})
	if got != "/Login" {
		t.Fatalf("expected sanitized redirect path, got %q", got)
	}
}

func TestNormalizeRedirectPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "login path lowercased",
			path: "/Login",
			want: "/login",
		},
		{
			name: "missing leading slash",
			path: "Login",
			want: "/login",
		},
		{
			name: "clean path",
			path: "/Account/../Login",
			want: "/login",
		},
		{
			name: "non login remains exact path",
			path: "/NotLogin",
			want: "/notlogin",
		},
		{
			name: "empty path becomes root",
			path: "",
			want: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeRedirectPath(tt.path)
			if got != tt.want {
				t.Fatalf("normalizeRedirectPath(%q) = %q, want %q", tt.path, got, tt.want)
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
	restyClient := resty.NewWithClient(client)
	restyClient.SetCookies([]*http.Cookie{
		{Name: "ASP.NET_SessionId", Value: "session-cookie"},
		{Name: ".ASPXAUTH", Value: "auth-cookie"},
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel() // cancel immediately

	_, err := FetchXSRFToken(ctx, restyClient)
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

	for key, expected := range BrowserHeaders {
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

func stubFindBrowserSession(
	t *testing.T,
	stub func(context.Context, ...browserauth.Option) (volumeleaders.Session, error),
) {
	t.Helper()
	original := findBrowserSession
	findBrowserSession = stub
	t.Cleanup(func() {
		findBrowserSession = original
	})
}
