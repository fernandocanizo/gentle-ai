# Tasks: Pi Harness Support

> Strict TDD mode active. Test runner: `go test ./...`

## Phase 1: Foundation

- [x] 1.1 Add `AgentPi AgentID = "pi"` constant to `internal/model/types.go` (after `AgentWindsurf`)
- [x] 1.2 Add `{ID: model.AgentPi, Name: "Pi", Tier: model.TierFull, ConfigPath: "~/.pi/agent"}` to `allAgents` slice in `internal/catalog/agents.go`

## Phase 2: RED — Write failing tests

- [x] 2.1 Create `internal/agents/pi/adapter_test.go` — `TestAgentIdentity` (Agent()="pi", Tier=TierFull) and `TestDetect` with 4 table cases: both present → installed=true; binary only → installed=false; dir only → installed=false; stat error → err bubbles up
- [x] 2.2 Add `TestInstallCommand` to `adapter_test.go` — 3 cases: darwin/NpmWritable=true → no sudo; linux/NpmWritable=false → sudo; linux/NpmWritable=true → no sudo
- [x] 2.3 Add `TestConfigPaths` to `adapter_test.go` — assert GlobalConfigDir, SystemPromptDir, SystemPromptFile, SkillsDir, SettingsPath, MCPConfigPath against expected values from spec path table
- [x] 2.4 Add `TestStrategies` to `adapter_test.go` — SystemPromptStrategy=StrategyFileReplace; `TestCapabilities` — SupportsSkills=true, SupportsSystemPrompt=true, SupportsMCP=false, SupportsAutoInstall=true, SupportsOutputStyles=false, SupportsSlashCommands=false
- [x] 2.5 Add Pi registry test to `internal/agents/registry_test.go`: Pi present in `NewDefaultRegistry().SupportedAgents()`; Pi absent from `NewMVPRegistry().SupportedAgents()`

## Phase 3: GREEN — Implement adapter

- [x] 3.1 Create `internal/agents/pi/adapter.go` — implement full `Adapter` interface; struct fields: `lookPath func(string) (string, error)`, `statPath func(string) statResult`; detection requires both binary found AND dir is directory; `MCPConfigPath` returns `""`; `MCPStrategy` returns `StrategyMergeIntoSettings` (sentinel, never exercised — add comment)
- [x] 3.2 Add Pi to `internal/agents/factory.go`: case in `NewAdapter` switch importing `pi` package; add `model.AgentPi` to the agent slice in `NewDefaultRegistry`
- [x] 3.3 Run `go test ./internal/agents/pi/... ./internal/agents/...` — all RED tests MUST be GREEN before proceeding

## Phase 4: Wiring

- [x] 4.1 Add `case string(model.AgentPi): agents = append(agents, model.AgentPi)` to `defaultAgentsFromDetection()` switch in `internal/cli/validate.go`
- [x] 4.2 Add `case string(model.AgentPi): selected = append(selected, model.AgentPi)` to the agent selection switch in `internal/tui/model.go` (around line 2236)

## Phase 5: Full verification

- [x] 5.1 Run `go test ./...` — zero failures
- [x] 5.2 Run `go vet ./...` — zero issues
- [x] 5.3 Run `golangci-lint run` — zero issues on changed files (pre-existing issues in backup/, restore.go, tui/screens/ untouched)
