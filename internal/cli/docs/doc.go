// Package docs provides cobra commands for documentation generation.
package docs

import (
	"fmt"

	"github.com/spf13/cobra/doc"
	"github.com/spf13/cobra"
)

// GenerateDocs generates markdown documentation for all CLI commands.
// It calls cobra's doc.GenMarkdownTree to create one .md file per command.
func GenerateDocs(rootCmd *cobra.Command, outputDir string) error {
	if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
		return fmt.Errorf("failed to generate markdown docs: %w", err)
	}
	return nil
}
