# v509 — Boss lane telegraphs (detached worktree handoff)

- **Status:** Integrated; focused checks and combined batch `make ci` passed (11m41s, 2026-10-01). Live reconnect and matched render-cost gates remain partial.
- **Base:** `5365832b9029e0d9e178d57f2a95a6a697b01acc`.
- **Product choice:** Each attack locks one server-authored world-space lane frame at warning start. Its safe index remains stable through both warnings and the strike; a shared rule sequence selects safe lanes between attacks.

## Implementation

`shifting_bulwark` is appended to the Cave Warden and Crypt Matron decks. Shared rules own lane geometry, safe sequence, bounded aim offsets, colors, warning intensity, phase timing, damage, and cooldown. The server tries the authored offsets until the safe corridor is walkable, retains the chosen world frame, emits the descriptor in phase events and entity/snapshot state, and resolves player-radius-aware hits against that frame with one hit per player per active phase. Go tests cover stage timing, blocked casts, safe/danger boundaries, co-op, deterministic replay, and rule overrides.

The client renders three lane strips and two safe-lane boundaries from the server descriptor. It retains world-space placement while the boss moves, restores from a phase snapshot, and removes the marker after the phase, death, entity removal, or level change. During the first real-camera pass, generic model tint replaced the new marker materials with opaque orange. The client now excludes the lane marker from model tint and verifies its material before reusing its mesh. A focused Godot test reproduces and repairs that overwrite.

## Focused checks

| Check | Result | Evidence |
|---|---|---|
| `make validate-shared` | PASS | 2,209 shared checks plus CODEMAP validation. |
| `go test ./internal/game -run '^TestBoss' -count=1` | PASS | Rule bounds, stage/state projection, safe/danger and co-op hits, blocked cast, fixed frame, and replay. |
| `.venv/bin/pytest tools/bot/test_boss_lane_queries.py tools/test_scenario_movement_audit.py tools/test_ci_pack.py -q` | PASS | 12 tests, including server-authored movement target and scenario audit. |
| `make bot scenario=boss_lane_telegraphs` | PASS | Pinned Crypt Matron lab; two warning stages, active/recovery order, safe-lane position, and no HP decrease during strike. |
| `make bot scenario=boss_lane_danger_hit` | PASS | Independent pinned lab attempt reaches a danger lane and loses HP after the active strike. |
| `make bot-client SCENARIO=boss_lane_telegraphs_visual HEADLESS=1` | PASS | Client marker and boss bar across both warning stages. |
| `godot --headless --path client --script res://tests/test_boss_lane_marker.gd` | PASS | Geometry, fixed frame, mesh reuse, tint overwrite repair, snapshot restore, and cleanup. |
| `make bot-visual scenario=boss_lane_telegraphs_visual` | PASS | Real Forward+ Metal window, 1920×1080, Balanced and Performance, stages 0 and 1; inspected all four PNGs. |
| `make lint-determinism`, `make maintainability`, `make client-unit` | PASS | Final source-state focused gates. |

The visual seed is `boss_lane_telegraphs`, world `boss_floor_gate_lab`, Crypt Matron on depth 5. Captures at ticks 293 and 306 (stage 0) and 320 and 335 (stage 1) record `safe_index=1`, `safe_color=#42e5af`, `safe_intensity=0.65`, danger indices `[0,2]`, five marker meshes, no legacy marker, and the pause menu closed. Stage 0 danger intensity is 0.35; stage 1 is 0.7. The safe corridor is visibly mint in each inspected frame:

| Tier / stage | Screenshot | Capture manifest |
|---|---|---|
| Balanced / 0 | [PNG](assets/v509/v509_lane_balanced.png) | [JSON](assets/v509/v509_lane_balanced.json) |
| Performance / 0 | [PNG](assets/v509/v509_lane_performance_stage0.png) | [JSON](assets/v509/v509_lane_performance_stage0.json) |
| Performance / 1 | [PNG](assets/v509/v509_lane_performance.png) | [JSON](assets/v509/v509_lane_performance.json) |
| Balanced / 1 | [PNG](assets/v509/v509_lane_balanced_stage1.png) | [JSON](assets/v509/v509_lane_balanced_stage1.json) |

## Partial gates and integration follow-up

- **Runtime reconnect:** The Go snapshot projection test and Godot snapshot-restoration test pass. A live reconnect attempt reopened the same session, but the boss lab then displayed prior deck phases while the client debug tick remained at 285. The attempted lane-specific reconnect scenario was removed rather than committing a failing extended test. The existing boss-floor reconnect behavior needs diagnosis before live restoration can be claimed. Raw attempts remain in ignored `.artifacts/v509/bot-client-reconnect*.log`.
- **Matched render cost:** The four real-camera captures establish readability and their capture context, not a performance comparison. No matched pre/post samples were collected at identical boss phase/tick/camera/quality with at least three valid runs per side and tier. Frame/process p50/p95 and draw-call budget remain unverified; screenshot FPS is not evidence for that gate.
- **Coordinator closeout:** Integrated paths were compared and combined `make ci` passed in 11m41s on 2026-10-01. Lifecycle and progress records are being closed out; the live-reconnect and matched-cost gaps remain open.

For a direct human visual check, run `make bot-visual scenario=boss_lane_telegraphs_visual` in an environment with a windowed Godot renderer.

## Handoff manifest

The 18 regenerated tracked `client/assets/**/*.glb.import` sidecars were restored after Godot. No branch, commit, push, main edit, or full CI was made in this slice worktree. The windowed Godot runner is released to the coordinator.

**Modified tracked files (24):**

```text
client/scripts/boss_visuals_controller.gd
client/scripts/bot_presentation_debug.gd
client/scripts/bot_wait_handlers.gd
client/scripts/main.gd
docs/CODEMAP.md
docs/progress/scenario-movement-audit.tsv
scripts/client_smoke.sh
server/internal/game/boss_pattern_rules.go
server/internal/game/boss_patterns.go
server/internal/game/entity_view.go
server/internal/game/rules.go
server/internal/game/sim.go
server/internal/game/sim_tick_monster_movement.go
server/internal/game/types.go
shared/protocol/session_snapshot.v8.schema.json
shared/protocol/state_delta.v8.schema.json
shared/rules/boss_patterns.v0.json
shared/rules/boss_patterns.v0.schema.json
shared/rules/boss_templates.v0.json
tools/bot/run.py
tools/bot/runtime_assertions.py
tools/bot/runtime_queries.py
tools/bot/scenario_movement_audit.py
tools/validate_boss_patterns.py
```

**Added untracked files (20):**

```text
client/scripts/boss_lane_marker.gd
client/tests/test_boss_lane_marker.gd
docs/as-built/assets/v509/v509_lane_balanced.json
docs/as-built/assets/v509/v509_lane_balanced.png
docs/as-built/assets/v509/v509_lane_balanced_stage1.json
docs/as-built/assets/v509/v509_lane_balanced_stage1.png
docs/as-built/assets/v509/v509_lane_performance.json
docs/as-built/assets/v509/v509_lane_performance.png
docs/as-built/assets/v509/v509_lane_performance_stage0.json
docs/as-built/assets/v509/v509_lane_performance_stage0.png
docs/as-built/v509_boss-lane-telegraphs.md
docs/plans/v509_2026-10-01-boss-lane-telegraphs.md
docs/specs/v509_spec-boss-lane-telegraphs.md
server/internal/game/boss_lanes.go
server/internal/game/boss_lanes_test.go
tools/bot/boss_lane_actions.py
tools/bot/scenarios/123_boss_lane_telegraphs.json
tools/bot/scenarios/124_boss_lane_danger_hit.json
tools/bot/scenarios/client/110_boss_lane_telegraphs_visual.json
tools/bot/test_boss_lane_queries.py
```

**Ignored local evidence:** `.artifacts/v509/` contains focused test, bot, visual, reconnect-attempt, and validation logs; `.artifacts/bot-captures/v509_lane_*.png` and matching JSON are the raw final captures. Selected four PNG/JSON pairs are copied above into `docs/as-built/assets/v509/`. Local `.godot/` and `.venv/` are tool caches, not handoff source.
