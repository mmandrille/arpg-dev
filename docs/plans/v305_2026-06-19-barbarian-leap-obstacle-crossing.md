# v305 Plan - Barbarian Leap Obstacle Crossing

Status: Complete
Goal: Let Barbarian Leap cross water and hole obstacle kinds while normal walking and other mobility remain blocked.
Architecture: Add a data-owned `ignore_obstacle_kinds` list to skill mobility rules and set it only on `leap`. Rework player mobility endpoint resolution so ignored obstacles are passable during the sweep but never valid landing tiles, while hard blockers like walls and closed doors still stop the move. Keep protocol output unchanged and reuse existing Leap presentation.
Tech stack: shared JSON rules/schema, Go deterministic sim mobility, Python protocol bot scenario, Godot visual replay, SDD docs.

## Baseline and Decisions

Baseline: v304 `flying-navigation-trait` is committed on `codex/world-detail-navigation` as `33aab019`.

Autoloop note: final batch `make ci` and the due review/refactor handoff remain deferred until the selected World Detail/Navigation queue completes.

Asset/plugin decision: reject external assets, imported Leap VFX, shaders, and Godot addons. Borrow the existing Leap skill cast visual, existing water/hole wall rendering, and compact world/bot lab patterns.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.json` | Mark `leap` as ignoring `water` and `hole` during mobility sweep. |
| Modify | `shared/rules/skills.v0.schema.json` | Validate optional mobility `ignore_obstacle_kinds`. |
| Modify | `shared/rules/worlds.v0.json` | Add compact Leap obstacle-crossing lab. |
| Modify | `server/internal/game/rogue_rules.go` | Load and validate mobility obstacle exceptions. |
| Modify | `server/internal/game/mobility_skills.go` | Use skill-owned mobility endpoint resolution for Leap/Charge. |
| Modify | `server/internal/game/rogue_skills.go` | Keep Dash on default obstacle blocking while sharing the resolver. |
| Add | `server/internal/game/mobility_obstacle_crossing_test.go` | Focused Leap-vs-Dash/walking obstacle proof. |
| Add | `tools/bot/scenarios/100_barbarian_leap_obstacle_crossing.json` | End-to-end Leap crossing proof. |
| Modify | `docs/specs/v305_spec-barbarian-leap-obstacle-crossing.md` | Mark complete after proof. |
| Modify | `docs/plans/v305_2026-06-19-barbarian-leap-obstacle-crossing.md` | Track execution checkboxes. |
| Add | `docs/as-built/v305_barbarian-leap-obstacle-crossing.md` | Record shipped behavior and proof. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v305 lifecycle row. |
| Modify | `PROGRESS.md` | Advance current status and keep autoloop review/refactor handoff due. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files expected:
- [x] `server/internal/game/rogue_skills.go` is under target but close enough to avoid growth where practical.
- [x] `server/internal/game/rules.go` should not need changes.
- [x] `server/internal/game/sim.go` should not be changed.
- [x] `tools/bot/run.py` should not need changes.
- [x] Other touched source/test/tool files must stay under their ratchet targets.

Decision:
- [x] Add a focused mobility obstacle test file and keep new resolver helpers out of large coordinators.

Verification:

```bash
make maintainability
```

## Task 1 - Shared Rules and Leap Lab

Files:
- Modify: `shared/rules/skills.v0.json`
- Modify: `shared/rules/skills.v0.schema.json`
- Modify: `shared/rules/worlds.v0.json`

- [x] Step 1.1: Add optional `mobility.ignore_obstacle_kinds` enum with values `water` and `hole`.
- [x] Step 1.2: Set `leap.mobility.ignore_obstacle_kinds` to `["water", "hole"]`.
- [x] Step 1.3: Add `barbarian_leap_obstacle_lab` with a water/hole strip and enough clear floor for a rank-1 Leap landing.

```bash
make validate-shared
```

## Task 2 - Server Mobility Semantics

Files:
- Modify: `server/internal/game/rogue_rules.go`
- Modify: `server/internal/game/mobility_skills.go`
- Modify: `server/internal/game/rogue_skills.go`

- [x] Step 2.1: Add `IgnoreObstacleKinds []string` to `SkillMobilityDef`.
- [x] Step 2.2: Validate unsupported ignored obstacle kinds and reject normal wall ignores.
- [x] Step 2.3: Resolve Leap endpoints with ignored obstacles passable during the sweep but invalid as landings.
- [x] Step 2.4: Keep Dash and Charge on the default hard-blocked resolver path.
- [x] Step 2.5: Keep normal walls and closed interactable barriers as hard blockers for all player mobility.

```bash
cd server && go test ./internal/game -run 'LeapObstacle|MobilityObstacle|RogueDash|GeneratedObstacleCollisionPaths'
```

## Task 3 - Focused Go Proof

Files:
- Add: `server/internal/game/mobility_obstacle_crossing_test.go`

- [x] Step 3.1: Prove ordinary movement stops at water/hole.
- [x] Step 3.2: Prove Leap crosses a water/hole strip and lands beyond it.
- [x] Step 3.3: Prove Leap does not land inside water/hole when range ends within the strip.
- [x] Step 3.4: Prove Leap still stops before normal walls.
- [x] Step 3.5: Prove Dash does not inherit Leap's obstacle exception.

```bash
cd server && go test ./internal/game -run 'LeapObstacle|MobilityObstacle|RogueDash'
```

## Task 4 - Bot and Visual Proof

Files:
- Add: `tools/bot/scenarios/100_barbarian_leap_obstacle_crossing.json`

- [x] Step 4.1: Add a protocol bot scenario that casts Leap east across the lab obstacle strip.
- [x] Step 4.2: Assert water and hole wall kinds exist in the scenario state.
- [x] Step 4.3: Assert the post-Leap player position lands beyond the obstacle strip.

```bash
ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=barbarian_leap_obstacle_crossing ./scripts/bot_local.sh
make client-unit
HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=barbarian_leap_obstacle_crossing ./scripts/bot_visual.sh
```

## Task 5 - Lifecycle Docs and Focused Gates

Files:
- Modify: `docs/specs/v305_spec-barbarian-leap-obstacle-crossing.md`
- Modify: `docs/plans/v305_2026-06-19-barbarian-leap-obstacle-crossing.md`
- Add: `docs/as-built/v305_barbarian-leap-obstacle-crossing.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark spec and plan complete after proof.
- [x] Step 5.2: Write as-built summary with Leap behavior, proof commands, and deferred scope.
- [x] Step 5.3: Update progress/current status and lifecycle row. Keep review/refactor handoff due after the selected queue unless a hard stop occurs.

```bash
make maintainability
```

## Final Verification

Focused autoloop slice gates:

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'LeapObstacle|MobilityObstacle|RogueDash|GeneratedObstacleCollisionPaths'`
- [x] `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=barbarian_leap_obstacle_crossing ./scripts/bot_local.sh`
- [x] `make client-unit`
- [x] `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=barbarian_leap_obstacle_crossing ./scripts/bot_visual.sh`
- [x] `make maintainability`

Autoloop batch gate:

- [ ] Final `make ci` is deferred to the selected batch after all requested world-detail/navigation slices are complete and committed.

## Deferred Scope

- General player pathing over water/holes, bridge/swim/fall/recovery systems, terrain costs, and
  normal walking exceptions remain out of scope.
- Dash, Charge, Teleport, monsters, companions, projectiles, fog/LOS, loot placement, and dungeon
  generation remain unchanged except for the new preset lab.
