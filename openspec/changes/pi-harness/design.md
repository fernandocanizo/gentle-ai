# Design: Pi Harness Support

## Technical Approach

Follow the established Adapter pattern exactly. Pi maps cleanly onto existing strategies: `StrategyFileReplace` for system prompt (shared with OpenCode, Codex) and no MCP. Detection requires both binary on PATH **and** config dir present — a tighter guard than Codex (which reports installed on binary alone) justified by `pi` being a short, collision-prone name.

## Architecture Decisions

### Decision: Dual-condition detection

**Choice**: `installed = lookPath("pi") == nil && statPath(~/.pi/agent).isDir`  
**Alternatives considered**: Binary-only check (like Codex); dir-only check (like Windsurf)  
**Rationale**: `pi` is a common Unix name (Python's `pi`, math libs, etc.). Binary-only would cause false positives. Dir-only misses the case where the dir was manually created but Pi isn't installed. Both together is safe.

### Decision: No assets directory for Pi

**Choice**: Rely on `generic/sdd-orchestrator.md` fallback in `sddOrchestratorAsset()`  
**Alternatives considered**: Create `internal/assets/pi/sdd-orchestrator.md` with Pi-specific content  
**Rationale**: `inject.go:859` already falls back to `"generic/sdd-orchestrator.md"` for unlisted agents. Pi doesn't need agent-specific orchestrator instructions. No change to `assets.go` or new directory needed — this simplifies the diff.

### Decision: MCPStrategy returns StrategyMergeIntoSettings as sentinel

**Choice**: `MCPStrategy()` returns `StrategyMergeIntoSettings`; `SupportsMCP()` returns `false`  
**Alternatives considered**: Panic; return a new `StrategyNone` constant  
**Rationale**: The interface requires a return value. Since `SupportsMCP() = false`, the MCP component never calls `MCPStrategy()` or `MCPConfigPath()`. Returning an existing constant is safe and avoids a model change. This is documented clearly in the adapter.

### Decision: Skip engram setup slug

**Choice**: Pi is not added to `SetupAgentSlug()` in `engram/setup.go`  
**Alternatives considered**: Add slug with `ShouldAttemptSetup` returning false  
**Rationale**: `SetupAgentSlug` exists to enable `engram setup` CLI, which works via MCP. Pi has no MCP. Adding a slug that always returns false is noise. The `default: return "", false` branch already handles it correctly.

## Data Flow

```
NewDefaultRegistry()
  └── NewAdapter(AgentPi) → pi.Adapter{}

pi.Adapter.Detect()
  ├── lookPath("pi")      → binaryPath / err
  └── statPath(~/.pi/agent) → isDir / err
      installed = binary found AND dir is a directory

Install flow (if auto-install):
  InstallCommand() → [["npm","install","-g","@mariozechner/pi-coding-agent"]]
  (Linux !NpmWritable) → [["sudo","npm","install","-g","@mariozechner/pi-coding-agent"]]

SDD inject:
  sddOrchestratorAsset("pi") → "generic/sdd-orchestrator.md"  (fallback, no Pi case)

Skills install:
  adapter.SkillsDir(home) → ~/.pi/agent/skills/

System prompt inject:
  adapter.SystemPromptFile(home) → ~/.pi/agent/AGENTS.md
  StrategyFileReplace → full file replacement
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/model/types.go` | Modify | Add `AgentPi AgentID = "pi"` constant |
| `internal/agents/pi/adapter.go` | Create | Full Adapter implementation |
| `internal/agents/pi/adapter_test.go` | Create | Tests for all spec scenarios |
| `internal/agents/factory.go` | Modify | Add Pi case in `NewAdapter` + `NewDefaultRegistry` |
| `internal/catalog/agents.go` | Modify | Add Pi entry in `allAgents` |
| `internal/cli/validate.go` | Modify | Add `AgentPi` case in both agent switches (lines ~171, ~186) |
| `internal/tui/model.go` | Modify | Add `AgentPi` case in agent selection switch (line ~2236) |

**No changes to**: `assets.go`, `engram/setup.go`, `sdd/inject.go`, `agentbuilder/engine.go`

## Interfaces / Contracts

```go
// internal/model/types.go
AgentPi AgentID = "pi"

// internal/catalog/agents.go
{ID: model.AgentPi, Name: "Pi", Tier: model.TierFull, ConfigPath: "~/.pi/agent"}

// internal/agents/pi/adapter.go — key method signatures
func (a *Adapter) Detect(_ context.Context, homeDir string) (bool, string, string, bool, error)
func (a *Adapter) SupportsAutoInstall() bool              { return true }
func (a *Adapter) SupportsMCP() bool                     { return false }
func (a *Adapter) SupportsSkills() bool                  { return true }
func (a *Adapter) SupportsSystemPrompt() bool            { return true }
func (a *Adapter) SystemPromptStrategy() model.SystemPromptStrategy { return model.StrategyFileReplace }
func (a *Adapter) MCPStrategy() model.MCPStrategy        { return model.StrategyMergeIntoSettings } // never exercised
func (a *Adapter) MCPConfigPath(_, _ string) string      { return "" }
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Detection (4 cases), install command (3 cases), path resolution, strategies, capabilities, identity | Table-driven tests in `adapter_test.go`; inject `lookPath` and `statPath` via struct fields |
| Integration | Pi appears in `NewDefaultRegistry()` and `AllAgents()`; absent from MVP variants | `registry_test.go` pattern |
| E2E | Not applicable — no E2E for adapters |

Test structure mirrors `codex/adapter_test.go` exactly: injectable `lookPath func` + `statPath func` in the struct, `const testHome = "/tmp/home"`.

## Migration / Rollout

No migration required. All changes are additive. Pi shows up in `AllAgents()` automatically — existing users are unaffected.

## Open Questions

None. The spec fully constrains the implementation.
