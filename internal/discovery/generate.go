package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/leodido/structcli/generate"

	"github.com/major/volumeleaders-agent/internal/cli"
)

const (
	// DefaultOutputDir keeps generated agent-facing files out of the repository
	// root so they do not overwrite the hand-maintained root AGENTS.md.
	DefaultOutputDir = "docs/llm"
	modulePath       = "github.com/major/volumeleaders-agent/cmd/volumeleaders-agent"
	defaultAuthor    = "major"
)

// Generate writes the structcli discovery files for the current command tree.
func Generate(outputDir, version string) error {
	rootCmd := cli.NewRootCmd(version)

	if err := generate.WriteAll(rootCmd, outputDir, generate.AllOptions{
		ModulePath: modulePath,
		Skill: generate.SkillOptions{
			Author:  defaultAuthor,
			Version: version,
		},
	}); err != nil {
		return fmt.Errorf("generate discovery files: %w", err)
	}
	if err := labelMarkdownFences(outputDir); err != nil {
		return err
	}

	return nil
}

func labelMarkdownFences(outputDir string) error {
	for _, fileName := range []string{"AGENTS.md", "SKILL.md", "llms.txt"} {
		path := filepath.Join(outputDir, fileName)
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read generated %s: %w", fileName, err)
		}

		labeled := labelPlainFences(string(contents))
		if err := os.WriteFile(path, []byte(labeled), 0o600); err != nil {
			return fmt.Errorf("write generated %s: %w", fileName, err)
		}
	}

	return nil
}

func labelPlainFences(contents string) string {
	var builder strings.Builder
	inFence := false
	for _, line := range strings.SplitAfter(contents, "\n") {
		trimmed := strings.TrimSuffix(line, "\n")
		if trimmed == "```" {
			if inFence {
				builder.WriteString(line)
			} else {
				builder.WriteString("```bash")
				if strings.HasSuffix(line, "\n") {
					builder.WriteString("\n")
				}
			}
			inFence = !inFence
			continue
		}
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
		}
		builder.WriteString(line)
	}

	return builder.String()
}
