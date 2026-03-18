package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		toml    string
		wantErr string
		check   func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid config with server",
			toml: `
[servers.local]
socket = "/run/tori/tori.sock"
`,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Servers["local"].Socket != "/run/tori/tori.sock" {
					t.Errorf("socket = %q", cfg.Servers["local"].Socket)
				}
			},
		},
		{
			name: "server missing socket and host",
			toml: `
[servers.broken]
port = 22
`,
			wantErr: "missing socket path",
		},
		{
			name: "empty display gets defaults",
			toml: `
[servers.local]
socket = "/tmp/tori.sock"
`,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Display.DateFormat != "2006-01-02" {
					t.Errorf("date_format = %q, want 2006-01-02", cfg.Display.DateFormat)
				}
				if cfg.Display.TimeFormat != "15:04:05" {
					t.Errorf("time_format = %q, want 15:04:05", cfg.Display.TimeFormat)
				}
			},
		},
		{
			name: "custom display formats",
			toml: `
[servers.local]
socket = "/tmp/tori.sock"

[display]
date_format = "02/01/2006"
time_format = "3:04 PM"
`,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Display.DateFormat != "02/01/2006" {
					t.Errorf("date_format = %q", cfg.Display.DateFormat)
				}
				if cfg.Display.TimeFormat != "3:04 PM" {
					t.Errorf("time_format = %q", cfg.Display.TimeFormat)
				}
			},
		},
		{
			name:    "invalid TOML",
			toml:    "this is not valid [[ toml",
			wantErr: "load config",
		},
		{
			name: "SSH server config",
			toml: `
[servers.prod]
host = "user@example.com"
port = 2222
socket = "/run/tori/tori.sock"
identity_file = "~/.ssh/id_ed25519"
auto_connect = true
`,
			check: func(t *testing.T, cfg *Config) {
				srv := cfg.Servers["prod"]
				if srv.Host != "user@example.com" {
					t.Errorf("host = %q", srv.Host)
				}
				if srv.Port != 2222 {
					t.Errorf("port = %d", srv.Port)
				}
				if !srv.AutoConnect {
					t.Error("auto_connect should be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.toml")
			if err := os.WriteFile(path, []byte(tt.toml), 0644); err != nil {
				t.Fatal(err)
			}
			cfg, err := LoadConfig(path)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestBuildTheme(t *testing.T) {
	defaults := TerminalTheme()

	t.Run("empty config returns defaults", func(t *testing.T) {
		got := BuildTheme(ThemeConfig{})
		if got != defaults {
			t.Error("empty ThemeConfig should return TerminalTheme defaults")
		}
	})

	t.Run("partial overrides", func(t *testing.T) {
		got := BuildTheme(ThemeConfig{
			Fg:       "#ffffff",
			Critical: "196",
		})
		if got.Fg != lipgloss.Color("#ffffff") {
			t.Errorf("Fg = %v, want #ffffff", got.Fg)
		}
		if got.Critical != lipgloss.Color("196") {
			t.Errorf("Critical = %v, want 196", got.Critical)
		}
		// Unset fields should keep defaults.
		if got.FgDim != defaults.FgDim {
			t.Errorf("FgDim = %v, want default %v", got.FgDim, defaults.FgDim)
		}
		if got.Healthy != defaults.Healthy {
			t.Errorf("Healthy = %v, want default %v", got.Healthy, defaults.Healthy)
		}
	})

	t.Run("full overrides", func(t *testing.T) {
		got := BuildTheme(ThemeConfig{
			Fg:         "#a9b1d6",
			FgDim:      "#3b4261",
			FgBright:   "#c0caf5",
			Border:     "#292e42",
			Accent:     "#7aa2f7",
			Healthy:    "#9ece6a",
			Warning:    "#e0af68",
			Critical:   "#f7768e",
			DebugLevel: "#414769",
			InfoLevel:  "#505a85",
			GraphCPU:   "#7dcfff",
			GraphMem:   "#bb9af7",
		})
		if got.Fg != lipgloss.Color("#a9b1d6") {
			t.Errorf("Fg = %v", got.Fg)
		}
		if got.GraphMem != lipgloss.Color("#bb9af7") {
			t.Errorf("GraphMem = %v", got.GraphMem)
		}
	})
}

func TestEnsureDefaultConfig(t *testing.T) {
	t.Run("creates file when missing", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "subdir", "config.toml")

		got, err := EnsureDefaultConfig(path)
		if err != nil {
			t.Fatal(err)
		}
		if got != path {
			t.Errorf("returned path = %q, want %q", got, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "Tori client configuration") {
			t.Error("default config file missing expected header")
		}
	})

	t.Run("returns existing path without overwrite", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.toml")

		content := "# my custom config\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := EnsureDefaultConfig(path)
		if err != nil {
			t.Fatal(err)
		}
		if got != path {
			t.Errorf("returned path = %q, want %q", got, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != content {
			t.Error("existing config was overwritten")
		}
	})
}
