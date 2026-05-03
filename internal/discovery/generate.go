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
	skillDescription = `  volumeleaders-agent queries institutional trade data from VolumeLeaders. Use it for trades, volume leaderboards, market data, alerts, and watchlists.

  Auth: reads browser cookies automatically. If auth fails with exit code 2 and "Authentication required: VolumeLeaders session has expired.", log in at https://www.volumeleaders.com in your browser, then retry.

  Output: compact JSON to stdout by default. Use --pretty before the command group for indented JSON. Use --jsonschema on any command for machine-readable input JSON Schema output, --jsonschema=tree on the root for the full CLI tree, outputschema for machine-readable stdout contracts, or --mcp on the root to serve leaf commands as MCP tools over stdio. Errors and logs go to stderr.`
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
	if err := normalizeGeneratedFiles(outputDir); err != nil {
		return err
	}

	return nil
}

func normalizeGeneratedFiles(outputDir string) error {
	for _, fileName := range []string{"AGENTS.md", "SKILL.md", "llms.txt"} {
		path := filepath.Join(outputDir, fileName)
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read generated %s: %w", fileName, err)
		}

		normalized := normalizeGeneratedFile(fileName, string(contents))
		// #nosec G703 - outputDir is the caller-selected generation target, and
		// fileName is constrained to the fixed discovery file list above.
		if err := os.WriteFile(path, []byte(normalized), 0o600); err != nil {
			return fmt.Errorf("write generated %s: %w", fileName, err)
		}
	}

	return nil
}

func normalizeGeneratedFile(fileName, contents string) string {
	normalized := labelPlainFences(contents)
	switch fileName {
	case "SKILL.md":
		return replaceSkillFrontmatterDescription(normalized)
	case "llms.txt":
		return strings.Replace(normalized, "https://github.com/major/volumeleaders-agent/cmd/volumeleaders-agent", "https://github.com/major/volumeleaders-agent", 1)
	default:
		return normalized
	}
}

func replaceSkillFrontmatterDescription(contents string) string {
	const descriptionStart = "description: |\n"
	const metadataStart = "metadata:\n"

	start := strings.Index(contents, descriptionStart)
	if start == -1 {
		return contents
	}
	metadata := strings.Index(contents[start:], metadataStart)
	if metadata == -1 {
		return contents
	}

	metadata += start
	var builder strings.Builder
	builder.Grow(len(contents))
	builder.WriteString(contents[:start+len(descriptionStart)])
	builder.WriteString(skillDescription)
	builder.WriteString("\n")
	builder.WriteString(contents[metadata:])
	return builder.String()
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
