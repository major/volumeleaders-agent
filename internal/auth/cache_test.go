package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCacheRoundTrip(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "cookies.json")
	cookies := map[string]string{
		"ASP.NET_SessionId": "session-value",
		".ASPXAUTH":         "auth-value",
	}

	if err := saveCacheFile(path, cookies); err != nil {
		t.Fatalf("saveCacheFile: %v", err)
	}

	got, err := loadCacheFile(path)
	if err != nil {
		t.Fatalf("loadCacheFile: %v", err)
	}

	for name, want := range cookies {
		if got[name] != want {
			t.Errorf("cookie %s: got %q, want %q", name, got[name], want)
		}
	}
}

func TestCacheFilePermissions(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "subdir", "cookies.json")
	cookies := map[string]string{
		"ASP.NET_SessionId": "session-value",
		".ASPXAUTH":         "auth-value",
	}

	if err := saveCacheFile(path, cookies); err != nil {
		t.Fatalf("saveCacheFile: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat cache file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("cache file permissions: got %o, want 0600", perm)
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat cache directory: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0o700 {
		t.Errorf("cache directory permissions: got %o, want 0700", perm)
	}
}

func TestLoadCacheMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nonexistent.json")
	_, err := loadCacheFile(path)
	if err == nil {
		t.Fatal("expected error for missing cache file")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("expected os.ErrNotExist, got: %v", err)
	}
}

func TestLoadCacheCorruptJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "cookies.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}

	_, err := loadCacheFile(path)
	if err == nil {
		t.Fatal("expected error for corrupt JSON")
	}
	if !strings.Contains(err.Error(), "decode cookie cache") {
		t.Fatalf("expected decode error, got: %v", err)
	}
}

func TestLoadCacheMissingRequiredCookies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cookies map[string]string
	}{
		{
			name:    "missing both",
			cookies: map[string]string{"other": "value"},
		},
		{
			name:    "missing session ID",
			cookies: map[string]string{".ASPXAUTH": "auth"},
		},
		{
			name:    "missing auth cookie",
			cookies: map[string]string{"ASP.NET_SessionId": "session"},
		},
		{
			name:    "empty session ID value",
			cookies: map[string]string{"ASP.NET_SessionId": "", ".ASPXAUTH": "auth"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "cookies.json")
			if err := saveCacheFile(path, tt.cookies); err != nil {
				t.Fatalf("saveCacheFile: %v", err)
			}

			_, err := loadCacheFile(path)
			if err == nil {
				t.Fatal("expected error for missing required cookies")
			}
			if !strings.Contains(err.Error(), "missing required values") {
				t.Fatalf("expected missing values error, got: %v", err)
			}
		})
	}
}

func TestInvalidateExistingCache(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "cookies.json")
	cookies := map[string]string{
		"ASP.NET_SessionId": "session-value",
		".ASPXAUTH":         "auth-value",
	}
	if err := saveCacheFile(path, cookies); err != nil {
		t.Fatalf("saveCacheFile: %v", err)
	}

	if err := invalidateCacheFile(path); err != nil {
		t.Fatalf("invalidateCacheFile: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("expected cache file to be removed")
	}
}

func TestInvalidateMissingCache(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nonexistent.json")
	if err := invalidateCacheFile(path); err != nil {
		t.Fatalf("invalidateCacheFile should be no-op for missing file: %v", err)
	}
}

func TestCachePreservesExtraCookies(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "cookies.json")
	cookies := map[string]string{
		"ASP.NET_SessionId":            "session-value",
		".ASPXAUTH":                    "auth-value",
		"__RequestVerificationToken":   "xsrf-cookie",
	}

	if err := saveCacheFile(path, cookies); err != nil {
		t.Fatalf("saveCacheFile: %v", err)
	}

	got, err := loadCacheFile(path)
	if err != nil {
		t.Fatalf("loadCacheFile: %v", err)
	}

	if len(got) != len(cookies) {
		t.Errorf("expected %d cookies, got %d", len(cookies), len(got))
	}
	for name, want := range cookies {
		if got[name] != want {
			t.Errorf("cookie %s: got %q, want %q", name, got[name], want)
		}
	}
}
