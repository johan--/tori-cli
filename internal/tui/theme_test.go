package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/thobiasn/tori-cli/internal/protocol"
)

func testTheme() *Theme {
	t := TerminalTheme()
	return &t
}

func TestHostUsageColor(t *testing.T) {
	theme := testTheme()
	tests := []struct {
		name    string
		percent float64
		want    lipgloss.Color
	}{
		{"0%", 0, theme.Fg},
		{"59%", 59, theme.Fg},
		{"60%", 60, theme.Warning},
		{"79%", 79, theme.Warning},
		{"80%", 80, theme.Critical},
		{"100%", 100, theme.Critical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hostUsageColor(tt.percent, theme)
			if got != tt.want {
				t.Errorf("hostUsageColor(%f) = %v, want %v", tt.percent, got, tt.want)
			}
		})
	}
}

func TestUsageColor(t *testing.T) {
	theme := TerminalTheme()
	tests := []struct {
		name    string
		percent float64
		want    lipgloss.Color
	}{
		{"0%", 0, theme.Healthy},
		{"59%", 59, theme.Healthy},
		{"60%", 60, theme.Warning},
		{"79%", 79, theme.Warning},
		{"80%", 80, theme.Critical},
		{"100%", 100, theme.Critical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := theme.UsageColor(tt.percent)
			if got != tt.want {
				t.Errorf("UsageColor(%f) = %v, want %v", tt.percent, got, tt.want)
			}
		})
	}
}

func TestStateColor(t *testing.T) {
	theme := TerminalTheme()
	tests := []struct {
		state string
		want  lipgloss.Color
	}{
		{"running", theme.Healthy},
		{"restarting", theme.Warning},
		{"unhealthy", theme.Warning},
		{"exited", theme.Critical},
		{"dead", theme.Critical},
		{"created", theme.FgDim},
		{"", theme.FgDim},
	}
	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			got := theme.StateColor(tt.state)
			if got != tt.want {
				t.Errorf("StateColor(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestStatusDotColor(t *testing.T) {
	theme := TerminalTheme()
	tests := []struct {
		name   string
		state  string
		health string
		want   lipgloss.Color
	}{
		{"running+healthy", "running", "healthy", theme.Healthy},
		{"running+unhealthy", "running", "unhealthy", theme.Warning},
		{"running+starting", "running", "starting", theme.Warning},
		{"running+no healthcheck", "running", "", theme.Healthy},
		{"exited", "exited", "", theme.Critical},
		{"dead", "dead", "healthy", theme.Critical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := theme.StatusDotColor(tt.state, tt.health)
			if got != tt.want {
				t.Errorf("StatusDotColor(%q, %q) = %v, want %v", tt.state, tt.health, got, tt.want)
			}
		})
	}
}

func TestContainerCPUColor(t *testing.T) {
	theme := testTheme()
	tests := []struct {
		name     string
		cpuPct   float64
		cpuLimit float64
		want     lipgloss.Color
	}{
		{"no limit", 150, 0, theme.FgDim},
		{"50% of limit", 50, 1, theme.FgDim},
		{"70% of limit", 70, 1, theme.Warning},
		{"89% of limit", 89, 1, theme.Warning},
		{"90% of limit", 90, 1, theme.Critical},
		{"over 100% of limit", 200, 1, theme.Critical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containerCPUColor(tt.cpuPct, tt.cpuLimit, theme)
			if got != tt.want {
				t.Errorf("containerCPUColor(%f, %f) = %v, want %v", tt.cpuPct, tt.cpuLimit, got, tt.want)
			}
		})
	}
}

func TestContainerMemColor(t *testing.T) {
	theme := testTheme()
	tests := []struct {
		name     string
		memPct   float64
		memLimit uint64
		want     lipgloss.Color
	}{
		{"no limit", 80, 0, theme.FgDim},
		{"50%", 50, 1 << 30, theme.FgDim},
		{"70%", 70, 1 << 30, theme.Warning},
		{"89%", 89, 1 << 30, theme.Warning},
		{"90%", 90, 1 << 30, theme.Critical},
		{"100%", 100, 1 << 30, theme.Critical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containerMemColor(tt.memPct, tt.memLimit, theme)
			if got != tt.want {
				t.Errorf("containerMemColor(%f, %d) = %v, want %v", tt.memPct, tt.memLimit, got, tt.want)
			}
		})
	}
}

func TestDetailCPUColor(t *testing.T) {
	theme := testTheme()
	// No limit: containerCPUColor returns FgDim, detailCPUColor replaces with Fg.
	got := detailCPUColor(50, 0, theme)
	if got != theme.Fg {
		t.Errorf("detailCPUColor(50, 0) = %v, want Fg %v", got, theme.Fg)
	}
	// Has limit, high usage: should stay Critical.
	got = detailCPUColor(95, 1, theme)
	if got != theme.Critical {
		t.Errorf("detailCPUColor(95, 1) = %v, want Critical %v", got, theme.Critical)
	}
}

func TestDetailMemColor(t *testing.T) {
	theme := testTheme()
	// No limit: FgDim → Fg.
	got := detailMemColor(50, 0, theme)
	if got != theme.Fg {
		t.Errorf("detailMemColor(50, 0) = %v, want Fg %v", got, theme.Fg)
	}
	// Has limit, high usage: Critical.
	got = detailMemColor(95, 1<<30, theme)
	if got != theme.Critical {
		t.Errorf("detailMemColor(95, 1<<30) = %v, want Critical %v", got, theme.Critical)
	}
}

func TestDiskSeverityColor(t *testing.T) {
	theme := testTheme()
	tests := []struct {
		name string
		pct  float64
		want lipgloss.Color
	}{
		{"low", 50, theme.Fg},
		{"69%", 69, theme.Fg},
		{"70%", 70, theme.Warning},
		{"89%", 89, theme.Warning},
		{"90%", 90, theme.Critical},
		{"100%", 100, theme.Critical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diskSeverityColor(tt.pct, theme)
			if got != tt.want {
				t.Errorf("diskSeverityColor(%f) = %v, want %v", tt.pct, got, tt.want)
			}
		})
	}
}

func TestLoadSeverityColor(t *testing.T) {
	theme := testTheme()
	tests := []struct {
		name  string
		load1 float64
		cpus  int
		want  lipgloss.Color
	}{
		{"low ratio", 0.5, 1, theme.Fg},
		{"warning ratio", 0.7, 1, theme.Warning},
		{"critical ratio", 1.1, 1, theme.Critical},
		{"multi-core low", 2.0, 4, theme.Fg},
		{"cpus=0 treated as 1", 0.8, 0, theme.Warning},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := loadSeverityColor(tt.load1, tt.cpus, theme)
			if got != tt.want {
				t.Errorf("loadSeverityColor(%f, %d) = %v, want %v", tt.load1, tt.cpus, got, tt.want)
			}
		})
	}
}

func TestColorRank(t *testing.T) {
	theme := testTheme()
	tests := []struct {
		name string
		c    lipgloss.Color
		want int
	}{
		{"FgDim", theme.FgDim, 0},
		{"Fg", theme.Fg, 1},
		{"Warning", theme.Warning, 2},
		{"Critical", theme.Critical, 3},
		{"unknown", lipgloss.Color("99"), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorRank(tt.c, theme)
			if got != tt.want {
				t.Errorf("colorRank(%v) = %d, want %d", tt.c, got, tt.want)
			}
		})
	}
}

func TestProjectStatColor(t *testing.T) {
	theme := testTheme()
	tests := []struct {
		name string
		g    containerGroup
		want lipgloss.Color
	}{
		{
			"all running healthy",
			containerGroup{
				running:    2,
				containers: []protocol.ContainerMetrics{{State: "running"}, {State: "running"}},
			},
			theme.Healthy,
		},
		{
			"none running",
			containerGroup{
				running:    0,
				containers: []protocol.ContainerMetrics{{State: "exited"}},
			},
			theme.Critical,
		},
		{
			"partial running",
			containerGroup{
				running:    1,
				containers: []protocol.ContainerMetrics{{State: "running"}, {State: "exited"}},
			},
			theme.Warning,
		},
		{
			"all running but unhealthy",
			containerGroup{
				running:    1,
				containers: []protocol.ContainerMetrics{{State: "running", Health: "unhealthy"}},
			},
			theme.Warning,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := projectStatColor(tt.g, theme)
			if got != tt.want {
				t.Errorf("projectStatColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasHealthcheck(t *testing.T) {
	tests := []struct {
		health string
		want   bool
	}{
		{"healthy", true},
		{"unhealthy", true},
		{"starting", true},
		{"", false},
		{"none", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		t.Run(tt.health, func(t *testing.T) {
			got := hasHealthcheck(tt.health)
			if got != tt.want {
				t.Errorf("hasHealthcheck(%q) = %v, want %v", tt.health, got, tt.want)
			}
		})
	}
}
