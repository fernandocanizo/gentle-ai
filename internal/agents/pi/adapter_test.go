package pi

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

const testHome = "/tmp/home"

// --- Identity ---

func TestAgentIdentity(t *testing.T) {
	a := NewAdapter()

	if got := a.Agent(); got != model.AgentPi {
		t.Fatalf("Agent() = %q, want %q", got, model.AgentPi)
	}

	if got := a.Tier(); got != model.TierFull {
		t.Fatalf("Tier() = %v, want %v", got, model.TierFull)
	}
}

// --- Detection ---

func TestDetect(t *testing.T) {
	tests := []struct {
		name            string
		lookPathPath    string
		lookPathErr     error
		stat            statResult
		wantInstalled   bool
		wantBinaryPath  string
		wantConfigPath  string
		wantConfigFound bool
		wantErr         bool
	}{
		{
			name:            "binary and config dir both present — installed",
			lookPathPath:    "/usr/local/bin/pi",
			stat:            statResult{isDir: true},
			wantInstalled:   true,
			wantBinaryPath:  "/usr/local/bin/pi",
			wantConfigPath:  filepath.Join(testHome, ".pi", "agent"),
			wantConfigFound: true,
		},
		{
			name:            "binary on PATH but config dir absent — not installed",
			lookPathPath:    "/usr/local/bin/pi",
			stat:            statResult{err: os.ErrNotExist},
			wantInstalled:   false,
			wantBinaryPath:  "/usr/local/bin/pi",
			wantConfigPath:  filepath.Join(testHome, ".pi", "agent"),
			wantConfigFound: false,
		},
		{
			name:            "config dir present but binary absent — not installed",
			lookPathErr:     errors.New("not found"),
			stat:            statResult{isDir: true},
			wantInstalled:   false,
			wantBinaryPath:  "",
			wantConfigPath:  filepath.Join(testHome, ".pi", "agent"),
			wantConfigFound: true,
		},
		{
			name:    "stat error bubbles up",
			stat:    statResult{err: errors.New("permission denied")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				lookPath: func(string) (string, error) {
					return tt.lookPathPath, tt.lookPathErr
				},
				statPath: func(string) statResult {
					return tt.stat
				},
			}

			installed, binaryPath, configPath, configFound, err := a.Detect(context.Background(), testHome)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Detect() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if installed != tt.wantInstalled {
				t.Fatalf("Detect() installed = %v, want %v", installed, tt.wantInstalled)
			}
			if binaryPath != tt.wantBinaryPath {
				t.Fatalf("Detect() binaryPath = %q, want %q", binaryPath, tt.wantBinaryPath)
			}
			if configPath != tt.wantConfigPath {
				t.Fatalf("Detect() configPath = %q, want %q", configPath, tt.wantConfigPath)
			}
			if configFound != tt.wantConfigFound {
				t.Fatalf("Detect() configFound = %v, want %v", configFound, tt.wantConfigFound)
			}
		})
	}
}

// --- Installation ---

func TestInstallCommand(t *testing.T) {
	a := NewAdapter()

	tests := []struct {
		name    string
		profile system.PlatformProfile
		want    [][]string
	}{
		{
			name:    "darwin — no sudo",
			profile: system.PlatformProfile{OS: "darwin"},
			want:    [][]string{{"npm", "install", "-g", "@mariozechner/pi-coding-agent"}},
		},
		{
			name:    "linux system npm — sudo required",
			profile: system.PlatformProfile{OS: "linux", NpmWritable: false},
			want:    [][]string{{"sudo", "npm", "install", "-g", "@mariozechner/pi-coding-agent"}},
		},
		{
			name:    "linux nvm/pnpm — no sudo",
			profile: system.PlatformProfile{OS: "linux", NpmWritable: true},
			want:    [][]string{{"npm", "install", "-g", "@mariozechner/pi-coding-agent"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := a.InstallCommand(tt.profile)
			if err != nil {
				t.Fatalf("InstallCommand() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("InstallCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Config paths ---

func TestConfigPaths(t *testing.T) {
	a := NewAdapter()
	home := testHome

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"GlobalConfigDir", a.GlobalConfigDir(home), filepath.Join(home, ".pi", "agent")},
		{"SystemPromptDir", a.SystemPromptDir(home), filepath.Join(home, ".pi", "agent")},
		{"SystemPromptFile", a.SystemPromptFile(home), filepath.Join(home, ".pi", "agent", "AGENTS.md")},
		{"SkillsDir", a.SkillsDir(home), filepath.Join(home, ".pi", "agent", "skills")},
		{"SettingsPath", a.SettingsPath(home), filepath.Join(home, ".pi", "agent", "settings.json")},
		{"MCPConfigPath", a.MCPConfigPath(home, "any-server"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

// --- Strategies ---

func TestStrategies(t *testing.T) {
	a := NewAdapter()

	if got := a.SystemPromptStrategy(); got != model.StrategyFileReplace {
		t.Fatalf("SystemPromptStrategy() = %v, want StrategyFileReplace", got)
	}
}

// --- Capabilities ---

func TestCapabilities(t *testing.T) {
	a := NewAdapter()

	if !a.SupportsSkills() {
		t.Fatal("SupportsSkills() = false, want true")
	}
	if !a.SupportsSystemPrompt() {
		t.Fatal("SupportsSystemPrompt() = false, want true")
	}
	if a.SupportsMCP() {
		t.Fatal("SupportsMCP() = true, want false — Pi is philosophically opposed to MCP")
	}
	if !a.SupportsAutoInstall() {
		t.Fatal("SupportsAutoInstall() = false, want true")
	}
	if a.SupportsOutputStyles() {
		t.Fatal("SupportsOutputStyles() = true, want false")
	}
	if a.SupportsSlashCommands() {
		t.Fatal("SupportsSlashCommands() = true, want false")
	}
}
