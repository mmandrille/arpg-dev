# v357 Spec: Town Night Perimeter

Status: Complete  
Date: 2026-06-27  
Codename: `town-night-perimeter`  
Baseline: v356 `player-navigation-guardrails`

## Purpose

Turn level-0 town in the production `dungeon_levels` world into a **night-time enclosed hub** that
matches the dungeon's darkness mood while keeping town non-combat and service-friendly.

Player-visible changes:

1. **Night / dungeon darkness** — town uses the same fog overlay, hero-light visibility, and
   suppressed ambient lighting as dungeon floors (not the current bright `TOWN_PROFILE` daylight).
2. **Wooden perimeter** — an authoritative ~15 m radius circular palisade blocks leaving town, with a
   **closed south gate** the player can click; the gate does not open and shows
   **"You can't leave for now."**
3. **Hub layout** — **stairs down** and **teleporter** sit at the town center; existing town services
   move to a wider outer ring with clearer spacing from the center travel anchors.

This slice reinforces ADR-0008's besieged-town premise: the hub is safe, but the outside world is
not yet accessible.

## Non-goals

- Opening the gate, overworld zones, or any travel beyond the town perimeter.
- Town combat, monster spawns, PvP, or safe-zone rule changes.
- Protocol schema version bump; reuse existing `action_intent` ack/reject and entity snapshots.
- Town torches, lanterns, fence torches, ambient audio, or production imported art.
- Changing lab-only worlds (`vendor_lab`, `quest_turn_in_lab`, etc.) unless a test hard-depends on
  production-town coordinates and breaks without a minimal fix.
- NavMesh, client-authoritative collision, or decorative cabin/campfire collision changes (v96 props
  stay presentation-only).
- Full circular-geometry PCG framework beyond this town's fixed perimeter recipe.

## Acceptance criteria

### Shared contracts — town layout and perimeter

- [ ] `dungeon_levels` level-0 entities place **stairs down** and **teleporter** at the town center
  anchor (default `(11, 12)`; adjacent tiles acceptable if interaction overlap must be avoided).
- [ ] All existing town service interactable IDs remain (`town_vendor`, `town_stash`,
  `town_bishop`, `town_market_board`, `town_blacksmith`, `town_mystery_seller`,
  `town_mercenary_board`, `town_quest_giver`, `town_unique_chest` when debug-enabled); only
  positions change.
- [ ] Each town service is on an outer ring with **at least 6 world units** from the center anchor and
  **at least 3 world units** from stairs/teleporter interaction cells.
- [ ] Player spawn (`dungeon_levels.player.position`) lands inside the enclosure, south of center,
  with a clear path to center and services.
- [ ] Perimeter geometry is **data-driven and deterministic** (committed JSON segments or a fixed
  rule-derived recipe from `center` + `radius_m` + `segment_count`); no runtime RNG in `game/`.
- [ ] Perimeter uses authoritative `wall` entities (~**15 m** radius, default **15.0** world units
  from center); walls block movement and participate in navigation bounds expansion
  (`townNavigationForWorld`).
- [ ] South opening contains one closed gate interactable (`town_exit_gate`) with a movement barrier
  when closed (same barrier pattern as `wooden_door`).

### Shared contracts — locked gate

- [ ] `shared/rules/interactables.v0.json` adds `town_exit_gate`:
  - `initial_state: "closed"`
  - `barrier_when_closed` sized for a doorway gap
  - `locked_exit_reason: "town_exit_locked"` (or equivalent data field consumed by server validation)
- [ ] Rules/schema validation accepts the new interactable and any new wall `kind` if introduced.

### Server — authority and interaction

- [ ] Players cannot walk through perimeter walls or the closed gate barrier.
- [ ] Clicking / `action_intent` on `town_exit_gate` while closed **does not** change interactable
  state to `open`; server rejects with reason **`town_exit_locked`** (existing reject envelope).
- [ ] Gate reject is deterministic and does not consume loot, shop, or travel flows.
- [ ] Stairs, teleporter, and all town services remain reachable via authoritative pathfinding from
  spawn after layout change.
- [ ] Focused Go tests prove: barrier blocks movement, gate reject reason, services still
  activatable.

### Client — night mood and feedback

- [ ] At `current_level == 0` with town perimeter walls present, fog overlay is active with the same
  presentation stack as dungeon floors (hero light, darkness feather, ambient suppression).
- [ ] Town lighting no longer uses bright `TOWN_PROFILE` when town fog is active; ambient/directional
  energies follow the same fog-suppressed dungeon profile used below level 0.
- [ ] Perimeter walls render as **wooden palisade** (~`TOWN_WALL_HEIGHT` / ~1 m), visually distinct
  from dungeon stone walls.
- [ ] Gate uses in-repo door presentation (`TownNodeFactory.make_door_node` or equivalent wood
  material); no external assets.
- [ ] On `intent_rejected` with reason `town_exit_locked`, the local player sees floating feedback
  **"You can't leave for now."** (same family as bag-full combat text / gesture hints).
- [ ] Existing town preview props (cabins, campfire, ambient silhouettes) still render; reposition
  client-only props if needed so they do not sit inside the gate gap or block service clicks.

### Regression / proof

- [ ] `make validate-shared` green.
- [ ] Focused Go + GDScript unit tests for gate reject, wall collision, fog-at-town lighting guard.
- [ ] Bot scenarios and tests that pin town coordinates are updated or converted to
  `interactable_def_id` targeting where possible.
- [ ] New **extended** client bot scenario `90_town_night_perimeter` proves: fog active in town,
  movement blocked at perimeter, gate click surfaces locked feedback, at least one service still
  reachable.
- [ ] `make ci` green.

## Scope and likely files

| Area | Files |
|------|-------|
| Shared rules | `shared/rules/worlds.v0.json`, `worlds.v0.schema.json` (if wall `kind` or perimeter metadata) |
| Shared interactables | `shared/rules/interactables.v0.json`, `interactables.v0.schema.json` |
| Shared presentation (optional) | `shared/assets/town_presentation.v0.json` + schema — `center`, `radius_m`, `segment_count`, wood tint |
| Server game | `server/internal/game/interactables.go`, `sim.go` (`populatePresetLevel`, `townNavigationForWorld`), focused `*_test.go` |
| Client lighting/fog | `client/scripts/dungeon_depth_lighting.gd`, `client/scripts/main.gd` (`_sync_fog_and_dungeon_lighting`, `_lab_world_fog_at_town_level` path) |
| Client walls/presentation | `client/scripts/wall_renderer.gd`, `client/scripts/ground_wall_factory.gd`, `client/scripts/town_node_factory.gd` |
| Client feedback | `client/scripts/main.gd`, `client/scripts/client_constants.gd` (message constant) |
| Client tests | `client/tests/test_factories.gd`, new or extended fog/lighting test |
| Bot | `tools/bot/scenarios/client/90_town_night_perimeter.json` (`ci_tier: extended`) |
| Showme (optional) | `skills/showme/scripts/visual_capture.gd` — refresh town focus framing for enclosure |
| Docs | plan + as-built on `/finish`; lifecycle row |

## Client asset / plugin decision

| Option | Decision |
|--------|----------|
| `TownNodeFactory.make_door_node()` wood mesh for south gate | **Borrow** |
| Procedural wood fence texture via `GroundWallFactory` / `WallRenderer` palette | **Adopt** (extend existing wall material path; new `wood` wall `kind` or `source: town_perimeter` tag) |
| Reuse fog compositor + `fog_presentation.v0.json` without gameplay changes | **Adopt** |
| External fence/gate GLBs, sky plugins, or night-post plugins | **Reject** |
| Town torches for fence readability | **Reject** (deferred; hero light only this slice) |

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'TownExit|TownPerimeter|townNavigation' -count=1
make client-unit
HEADLESS=1 make bot-client SCENARIO=90_town_night_perimeter
python3 skills/showme/scripts/render_focus.py --focus town   # visual sanity
make maintainability
make ci
```

Manual check:

```bash
make play
# Town should be dark like dungeon; perimeter visible; south gate click shows lock message; services reachable.
```

**CI pack:** scenario `90_town_night_perimeter` stays **extended** unless merge CI lacks town layout
coverage after coordinate churn; if promoted, demote a redundant extended scenario per pack policy.

## ADR alignment

- **ADR-0008:** Town remains the safe hub; perimeter enforces "cannot leave yet" without overworld
  content.
- **ADR-0007:** Gate message and wood presentation are client-only; server only rejects with reason.
- **ADR-0001 D2:** Walls, barriers, and interaction outcomes stay server-authoritative.
- **ADR-0006:** Wood look stays glTF-free / procedural per manifest policy.

## Open questions and risks

| Item | Default for planning |
|------|----------------------|
| Town center anchor | `(11, 12)` world units |
| Perimeter radius | `15.0` m (world units) |
| Fence height | `TOWN_WALL_HEIGHT` (~1 m palisade) |
| Fog LOS on town walls/gate | **Yes** — full fog stack like dungeon |
| Gate def vs reusing `wooden_door` | **New `town_exit_gate`** to avoid accidental open-toggle path |
| Reject reason string | `town_exit_locked` |
| Perimeter authoring | Prefer one committed recipe in shared JSON; optional `town_presentation.v0.json` for tuning |
| `_lab_world_fog_at_town_level()` naming | Behavior already enables fog when town has walls; extend lighting path, rename only if plan finds it misleading |
| Layout churn breaks bots/tests | Audit scenarios using hardcoded town coordinates; prefer `interactable_def_id` steps |
| `main.gd` / `ground_wall_factory.gd` size | Grandfathered; extract only if maintainability gate forces it |
| Campfire/cabin props vs new center | Reposition client-only props so center reads as travel hub, not blocked by decor |

No blocking questions remain for `/plan` under the defaults above.
