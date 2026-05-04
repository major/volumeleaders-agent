package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const cacheSubdir = "volumeleaders-agent"
const cacheFileName = "cookies.json"

// cookieCache is the on-disk format for cached authentication cookies.
type cookieCache struct {
	Cookies  map[string]string `json:"cookies"`
	CachedAt time.Time         `json:"cached_at"`
}

// cachePath returns the full path to the cookie cache file using the
// platform-appropriate user cache directory.
func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("user cache directory: %w", err)
	}
	return filepath.Join(dir, cacheSubdir, cacheFileName), nil
}

// loadCacheFile reads cached cookies from the given path. Returns
// os.ErrNotExist on cache miss, or an error if the file is corrupt
// or missing required cookie values.
func loadCacheFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache cookieCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("decode cookie cache: %w", err)
	}

	if cache.Cookies["ASP.NET_SessionId"] == "" || cache.Cookies[".ASPXAUTH"] == "" {
		return nil, fmt.Errorf("cached cookies missing required values")
	}

	return cache.Cookies, nil
}

// saveCacheFile writes cookies to the given path, creating parent
// directories as needed. The file is written with 0o600 permissions
// since it contains authentication cookies.
func saveCacheFile(path string, cookies map[string]string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	cache := cookieCache{
		Cookies:  cookies,
		CachedAt: time.Now().UTC(),
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("encode cookie cache: %w", err)
	}

	return os.WriteFile(path, data, 0o600)
}

// invalidateCacheFile removes a cache file. Returns nil if the file
// does not exist.
func invalidateCacheFile(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// loadCache reads cached cookies from the default cache location.
func loadCache() (map[string]string, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}
	return loadCacheFile(path)
}

// saveCache writes cookies to the default cache location.
func saveCache(cookies map[string]string) error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	return saveCacheFile(path, cookies)
}

// InvalidateCache removes cached authentication cookies so the next
// ExtractCookies call reads directly from browser stores.
func InvalidateCache() error {
	path, err := cachePath()
	if err != nil {
		return err
	}
	return invalidateCacheFile(path)
}
