# v62 Spec - Monster Depth Stat Scaling

Status: Complete
Date: 2026-06-11
Codename: `monster-depth-stat-scaling`

## Purpose

Generated dungeon monsters should grow through the broader combat system, not only through HP,
damage, XP, and loot depth. Dungeon depth should raise baseline pressure, while rarity should add
noticeable stat identity for champion, rare, and unique monsters.

## Non-goals

- No new monster art, animations, UI panels, or client-only presentation work.
- No protocol schema version bump; combat outcomes remain server-authoritative through existing
  state deltas and events.
- No full encounter rebalance, item-level system, affix system, or new monster archetype roster.
- No changes to static lab monsters unless tests deliberately construct generated dungeon cases.

## Acceptance Criteria

1. Generated dungeon monsters scale HP and attack damage by both dungeon depth and rarity.
2. Generated dungeon monsters also scale defensive/offensive derived stats: `armor`,
   `hit_chance`, `crit_chance`, `block_percent`, and attack cadence.
3. Scaling is data-driven from `shared/rules/dungeon_generation.v0.json` and validated by the
   shared schema.
4. Volatile stats have safe caps/floors so deeper or rarer monsters do not become unfair:
   hit chance and crit chance cap below certainty, block uses existing percent semantics, and
   attack cooldown cannot drop below a configured minimum.
5. Existing boss template scaling and lab/static monster behavior remain unchanged unless they
   already use generated monster placement.
6. Golden fixtures prove common/champion/rare/unique effective monster stats for pinned cases.
7. Go tests cover the stat-scaling helper and a generated dungeon spawn path.
8. `make ci` passes.

## Scope and Files Likely Touched

```text
shared/rules/dungeon_generation.v0.json        - add depth and rarity stat-scaling data
shared/rules/dungeon_generation.v0.schema.json - validate new data shape
shared/golden/monster_rarity.json              - expected effective stat outputs
shared/golden/monster_rarity.v0.schema.json    - fixture schema for new fields
server/internal/game/rules.go                  - structs and validation for new data
server/internal/game/sim.go                    - generated-monster effective stat application
server/internal/game/game_test.go              - unit/golden coverage
docs/plans/v62_2026-06-11-monster-depth-stat-scaling.md - implementation plan
docs/as-built/v62_monster-depth-stat-scaling.md - close-out summary
PROGRESS.md                                    - lifecycle update when complete
```

## Design

Monster definitions remain the base archetype. Generated dungeon placement applies two data-driven
layers:

```text
effective stat = base monster stat
  * depth multiplier for absolute dungeon depth
  * rarity multiplier
  + rarity flat bonus where useful
```

Attack cadence uses cooldown ticks, so lower is stronger:

```text
effective cooldown = max(min_cooldown_ticks, round(base_cooldown * depth_multiplier * rarity_multiplier))
```

Depth scaling should be conservative. Rarity should be more visible but still capped.

## Test and Bot Proof

- `make validate-shared` validates new rules and golden fixture shape.
- `cd server && go test ./internal/game/... -run 'TestMonsterRarity|TestGeneratedDungeonMonster'`
  proves helper behavior and generated-spawn integration.
- Existing protocol bot coverage remains sufficient because this slice does not add a new intent,
  world preset, or wire contract; combat scenarios should continue passing under `make bot`.
- `make ci` is the final gate.

## Open Questions and Risks

| # | Question | Status |
|---|----------|--------|
| 1 | Should this include new client stat readouts? | Deferred; no UI in this slice. |
| 2 | Should boss templates also get depth stat scaling? | Deferred; boss templates already own bespoke scaling. |
| 3 | Are exact tuning numbers final? | No; this slice installs safe, conservative data hooks and goldens. |
