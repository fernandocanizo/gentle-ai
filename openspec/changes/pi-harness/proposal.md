# Proposal: Pi Harness Support

## Intent

Pi (https://github.com/badlogic/pi-mono) is an open-source AI coding agent with a growing user base. gentle-ai currently supports 8 harnesses; Pi users cannot benefit from the ecosystem. This change adds Pi as a fully supported harness — giving Pi users SDD orchestration, skill installation, and system prompt injection — without MCP (Pi's author considers it philosophically wrong; shell scripts are the answer).

## Scope

### In Scope
- New `internal/agents/pi/` adapter package implementing the full `Adapter` interface
- `AgentPi` constant in `internal/model/types.go`
- Pi entry in `internal/agents/factory.go` and `internal/catalog/agents.go`
- Embedded asset directory `internal/assets/pi/` (uses generic SDD orchestrator — no Pi-specific variant)
- Detection: `pi` binary on PATH
- Auto-install: `npm install -g @mariozechner/pi-coding-agent`
- System prompt: `~/.pi/agent/AGENTS.md` via `StrategyFileReplace`
- Skills: `~/.pi/agent/skills/`
- Registration in CLI agent switches (`validate.go`, `model.go`)

### Out of Scope
- MCP support (permanent — Pi author's philosophical position)
- Engram `engram setup` slug (requires MCP)
- `PiEngine` in agentbuilder (headless mode not confirmed)
- Multi-named Pi agent dirs (`~/.pi/{custom-name}/`) — target `agent` only

## Capabilities

### New Capabilities
- `pi-agent-adapter` — full Adapter implementation for Pi coding agent

### Modified Capabilities
- `agent-registry` — Pi added to `NewDefaultRegistry` and `AllAgents`

## Option Chosen

**Option A** — skills + system prompt, no MCP. Safe, complete, honest. `SupportsMCP()` returns `false` permanently; the adapter is written so this can never silently break.

## Rollback Plan

All changes are additive. Removing Pi requires reverting the constant, the package, and the wiring lines — no other agents are affected.

## Risks
- `pi` is a short binary name; detection uses binary + config dir presence to avoid false positives.
- Pi may change its config path in future versions; paths are isolated in the adapter.
