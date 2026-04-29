package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/major/volumeleaders-agent/internal/auth"
)

func newTestClient(baseURL string) *Client {
	return &Client{
		http:      &http.Client{Timeout: 5 * time.Second},
		baseURL:   baseURL,
		cookies:   map[string]string{"ASP.NET_SessionId": "test", ".ASPXAUTH": "test"},
		xsrfToken: "test-token",
	}
}

func TestPostDataTables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       string
		gzipBody   bool
		want       []string
		wantErr    string
	}{
		{
			name:       "valid response",
			statusCode: http.StatusOK,
			body:       `{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":["AAPL","NVDA"]}`,
			want:       []string{"AAPL", "NVDA"},
		},
		{
			name:       "empty data field",
			statusCode: http.StatusOK,
			body:       `{"draw":1}`,
			want:       []string{},
		},
		{
			name:       "null data field",
			statusCode: http.StatusOK,
			body:       `{"draw":1,"data":null}`,
			want:       []string{},
		},
		{
			name:       "non 200 status",
			statusCode: http.StatusBadGateway,
			body:       `upstream failed`,
			wantErr:    "status 502: upstream failed",
		},
		{
			name:       "gzip response",
			statusCode: http.StatusOK,
			body:       `{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":["MSFT"]}`,
			gzipBody:   true,
			want:       []string{"MSFT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/datatable" {
					t.Errorf("expected path /datatable, got %s", r.URL.Path)
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("read request body: %v", err)
				}
				if string(body) != "draw=1" {
					t.Errorf("expected request body draw=1, got %q", string(body))
				}
				if tt.gzipBody {
					w.Header().Set("Content-Encoding", "gzip")
				}
				w.WriteHeader(tt.statusCode)
				writeResponse(t, w, tt.body, tt.gzipBody)
			}))
			t.Cleanup(server.Close)

			client := newTestClient(server.URL)
			var got []string
			err := client.PostDataTables(t.Context(), "/datatable", "draw=1", &got)
			assertErrorContains(t, err, tt.wantErr)
			if tt.wantErr != "" {
				return
			}
			if fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestPostJSON(t *testing.T) {
	t.Parallel()

	t.Run("valid response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type: expected application/json, got %q", got)
			}
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode request payload: %v", err)
			}
			if payload["symbol"] != "AAPL" {
				t.Errorf("expected symbol AAPL, got %q", payload["symbol"])
			}
			fmt.Fprint(w, `{"ok":true}`)
		}))
		t.Cleanup(server.Close)

		client := newTestClient(server.URL)
		var got struct {
			OK bool `json:"ok"`
		}
		if err := client.PostJSON(t.Context(), "/json", map[string]string{"symbol": "AAPL"}, &got); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.OK {
			t.Errorf("expected OK response")
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		t.Parallel()

		client := newTestClient("http://example.test")
		err := client.PostJSON(t.Context(), "/json", map[string]any{"bad": make(chan int)}, &struct{}{})
		assertErrorContains(t, err, "marshal JSON request")
	})

	t.Run("non 200 status", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "nope", http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)

		client := newTestClient(server.URL)
		err := client.PostJSON(t.Context(), "/json", map[string]string{"ok": "true"}, &struct{}{})
		assertErrorContains(t, err, "status 500")
	})
}

func TestPostForm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       string
		result     any
		wantName   string
		wantErr    string
	}{
		{
			name:       "valid response",
			statusCode: http.StatusOK,
			body:       `{"name":"AAPL"}`,
			result: &struct {
				Name string `json:"name"`
			}{},
			wantName: "AAPL",
		},
		{
			name:       "nil result skips decode",
			statusCode: http.StatusOK,
			body:       `not json`,
		},
		{
			name:       "non 200 status",
			statusCode: http.StatusBadRequest,
			body:       `bad form`,
			result:     &struct{}{},
			wantErr:    "status 400: bad form",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("read request body: %v", err)
				}
				if string(body) != "ticker=AAPL" {
					t.Errorf("expected form body ticker=AAPL, got %q", string(body))
				}
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.body)
			}))
			t.Cleanup(server.Close)

			client := newTestClient(server.URL)
			err := client.PostForm(t.Context(), "/form", url.Values{"ticker": {"AAPL"}}, tt.result)
			assertErrorContains(t, err, tt.wantErr)
			if tt.wantName != "" {
				got := tt.result.(*struct {
					Name string `json:"name"`
				})
				if got.Name != tt.wantName {
					t.Errorf("expected name %q, got %q", tt.wantName, got.Name)
				}
			}
		})
	}
}

func TestPostMultipart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		wantErr    string
	}{
		{name: "successful post", statusCode: http.StatusOK},
		{name: "redirect success", statusCode: http.StatusFound},
		{name: "error status", statusCode: http.StatusBadRequest, wantErr: "status 400: multipart failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
					t.Errorf("expected multipart content type, got %q", r.Header.Get("Content-Type"))
				}
				if err := r.ParseMultipartForm(1024); err != nil {
					t.Errorf("parse multipart form: %v", err)
				}
				if got := r.FormValue("ticker"); got != "AAPL" {
					t.Errorf("expected ticker AAPL, got %q", got)
				}
				if tt.statusCode == http.StatusFound {
					w.Header().Set("Location", "/next")
				}
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, "multipart failed")
			}))
			t.Cleanup(server.Close)

			client := newTestClient(server.URL)
			err := client.PostMultipart(t.Context(), "/multipart", map[string]string{"ticker": "AAPL"})
			assertErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestReadResponseBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		gzipBody bool
	}{
		{name: "plain body", body: "plain response"},
		{name: "gzip body", body: "gzip response", gzipBody: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var body bytes.Buffer
			if tt.gzipBody {
				gz := gzip.NewWriter(&body)
				if _, err := gz.Write([]byte(tt.body)); err != nil {
					t.Fatalf("write gzip body: %v", err)
				}
				if err := gz.Close(); err != nil {
					t.Fatalf("close gzip body: %v", err)
				}
			} else {
				body.WriteString(tt.body)
			}

			resp := &http.Response{Body: io.NopCloser(&body), Header: make(http.Header)}
			if tt.gzipBody {
				resp.Header.Set("Content-Encoding", "gzip")
			}

			got, err := readResponseBody(resp)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.body {
				t.Errorf("expected %q, got %q", tt.body, string(got))
			}
		})
	}
}

func TestSetHeaders(t *testing.T) {
	t.Parallel()

	client := newTestClient("http://example.test")
	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	req.Header.Set("Content-Type", "application/json")
	client.setHeaders(req)

	checks := map[string]string{
		"User-Agent":         auth.UserAgent,
		"x-xsrf-token":       "test-token",
		"x-requested-with":   "XMLHttpRequest",
		"Accept":             "application/json, text/javascript, */*; q=0.01",
		"Content-Type":       "application/json",
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
		if got := req.Header.Get(key); got != expected {
			t.Errorf("%s: expected %q, got %q", key, expected, got)
		}
	}
}

func TestSetHeadersDefaultContentType(t *testing.T) {
	t.Parallel()

	client := newTestClient("http://example.test")
	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	client.setHeaders(req)

	if got := req.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded; charset=UTF-8" {
		t.Errorf("expected default Content-Type, got %q", got)
	}
}

func TestSetCookies(t *testing.T) {
	t.Parallel()

	client := newTestClient("http://example.test")
	req := httptest.NewRequest(http.MethodPost, "/", http.NoBody)
	client.setCookies(req)

	checks := map[string]string{
		"ASP.NET_SessionId": "test",
		".ASPXAUTH":         "test",
	}
	for name, expected := range checks {
		cookie, err := req.Cookie(name)
		if err != nil {
			t.Errorf("missing cookie %s: %v", name, err)
			continue
		}
		if cookie.Value != expected {
			t.Errorf("cookie %s: expected %q, got %q", name, expected, cookie.Value)
		}
	}
}

func writeResponse(t *testing.T, w http.ResponseWriter, body string, gzipBody bool) {
	t.Helper()

	if !gzipBody {
		fmt.Fprint(w, body)
		return
	}
	w.Header().Set("Content-Encoding", "gzip")
	gz := gzip.NewWriter(w)
	defer gz.Close()
	if _, err := gz.Write([]byte(body)); err != nil {
		t.Errorf("write gzip response: %v", err)
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
}

func TestPostDataTablesDataDecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"draw":1,"data":"not an array"}`)
	}))
	t.Cleanup(server.Close)

	c := newTestClient(server.URL)
	var got []string
	err := c.PostDataTables(t.Context(), "/test", "draw=1", &got)
	assertErrorContains(t, err, "decode DataTables data")
}

func TestPostJSONDecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `not json at all`)
	}))
	t.Cleanup(server.Close)

	c := newTestClient(server.URL)
	var got struct{ Name string }
	err := c.PostJSON(t.Context(), "/test", struct{}{}, &got)
	assertErrorContains(t, err, "decode JSON response")
}

func TestPostFormDecodeError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `not json`)
	}))
	t.Cleanup(server.Close)

	c := newTestClient(server.URL)
	var got struct{ Name string }
	err := c.PostForm(t.Context(), "/test", url.Values{"key": {"val"}}, &got)
	assertErrorContains(t, err, "decode form response")
}

func TestDoRequestConnectionError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	server.Close()

	c := newTestClient(server.URL)
	err := c.PostJSON(t.Context(), "/test", struct{}{}, &struct{}{})
	assertErrorContains(t, err, "post JSON request")
}

func TestReadResponseBodyGzipError(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		Body:   io.NopCloser(strings.NewReader("not-gzip-data")),
		Header: http.Header{"Content-Encoding": []string{"gzip"}},
	}
	_, err := readResponseBody(resp)
	assertErrorContains(t, err, "open gzip response body")
}

func TestPostMultipartConnectionError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	server.Close()

	c := newTestClient(server.URL)
	err := c.PostMultipart(t.Context(), "/test", map[string]string{"key": "val"})
	assertErrorContains(t, err, "post multipart request")
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()

	if want == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected error containing %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}
