package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/major/volumeleaders-agent/internal/client"
)

// contextWithTestClient returns a context carrying a test Client that targets
// the given base URL. The returned context bypasses browser authentication.
func contextWithTestClient(t *testing.T, baseURL string) context.Context {
	t.Helper()
	httpClient := &http.Client{Timeout: 5 * time.Second}
	c := client.NewForTesting(httpClient, baseURL)
	return context.WithValue(t.Context(), testClientKey, c)
}

// captureStdout calls fn and returns everything written to os.Stdout during
// its execution. It is not safe for parallel use.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()
	w.Close()
	<-done
	return buf.String()
}

// dataTablesJSON wraps data in a valid DataTables response envelope.
func dataTablesJSON(data string) string {
	return `{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":` + data + `}`
}

// dataTablesJSONPage wraps data in a DataTables response with explicit
// recordsFiltered count, used by pagination tests.
func dataTablesJSONPage(data string, recordsFiltered int) string {
	return fmt.Sprintf(`{"draw":1,"recordsTotal":%d,"recordsFiltered":%d,"data":%s}`,
		recordsFiltered, recordsFiltered, data)
}

// addPrettyJSON returns a context with the pretty JSON flag set.
func addPrettyJSON(ctx context.Context) context.Context {
	return context.WithValue(ctx, prettyJSONKey, true)
}

// assertErrContains checks that err is non-nil and its message contains want.
// If want is empty, err must be nil.
func assertErrContains(t *testing.T, err error, want string) {
	t.Helper()
	if want == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got: %v", want, err)
	}
}
