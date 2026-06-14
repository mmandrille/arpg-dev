# v142 Spec: Sim Load And Players Extraction

Status: Complete
Date: 2026-06-13
Codename: `sim-load-and-players-extraction`

## Purpose

Move persistence-load helpers and player lifecycle/context helpers out of
`server/internal/game/sim.go` into focused package files. This is a behavior-preserving
maintainability slice that shrinks a determinism-critical coordinator without changing gameplay,
protocol, replay, or data rules.

## Non-goals

- No gameplay behavior, combat, movement, spawn-position, co-op, replay, protocol, or shared rule
  changes.
- No new bot scenario or world preset.
- No test rewrites beyond any compile-only import fallout.
- No broad `game_test.go` split in this slice.

## Acceptance Criteria

- `LoadInventory`, `LoadHotbar`, `LoadSkillBindings`, `LoadShopStock`, `LoadAccountStash`,
  `LoadDiscoveredTeleporters`, and their `ForPlayer` variants move from `sim.go` into
  `server/internal/game/sim_load.go` with unchanged signatures and behavior.
- Persisted item/stash/hotbar/skill binding structs and payload clone helpers move with the load
  code where practical, staying package-private or public exactly as before.
- `AddGuestPlayer`, player metadata/connectivity/query helpers, removal/respawn, town spawn
  selection/blocking helpers, player context switching, save/default helpers, and deterministic
  `sortedPlayerIDs` move into `server/internal/game/sim_players.go`.
- `server/internal/game/sim.go` shrinks materially and its ratchet baseline is lowered to the new
  line count.
- New Go files stay under the 600-line target.
- Determinism lint and replay-sensitive game tests remain green.
- `docs/CODEMAP.md` points relevant domains at the new sim load/player files.

## Scope And Likely Files

- `server/internal/game/sim.go` - remove moved load/player code.
- `server/internal/game/sim_load.go` - persisted load and payload clone helpers.
- `server/internal/game/sim_players.go` - player lifecycle, spawn, and context helpers.
- `.maintainability/file-size-baseline.tsv` - lower `sim.go` baseline.
- `docs/CODEMAP.md` - update Session/Combat domain file map as needed.
- `PROGRESS.md`, plan, and as-built docs.

## Test And Bot Proof

This slice is a package-internal refactor. Bot proof is covered by existing replay and CI gates; no
new scenario is required.

- `make lint-determinism`
- `cd server && go test ./internal/game ./internal/realtime`
- `make maintainability`
- `make ci`

## Open Questions And Risks

- No blocking product/design questions.
- Risk: moving helpers with broad package usage can leave imports stale. Mitigation: `gofmt`, Go
  package tests, determinism lint, and full CI.
