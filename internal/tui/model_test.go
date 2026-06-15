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
