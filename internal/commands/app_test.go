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

	// Verify all top-level command groups are registered.
	expected := map[string]bool{
		"trade": false, "volume": false,
		"market": false, "alert": false, "watchlist": false,
		"schema": false,
	}
	if got, want := len(app.Commands), len(expected); got != want {
		t.Errorf("expected %d command groups, got %d", want, got)
	}
	for _, cmd := range app.Commands {
		if _, ok := expected[cmd.Name]; ok {
			expected[cmd.Name] = true
		} else {
			t.Errorf("unexpected command group: %s", cmd.Name)
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
		"trade":     10,
		"volume":    3,
		"market":    3,
		"alert":     4,
		"watchlist": 6,
		"schema":    0,
	}

	for _, cmd := range app.Commands {
		want, ok := expected[cmd.Name]
		if !ok {
			t.Errorf("unexpected command group: %s", cmd.Name)
			continue
		}
		if got := len(cmd.Commands); got != want {
			t.Errorf("command %q: expected %d subcommands, got %d", cmd.Name, want, got)
		}
	}
}
