# v509 Plan — Boss Lane Telegraphs

- **Status:** Implemented and combined-batch-CI verified; live reconnect and matched render-cost evidence remain partial.
- **Date:** 2026-10-01
- **Spec:** [v509 boss lane telegraphs](../specs/v509_spec-boss-lane-telegraphs.md)
- **Goal:** Add a staged, deterministic boss lane warning with a visible safe route and a
  server-authored strike matching the warning.
- **Baseline:** detached batch worktree at `5365832b9029e0d9e178d57f2a95a6a697b01acc`.
- **Prerequisites:** v250, v282, v287, and v498 are complete at the base. No v501–v508 or v510
  runtime change is required to draft or implement this slice. Recheck changed shared/client paths
  against the coordinator's latest integrated state before implementation and again at handoff.

## Spec review gate

The spec bounds this to one new pattern in both current boss decks, with a stable safe lane for an
individual warning. It maps acceptance to rule validation, Go hit/replay tests, v8 wire schemas,
protocol and Godot client bots, reconnect, and real-camera captures. Timing and geometry remain
shared data; collision, lane choice, and damage remain server-owned. The in-repo asset decision is
recorded. No new golden formula is proposed; if execution introduces a cross-language evaluator,
add paired Go/GDScript golden coverage. Current `BossVisualsController` draws a marker attached to
the boss, and current `BossPhaseView` holds timing but no geometry, so this plan includes a locked
world-space descriptor in events and snapshots. The rules validator currently compares one active
shape with one prior telegraph; it needs a narrow lane-pattern branch without weakening existing
patterns.

**Resolved decision:** The owner chose a lane fixed at its telegraph start position. It does not
track the player during the warning. The existing stable-lane design applies; separate attacks
may choose another safe lane through the authored sequence.

## Design and file ownership

| Files | Change | Owner / integration note |
|---|---|---|
| `shared/rules/boss_patterns.v0.json`, `.schema.json`, `boss_templates.v0.json` | Add one pattern and append to Cave Warden and Crypt Matron decks; encode lane dimensions, stage timing/intensity, safe-lane sequence, strike damage | Shared tuning; compare sibling rule edits before integration |
| `server/internal/game/rules.go`, `boss_pattern_rules.go`, new focused `boss_lanes.go` | Decode/validate rule data; lock world-space frame and safe lane; calculate lane collision without RNG or wall clock | Server authority; avoid growing the 432-line `boss_patterns.go` with the new geometry domain |
| `server/internal/game/boss_patterns.go`, `sim.go`, `sim_tick_monster_movement.go`, `entity_view.go`, `types.go` | Narrow scheduling/state hooks, stable frame lifetime, pattern-specific movement rule, event/snapshot projection | Keep existing boss patterns and event ordering stable |
| `shared/protocol/state_delta.v8.schema.json`, `session_snapshot.v8.schema.json` | Add a bounded lane descriptor to current boss phase/event shape | Coordinate exact JSON tags with Go; no unrelated protocol bump |
| `client/scripts/boss_visuals_controller.gd`, new focused `boss_lane_marker.gd`, client tests | Reuse marker lifecycle and material; draw world-space lane bands and safe corridor from server descriptor, reflect stages and cleanup | Client display only; avoid edits to `main.gd` unless routing truly requires them |
| `tools/bot/scenarios/NN_boss_lane_telegraphs.json`, `tools/bot/scenarios/client/NN_boss_lane_telegraphs_visual.json` | Pinned boss-floor lab proof, event/damage assertions, real camera and reconnect marker check | Choose unused numeric prefixes at execution; mark extended unless CI-pack curation justifies promotion |
| `docs/as-built/v509_boss-lane-telegraphs.md`, `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Final evidence and lifecycle | Coordinator closes after integration and combined CI |

The authoritative descriptor should include locked world-space origin, forward/right unit axes,
lane count/width/length, selected safe-lane index, warning-stage index, and the currently energized
danger lanes. Its single source is server state retained for the complete attack. Active collision
reads that retained frame, not the boss's later position or a client reconstruction. Event and
snapshot serialization both use the same projection helper. Keep stage durations and intensity in
shared JSON; use tick offsets only. Validation must ensure the safe corridor is wide enough for a
player collision footprint and stage durations meet the configured minimum warning.
Try the data-authored aim offsets in order at cast start and use the first frame with a walkable
safe corridor; skip the cast when none qualifies. The chosen frame then stays fixed.

The current bot lab is `boss_floor_gate_lab`; setup should start there directly and avoid
incidental navigation. Append the deck entry after current bot-proven entries so existing
`boss_floor_gate` and `boss_telegraph_decals` sequences remain intact. A focused test should
select the new pattern directly or advance under a bounded, rule-derived tick budget rather than
copying today's cooldown values into assertions.

## Asset and plugin choice

- **Adopt:** existing `BossVisualsController` marker lifecycle, boss bar, boss-floor lab, and
  in-repo Godot procedural meshes/materials.
- **Borrow:** rectangle predicate and v498 real-renderer warning capture method.
- **Reject:** external assets/plugins and a separate imported decal or shader pipeline.

## Ordered tasks

### 0. Resolve the product and integration gates

- [x] Record the owner's fixed-lane answer and coordinator authorization to execute.
- [x] Compare coordinator integrated HEAD and sibling changed paths with this file map; resolve
  any overlap before execution. Keep all experiments inside this detached v509 worktree.

Check: document review plus `git status --short` in the assigned worktree. No runtime changes.

### 1. Author and validate the shared pattern

- [x] Add lane and warning-stage schema fields with explicit bounds and a safe-lane sequence.
  Encode every duration, width, length, color/intensity, and cooldown in rules, not Go or GDScript.
- [x] Append the pattern to both boss decks after the existing sequence, and validate that every
  active damage zone is a subset of the displayed danger lanes. Preserve old pattern validation.
- [x] Add focused invalid-rule fixtures and a fixture that varies stage duration and lane width
  to prove runtime follows the modified data.

Smallest checks: `make validate-shared` and
`(cd server && go test ./internal/game -run 'TestBossLane.*Rules' -count=1)`.

### 2. Implement server-authoritative lane state, strike, and wire projection

- [x] At pattern start, choose the safe index from the data sequence using stable attack/deck
  order, lock world-space frame, and retain it across stages/active. Check the arena's walkable
  floor for a contiguous safe corridor; skip the cast if no valid frame exists. Do not consume
  unrelated RNG.
- [x] Progress warning stages by fixed ticks using the existing phase-start event for each stage,
  stop boss motion or keep the
  fixed zone independent of motion, and guarantee damage starts only after the last stage.
- [x] Hit-test player-radius-aware lane boundaries against the locked frame; one normal combat
  outcome at most per player per active phase. Cover safe, danger, exact edge, and co-op positions.
- [x] Expose the same frame/stage in `boss_phase_started` events, state deltas, and reconnect
  snapshots. Update Go wire types and current v8 schemas together.
- [x] Add pinned-seed two-simulation/replay equality tests for lane choice, stage/event order,
  combat outputs, and end state under identical ordered inputs. Add snapshot/reconnect projection
  test. No new client or bot damage authority.

Smallest checks: `(cd server && go test ./internal/game -run 'TestBossLane' -count=1)`,
`make validate-shared`, and `make lint-determinism`.

### 3. Reuse the client telegraph path for a lane marker

- [x] Extract a focused lane marker helper called by `BossVisualsController`, with world-space
  placement from the authoritative descriptor; preserve existing marker/boss bar behavior.
- [x] Show the full footprint and stable safe corridor at first warning, then intensify the
  configured danger lanes on stage events. Restore the marker from a warning snapshot after
  reconnect; remove it on phase end, entity removal, level change, and boss death.
- [x] Expose stage, safe-lane, marker geometry, and cleanup in client debug state, with focused
  Godot tests. Verify Balanced and Performance legibility against fog and dungeon lighting.

Smallest checks: focused Godot lane-marker test, then `make client-unit` and
`make bot-client SCENARIO=boss_lane_telegraphs_visual HEADLESS=1`.

### 4. Prove gameplay and real-camera behavior

- [x] Add protocol scenarios `boss_lane_telegraphs` and `boss_lane_danger_hit` with pinned seed/lab,
  bounded waits, stage event order, a safe-lane dodge without strike damage, and a separate
  danger-lane hit. Go tests cover deterministic replay and varied rule fixtures.
- [x] Add client scenario `boss_lane_telegraphs_visual` that asserts marker presence and stage
  changes, and captures the active player-camera warning with HUD and fog. Inspect captures at
  native size in Balanced and Performance; retain selected PNGs. Snapshot restoration passes a
  focused Godot test. Live boss-floor reconnect was attempted and remains a documented partial
  gate because the session resumed on earlier deck phases with a frozen client debug tick.
- [ ] PARTIAL — the new geometry adds visible mesh cost; a fresh matched baseline/candidate comparison remains unverified. Compare at the
  same seed, boss phase/tick, camera, renderer, resolution, quality tier, and machine. Keep raw
  frame p50/p95, process p95, and draw-call samples (at least three valid runs per side/tier).
  Budget: no more than `max(0.5 ms, 5%)` added frame/process p95 or `max(10 calls, 5%)` added
  draw calls. If control variance exceeds a limit, record it and revise the budget before tuning.

Smallest checks: `make bot scenario=boss_lane_telegraphs`,
`make bot-client SCENARIO=boss_lane_telegraphs_visual HEADLESS=1`, and exact human visual command:

```bash
make bot-visual scenario=boss_lane_telegraphs_visual
```

This command dispatches the client scenario through the existing visible `bot-visual` route.
Save the chosen capture manifest and PNGs under `.artifacts/` while iterating; copy only cited
final evidence into `docs/as-built/assets/v509/` at handoff.

### 5. Focused handoff and coordinator closeout

- [x] Run `make validate-shared`, focused Go/Go replay tests, `make lint-determinism`, focused
  Godot tests / `make client-unit`, both focused bot scenarios, and `make maintainability`.
- [x] Record the exact commands, results, seed, stage timeline, safe/danger outcome, snapshot
  restoration, visual captures, and performance limits in the v509 as-built draft.
- [x] Report every changed and untracked path plus any ignored evidence to the coordinator. Do
  not commit, push, clean up this worktree, edit `main`, or run `make ci`/`make ci-full` here.

After all accepted slices are integrated, the **coordinator** compares the integrated files and
behavior with this handoff, runs combined `make ci`, closes PROGRESS/lifecycle/as-built, and commits.

## Maintainability and conflict boundaries

`boss_patterns.go` is 432 lines and `boss_visuals_controller.gd` is 291 at this base; add lane
geometry in focused new files instead of bloating either coordinator. `tools/bot/run.py` is an
over-limit grandfathered file; avoid touching it by using existing scenario actions. If a shared
over-limit file must change, follow touch-to-shrink and update the ratchet baseline only for a
real extraction or documented tightly coupled exception. Keep any new source/test/tool file at or
below 600 lines. Expected shared-file overlap is boss/presentation schemas or Godot helpers from
other graphics slices; resolve against the integrated coordinator state, preserving both changes.
