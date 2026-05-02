// Package testutil provides shared test helpers for cobra command package tests.
// It is an importable (non-test) package so that all command packages can share
// the same utilities without duplicating code.
package testutil

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/leodido/structcli"
	"github.com/spf13/cobra"

	"github.com/major/volumeleaders-agent/internal/cli/common"
	"github.com/major/volumeleaders-agent/internal/client"
)

// ContextWithTestClient creates a context with a test HTTP client injected.
// The returned context bypasses browser authentication and targets baseURL.
func ContextWithTestClient(t *testing.T, baseURL string) context.Context {
	t.Helper()
	httpClient := &http.Client{Timeout: 5 * time.Second}
	c := client.NewForTesting(httpClient, baseURL)
	return context.WithValue(t.Context(), common.TestClientKey, c)
}

// AddPrettyJSON returns a context with the pretty-print JSON flag set to true.
func AddPrettyJSON(ctx context.Context) context.Context {
	return context.WithValue(ctx, common.PrettyJSONKey, true)
}

// DataTablesJSON wraps data in a valid DataTables response envelope with a
// single record count. Use DataTablesJSONPage when you need a specific count.
func DataTablesJSON(data string) string {
	return `{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":` + data + `}`
}

// DataTablesJSONPage wraps data in a DataTables response envelope with an
// explicit recordsFiltered count, used by pagination tests.
func DataTablesJSONPage(data string, recordsFiltered int) string {
	return fmt.Sprintf(`{"draw":1,"recordsTotal":%d,"recordsFiltered":%d,"data":%s}`,
		recordsFiltered, recordsFiltered, data)
}

// AssertErrContains fails the test if err is nil or its message does not
// contain want. If want is empty, err must be nil.
func AssertErrContains(t *testing.T, err error, want string) {
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

// ExecuteCommand runs a cobra command in-process and returns captured stdout
// and stderr separately. It sets args, output buffers, and context on the
// command before executing. This approach is parallel-safe because it uses
// cmd.SetOut/SetErr rather than swapping os.Stdout/os.Stderr globally.
func ExecuteCommand(t *testing.T, cmd *cobra.Command, ctx context.Context, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	cmd.SetContext(ctx)
	propagateTestContext(cmd, ctx)
	_, err = structcli.ExecuteC(cmd)
	return outBuf.String(), errBuf.String(), err
}

// propagateTestContext layers test context values onto child commands that
// already have a context set by structcli.Bind. During Bind, structcli
// calls cmd.SetContext to attach its internal scope, which prevents cobra
// from inheriting the parent context during execution. This helper walks
// the command tree and layers TestClientKey and PrettyJSONKey onto each
// child's existing context so that RunE handlers find the test client.
func propagateTestContext(parent *cobra.Command, ctx context.Context) {
	testClient := ctx.Value(common.TestClientKey)
	prettyJSON := ctx.Value(common.PrettyJSONKey)
	for _, sub := range parent.Commands() {
		if subCtx := sub.Context(); subCtx != nil {
			if testClient != nil {
				subCtx = context.WithValue(subCtx, common.TestClientKey, testClient)
			}
			if prettyJSON != nil {
				subCtx = context.WithValue(subCtx, common.PrettyJSONKey, prettyJSON)
			}
			sub.SetContext(subCtx)
		}
		propagateTestContext(sub, ctx)
	}
}
