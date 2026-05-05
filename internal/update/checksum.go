package update

import (
	"fmt"
	"strings"

	selfupdate "github.com/creativeprojects/go-selfupdate"
)

const projectName = "volumeleaders-agent"

// GoReleaserChecksumValidator maps versioned GoReleaser archives to the matching checksum asset.
type GoReleaserChecksumValidator struct{}

// Validate verifies release bytes against the versioned GoReleaser checksum file.
func (GoReleaserChecksumValidator) Validate(filename string, release, asset []byte) error {
	validator := selfupdate.ChecksumValidator{UniqueFilename: checksumAssetName(filename)}
	if err := validator.Validate(filename, release, asset); err != nil {
		return fmt.Errorf("go-releaser checksum validation for %s: %w", filename, err)
	}
	return nil
}

// GetValidationAssetName returns the checksum file uploaded by this repo's GoReleaser config.
func (GoReleaserChecksumValidator) GetValidationAssetName(releaseFilename string) string {
	return checksumAssetName(releaseFilename)
}

func checksumAssetName(releaseFilename string) string {
	prefix := projectName + "_"
	if !strings.HasPrefix(releaseFilename, prefix) {
		return fmt.Sprintf("%s_checksums.txt", projectName)
	}
	versionAndTarget := strings.TrimPrefix(releaseFilename, prefix)
	version, _, found := strings.Cut(versionAndTarget, "_")
	if !found || version == "" {
		return fmt.Sprintf("%s_checksums.txt", projectName)
	}
	return fmt.Sprintf("%s_%s_checksums.txt", projectName, version)
}
