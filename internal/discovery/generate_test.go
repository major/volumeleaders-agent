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
	if err := Generate(outputDir, "test"); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	for _, fileName := range []string{"AGENTS.md", "SKILL.md", "llms.txt"} {
		fileName := fileName
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()

			contents, err := os.ReadFile(filepath.Join(outputDir, fileName))
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
	if err := Generate(outputDir, "dev"); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	for _, fileName := range []string{"AGENTS.md", "SKILL.md", "llms.txt"} {
		fileName := fileName
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()

			generated, err := os.ReadFile(filepath.Join(outputDir, fileName))
			if err != nil {
				t.Fatalf("read generated %s: %v", fileName, err)
			}

			committed, err := os.ReadFile(filepath.Join(repoRoot, DefaultOutputDir, fileName))
			if err != nil {
				t.Fatalf("read committed %s: %v", fileName, err)
			}

			if !bytes.Equal(generated, committed) {
				t.Fatalf("%s is stale; run `make generate-discovery`", filepath.Join(DefaultOutputDir, fileName))
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
