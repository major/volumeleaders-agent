// Package auth extracts browser cookies and XSRF tokens needed to
// authenticate with the VolumeLeaders web application.
package auth

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
)

// UserAgent mimics Chrome 147 on Windows for authenticated requests.
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"

const volumeLeadersDomain = "volumeleaders.com"

// ErrSessionExpired marks auth failures caused by an expired browser session.
var ErrSessionExpired = errors.New("session expired")

// SessionExpiredMessage is the user-facing remediation for expired sessions.
const SessionExpiredMessage = "Authentication required: VolumeLeaders session has expired. Log in at https://www.volumeleaders.com in your browser, then retry."

var requiredCookieNames = []string{"ASP.NET_SessionId", ".ASPXAUTH"}

var xsrfTokenPattern = regexp.MustCompile(`<input\s+name="__RequestVerificationToken"\s+type="hidden"\s+value="([^"]+)"`)

type sessionExpiredError struct {
	redirectPath string
}

func (e sessionExpiredError) Error() string {
	return SessionExpiredMessage
}

func (e sessionExpiredError) Unwrap() error {
	return ErrSessionExpired
}

func (e sessionExpiredError) Detail() string {
	return fmt.Sprintf("requested host www.%s redirected to %s", volumeLeadersDomain, e.redirectPath)
}

// IsSessionExpired reports whether err indicates an expired VolumeLeaders session.
func IsSessionExpired(err error) bool {
	return errors.Is(err, ErrSessionExpired)
}

// ExtractCookies reads required VolumeLeaders cookies from supported browsers.
//
// Kooky scans all registered browsers and returns accumulated errors for
// browsers that could not be read (uninstalled, locked, etc.). We ignore
// those errors as long as the required auth cookies were found in at least
// one browser.
func ExtractCookies(ctx context.Context) (map[string]string, error) {
	// ReadCookies returns cookies it could find plus errors from browsers
	// it could not read. Errors from missing browsers are expected.
	validCookies, _ := kooky.ReadCookies(ctx, kooky.Valid, kooky.DomainHasSuffix(volumeLeadersDomain))
	found := authCookies(validCookies)

	if found["ASP.NET_SessionId"] == "" || found[".ASPXAUTH"] == "" {
		allCookies, _ := kooky.ReadCookies(ctx, kooky.DomainHasSuffix(volumeLeadersDomain))
		return nil, fmt.Errorf("required browser cookies unavailable: %s", cookieDiagnostic(found, allCookies, validCookies))
	}
	return found, nil
}

// FetchXSRFToken retrieves the hidden request verification token from ExecutiveSummary.
func FetchXSRFToken(ctx context.Context, httpClient *http.Client, cookies map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.volumeleaders.com/ExecutiveSummary", http.NoBody)
	if err != nil {
		return "", fmt.Errorf("create XSRF token request: %w", err)
	}
	setBrowserHeaders(req)
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch XSRF token page: %w", err)
	}
	defer resp.Body.Close()

	redirectPath := safeRedirectPath(resp)
	if normalizeRedirectPath(redirectPath) == "/login" {
		return "", sessionExpiredError{redirectPath: redirectPath}
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch XSRF token page: status %d", resp.StatusCode)
	}

	// When Accept-Encoding is set explicitly, Go's net/http does not
	// auto-decompress. Handle gzip manually.
	var bodyReader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gr, gzErr := gzip.NewReader(resp.Body)
		if gzErr != nil {
			return "", fmt.Errorf("decompress XSRF token page: %w", gzErr)
		}
		defer gr.Close()
		bodyReader = gr
	}

	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return "", fmt.Errorf("read XSRF token page: %w", err)
	}
	matches := xsrfTokenPattern.FindSubmatch(body)
	if matches == nil {
		return "", fmt.Errorf("XSRF token not found in HTML")
	}
	return string(matches[1]), nil
}

func authCookies(cookies kooky.Cookies) map[string]string {
	found := make(map[string]string, 3)
	for _, cookie := range cookies {
		switch cookie.Name {
		case "ASP.NET_SessionId", ".ASPXAUTH", "__RequestVerificationToken":
			found[cookie.Name] = cookie.Value
		}
	}
	return found
}

func cookieDiagnostic(found map[string]string, allCookies, validCookies kooky.Cookies) string {
	missing := missingRequiredCookies(found)
	rejected := rejectedRequiredCookies(found, allCookies)
	parts := []string{
		fmt.Sprintf("searched browser cookie stores for domain suffix %q", volumeLeadersDomain),
		"required cookies: ASP.NET_SessionId, .ASPXAUTH",
		fmt.Sprintf("valid VolumeLeaders cookies found: %d", len(validCookies)),
		fmt.Sprintf("browser stores with VolumeLeaders cookies: %d", browserStoreCount(allCookies)),
		fmt.Sprintf("missing valid cookies: %s", strings.Join(missing, ", ")),
		"only cookie storage is inspected; local storage, session storage, and IndexedDB are not inspected",
	}
	if len(rejected) > 0 {
		parts = append(parts, fmt.Sprintf("matching required cookies found but not usable as valid cookies: %s", strings.Join(rejected, ", ")))
	}
	return strings.Join(parts, "; ")
}

func missingRequiredCookies(found map[string]string) []string {
	missing := make([]string, 0, len(requiredCookieNames))
	for _, name := range requiredCookieNames {
		if found[name] == "" {
			missing = append(missing, name)
		}
	}
	return missing
}

func rejectedRequiredCookies(found map[string]string, allCookies kooky.Cookies) []string {
	rejected := make([]string, 0, len(requiredCookieNames))
	for _, name := range requiredCookieNames {
		if found[name] != "" || !containsCookieName(allCookies, name) {
			continue
		}
		rejected = append(rejected, name)
	}
	return rejected
}

func containsCookieName(cookies kooky.Cookies, name string) bool {
	return slices.ContainsFunc(cookies, func(cookie *kooky.Cookie) bool {
		return cookie.Name == name
	})
}

func browserStoreCount(cookies kooky.Cookies) int {
	stores := make(map[string]struct{})
	for _, cookie := range cookies {
		if cookie.Browser == nil {
			stores["unknown"] = struct{}{}
			continue
		}
		stores[cookie.Browser.Browser()+":"+cookie.Browser.Profile()] = struct{}{}
	}
	return len(stores)
}

func safeRedirectPath(resp *http.Response) string {
	if resp == nil || resp.Request == nil || resp.Request.URL == nil {
		return "unknown redirect target"
	}
	escapedPath := resp.Request.URL.EscapedPath()
	if escapedPath == "" {
		return "/"
	}
	return escapedPath
}

func normalizeRedirectPath(redirectPath string) string {
	if redirectPath == "" {
		return "/"
	}
	if !strings.HasPrefix(redirectPath, "/") {
		redirectPath = "/" + redirectPath
	}
	return strings.ToLower(path.Clean(redirectPath))
}

func setBrowserHeaders(req *http.Request) {
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="147", "Not A(Brand";v="24", "Google Chrome";v="147"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
}
