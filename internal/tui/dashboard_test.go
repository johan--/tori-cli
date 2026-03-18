package tui

import (
	"strings"
	"testing"

	"github.com/thobiasn/tori-cli/internal/protocol"
)

func TestBuildGroups(t *testing.T) {
	containers := []protocol.ContainerMetrics{
		{ID: "a1", Name: "web", State: "running", Project: "myapp", Service: "web"},
		{ID: "a2", Name: "api", State: "running", Project: "myapp", Service: "api"},
		{ID: "b1", Name: "db", State: "running", Project: "infra", Service: "db"},
		{ID: "c1", Name: "standalone", State: "exited"},
	}
	contInfo := []protocol.ContainerInfo{
		{ID: "a1", Name: "web", Project: "myapp", Service: "web"},
		{ID: "a2", Name: "api", Project: "myapp", Service: "api"},
		{ID: "b1", Name: "db", Project: "infra", Service: "db"},
		{ID: "c1", Name: "standalone"},
	}

	groups := buildGroups(containers, contInfo)

	// Should be sorted: infra, myapp, other (other always last).
	if len(groups) != 3 {
		t.Fatalf("got %d groups, want 3", len(groups))
	}
	if groups[0].name != "infra" {
		t.Errorf("groups[0].name = %q, want infra", groups[0].name)
	}
	if groups[1].name != "myapp" {
		t.Errorf("groups[1].name = %q, want myapp", groups[1].name)
	}
	if groups[2].name != "other" {
		t.Errorf("groups[2].name = %q, want other", groups[2].name)
	}

	// myapp has 2 running.
	if groups[1].running != 2 {
		t.Errorf("myapp running = %d, want 2", groups[1].running)
	}
	// other has 0 running (exited).
	if groups[2].running != 0 {
		t.Errorf("other running = %d, want 0", groups[2].running)
	}
}

func TestBuildGroupsUntrackedStubs(t *testing.T) {
	// Containers from metrics (tracked).
	containers := []protocol.ContainerMetrics{
		{ID: "a1", Name: "web", State: "running", Project: "myapp"},
	}
	// contInfo includes an untracked container.
	contInfo := []protocol.ContainerInfo{
		{ID: "a1", Name: "web", Project: "myapp"},
		{ID: "a2", Name: "worker", Project: "myapp", State: "running"},
	}

	groups := buildGroups(containers, contInfo)
	if len(groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(groups))
	}
	if len(groups[0].containers) != 2 {
		t.Errorf("myapp containers = %d, want 2 (including stub)", len(groups[0].containers))
	}
}

func TestBuildSelectableItems(t *testing.T) {
	groups := []containerGroup{
		{name: "myapp", containers: make([]protocol.ContainerMetrics, 2)},
		{name: "other", containers: make([]protocol.ContainerMetrics, 1)},
	}

	t.Run("all expanded", func(t *testing.T) {
		items := buildSelectableItems(groups, map[string]bool{})
		// 2 project headers + 2 containers + 1 container = 5
		if len(items) != 5 {
			t.Fatalf("got %d items, want 5", len(items))
		}
		if !items[0].isProject || items[0].groupIdx != 0 {
			t.Error("items[0] should be project 0")
		}
		if items[1].isProject {
			t.Error("items[1] should be container")
		}
	})

	t.Run("first group collapsed", func(t *testing.T) {
		items := buildSelectableItems(groups, map[string]bool{"myapp": true})
		// 2 project headers + 0 (collapsed) + 1 container = 3
		if len(items) != 3 {
			t.Fatalf("got %d items, want 3", len(items))
		}
		if !items[0].isProject {
			t.Error("items[0] should be project")
		}
		if !items[1].isProject {
			t.Error("items[1] should be project (other)")
		}
	})
}

func TestInstanceKeyContainerID(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"high_cpu:abc123", "abc123"},
		{"host_rule", ""},
		{"rule:container:extra", "container:extra"},
		{":", ""},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := instanceKeyContainerID(tt.key)
			if got != tt.want {
				t.Errorf("instanceKeyContainerID(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestInstanceDisplayName(t *testing.T) {
	contInfo := []protocol.ContainerInfo{
		{ID: "abc123", Name: "web-1", Service: "web"},
		{ID: "def456", Name: "db-1"},
	}

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"finds service name", "rule:abc123", "web"},
		{"falls back to container name", "rule:def456", "db-1"},
		{"host-scoped key", "high_cpu", "high_cpu"},
		{"unknown container truncated", "rule:abcdef1234567890", "abcdef123456"},
		{"unknown short ID", "rule:short", "short"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := instanceDisplayName(tt.key, contInfo)
			if got != tt.want {
				t.Errorf("instanceDisplayName(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestContainerAlertSeverity(t *testing.T) {
	alerts := map[int64]*protocol.AlertEvent{
		1: {State: "firing", InstanceKey: "rule1:abc", Severity: "warning"},
		2: {State: "firing", InstanceKey: "rule2:abc", Severity: "critical"},
		3: {State: "firing", InstanceKey: "rule1:def", Severity: "warning"},
		4: {State: "resolved", InstanceKey: "rule3:abc", Severity: "critical"},
	}

	tests := []struct {
		name        string
		containerID string
		want        string
	}{
		{"critical overrides warning", "abc", "critical"},
		{"warning only", "def", "warning"},
		{"no alerts", "xyz", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containerAlertSeverity(alerts, tt.containerID)
			if got != tt.want {
				t.Errorf("containerAlertSeverity(%q) = %q, want %q", tt.containerID, got, tt.want)
			}
		})
	}
}

func TestProjectAlertSeverity(t *testing.T) {
	alerts := map[int64]*protocol.AlertEvent{
		1: {State: "firing", InstanceKey: "rule:c1", Severity: "warning"},
		2: {State: "firing", InstanceKey: "rule:c2", Severity: "critical"},
	}

	g := containerGroup{
		containers: []protocol.ContainerMetrics{
			{ID: "c1"},
			{ID: "c2"},
		},
	}
	got := projectAlertSeverity(g, alerts)
	if got != "critical" {
		t.Errorf("projectAlertSeverity = %q, want critical", got)
	}

	// Group with no alerts.
	g2 := containerGroup{
		containers: []protocol.ContainerMetrics{{ID: "c3"}},
	}
	got = projectAlertSeverity(g2, alerts)
	if got != "" {
		t.Errorf("projectAlertSeverity(no alerts) = %q, want empty", got)
	}
}

func TestContainerAlertIndicator(t *testing.T) {
	theme := testTheme()

	t.Run("no alerts", func(t *testing.T) {
		got := containerAlertIndicator(nil, "abc", theme)
		if got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})

	t.Run("with alerts", func(t *testing.T) {
		alerts := map[int64]*protocol.AlertEvent{
			1: {State: "firing", InstanceKey: "r:abc", Severity: "critical"},
			2: {State: "firing", InstanceKey: "r2:abc", Severity: "warning"},
		}
		got := containerAlertIndicator(alerts, "abc", theme)
		stripped := stripANSI(got)
		if !strings.Contains(stripped, "▲") {
			t.Errorf("expected ▲ markers, got %q", stripped)
		}
	})
}

func TestProjectAlertIndicator(t *testing.T) {
	theme := testTheme()

	alerts := map[int64]*protocol.AlertEvent{
		1: {State: "firing", InstanceKey: "r:c1", Severity: "critical"},
	}
	g := containerGroup{
		containers: []protocol.ContainerMetrics{{ID: "c1"}},
	}

	got := projectAlertIndicator(g, alerts, theme)
	stripped := stripANSI(got)
	if !strings.Contains(stripped, "▲") {
		t.Errorf("expected ▲, got %q", stripped)
	}

	// No alerts.
	got = projectAlertIndicator(g, nil, theme)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestSeverityLabel(t *testing.T) {
	tests := []struct {
		severity string
		want     string
	}{
		{"critical", "CRIT"},
		{"CRITICAL", "CRIT"},
		{"warning", "WARN"},
		{"", "WARN"},
	}
	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			got := severityLabel(tt.severity)
			if got != tt.want {
				t.Errorf("severityLabel(%q) = %q, want %q", tt.severity, got, tt.want)
			}
		})
	}
}

func TestSeverityColor(t *testing.T) {
	theme := testTheme()
	tests := []struct {
		severity string
		want     string
	}{
		{"critical", string(theme.Critical)},
		{"warning", string(theme.Warning)},
		{"", string(theme.Warning)},
	}
	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			got := severityColor(tt.severity, theme)
			if string(got) != tt.want {
				t.Errorf("severityColor(%q) = %v, want %v", tt.severity, got, tt.want)
			}
		})
	}
}

func TestContainerCount(t *testing.T) {
	groups := []containerGroup{
		{containers: make([]protocol.ContainerMetrics, 3)},
		{containers: make([]protocol.ContainerMetrics, 2)},
	}
	if got := containerCount(groups); got != 5 {
		t.Errorf("containerCount = %d, want 5", got)
	}
	if got := containerCount(nil); got != 0 {
		t.Errorf("containerCount(nil) = %d, want 0", got)
	}
}

func TestContainerAtCursor(t *testing.T) {
	groups := []containerGroup{
		{name: "proj", containers: []protocol.ContainerMetrics{{ID: "a"}, {ID: "b"}}},
	}
	items := buildSelectableItems(groups, map[string]bool{})
	// items[0] = project, items[1] = container a, items[2] = container b

	// Project row returns nil.
	if got := containerAtCursor(groups, items, 0); got != nil {
		t.Error("expected nil for project row")
	}
	// Container row returns metrics.
	if got := containerAtCursor(groups, items, 1); got == nil || got.ID != "a" {
		t.Error("expected container a")
	}
	// Out of bounds returns nil.
	if got := containerAtCursor(groups, items, 99); got != nil {
		t.Error("expected nil for out of bounds")
	}
	if got := containerAtCursor(groups, items, -1); got != nil {
		t.Error("expected nil for negative index")
	}
}

func TestGroupAtCursor(t *testing.T) {
	groups := []containerGroup{
		{name: "proj", containers: []protocol.ContainerMetrics{{ID: "a"}}},
	}
	items := buildSelectableItems(groups, map[string]bool{})

	if got := groupAtCursor(groups, items, 0); got != "proj" {
		t.Errorf("groupAtCursor(0) = %q, want proj", got)
	}
	if got := groupAtCursor(groups, items, 1); got != "proj" {
		t.Errorf("groupAtCursor(1) = %q, want proj", got)
	}
	if got := groupAtCursor(groups, items, -1); got != "" {
		t.Errorf("groupAtCursor(-1) = %q, want empty", got)
	}
}

func TestServiceNameByID(t *testing.T) {
	contInfo := []protocol.ContainerInfo{
		{ID: "a", Name: "web-1", Service: "web"},
		{ID: "b", Name: "db-1"},
	}

	tests := []struct {
		id   string
		want string
	}{
		{"a", "web"},
		{"b", "db-1"},
		{"unknown", ""},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := serviceNameByID(tt.id, contInfo)
			if got != tt.want {
				t.Errorf("serviceNameByID(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
