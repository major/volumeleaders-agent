package update

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestGoReleaserChecksumValidatorAssetName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
	}{
		{"volumeleaders-agent_0.8.1_linux_amd64.tar.gz", "volumeleaders-agent_0.8.1_checksums.txt"},
		{"volumeleaders-agent_0.8.1_windows_arm64.zip", "volumeleaders-agent_0.8.1_checksums.txt"},
		{"unexpected_linux_amd64.tar.gz", "volumeleaders-agent_checksums.txt"},
	}
	validator := GoReleaserChecksumValidator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := validator.GetValidationAssetName(tt.name); got != tt.want {
				t.Fatalf("checksum asset = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsNewerVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		latest  string
		current string
		want    bool
	}{
		{"newer", "0.8.1", "0.8.0", true},
		{"same", "0.8.1", "0.8.1", false},
		{"older", "0.8.0", "0.8.1", false},
		{"dev current", "0.8.1", "dev", true},
		{"invalid latest", "dev", "0.8.1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isNewerVersion(tt.latest, tt.current); got != tt.want {
				t.Fatalf("isNewerVersion(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
			}
		})
	}
}

func TestGoReleaserChecksumValidatorValidate(t *testing.T) {
	t.Parallel()
	releaseName := "volumeleaders-agent_0.8.1_linux_amd64.tar.gz"
	release := []byte("release archive bytes")
	validHash := sha256.Sum256(release)
	tests := []struct {
		name    string
		asset   []byte
		wantErr bool
	}{
		{
			name:  "valid checksum",
			asset: []byte(fmt.Sprintf("%x  %s\n", validHash, releaseName)),
		},
		{
			name:    "mismatched checksum",
			asset:   []byte(fmt.Sprintf("%064x  %s\n", 0, releaseName)),
			wantErr: true,
		},
		{
			name:    "missing archive entry",
			asset:   []byte(fmt.Sprintf("%x  other_archive.tar.gz\n", validHash)),
			wantErr: true,
		},
	}
	validator := GoReleaserChecksumValidator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validator.Validate(releaseName, release, tt.asset)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
