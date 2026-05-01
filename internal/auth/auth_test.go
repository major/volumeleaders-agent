package auth

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/browserutils/kooky"
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
		{
			name: "invalid gzip response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Encoding", "gzip")
				fmt.Fprint(w, "not gzip")
			},
			wantErr: "decompress XSRF token page",
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

func TestSessionExpiredErrorDetail(t *testing.T) {
	t.Parallel()

	err := sessionExpiredError{redirectPath: "/Login"}
	if got := err.Detail(); got != "requested host www.volumeleaders.com redirected to /Login" {
		t.Fatalf("Detail() = %q", got)
	}
}

func TestAuthCookies(t *testing.T) {
	t.Parallel()

	cookies := kooky.Cookies{
		{Cookie: http.Cookie{Name: "ASP.NET_SessionId", Value: "session-cookie"}},
		{Cookie: http.Cookie{Name: ".ASPXAUTH", Value: "auth-cookie"}},
		{Cookie: http.Cookie{Name: "__RequestVerificationToken", Value: "cookie-token"}},
		{Cookie: http.Cookie{Name: "unrelated", Value: "ignored"}},
	}

	got := authCookies(cookies)
	want := map[string]string{
		"ASP.NET_SessionId":          "session-cookie",
		".ASPXAUTH":                  "auth-cookie",
		"__RequestVerificationToken": "cookie-token",
	}
	if len(got) != len(want) {
		t.Fatalf("authCookies() len = %d, want %d: %v", len(got), len(want), got)
	}
	for name, wantValue := range want {
		if got[name] != wantValue {
			t.Fatalf("authCookies()[%q] = %q, want %q", name, got[name], wantValue)
		}
	}
}

func TestCookieDiagnostic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		found        map[string]string
		allCookies   kooky.Cookies
		validCookies kooky.Cookies
		wantParts    []string
		forbidParts  []string
	}{
		{
			name:  "missing all required cookies",
			found: map[string]string{},
			wantParts: []string{
				`searched browser cookie stores for domain suffix "volumeleaders.com"`,
				"required cookies: ASP.NET_SessionId, .ASPXAUTH",
				"valid VolumeLeaders cookies found: 0",
				"browser stores with VolumeLeaders cookies: 0",
				"missing valid cookies: ASP.NET_SessionId, .ASPXAUTH",
				"only cookie storage is inspected; local storage, session storage, and IndexedDB are not inspected",
			},
		},
		{
			name: "reports required cookies not usable as valid cookies",
			found: map[string]string{
				"ASP.NET_SessionId": "valid-session-cookie",
			},
			allCookies: kooky.Cookies{
				cookieWithBrowser("ASP.NET_SessionId", "valid-session-cookie", "Firefox", "default-release"),
				cookieWithBrowser(".ASPXAUTH", "expired-auth-cookie", "Firefox", "default-release"),
			},
			validCookies: kooky.Cookies{
				cookieWithBrowser("ASP.NET_SessionId", "valid-session-cookie", "Firefox", "default-release"),
			},
			wantParts: []string{
				"valid VolumeLeaders cookies found: 1",
				"browser stores with VolumeLeaders cookies: 1",
				"missing valid cookies: .ASPXAUTH",
				"matching required cookies found but not usable as valid cookies: .ASPXAUTH",
			},
			forbidParts: []string{
				"valid-session-cookie",
				"expired-auth-cookie",
				"default-release",
				"/home/",
			},
		},
		{
			name:  "counts cookies without browser metadata as one unknown store",
			found: map[string]string{},
			allCookies: kooky.Cookies{
				{Cookie: http.Cookie{Name: "ASP.NET_SessionId", Value: "expired-session-cookie"}},
				{Cookie: http.Cookie{Name: ".ASPXAUTH", Value: "expired-auth-cookie"}},
			},
			wantParts: []string{
				"browser stores with VolumeLeaders cookies: 1",
				"matching required cookies found but not usable as valid cookies: ASP.NET_SessionId, .ASPXAUTH",
			},
			forbidParts: []string{
				"expired-session-cookie",
				"expired-auth-cookie",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			diagnostic := cookieDiagnostic(tt.found, tt.allCookies, tt.validCookies)
			for _, want := range tt.wantParts {
				if !strings.Contains(diagnostic, want) {
					t.Errorf("expected diagnostic to contain %q, got %q", want, diagnostic)
				}
			}
			for _, forbidden := range tt.forbidParts {
				if strings.Contains(diagnostic, forbidden) {
					t.Errorf("expected diagnostic not to contain %q, got %q", forbidden, diagnostic)
				}
			}
		})
	}
}

func TestSafeRedirectPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		resp *http.Response
		want string
	}{
		{
			name: "sanitizes query from redirect URL",
			resp: &http.Response{Request: httptest.NewRequest(http.MethodGet, "https://www.volumeleaders.com/Login?returnUrl=%2FAccount", http.NoBody)},
			want: "/Login",
		},
		{
			name: "unknown when response is nil",
			resp: nil,
			want: "unknown redirect target",
		},
		{
			name: "unknown when request is nil",
			resp: &http.Response{},
			want: "unknown redirect target",
		},
		{
			name: "root path remains slash",
			resp: &http.Response{Request: httptest.NewRequest(http.MethodGet, "https://www.volumeleaders.com", http.NoBody)},
			want: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := safeRedirectPath(tt.resp); got != tt.want {
				t.Fatalf("safeRedirectPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetBrowserHeadersOverwritesExistingValues(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "https://www.volumeleaders.com/ExecutiveSummary", bytes.NewReader(nil))
	req.Header.Set("User-Agent", "old-agent")
	req.Header.Set("Accept-Encoding", "identity")

	setBrowserHeaders(req)

	if got := req.Header.Get("User-Agent"); got != UserAgent {
		t.Fatalf("User-Agent = %q, want %q", got, UserAgent)
	}
	if got := req.Header.Get("Accept-Encoding"); got != "gzip, deflate, br" {
		t.Fatalf("Accept-Encoding = %q", got)
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

type testBrowserInfo struct {
	browser string
	profile string
}

func (b testBrowserInfo) Browser() string { return b.browser }

func (b testBrowserInfo) Profile() string { return b.profile }

func (b testBrowserInfo) IsDefaultProfile() bool { return true }

func (b testBrowserInfo) FilePath() string {
	return "/home/example/.mozilla/firefox/profile/cookies.sqlite"
}

func cookieWithBrowser(name, value, browser, profile string) *kooky.Cookie {
	return &kooky.Cookie{
		Cookie: http.Cookie{
			Name:  name,
			Value: value,
		},
		Browser: testBrowserInfo{browser: browser, profile: profile},
	}
}
