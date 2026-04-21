# Pi Agent Adapter Specification

## Purpose

Defines the behavior of the Pi harness adapter: detection, path resolution, system prompt injection, skill installation, and capability declarations. Pi intentionally has no MCP support.

## Requirements

### Requirement: Agent Identity

The adapter MUST return `AgentID = "pi"` and `Tier = TierFull`.

#### Scenario: Identity is stable

- GIVEN a Pi adapter instance
- WHEN `Agent()` and `Tier()` are called
- THEN they return `"pi"` and `TierFull` respectively

---

### Requirement: Detection

The adapter MUST report Pi as installed when the `pi` binary is on PATH AND the `~/.pi/agent/` directory exists. Either condition alone MUST NOT be sufficient.

#### Scenario: Both binary and config dir present

- GIVEN `pi` is found on PATH
- AND `~/.pi/agent/` exists as a directory
- WHEN `Detect()` is called
- THEN `installed = true`, `binaryPath` is non-empty, `configPath = ~/.pi/agent`, `configFound = true`

#### Scenario: Binary on PATH but config dir absent

- GIVEN `pi` is found on PATH
- AND `~/.pi/agent/` does not exist
- WHEN `Detect()` is called
- THEN `installed = false`, `configFound = false`

#### Scenario: Config dir present but binary absent

- GIVEN `pi` is NOT found on PATH
- AND `~/.pi/agent/` exists
- WHEN `Detect()` is called
- THEN `installed = false`, `configFound = true`

#### Scenario: Stat error on config dir

- GIVEN `os.Stat` returns a non-`ErrNotExist` error for `~/.pi/agent/`
- WHEN `Detect()` is called
- THEN the error is returned and `installed = false`

---

### Requirement: Auto-Install

The adapter MUST support auto-install via `npm install -g @mariozechner/pi-coding-agent`. On Linux without npm write access, `sudo` MUST be prepended.

#### Scenario: Standard install

- GIVEN `profile.NpmWritable = true` (or non-Linux)
- WHEN `InstallCommand()` is called
- THEN returns `[["npm", "install", "-g", "@mariozechner/pi-coding-agent"]]`

#### Scenario: Linux without npm write access

- GIVEN `profile.OS = "linux"` AND `profile.NpmWritable = false`
- WHEN `InstallCommand()` is called
- THEN returns `[["sudo", "npm", "install", "-g", "@mariozechner/pi-coding-agent"]]`

---

### Requirement: Config Path Resolution

The adapter MUST resolve all config paths relative to `~/.pi/agent/`.

| Method | Expected Path |
|--------|--------------|
| `GlobalConfigDir(home)` | `~/.pi/agent` |
| `SystemPromptDir(home)` | `~/.pi/agent` |
| `SystemPromptFile(home)` | `~/.pi/agent/AGENTS.md` |
| `SkillsDir(home)` | `~/.pi/agent/skills` |
| `SettingsPath(home)` | `~/.pi/agent/settings.json` |
| `MCPConfigPath(home, _)` | `""` (MCP unsupported) |

#### Scenario: Paths are resolved correctly

- GIVEN `homeDir = "/home/user"`
- WHEN each path method is called
- THEN it returns the path from the table above (with `/home/user` substituted for `~`)

---

### Requirement: Config Strategies

The adapter MUST use `StrategyFileReplace` for system prompt injection (Pi's `AGENTS.md` is a full-replace file like OpenCode). The MCP strategy MUST be declared but MUST NOT be exercised since `SupportsMCP()` returns `false`.

#### Scenario: System prompt strategy

- GIVEN a Pi adapter
- WHEN `SystemPromptStrategy()` is called
- THEN it returns `StrategyFileReplace`

---

### Requirement: Capability Declarations

| Capability | Value | Reason |
|-----------|-------|--------|
| `SupportsSkills()` | `true` | `~/.pi/agent/skills/` is directory-based, SKILL.md compatible |
| `SupportsSystemPrompt()` | `true` | `AGENTS.md` present |
| `SupportsMCP()` | `false` | Permanent — author's philosophical position |
| `SupportsAutoInstall()` | `true` | npm-installable |
| `SupportsOutputStyles()` | `false` | Not supported |
| `SupportsSlashCommands()` | `false` | Not supported |

#### Scenario: Capabilities match declarations

- GIVEN a Pi adapter
- WHEN each `Supports*()` method is called
- THEN it returns the value from the table above

---

### Requirement: Registry and Catalog Registration

`AgentPi` MUST be included in `NewDefaultRegistry()` and `catalog.AllAgents()`. It MUST NOT be in `NewMVPRegistry()` or `catalog.MVPAgents()`.

#### Scenario: Pi is in the default registry

- GIVEN `NewDefaultRegistry()` is called
- WHEN `registry.Get("pi")` is called
- THEN it returns a valid adapter and `ok = true`

#### Scenario: Pi is not in the MVP registry

- GIVEN `NewMVPRegistry()` is called
- WHEN `registry.SupportedAgents()` is examined
- THEN `"pi"` is NOT in the list
