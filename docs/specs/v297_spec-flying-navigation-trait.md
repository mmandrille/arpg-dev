# v297 Spec - Flying Navigation Trait

Status: Complete
Date: 2026-06-19
Codename: flying-navigation-trait

## Purpose

Add a data-owned monster navigation trait so flying monsters, starting with `dungeon_bat`, can
ignore floor-feature obstacle kinds such as water and holes while grounded monsters continue to
route around them. The trait should affect server-authoritative monster pathfinding and final
movement collision only; walls, closed doors, players, and living entities remain blockers.

This slice builds on v295 water and v296 holes so later slices can add player class mobility
exceptions without overloading monster behavior.

## Non-goals

- No player flying, barbarian leap, bridge, swim, fall, damage, knockback, rescue, or terrain-cost
  mechanics.
- No projectile, fog/LOS, loot/corpse placement, player auto-navigation, companion navigation, or
  boss-floor generation behavior changes.
- No protocol schema change for entity views; the trait is rules-owned server behavior.
- No new imported assets, shaders, Godot plugins, animation changes, or production bat behavior
  polish.

## Acceptance Criteria

- `shared/rules/monsters.v0.json` marks `dungeon_bat` with a schema-backed flying navigation trait;
  omitted trait defaults to grounded.
- Server rule validation rejects unknown monster navigation traits.
- Flying monster pathfinding treats `kind: "water"` and `kind: "hole"` layout obstacles as
  traversable, but still treats normal walls, closed interactable barriers, players, and living
  entities as blocked.
- Grounded monsters still treat water and holes as blocked.
- Final monster movement collision uses the same trait behavior as pathfinding so flying monsters
  do not plan through water/hole and then get stopped by the resolver.
- Preset world wall entities can carry optional `kind` values for compact movement labs, and those
  kinds reach server snapshots and the existing Godot wall renderer.
- A focused Go test proves flying and grounded monsters produce different path behavior over the
  same water/hole strip, and that flying monsters still cannot pass through normal walls.
- A protocol bot scenario proves a `dungeon_bat` moves across a lab with water/hole obstacles and
  remains able to approach the player.
- Existing water/hole generation, monster movement, pathfinding, shared validation, client unit,
  and reachable obstacle proofs remain green.

## Scope and Likely Files

- Shared monster/world rules:
  - `shared/rules/monsters.v0.json`
  - `shared/rules/monsters.v0.schema.json`
  - `shared/rules/worlds.v0.json`
  - `shared/rules/worlds.v0.schema.json`
- Server rules/navigation:
  - `server/internal/game/rules.go`
  - `server/internal/game/sim.go`
  - `server/internal/game/monster_navigation_traits.go`
  - `server/internal/game/monster_navigation_traits_test.go`
  - `server/internal/game/monster_ranged_positioning.go`
  - `server/internal/game/monster_navigation_budget.go`
  - `server/internal/game/elite_minion_ai.go` if follow-slot checks need trait-aware blocking
- Client presentation:
  - `client/scripts/wall_renderer.gd`
  - `client/tests/test_factories.gd`
- Bot proof:
  - `tools/bot/scenarios/99_flying_navigation_trait.json`
- Docs:
  - `docs/plans/v297_2026-06-19-flying-navigation-trait.md`
  - `docs/as-built/v297_flying-navigation-trait.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

Asset/plugin decision: reject external assets, imported bat/flying VFX, shaders, and plugins. Borrow
the existing bat monster definition/model, the v295/v296 obstacle-kind rendering path, and the
code-native preset lab pattern.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'FlyingNavigationTrait|FlyingNavigationLab|Path'`
- `ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=flying_navigation_trait ./scripts/bot_local.sh`
- `make client-unit`
- `HEADLESS=1 ARPG_ADDR=:18081 BASE_URL=http://localhost:18081 SCENARIO=flying_navigation_trait ./scripts/bot_visual.sh`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=flying_navigation_trait
```

## Open Questions and Risks

- No required questions for this run. Defaults: only `dungeon_bat` gets the flying trait, flying
  ignores water and holes only, and all non-floor blockers stay authoritative.
- Risk: path cache validity uses obstacle checks. The cache invalidation check must use the same
  trait-aware blocking as the path planner.
- Risk: preset `kind` support touches shared world rules plus server/client render adapters. Keep
  it additive and optional so existing labs remain unchanged.
