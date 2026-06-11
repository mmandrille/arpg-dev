# v82 Plan - Realtime Fanout Level Snapshot

Status: Ready for implementation
Goal: Close the realtime fanout review finding by snapshotting client levels under the session loop lock.
Architecture: The session loop already copies connected clients while holding `sessionLoop.mu`.
This slice extends that tick-time snapshot with a `playerID -> level` map and passes it into fanout.
`fanoutResult` stays a pure distributor over the tick result, client list, input types, and captured
client levels instead of re-reading mutable sim state after unlock.
Tech stack: Go realtime loop and focused Go tests.

## Baseline and shortcut decision

Builds on v81 `paladin-holy-shield`. No Godot plugin/adoption decision is needed because this is
server-only realtime infrastructure.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/realtime/session_loop.go` | Capture client levels under lock and pass them to fanout. |
| Modify | `server/internal/realtime/session_loop_test.go` | Update direct fanout tests to provide explicit level snapshots. |
| Modify | `tools/bot/scenarios/32_skill_points_and_magic_bolt.json` | Remove stale attack-interval-derived tuning pins found by CI. |
| Modify | `tools/bot/scenarios/39_rage_and_heal_skills.json` | Remove stale exact skill cooldown tick pin found by CI. |
| Modify | `tools/bot/scenarios/40_paladin_heal_skill.json` | Remove stale exact skill cooldown tick pin found by CI. |
| Modify | `tools/bot/scenarios/43_paladin_holy_shield.json` | Remove stale exact skill cooldown tick pin found by CI. |
| Modify | `tools/bot/scenarios/client/19_skill_points_and_magic_bolt.json` | Keep the matching client scenario from pinning attack interval tuning. |
| Modify | `shared/rules/worlds.v0.json` | Add a low-HP safe dummy to `combat_stat_lab` for deterministic death-reaction proof. |
| Modify | `tools/bot/scenarios/client/12_model_reaction_polish.json` | Keep model reaction proof semantic under current combat cadence. |
| Add | `docs/as-built/v82_realtime-fanout-level-snapshot.md` | Summarize shipped behavior. |
| Modify | `PROGRESS.md` | Mark v82 complete and record deferred scope. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/realtime/session_loop.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected

Decision:
- [x] Defer extraction with rationale: this slice changes a narrow concurrency boundary in an
  already-grandfathered realtime loop; extracting while changing lock/fanout semantics would add
  unnecessary risk.

Verification:
```bash
make maintainability
```

## Task 1 - Snapshot client levels in fanout

Files:
- Modify: `server/internal/realtime/session_loop.go`

- [x] Step 1.1: Add a client-level snapshot map in `doTick` while `sessionLoop.mu` is held.
- [x] Step 1.2: Pass the map into `fanoutResult`.
- [x] Step 1.3: Change `fanoutResult` to use the captured level map and stop reading from `l.sim`.

```bash
cd server && go test ./internal/realtime/...
```

## Task 2 - Update focused realtime tests

Files:
- Modify: `server/internal/realtime/session_loop_test.go`

- [x] Step 2.1: Update direct `fanoutResult` callers to pass explicit client-level snapshots.
- [x] Step 2.2: Preserve existing same-level/cross-level fanout expectations.

```bash
cd server && go test ./internal/realtime/...
```

## Task 3 - Lifecycle docs and CI

Files:
- Add: `docs/as-built/v82_realtime-fanout-level-snapshot.md`
- Modify: `PROGRESS.md`
- Modify: `docs/specs/v82_spec-realtime-fanout-level-snapshot.md`
- Modify: `docs/plans/v82_2026-06-11-realtime-fanout-level-snapshot.md`

- [x] Step 3.1: Mark the spec complete and all plan tasks complete.
- [x] Step 3.2: Add the as-built note and lifecycle/progress updates.
- [x] Step 3.3: Remove stale attack-interval-derived exact bot expectations uncovered by CI.
- [x] Step 3.4: Run final verification.

```bash
make maintainability
make ci
```

## Final verification

- [x] `cd server && go test ./internal/realtime/...`
- [x] `VERBOSE=1 make bot scenario=32_skill_points_and_magic_bolt`
- [x] `VERBOSE=1 make bot scenario=39_rage_and_heal_skills,40_paladin_heal_skill`
- [x] `VERBOSE=1 make bot scenario=43_paladin_holy_shield`
- [x] `make validate-shared`
- [x] `VERBOSE=1 HEADLESS=1 make bot-client scenario=12_model_reaction_polish`
- [x] `make maintainability`
- [x] `make ci`

Deferred scope: broader realtime loop extraction, protocol changes, and gameplay behavior changes.
