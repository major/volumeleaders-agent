package main

import (
	"fmt"
	"testing"

	"github.com/major/volumeleaders-agent/internal/auth"
)

func TestUserFacingError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "session expired",
			err:  fmt.Errorf("create client: %w", auth.ErrSessionExpired),
			want: auth.SessionExpiredMessage,
		},
		{
			name: "generic error",
			err:  fmt.Errorf("create client: missing cookies"),
			want: "create client: missing cookies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := userFacingError(tt.err); got != tt.want {
				t.Fatalf("userFacingError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "session expired",
			err:  fmt.Errorf("fetch XSRF token: %w", auth.ErrSessionExpired),
			want: 2,
		},
		{
			name: "generic error",
			err:  fmt.Errorf("fetch XSRF token: status 403"),
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := exitCode(tt.err); got != tt.want {
				t.Fatalf("exitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}
