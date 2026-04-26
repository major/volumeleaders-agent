package commands

import (
	"testing"
)

func TestNewAppStructure(t *testing.T) {
	t.Parallel()

	app := NewApp("1.0.0-test")

	if app.Name != "volumeleaders-agent" {
		t.Errorf("expected app name %q, got %q", "volumeleaders-agent", app.Name)
	}
	if app.Version != "1.0.0-test" {
		t.Errorf("expected version %q, got %q", "1.0.0-test", app.Version)
	}

	// Verify all 6 command groups are registered.
	expected := map[string]bool{
		"trade": false, "volume": false, "chart": false,
		"market": false, "alert": false, "watchlist": false,
	}
	for _, cmd := range app.Commands {
		if _, ok := expected[cmd.Name]; ok {
			expected[cmd.Name] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("missing command group: %s", name)
		}
	}

	// Verify --pretty flag exists.
	var foundPretty bool
	for _, flag := range app.Flags {
		if flag.Names()[0] == "pretty" {
			foundPretty = true
		}
	}
	if !foundPretty {
		t.Error("missing --pretty flag")
	}
}

func TestNewAppSubcommandCounts(t *testing.T) {
	t.Parallel()

	app := NewApp("0.0.0")

	expected := map[string]int{
		"trade":     9,
		"volume":    3,
		"chart":     4,
		"market":    3,
		"alert":     4,
		"watchlist": 6,
	}

	for _, cmd := range app.Commands {
		want, ok := expected[cmd.Name]
		if !ok {
			continue
		}
		if got := len(cmd.Commands); got != want {
			t.Errorf("command %q: expected %d subcommands, got %d", cmd.Name, want, got)
		}
	}
}
