package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/major/volumeleaders-go/volumeleaders"
	"github.com/major/volumeleaders-go/volumeleaders/browserauth"
)

func TestExtractCookiesUsesCachedCookies(t *testing.T) {
	cachePath := setTestCacheHome(t)

	want := map[string]string{
		"ASP.NET_SessionId": "cached-session",
		".ASPXAUTH":         "cached-auth",
	}
	if err := saveCacheFile(cachePath, want); err != nil {
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
	cachePath := setTestCacheHome(t)
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

	cached, err := loadCacheFile(cachePath)
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
	setTestCacheHome(t)
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

func setTestCacheHome(t *testing.T) string {
	t.Helper()

	cacheDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheDir)
	t.Setenv("HOME", cacheDir)
	t.Setenv("LOCALAPPDATA", cacheDir)
	t.Setenv("LocalAppData", cacheDir)

	path, err := cachePath()
	if err != nil {
		t.Fatalf("cachePath() error = %v", err)
	}
	return path
}
