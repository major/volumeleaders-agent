package discovery

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateWritesDiscoveryFiles(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	skillPath := filepath.Join(outputDir, "SKILL.md")
	if err := Generate(outputDir, skillPath, "test"); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for fileName, path := range map[string]string{
		"AGENTS.md": filepath.Join(outputDir, "AGENTS.md"),
		"SKILL.md":  skillPath,
		"llms.txt":  filepath.Join(outputDir, "llms.txt"),
	} {
		fileName, path := fileName, path
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()

			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read generated %s: %v", fileName, err)
			}
			if !bytes.Contains(contents, []byte("volumeleaders-agent")) {
				t.Fatalf("generated %s does not mention volumeleaders-agent", fileName)
			}
		})
	}
}

func TestGeneratedFilesAreCurrent(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	skillPath := filepath.Join(outputDir, "SKILL.md")
	if err := Generate(outputDir, skillPath, "dev"); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	for fileName, paths := range map[string]struct{ generated, committed string }{
		"AGENTS.md": {generated: filepath.Join(outputDir, "AGENTS.md"), committed: filepath.Join(repoRoot, DefaultOutputDir, "AGENTS.md")},
		"SKILL.md":  {generated: skillPath, committed: filepath.Join(repoRoot, DefaultSkillPath)},
		"llms.txt":  {generated: filepath.Join(outputDir, "llms.txt"), committed: filepath.Join(repoRoot, DefaultOutputDir, "llms.txt")},
	} {
		fileName, paths := fileName, paths
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()

			generated, err := os.ReadFile(paths.generated)
			if err != nil {
				t.Fatalf("read generated %s: %v", fileName, err)
			}

			committed, err := os.ReadFile(paths.committed)
			if err != nil {
				t.Fatalf("read committed %s: %v", fileName, err)
			}

			if !bytes.Equal(generated, committed) {
				t.Fatalf("%s is stale; run `make generate-discovery`", paths.committed)
			}
		})
	}
}

func TestLabelPlainFences(t *testing.T) {
	t.Parallel()

	input := "before\n```\nvolumeleaders-agent --help\n```\n```json\n{}\n```\nafter\n"
	want := "before\n```bash\nvolumeleaders-agent --help\n```\n```json\n{}\n```\nafter\n"
	if got := labelPlainFences(input); got != want {
		t.Fatalf("labelPlainFences() = %q, want %q", got, want)
	}
}

func TestNormalizeGeneratedFileRewritesLLMsURL(t *testing.T) {
	t.Parallel()

	input := "# volumeleaders-agent\n\nhttps://github.com/major/volumeleaders-agent/cmd/volumeleaders-agent\n"
	want := "# volumeleaders-agent\n\nhttps://github.com/major/volumeleaders-agent\n"
	if got := normalizeGeneratedFile("llms.txt", input); got != want {
		t.Fatalf("normalizeGeneratedFile() = %q, want %q", got, want)
	}
}

func TestNormalizeGeneratedFileReplacesSkillFrontmatterDescription(t *testing.T) {
	t.Parallel()

	input := "---\nname: volumeleaders-agent\ndescription: |\n  Find individual institutional pr...\nmetadata:\n  author: major\n---\n\n# volumeleaders-agent\n"
	got := normalizeGeneratedFile("SKILL.md", input)
	if bytes.Contains([]byte(got), []byte("Find individual institutional pr...")) {
		t.Fatalf("normalizeGeneratedFile() kept truncated description: %q", got)
	}
	if !bytes.Contains([]byte(got), []byte("Output: compact JSON to stdout by default.")) {
		t.Fatalf("normalizeGeneratedFile() missing replacement description: %q", got)
	}
	if !bytes.Contains([]byte(got), []byte("metadata:\n  author: major")) {
		t.Fatalf("normalizeGeneratedFile() corrupted frontmatter metadata: %q", got)
	}
}
