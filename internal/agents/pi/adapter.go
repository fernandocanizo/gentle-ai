package pi

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/system"
)

var LookPathOverride = exec.LookPath

type statResult struct {
	isDir bool
	err   error
}

// Adapter implements agents.Adapter for Pi coding agent.
//
// Config path summary:
//   - Global AI config: ~/.pi/agent/
//     → AGENTS.md       (system prompt — full-replace)
//     → skills/         (directory-based skills, SKILL.md compatible)
//     → settings.json   (agent settings)
//
// Detection: Pi is detected when the `pi` binary is on PATH AND
// ~/.pi/agent/ exists as a directory. Either condition alone is
// insufficient — `pi` is a short, collision-prone binary name.
//
// MCP: Pi intentionally has no MCP support. The author's position
// is that shell scripts solve what MCP attempts to solve. SupportsMCP()
// returns false permanently. MCPStrategy() returns a sentinel value
// that is never exercised because the MCP component gates on SupportsMCP().
type Adapter struct {
	lookPath func(string) (string, error)
	statPath func(string) statResult
}

func NewAdapter() *Adapter {
	return &Adapter{
		lookPath: LookPathOverride,
		statPath: defaultStat,
	}
}

// --- Identity ---

func (a *Adapter) Agent() model.AgentID    { return model.AgentPi }
func (a *Adapter) Tier() model.SupportTier { return model.TierFull }

// --- Detection ---

// Detect reports Pi as installed only when BOTH conditions are true:
// the `pi` binary is on PATH and ~/.pi/agent/ exists as a directory.
func (a *Adapter) Detect(_ context.Context, homeDir string) (bool, string, string, bool, error) {
	configPath := a.GlobalConfigDir(homeDir)

	binaryPath, lookErr := a.lookPath("pi")
	binaryFound := lookErr == nil

	stat := a.statPath(configPath)
	if stat.err != nil {
		if os.IsNotExist(stat.err) {
			return false, binaryPath, configPath, false, nil
		}
		return false, "", "", false, stat.err
	}

	configFound := stat.isDir
	installed := binaryFound && configFound

	return installed, binaryPath, configPath, configFound, nil
}

// --- Installation ---

func (a *Adapter) SupportsAutoInstall() bool { return true }

func (a *Adapter) InstallCommand(profile system.PlatformProfile) ([][]string, error) {
	if profile.OS == "linux" && !profile.NpmWritable {
		return [][]string{{"sudo", "npm", "install", "-g", "@mariozechner/pi-coding-agent"}}, nil
	}
	return [][]string{{"npm", "install", "-g", "@mariozechner/pi-coding-agent"}}, nil
}

// --- Config paths ---

// GlobalConfigDir returns ~/.pi/agent — Pi's default agent config root.
func (a *Adapter) GlobalConfigDir(homeDir string) string {
	return filepath.Join(homeDir, ".pi", "agent")
}

// SystemPromptDir returns the same root as GlobalConfigDir.
// Pi stores AGENTS.md directly in the agent root, not a subdirectory.
func (a *Adapter) SystemPromptDir(homeDir string) string {
	return a.GlobalConfigDir(homeDir)
}

// SystemPromptFile returns ~/.pi/agent/AGENTS.md.
func (a *Adapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(a.GlobalConfigDir(homeDir), "AGENTS.md")
}

// SkillsDir returns ~/.pi/agent/skills.
// Pi uses directory-based skills (one directory per skill with SKILL.md),
// which is compatible with the gentle-ai skill format.
func (a *Adapter) SkillsDir(homeDir string) string {
	return filepath.Join(a.GlobalConfigDir(homeDir), "skills")
}

// SettingsPath returns ~/.pi/agent/settings.json.
func (a *Adapter) SettingsPath(homeDir string) string {
	return filepath.Join(a.GlobalConfigDir(homeDir), "settings.json")
}

// --- Config strategies ---

// SystemPromptStrategy uses FileReplace: Pi's AGENTS.md is a full-replace
// file, consistent with OpenCode and Codex.
func (a *Adapter) SystemPromptStrategy() model.SystemPromptStrategy {
	return model.StrategyFileReplace
}

// MCPStrategy returns a sentinel value that is never exercised.
// SupportsMCP() returns false, so the MCP component never calls this method.
func (a *Adapter) MCPStrategy() model.MCPStrategy {
	return model.StrategyMergeIntoSettings
}

// --- MCP ---

// MCPConfigPath returns empty string — Pi has no MCP support.
func (a *Adapter) MCPConfigPath(_ string, _ string) string { return "" }

// --- Optional capabilities ---

func (a *Adapter) SupportsOutputStyles() bool     { return false }
func (a *Adapter) OutputStyleDir(_ string) string { return "" }
func (a *Adapter) SupportsSlashCommands() bool    { return false }
func (a *Adapter) CommandsDir(_ string) string    { return "" }
func (a *Adapter) SupportsSkills() bool           { return true }
func (a *Adapter) SupportsSystemPrompt() bool     { return true }
func (a *Adapter) SupportsMCP() bool              { return false }

func defaultStat(path string) statResult {
	info, err := os.Stat(path)
	if err != nil {
		return statResult{err: err}
	}
	return statResult{isDir: info.IsDir()}
}
