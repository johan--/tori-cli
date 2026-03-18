package tui

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{"short string", "abc", 5, "abc"},
		{"exact length", "abcde", 5, "abcde"},
		{"long string", "abcdefgh", 5, "abcd…"},
		{"maxLen=1", "abcdef", 1, "…"},
		{"maxLen=0", "abcdef", 0, ""},
		{"negative maxLen", "abc", -1, ""},
		{"empty string", "", 5, ""},
		{"CJK runes", "日本語テスト", 4, "日本語…"},
		{"emoji", "👋🌍🎉💻🔥", 3, "👋🌍…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.s, tt.maxLen)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"plain text", "hello world", "hello world"},
		{"single color", "\x1b[31mred\x1b[0m", "red"},
		{"multiple escapes", "\x1b[1m\x1b[31mbold red\x1b[0m", "bold red"},
		{"nested", "\x1b[38;5;196mhello\x1b[0m \x1b[32mworld\x1b[0m", "hello world"},
		{"empty string", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripANSI(tt.s)
			if got != tt.want {
				t.Errorf("stripANSI(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		name    string
		seconds float64
		want    string
	}{
		{"zero", 0, "0m"},
		{"59 minutes", 59 * 60, "59m"},
		{"1 hour", 3600, "1h 0m"},
		{"1 hour 30 min", 5400, "1h 30m"},
		{"25 hours", 25 * 3600, "1d 1h"},
		{"5 days 11 hours", (5*24 + 11) * 3600, "5d 11h"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatUptime(tt.seconds)
			if got != tt.want {
				t.Errorf("FormatUptime(%f) = %q, want %q", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name string
		b    uint64
		want string
	}{
		{"zero", 0, "0B"},
		{"512 bytes", 512, "512B"},
		{"1 KB", 1024, "1.00K"},
		{"15.5 KB", 15872, "15.5K"},
		{"512 KB", 512 * 1024, "512K"},
		{"1 MB", 1 << 20, "1.00M"},
		{"30.9 MB", 32399770, "30.9M"},
		{"512 MB", 512 * (1 << 20), "512M"},
		{"1 GB", 1 << 30, "1.00G"},
		{"15.5 GB", uint64(15.5 * float64(1<<30)), "15.5G"},
		{"100 GB", 100 * (1 << 30), "100G"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.b)
			if got != tt.want {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.b, got, tt.want)
			}
		})
	}
}

func TestFormatCompactUptime(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		want    string
	}{
		{"zero", 0, "0m"},
		{"negative", -10, "0m"},
		{"under 1 min", 45, "0m"},
		{"5 minutes", 300, "5m"},
		{"3 hours", 3 * 3600, "3h"},
		{"5 days", 5 * 86400, "5d"},
		{"1 day (exact)", 86400, "1d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCompactUptime(tt.seconds)
			if got != tt.want {
				t.Errorf("formatCompactUptime(%d) = %q, want %q", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestRightAlign(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		width int
		want  string
	}{
		{"shorter", "abc", 6, "   abc"},
		{"equal", "abcdef", 6, "abcdef"},
		{"longer", "abcdefgh", 6, "abcdefgh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rightAlign(tt.s, tt.width)
			if got != tt.want {
				t.Errorf("rightAlign(%q, %d) = %q, want %q", tt.s, tt.width, got, tt.want)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		width int
		want  []string
	}{
		{"short line", "hello", 10, []string{"hello"}},
		{"exact width", "hello", 5, []string{"hello"}},
		{"wraps", "abcdefghij", 4, []string{"abcd", "efgh", "ij"}},
		{"embedded newlines", "abc\ndef", 10, []string{"abc", "def"}},
		{"width=0", "abc", 0, nil},
		{"empty string", "", 10, []string{""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.s, tt.width)
			if tt.want == nil {
				if got != nil {
					t.Errorf("wrapText(%q, %d) = %v, want nil", tt.s, tt.width, got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("wrapText(%q, %d) = %v (len %d), want %v (len %d)",
					tt.s, tt.width, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("wrapText(%q, %d)[%d] = %q, want %q", tt.s, tt.width, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCenterText(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		totalW int
		want   string
	}{
		{"narrower", "hi", 10, "    hi"},
		{"equal", "hello", 5, "hello"},
		{"wider", "hello world", 5, "hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := centerText(tt.s, tt.totalW)
			if got != tt.want {
				t.Errorf("centerText(%q, %d) = %q, want %q", tt.s, tt.totalW, got, tt.want)
			}
		})
	}
}
