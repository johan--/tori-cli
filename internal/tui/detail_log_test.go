package tui

import (
	"testing"
	"time"

	"github.com/thobiasn/tori-cli/internal/protocol"
)

func TestSanitizeLogMsg(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"plain text", "hello world", "hello world"},
		{"strips ANSI", "\x1b[31mred\x1b[0m text", "red text"},
		{"replaces tab", "col1\tcol2", "col1    col2"},
		{"preserves newline", "line1\nline2", "line1\nline2"},
		{"drops control chars", "hello\x01\x02world", "helloworld"},
		{"combined", "\x1b[1mhi\x1b[0m\there\x00", "hi    here"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeLogMsg(tt.s)
			if got != tt.want {
				t.Errorf("sanitizeLogMsg(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{-5000, "-5,000"},
		{100, "100"},
		{10000, "10,000"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatNumber(tt.n)
			if got != tt.want {
				t.Errorf("formatNumber(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestFormatBytesRate(t *testing.T) {
	tests := []struct {
		name        string
		bytesPerSec float64
		want        string
	}{
		{"bytes", 500, "500B/s"},
		{"kilobytes", 5000, "5.0KB/s"},
		{"megabytes", 5e6, "5.0MB/s"},
		{"gigabytes", 1e9, "1.0GB/s"},
		{"zero", 0, "0B/s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytesRate(tt.bytesPerSec)
			if got != tt.want {
				t.Errorf("formatBytesRate(%f) = %q, want %q", tt.bytesPerSec, got, tt.want)
			}
		})
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"", 0},
		{"hello", 1},
		{"hello\nworld", 2},
		{"a\nb\nc", 3},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := countLines(tt.s)
			if got != tt.want {
				t.Errorf("countLines(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestFormatTimestamp(t *testing.T) {
	// 2009-02-13 23:31:30 UTC
	ts := int64(1234567890)
	got := formatTimestamp(ts, "2006-01-02")
	want := time.Unix(ts, 0).Format("2006-01-02")
	if got != want {
		t.Errorf("formatTimestamp = %q, want %q", got, want)
	}
}

func TestContainerAlerts(t *testing.T) {
	alerts := map[int64]*protocol.AlertEvent{
		1: {ID: 1, InstanceKey: "high_cpu:abc", State: "firing"},
		2: {ID: 2, InstanceKey: "high_mem:abc", State: "firing"},
		3: {ID: 3, InstanceKey: "high_cpu:def", State: "firing"},
		4: {ID: 4, InstanceKey: "host_rule", State: "firing"},
	}

	got := containerAlerts(alerts, "abc")
	if len(got) != 2 {
		t.Fatalf("got %d alerts, want 2", len(got))
	}
	// Should be sorted by ID.
	if got[0].ID != 1 || got[1].ID != 2 {
		t.Errorf("alerts not sorted: [%d, %d]", got[0].ID, got[1].ID)
	}

	// No matching alerts.
	got = containerAlerts(alerts, "xyz")
	if len(got) != 0 {
		t.Errorf("got %d alerts for xyz, want 0", len(got))
	}
}

func TestFindContainer(t *testing.T) {
	containers := []protocol.ContainerMetrics{
		{ID: "a", Name: "web"},
		{ID: "b", Name: "api"},
	}

	got := findContainer("a", containers)
	if got == nil || got.Name != "web" {
		t.Error("expected to find web")
	}

	got = findContainer("c", containers)
	if got != nil {
		t.Error("expected nil for unknown ID")
	}
}

func TestFindContInfo(t *testing.T) {
	info := []protocol.ContainerInfo{
		{ID: "a", Name: "web"},
		{ID: "b", Name: "api"},
	}

	got := findContInfo("b", info)
	if got == nil || got.Name != "api" {
		t.Error("expected to find api")
	}

	got = findContInfo("c", info)
	if got != nil {
		t.Error("expected nil for unknown ID")
	}
}

func TestCountDateChanges(t *testing.T) {
	day1 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC).Unix()
	day2 := time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC).Unix()
	day3 := time.Date(2025, 1, 3, 10, 0, 0, 0, time.UTC).Unix()

	tests := []struct {
		name    string
		entries []protocol.LogEntryMsg
		want    int
	}{
		{"empty", nil, 0},
		{"single entry", []protocol.LogEntryMsg{{Timestamp: day1}}, 1},
		{
			"same day",
			[]protocol.LogEntryMsg{{Timestamp: day1}, {Timestamp: day1 + 3600}},
			1,
		},
		{
			"three different days",
			[]protocol.LogEntryMsg{
				{Timestamp: day1}, {Timestamp: day2}, {Timestamp: day3},
			},
			3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countDateChanges(tt.entries, "2006-01-02")
			if got != tt.want {
				t.Errorf("countDateChanges = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestInjectDeploySeparators(t *testing.T) {
	t.Run("no redeploy", func(t *testing.T) {
		entries := []protocol.LogEntryMsg{
			{ContainerName: "web", ContainerID: "a1", Stream: "stdout", Message: "line1"},
			{ContainerName: "web", ContainerID: "a1", Stream: "stdout", Message: "line2"},
		}
		got := injectDeploySeparators(entries)
		if len(got) != 2 {
			t.Errorf("got %d entries, want 2", len(got))
		}
	})

	t.Run("redeploy detected", func(t *testing.T) {
		entries := []protocol.LogEntryMsg{
			{ContainerName: "web", ContainerID: "old", Stream: "stdout", Message: "line1"},
			{ContainerName: "web", ContainerID: "new", Stream: "stdout", Message: "line2"},
		}
		got := injectDeploySeparators(entries)
		if len(got) != 3 {
			t.Fatalf("got %d entries, want 3 (separator injected)", len(got))
		}
		if got[1].Stream != "event" {
			t.Errorf("separator stream = %q, want event", got[1].Stream)
		}
		if got[1].Message != "── web redeployed ──" {
			t.Errorf("separator message = %q", got[1].Message)
		}
	})

	t.Run("event entries pass through", func(t *testing.T) {
		entries := []protocol.LogEntryMsg{
			{ContainerName: "web", ContainerID: "a", Stream: "event", Message: "started"},
			{ContainerName: "web", ContainerID: "b", Stream: "stdout", Message: "log"},
		}
		got := injectDeploySeparators(entries)
		// Event entry doesn't track ID, so "b" is first seen for "web" — no separator.
		if len(got) != 2 {
			t.Errorf("got %d entries, want 2", len(got))
		}
	})

	t.Run("empty", func(t *testing.T) {
		got := injectDeploySeparators(nil)
		if len(got) != 0 {
			t.Errorf("got %d entries for nil input, want 0", len(got))
		}
	})

	t.Run("different containers no separator", func(t *testing.T) {
		entries := []protocol.LogEntryMsg{
			{ContainerName: "web", ContainerID: "a", Stream: "stdout"},
			{ContainerName: "api", ContainerID: "b", Stream: "stdout"},
		}
		got := injectDeploySeparators(entries)
		if len(got) != 2 {
			t.Errorf("got %d entries, want 2 (different names, no separator)", len(got))
		}
	})
}

func TestParseFilterBound(t *testing.T) {
	dateFormat := "2006-01-02"
	timeFormat := "15:04:05"

	t.Run("both empty", func(t *testing.T) {
		got := parseFilterBound("", "", dateFormat, timeFormat, false)
		if got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("date and time", func(t *testing.T) {
		got := parseFilterBound("2025-03-18", "14:30:00", dateFormat, timeFormat, false)
		if got == 0 {
			t.Error("expected non-zero timestamp for valid date+time")
		}
		parsed := time.Unix(got, 0)
		if parsed.Hour() != 14 || parsed.Minute() != 30 {
			t.Errorf("parsed time = %v, want 14:30", parsed)
		}
	})

	t.Run("date only isTo=false", func(t *testing.T) {
		got := parseFilterBound("2025-06-15", "", dateFormat, timeFormat, false)
		if got == 0 {
			t.Error("expected non-zero")
		}
		parsed := time.Unix(got, 0)
		if parsed.Hour() != 0 || parsed.Minute() != 0 {
			t.Errorf("from bound should be start of day, got %v", parsed)
		}
	})

	t.Run("date only isTo=true", func(t *testing.T) {
		got := parseFilterBound("2025-06-15", "", dateFormat, timeFormat, true)
		if got == 0 {
			t.Error("expected non-zero")
		}
		parsed := time.Unix(got, 0)
		if parsed.Hour() != 23 || parsed.Minute() != 59 {
			t.Errorf("to bound should be end of day, got %v", parsed)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		got := parseFilterBound("not-a-date", "", dateFormat, timeFormat, false)
		if got != 0 {
			t.Errorf("expected 0 for invalid date, got %d", got)
		}
	})

	t.Run("time only", func(t *testing.T) {
		got := parseFilterBound("", "10:30:00", dateFormat, timeFormat, false)
		if got == 0 {
			t.Error("expected non-zero for time-only")
		}
		parsed := time.Unix(got, 0)
		if parsed.Hour() != 10 || parsed.Minute() != 30 {
			t.Errorf("parsed = %v, want 10:30", parsed)
		}
	})
}
