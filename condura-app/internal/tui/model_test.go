package tui

import "testing"

func TestTabNamesCount(t *testing.T) {
	if len(tabNames) != int(tabCount) {
		t.Fatalf("tabNames len %d != tabCount %d", len(tabNames), tabCount)
	}
}

func TestViewTabOrder(t *testing.T) {
	if tabNames[0] != "Chat" || tabNames[len(tabNames)-1] != "Health" {
		t.Fatalf("unexpected tab order: %v", tabNames)
	}
}

// TestHubSyncSkillsTabsPresent ensures all Phase 12 tabs exist
// and appear in the right order.
func TestHubSyncSkillsTabsPresent(t *testing.T) {
	want := []string{"Hub", "Sync", "Skills"}
	idx := map[string]int{}
	for i, n := range tabNames {
		idx[n] = i
	}
	for _, w := range want {
		if _, ok := idx[w]; !ok {
			t.Errorf("missing tab %q", w)
		}
	}
	// Hub before Sync before Skills (so the natural tab order makes sense)
	if idx["Hub"] >= idx["Sync"] || idx["Sync"] >= idx["Skills"] {
		t.Errorf("tab order wrong: %v", tabNames)
	}
}
