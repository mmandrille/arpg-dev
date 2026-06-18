# v256 Plan - Discovery Minimap

Status: Complete
Goal: Replace the objective-only minimap with a larger transparent discovery minimap toggled by
`TAB`.
Architecture: Keep discovery map state entirely client-presentational and session-local. A focused
state helper owns explored-cell memory by level and projection math; a focused Control owns drawing
and debug state. `main.gd` only creates the widget, toggles it, synchronizes existing player/fog/wall
state, and exposes bot debug data. No protocol, shared rules, server, or replay contract changes are
required.
Tech stack: Godot GDScript client, Godot client bot scenario, docs.

## Baseline and Shortcut Decision

Builds on v253-v255 fog-of-war presentation work and v176's compact elite-objective minimap. The
server already sends the player's authoritative position, progression-derived light radius, current
level, and normalized wall layout through existing client state. This slice borrows that data to
draw a player-facing discovery map without changing authority.

Asset/plugin decision: reject external assets, imported minimap art, shader plugins, and Godot
addons. Borrow existing in-repo `EliteObjectiveMinimap` HUD placement/drawing conventions,
`FogOfWarOverlay` light-radius debug inputs, `WallRenderer` normalized wall-layout data, and client
bot assertion patterns.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/discovery_minimap_state.gd` | Track session-local explored cells per level and derive render/debug data |
| Add | `client/scripts/discovery_minimap.gd` | Draw the larger transparent minimap, player marker, known walls, optional objective pin, and debug state |
| Modify | `client/scripts/main.gd` | Replace old objective minimap wiring with discovery minimap creation, `TAB` toggle, sync, and bot debug |
| Add | `client/tests/test_discovery_minimap.gd` | Unit proof for toggle/defaults, exploration, level scope, wall reveal, size/opacity, and objective pin |
| Modify | `scripts/client_smoke.sh` | Run the new discovery minimap unit gate instead of the old objective minimap gate |
| Modify | `client/scripts/bot_step_catalog.gd` | Register `assert_discovery_minimap` |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Assert minimap debug visibility, toggle, explored count, size, opacity, and pin state |
| Add | `tools/bot/scenarios/client/69_discovery_minimap_toggle.json` | Client visual proof for `TAB` toggle and explored minimap state |
| Modify | `docs/specs/v256_spec-discovery-minimap.md` | Mark complete during close-out |
| Modify | `docs/progress/slice-lifecycle.md` | Add v256 lifecycle row |
| Add | `docs/as-built/v256_discovery-minimap.md` | Record shipped behavior and proof |
| Modify | `PROGRESS.md` | Update current status and remaining minimap/durable-map deferrals |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_controller.gd`
- [x] `client/scripts/bot_scenario_runner.gd`: not planned
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none planned
- [x] Did every touched grandfathered file stay at or below its baseline allowance?

Decision:
- [x] Keep discovery state and drawing in new focused scripts.
- [x] Keep `client/scripts/main.gd` changes narrow and within the grandfathered baseline allowance.

Verification:
```bash
make maintainability
```

## Task 1 - Discovery Minimap State and UI

Files:
- Add: `client/scripts/discovery_minimap_state.gd`
- Add: `client/scripts/discovery_minimap.gd`
- Add: `client/tests/test_discovery_minimap.gd`

- [x] Step 1.1: Implement a pure state helper that records explored grid cells by active level from
  player position and light radius.
- [x] Step 1.2: Derive discovered wall rectangles and an optional elite-objective pin from current
  wall layout/entity data without owning gameplay decisions.
- [x] Step 1.3: Draw a roughly 208px minimap with a slightly transparent panel/background, explored
  floor cells, known walls, centered player marker, and optional objective pin.
- [x] Step 1.4: Expose debug state for visibility, toggle state, map size, opacity, explored cell
  count, wall count, player marker, and objective pin state.
- [x] Step 1.5: Cover default hidden state, toggle behavior, exploration accumulation, level scoping,
  wall reveal, objective pin, size, and opacity in unit tests.

```bash
make client-unit
```

## Task 2 - Client Wiring and TAB Toggle

Files:
- Modify: `client/scripts/main.gd`
- Modify: `scripts/client_smoke.sh`

- [x] Step 2.1: Replace gameplay HUD creation of `EliteObjectiveMinimap` with the discovery minimap.
- [x] Step 2.2: Add a `TAB` key handler that toggles the minimap during gameplay and marks the input
  handled.
- [x] Step 2.3: Synchronize minimap state after snapshots, deltas, wall-layout changes, level
  changes, and progression updates using existing player position, light radius, wall layout, and
  entity data.
- [x] Step 2.4: Expose `discovery_minimap` debug state from `get_bot_state()` while preserving
  focused compatibility fields only if needed for existing bot helpers.
- [x] Step 2.5: Replace the old smoke unit gate with the new discovery minimap test.

```bash
make client-unit
```

## Task 3 - Client Bot Proof

Files:
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Add: `tools/bot/scenarios/client/69_discovery_minimap_toggle.json`

- [x] Step 3.1: Register and implement `assert_discovery_minimap`.
- [x] Step 3.2: Add a focused client scenario that starts with the minimap hidden, presses `TAB`,
  and asserts the map is visible with explored cells and the expected doubled size/opacity range.
- [x] Step 3.3: Prove the scenario with headless bot-visual.

```bash
HEADLESS=1 make bot-visual scenario=69_discovery_minimap_toggle
```

## Task 4 - Lifecycle Docs

Files:
- Modify: `docs/specs/v256_spec-discovery-minimap.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v256_discovery-minimap.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark the v256 spec complete.
- [x] Step 4.2: Add v256 lifecycle and as-built notes.
- [x] Step 4.3: Update `PROGRESS.md` current status and defer durable explored-map memory, full-screen
  map overlay, route/compass/pathing, and controls remapping.

```bash
make maintainability
```

## Final Verification

- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-visual scenario=69_discovery_minimap_toggle`
- [x] `HEADLESS=1 make bot-visual scenario=45_elite_objective_minimap_pin`
- [x] `make maintainability`
- [ ] Autoloop final batch gate: `make ci`

Manual visual proof, if desired:

```bash
make bot-visual scenario=69_discovery_minimap_toggle
```
