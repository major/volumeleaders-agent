package common

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/major/volumeleaders-agent/internal/client"
)

func TestNewCommandClientWithTestClient(t *testing.T) {
	t.Parallel()
	want := client.NewForTesting(&http.Client{Timeout: 5 * time.Second}, "http://example.test")
	got, err := NewCommandClient(context.WithValue(t.Context(), TestClientKey, want))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected injected client %p, got %p", want, got)
	}
}
