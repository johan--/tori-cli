package tui

import (
	"strings"
	"testing"
)

func TestScrollAndPad(t *testing.T) {
	t.Run("fewer lines than maxH", func(t *testing.T) {
		lines := []string{"a", "b"}
		got := scrollAndPad(lines, 0, 5)
		result := strings.Split(got, "\n")
		if len(result) != 5 {
			t.Errorf("got %d lines, want 5", len(result))
		}
		if result[0] != "a" || result[1] != "b" {
			t.Error("content should be preserved")
		}
	})

	t.Run("more lines scrolled to cursor", func(t *testing.T) {
		lines := make([]string, 20)
		for i := range lines {
			lines[i] = strings.Repeat("x", i)
		}
		got := scrollAndPad(lines, 10, 5)
		result := strings.Split(got, "\n")
		if len(result) != 5 {
			t.Errorf("got %d lines, want 5", len(result))
		}
	})

	t.Run("cursor at start", func(t *testing.T) {
		lines := make([]string, 10)
		for i := range lines {
			lines[i] = "line"
		}
		got := scrollAndPad(lines, 0, 5)
		result := strings.Split(got, "\n")
		if len(result) != 5 {
			t.Errorf("got %d lines, want 5", len(result))
		}
	})
}

func TestPadLines(t *testing.T) {
	got := padLines("hello", 3)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Errorf("got %d lines, want 3", len(lines))
	}
	if lines[0] != "hello" {
		t.Errorf("lines[0] = %q, want hello", lines[0])
	}
}

func TestSectionLabel(t *testing.T) {
	theme := testTheme()

	active := sectionLabel("Alerts", true, theme)
	if stripped := stripANSI(active); stripped != "Alerts" {
		t.Errorf("active sectionLabel = %q, want Alerts", stripped)
	}

	inactive := sectionLabel("Rules", false, theme)
	if stripped := stripANSI(inactive); stripped != "Rules" {
		t.Errorf("inactive sectionLabel = %q, want Rules", stripped)
	}
}
