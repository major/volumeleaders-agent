package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultOutputDir keeps extended generated agent-facing files out of the
	// repository root so they do not overwrite the hand-maintained root AGENTS.md.
	DefaultOutputDir = "docs/llm"
	// DefaultSkillPath keeps the primary skill file at the repository root, which
	// matches the sibling agent repositories and makes it easy for users and tools
	// to discover without knowing each repo's extended documentation layout.
	DefaultSkillPath = "SKILL.md"
	skillDescription = `  volumeleaders-agent queries institutional trade data from VolumeLeaders. Use it for trades, volume leaderboards, market data, alerts, and watchlists.

  Auth: reads browser cookies automatically. If auth fails with exit code 2 and "Authentication required: VolumeLeaders session has expired.", log in at https://www.volumeleaders.com in your browser, then retry.

  Output: compact JSON to stdout by default. Use --pretty before the command group for indented JSON. Use --jsonschema on any command for machine-readable input JSON Schema output, --jsonschema=tree on the root for the full CLI tree, outputschema for machine-readable stdout contracts, or --mcp on the root to serve leaf commands as MCP tools over stdio. Errors and logs go to stderr.`
)

// Generate refreshes the CLI discovery files from the checked-in Cobra discovery templates.
// The live `--jsonschema=tree` and `outputschema` outputs remain the source of truth for command and stdout contracts; these files package that guidance for LLM tools.
func Generate(outputDir, skillPath, version string) error {
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return fmt.Errorf("create discovery directory: %w", err)
	}

	if err := writeGeneratedFile(filepath.Join(outputDir, "AGENTS.md"), func() ([]byte, error) {
		return readDiscoveryTemplate(filepath.Join(DefaultOutputDir, "AGENTS.md"))
	}); err != nil {
		return err
	}
	if err := writeGeneratedFile(skillPath, func() ([]byte, error) {
		return readDiscoveryTemplate(DefaultSkillPath)
	}); err != nil {
		return err
	}
	if err := writeGeneratedFile(filepath.Join(outputDir, "llms.txt"), func() ([]byte, error) {
		return readDiscoveryTemplate(filepath.Join(DefaultOutputDir, "llms.txt"))
	}); err != nil {
		return err
	}
	if err := normalizeGeneratedFiles(outputDir, skillPath); err != nil {
		return err
	}

	return nil
}

func readDiscoveryTemplate(path string) ([]byte, error) {
	root, err := repoRoot()
	if err != nil {
		return nil, err
	}
	contents, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		return nil, fmt.Errorf("read discovery template %s: %w", path, err)
	}
	return contents, nil
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("find repository root from %s", dir)
		}
		dir = parent
	}
}

func writeGeneratedFile(path string, generateFile func() ([]byte, error)) error {
	contents, err := generateFile()
	if err != nil {
		return fmt.Errorf("generate %s: %w", filepath.Base(path), err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", path, err)
	}
	// #nosec G703 - path is a caller-selected documentation output path.
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		return fmt.Errorf("write generated %s: %w", filepath.Base(path), err)
	}
	return nil
}

func normalizeGeneratedFiles(outputDir, skillPath string) error {
	for fileName, path := range map[string]string{
		"AGENTS.md": filepath.Join(outputDir, "AGENTS.md"),
		"SKILL.md":  skillPath,
		"llms.txt":  filepath.Join(outputDir, "llms.txt"),
	} {
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
