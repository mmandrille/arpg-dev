# v57 Plan - Boss Phase Readability

Status: Complete
Goal: Make existing boss telegraph timing readable through the boss health bar and an in-world
warning marker, without changing server combat or protocol.
Architecture: Reuse server-authored `boss_phase_started` / `boss_phase_ended` events and boss
entity metadata. The client owns only presentation: a local phase countdown seeded from
`duration_ticks`, boss bar display fields, and a primitive telegraph marker under the boss node.
Tech stack: Godot GDScript client, client bot scenarios, existing Python protocol bot regression,
SDD docs.

## Baseline and shortcut decision

Baseline is v56 `monster-attack-cadence` on `main`.

Adoption checklist:
- Adopt: existing in-repo `BossHealthBar`, `get_bot_state()`, procedural primitive marker, and
  client bot assertion patterns.
- Borrow: current boss tint behavior and archer marker primitive style.
- Reject: external plugins/assets. The slice is a narrow presentation layer over existing server
  phase facts; adding dependencies would increase surface area without reducing risk.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/boss_health_bar.gd` | Phase label/progress/debug state |
| Modify | `client/scripts/main.gd` | Seed local countdown, sync boss bar phase, telegraph marker |
| Modify | `client/scripts/bot_scenario_runner.gd` | Phase and presentation assertion support |
| Modify | `client/tests/test_boss_health_bar.gd` | Unit coverage for phase display/debug state |
| Modify | `client/tests/test_client_bot.gd` | Runner validation/matching coverage |
| Add | `tools/bot/scenarios/client/28_boss_phase_readability.json` | Client-bot proof |
| Add | `docs/specs/v57_spec-boss-phase-readability.md` | Slice spec |
| Add | `docs/plans/v57_2026-06-10-boss-phase-readability.md` | This plan |
| Modify | `PROGRESS.md` | Lifecycle close-out |
| Add | `docs/as-built/v57_boss-phase-readability.md` | As-built proof |

## Task 1 - Boss health bar phase state

Files:
- Modify: `client/scripts/boss_health_bar.gd`
- Modify: `client/tests/test_boss_health_bar.gd`

- [x] Step 1.1: Add `set_phase_state(phase: Dictionary)` and `clear_phase_state()` methods.
- [x] Step 1.2: Render a compact phase label and progress strip under the existing hp bar.
- [x] Step 1.3: Extend `get_debug_state()` with phase fields.
- [x] Step 1.4: Add unit assertions for telegraph phase state, clamping, ratio, and clearing.

```bash
make client-unit
```

## Task 2 - Main client phase countdown and telegraph marker

Files:
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Add a local `BOSS_PHASE_TICK_RATE` constant for display-only countdown math.
- [x] Step 2.2: On `boss_phase_started`, store `remaining_ticks`, `duration_ticks`, `pattern_id`,
  `phase_index`, `phase_kind`, and telegraph/hit-shape metadata on the boss record.
- [x] Step 2.3: During `_process(delta)`, decrement active phase `remaining_ticks` locally and
  refresh the boss bar.
- [x] Step 2.4: Add/remove a `BossTelegraphMarker` child on telegraph start/end using the
  authoritative radius/color.
- [x] Step 2.5: Clear phase state and marker on non-telegraph phase, phase end, boss removal,
  level clear, and death.
- [x] Step 2.6: Extend bot presentation debug state with marker presence, radius, and color.

```bash
make client-unit
```

## Task 3 - Client bot assertions

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/tests/test_client_bot.gd`

- [x] Step 3.1: Extend `wait_boss_health_bar` / `assert_boss_health_bar` expectations with
  `phase_kind`, `pattern_id`, `phase_index`, `duration_ticks`, `remaining_ticks_min/max`, and
  `phase_ratio_min/max`.
- [x] Step 3.2: Add assertions for boss presentation rows that can require
  `boss_telegraph_active`, `has_boss_telegraph_marker`, `telegraph_radius_min/max`, and
  `telegraph_tint`.
- [x] Step 3.3: Add runner unit coverage for phase health-bar and telegraph presentation matching.

```bash
make client-unit
```

## Task 4 - Focused client-bot scenario

Files:
- Add: `tools/bot/scenarios/client/28_boss_phase_readability.json`

- [x] Step 4.1: Add a scenario that descends to level `-5`, waits for the `cave_warden` boss bar,
  then waits for telegraph phase state on the bar.
- [x] Step 4.2: Assert the in-world boss presentation has an active telegraph marker during the
  telegraph phase.
- [x] Step 4.3: Run the focused client scenario.

```bash
make bot-client scenario=28_boss_phase_readability.json HEADLESS=1
```

## Task 5 - Regression and lifecycle

Files:
- Modify: `docs/plans/v57_2026-06-10-boss-phase-readability.md`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v57_boss-phase-readability.md`

- [x] Step 5.1: Run protocol boss-floor gate to prove server behavior remains unchanged.
- [x] Step 5.2: Update lifecycle docs and as-built summary.
- [x] Step 5.3: Run full CI.

```bash
make bot scenario=24_boss_floor_gate.json
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make bot-client scenario=28_boss_phase_readability.json HEADLESS=1`
- [x] `make bot scenario=24_boss_floor_gate.json`
- [x] `make ci`

## Deferred scope

- New boss pattern variety and additional telegraph shapes.
- Production VFX/audio, boss portraits, exact authoritative countdown sync, and multi-boss UI.
