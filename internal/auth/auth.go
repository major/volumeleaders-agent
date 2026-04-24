// Package auth extracts browser cookies and XSRF tokens needed to
// authenticate with the VolumeLeaders web application.
package auth

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
)

// UserAgent mimics Chrome 147 on Windows for authenticated requests.
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"

var xsrfTokenPattern = regexp.MustCompile(`<input\s+name="__RequestVerificationToken"\s+type="hidden"\s+value="([^"]+)"`)

// ExtractCookies reads required VolumeLeaders cookies from supported browsers.
//
// Kooky scans all registered browsers and returns accumulated errors for
// browsers that could not be read (uninstalled, locked, etc.). We ignore
// those errors as long as the required auth cookies were found in at least
// one browser.
func ExtractCookies(ctx context.Context) (map[string]string, error) {
	found := make(map[string]string, 3)

	// ReadCookies returns cookies it could find plus errors from browsers
	// it could not read. Errors from missing browsers are expected.
	cookies, _ := kooky.ReadCookies(ctx, kooky.Valid, kooky.DomainHasSuffix("volumeleaders.com"))
	for _, cookie := range cookies {
		switch cookie.Name {
		case "ASP.NET_SessionId", ".ASPXAUTH", "__RequestVerificationToken":
			found[cookie.Name] = cookie.Value
		}
	}

	if found["ASP.NET_SessionId"] == "" || found[".ASPXAUTH"] == "" {
		return nil, fmt.Errorf(
			"required cookies (ASP.NET_SessionId, .ASPXAUTH) not found in any browser; " +
				"log in to volumeleaders.com and try again",
		)
	}
	return found, nil
}

// FetchXSRFToken retrieves the hidden request verification token from ExecutiveSummary.
func FetchXSRFToken(httpClient *http.Client, cookies map[string]string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.volumeleaders.com/ExecutiveSummary", http.NoBody)
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

	if strings.Contains(resp.Request.URL.Path, "/Login") {
		return "", fmt.Errorf("session expired")
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
