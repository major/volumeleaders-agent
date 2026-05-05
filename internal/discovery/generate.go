package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/major/volumeleaders-agent/internal/cli"
	"github.com/major/volumeleaders-agent/internal/cli/common"
)

const (
	// DefaultOutputDir keeps extended generated agent-facing files out of the
	// repository root so they do not overwrite the hand-maintained root AGENTS.md.
	DefaultOutputDir = "docs/llm"
	// DefaultSkillPath keeps the primary skill file at the repository root, which
	// matches the sibling agent repositories and makes it easy for users and tools
	// to discover without knowing each repo's extended documentation layout.
	DefaultSkillPath = "SKILL.md"
	modulePath       = "github.com/major/volumeleaders-agent/cmd/volumeleaders-agent"
	defaultAuthor    = "major"
	skillDescription = `  volumeleaders-agent queries institutional trade data from VolumeLeaders. Use it for trades, volume leaderboards, market data, alerts, and watchlists.

  Auth: reads browser cookies automatically. If auth fails with exit code 2 and "Authentication required: VolumeLeaders session has expired.", log in at https://www.volumeleaders.com in your browser, then retry.

  Output: compact JSON to stdout by default. Use --pretty before the command group for indented JSON. Use --jsonschema on any command for machine-readable input JSON Schema output, --jsonschema=tree on the root for the full CLI tree, outputschema for machine-readable stdout contracts, or --mcp on the root to serve leaf commands as MCP tools over stdio. Errors and logs go to stderr.`
)

// Generate writes the discovery files for the current command tree.
func Generate(outputDir, skillPath, version string) error {
	rootCmd := cli.NewRootCmd(version)

	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return fmt.Errorf("create discovery directory: %w", err)
	}

	if err := writeGeneratedFile(filepath.Join(outputDir, "AGENTS.md"), func() ([]byte, error) {
		return generateAgents(rootCmd), nil
	}); err != nil {
		return err
	}
	if err := writeGeneratedFile(skillPath, func() ([]byte, error) {
		return generateSkill(rootCmd, version), nil
	}); err != nil {
		return err
	}
	if err := writeGeneratedFile(filepath.Join(outputDir, "llms.txt"), func() ([]byte, error) {
		return generateLLMsTxt(rootCmd), nil
	}); err != nil {
		return err
	}
	if err := normalizeGeneratedFiles(outputDir, skillPath); err != nil {
		return err
	}

	return nil
}

func generateAgents(root *cobra.Command) []byte {
	var builder strings.Builder
	writeAgentIntro(&builder, root)
	builder.WriteString("\n## Installation\n\n```bash\ngo install ")
	builder.WriteString(modulePath)
	builder.WriteString("@latest\n```\n\n## Commands\n\n")
	builder.WriteString("| Command | Description | Required Flags |\n|---------|-------------|---------------|\n")
	for _, cmd := range runnableCommands(root) {
		fmt.Fprintf(&builder, "| `%s` | %s | %s |\n", cmd.CommandPath(), tableText(commandDescription(cmd)), requiredFlagList(cmd))
	}
	builder.WriteString("\n## Configuration\n\n### Flags\n")
	writeFlagReference(&builder, root)
	builder.WriteString("\n## Machine Interface\n\n- JSON Schema: `volumeleaders-agent --jsonschema`\n- MCP: `volumeleaders-agent --mcp`\n")
	return []byte(builder.String())
}

func generateSkill(root *cobra.Command, version string) []byte {
	var builder strings.Builder
	fmt.Fprintf(&builder, "---\nname: volumeleaders-agent\ndescription: |\n%s\nmetadata:\n  author: %s\n  version: %s\n---\n\n", skillDescription, defaultAuthor, version)
	builder.WriteString("# volumeleaders-agent\n\n## Instructions\n\n### Available Commands\n")
	for _, cmd := range runnableCommands(root) {
		fmt.Fprintf(&builder, "\n#### `%s`\n\n%s\n", cmd.CommandPath(), commandDescription(cmd))
		writeCommandFlags(&builder, cmd, true)
		if cmd.Example != "" {
			fmt.Fprintf(&builder, "\n**Example:**\n\n```bash\n%s\n```\n", strings.TrimSpace(cmd.Example))
		}
	}
	return []byte(builder.String())
}

func generateLLMsTxt(root *cobra.Command) []byte {
	var builder strings.Builder
	writeAgentIntro(&builder, root)
	fmt.Fprintf(&builder, "\n## Commands\n\n")
	for _, cmd := range runnableCommands(root) {
		anchor := strings.ReplaceAll(strings.ToLower(cmd.CommandPath()), " ", "-")
		fmt.Fprintf(&builder, "- [%s](#%s): %s\n", cmd.CommandPath(), anchor, commandDescription(cmd))
	}
	builder.WriteString("\n## Configuration\n\n### Flags\n")
	writeFlagReference(&builder, root)
	return []byte(builder.String())
}

func writeAgentIntro(builder *strings.Builder, root *cobra.Command) {
	fmt.Fprintf(builder, "# %s\n\nhttps://github.com/major/volumeleaders-agent\n\n> %s\n\n%s\n", root.Name(), root.Short, root.Long)
}

// runnableCommands returns all runnable, non-hidden commands in the tree. This
// includes parent commands that are both runnable and have subcommands (e.g.
// "update" is runnable and also has "update check" and "update config").
func runnableCommands(root *cobra.Command) []*cobra.Command {
	var commands []*cobra.Command
	walkCommands(root, func(cmd *cobra.Command) {
		if cmd.Runnable() && !cmd.Hidden && !isGeneratedBuiltin(cmd) {
			commands = append(commands, cmd)
		}
	})
	slices.SortFunc(commands, func(a, b *cobra.Command) int { return strings.Compare(a.CommandPath(), b.CommandPath()) })
	return commands
}

func walkCommands(cmd *cobra.Command, visit func(*cobra.Command)) {
	visit(cmd)
	for _, sub := range cmd.Commands() {
		walkCommands(sub, visit)
	}
}

func isGeneratedBuiltin(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "help", "completion", "bash", "fish", "powershell", "zsh":
		return true
	default:
		return false
	}
}

func commandDescription(cmd *cobra.Command) string {
	if cmd.Long != "" {
		return strings.TrimSpace(cmd.Long)
	}
	return strings.TrimSpace(cmd.Short)
}

func tableText(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	return value
}

func requiredFlagList(cmd *cobra.Command) string {
	var names []string
	visitFlags(cmd, func(flag *pflag.Flag) {
		if common.IsFlagRequired(flag) {
			names = append(names, "`--"+flag.Name+"`")
		}
	})
	slices.Sort(names)
	return strings.Join(names, ", ")
}

func writeFlagReference(builder *strings.Builder, root *cobra.Command) {
	for _, cmd := range runnableCommands(root) {
		fmt.Fprintf(builder, "\n#### `%s`\n", cmd.CommandPath())
		writeCommandFlags(builder, cmd, false)
	}
}

func writeCommandFlags(builder *strings.Builder, cmd *cobra.Command, includeRequired bool) {
	flags := commandFlags(cmd)
	if len(flags) == 0 {
		return
	}
	if includeRequired {
		builder.WriteString("\n**Flags:**\n\n| Flag | Type | Default | Required | Description |\n|------|------|---------|----------|-------------|\n")
	} else {
		builder.WriteString("\n| Flag | Type | Default | Description |\n|------|------|---------|-------------|\n")
	}
	for _, flag := range flags {
		defaultValue := flag.DefValue
		if defaultValue == "" {
			defaultValue = "-"
		}
		description := flag.Usage
		if enum := common.FlagEnumValues(flag); len(enum) > 0 {
			description = fmt.Sprintf("%s (%s)", description, strings.Join(enum, ", "))
		}
		if includeRequired {
			required := "no"
			if common.IsFlagRequired(flag) {
				required = "yes"
			}
			fmt.Fprintf(builder, "| `--%s` | %s | %s | %s | %s |\n", flag.Name, flag.Value.Type(), defaultValue, required, tableText(description))
		} else {
			fmt.Fprintf(builder, "| `--%s` | %s | %s | %s |\n", flag.Name, flag.Value.Type(), defaultValue, tableText(description))
		}
	}
}

func commandFlags(cmd *cobra.Command) []*pflag.Flag {
	var flags []*pflag.Flag
	visitFlags(cmd, func(flag *pflag.Flag) {
		if !flag.Hidden {
			flags = append(flags, flag)
		}
	})
	slices.SortFunc(flags, func(a, b *pflag.Flag) int { return strings.Compare(a.Name, b.Name) })
	return flags
}

func visitFlags(cmd *cobra.Command, visit func(*pflag.Flag)) {
	seen := make(map[string]struct{})
	for _, flags := range []*pflag.FlagSet{cmd.InheritedFlags(), cmd.NonInheritedFlags()} {
		flags.VisitAll(func(flag *pflag.Flag) {
			if _, ok := seen[flag.Name]; ok {
				return
			}
			seen[flag.Name] = struct{}{}
			visit(flag)
		})
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
