package tui

import (
	"strings"
	"testing"
)

func TestPageFrame(t *testing.T) {
	t.Run("centers content", func(t *testing.T) {
		got := pageFrame("hi", 2, 10, 3)
		lines := strings.Split(got, "\n")
		if len(lines) != 3 {
			t.Fatalf("got %d lines, want 3", len(lines))
		}
		// Content should be padded with 4 leading spaces ((10-2)/2).
		if !strings.HasPrefix(lines[0], "    ") {
			t.Errorf("line[0] = %q, expected 4-space prefix", lines[0])
		}
	})

	t.Run("no centering when narrow", func(t *testing.T) {
		got := pageFrame("hello", 5, 5, 2)
		lines := strings.Split(got, "\n")
		if lines[0] != "hello" {
			t.Errorf("line[0] = %q, want hello", lines[0])
		}
	})

	t.Run("pads to height", func(t *testing.T) {
		got := pageFrame("line1", 5, 5, 4)
		lines := strings.Split(got, "\n")
		if len(lines) != 4 {
			t.Errorf("got %d lines, want 4", len(lines))
		}
	})

	t.Run("trims to height", func(t *testing.T) {
		content := "a\nb\nc\nd\ne"
		got := pageFrame(content, 1, 1, 3)
		lines := strings.Split(got, "\n")
		if len(lines) != 3 {
			t.Errorf("got %d lines, want 3", len(lines))
		}
	})
}

func TestHealthLabel(t *testing.T) {
	theme := testTheme()

	t.Run("no healthcheck singular", func(t *testing.T) {
		got := healthLabel("", false, theme)
		if !strings.Contains(stripANSI(got), "no check") {
			t.Errorf("got %q, expected 'no check'", stripANSI(got))
		}
	})

	t.Run("no healthcheck plural", func(t *testing.T) {
		got := healthLabel("", true, theme)
		if !strings.Contains(stripANSI(got), "no checks") {
			t.Errorf("got %q, expected 'no checks'", stripANSI(got))
		}
	})

	t.Run("healthy", func(t *testing.T) {
		got := healthLabel("healthy", false, theme)
		if !strings.Contains(stripANSI(got), "healthy") {
			t.Errorf("got %q, expected 'healthy'", stripANSI(got))
		}
	})

	t.Run("unhealthy", func(t *testing.T) {
		got := healthLabel("unhealthy", false, theme)
		if !strings.Contains(stripANSI(got), "unhealthy") {
			t.Errorf("got %q, expected 'unhealthy'", stripANSI(got))
		}
	})

	t.Run("starting", func(t *testing.T) {
		got := healthLabel("starting", false, theme)
		if !strings.Contains(stripANSI(got), "starting") {
			t.Errorf("got %q, expected 'starting'", stripANSI(got))
		}
	})
}

func TestBirdIcon(t *testing.T) {
	theme := testTheme()
	open := stripANSI(birdIcon(false, theme))
	if open != "—(•)>" {
		t.Errorf("birdIcon(false) = %q, want —(•)>", open)
	}
	blinked := stripANSI(birdIcon(true, theme))
	if blinked != "—(-)>" {
		t.Errorf("birdIcon(true) = %q, want —(-)>", blinked)
	}
}

func TestCursorRow(t *testing.T) {
	got := cursorRow("hello world", 8)
	stripped := stripANSI(got)
	if stripped != "hello w…" {
		t.Errorf("cursorRow truncated = %q, want 'hello w…'", stripped)
	}
	// Exact fit.
	got = cursorRow("hi", 10)
	stripped = stripANSI(got)
	if stripped != "hi" {
		t.Errorf("cursorRow = %q, want 'hi'", stripped)
	}
}

func TestRenderHelpBar(t *testing.T) {
	theme := testTheme()
	bindings := []helpBinding{
		{"j/k", "up/down"},
		{"q", "quit"},
	}
	got := renderHelpBar(bindings, 40, theme)
	stripped := stripANSI(got)
	if !strings.Contains(stripped, "j/k") || !strings.Contains(stripped, "quit") {
		t.Errorf("renderHelpBar missing content: %q", stripped)
	}
}

func TestStyledSep(t *testing.T) {
	theme := testTheme()
	got := stripANSI(styledSep(theme))
	if got != " · " {
		t.Errorf("styledSep = %q, want ' · '", got)
	}
}
