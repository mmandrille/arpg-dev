# Spec: `monster-rarity-and-loot-scaling`

Status: Implemented
Branch: `main`
Slice: v30 - generated monster rarity, pastel visual tiers, and loot-depth offsets
Baseline: v29 `dungeon-equipment-drop-expansion`
Related:

- [`v21_spec-dungeon-monster-combat.md`](v21_spec-dungeon-monster-combat.md) - generated hostile dungeon mobs
- [`v23_spec-item-templates-and-rolled-drops.md`](v23_spec-item-templates-and-rolled-drops.md) - rolled item payloads and rarity metadata for items
- [`v25_spec-treasure-classes-and-guarded-chests.md`](v25_spec-treasure-classes-and-guarded-chests.md) - treasure class rolls and guarded chest source behavior
- [`v26_spec-character-stats-and-leveling.md`](v26_spec-character-stats-and-leveling.md) - XP and character level progression
- [`v29_spec-dungeon-equipment-drop-expansion.md`](v29_spec-dungeon-equipment-drop-expansion.md) - depth-banded dungeon monster/chest drops
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - seeded procedural dungeon generation and better loot by depth
- [`../../PROGRESS.md`](../../PROGRESS.md)

## 1. Purpose

v29 made generated dungeon monsters and guarded chests use the expanded equipment catalog through
temporary depth bands. Dungeon monster population is still visually and mechanically flat: every
generated monster is the same `dungeon_mob`, has the same challenge, and uses the loot table for the
floor depth where it spawned.

This slice adds **server-authoritative monster rarity** for generated dungeon mobs. During
procedural level generation, every generated dungeon monster rolls one rarity tier from shared data:

| Rarity | Visual color | Challenge | Loot depth offset |
|--------|--------------|-----------|-------------------|
| `common` | pastel white | baseline | `+0` |
| `champion` | pastel blue | harder | `+1` |
| `rare` | pastel red | harder | `+2` |
| `unique` | pastel golden | hardest | `+3` |

Rarity affects:

- generated monster HP,
- generated monster attack damage,
- XP reward,
- loot depth used when resolving that monster's drop source,
- client presentation tint.

Example: a `unique` monster generated on dungeon level `-5` uses effective loot depth `8`
(`abs(-5) + 3`) while remaining physically on level `-5`.

The player character is tinted green for now. Enemies reuse the same current character/monster model
with different material colors. This is a visual placeholder and not a production art direction.

The proof is: seeded dungeon generation rolls rarity -> server applies rarity challenge and loot
offset -> protocol exposes rarity for snapshot/delta/replay -> Godot tints player/enemies -> bot
proves kill, loot, state, reconnect, replay, and persistence.

## 2. Non-goals

- No unique item catalog, set items, or special unique-only drop pool. `unique` monster rarity has
  no relationship to unique item drops.
- No affixes, procedural item names, named elite monsters, rare packs, minions, aura modifiers, or
  boss packs.
- No Magic Find, player-driven rarity modifiers, or final dungeon economy model.
- No boss-floor chest integration or guaranteed unique spawn per floor.
- No production monster art, new monster model family, texture pipeline, VFX, audio, or animation
  polish.
- No rarity for static/lab monsters in v30. Existing world-preset monsters keep their current
  behavior unless a future slice opts them into rarity.
- No client-side gameplay logic. The client renders the rarity and sends the same intents as today.
- No Protobuf migration.

## 3. Files to create or modify

```text
docs/specs/v30_spec-monster-rarity-and-loot-scaling.md       - this slice contract
docs/plans/v30_<YYYY-MM-DD>-monster-rarity-and-loot-scaling.md - implementation plan
shared/rules/dungeon_generation.v0.schema.json               - generated monster rarity rules
shared/rules/dungeon_generation.v0.json                      - rarity weights, colors, challenge, loot offsets
shared/rules/loot_tables.v0.json                             - if effective-depth loot tables need explicit routes
shared/rules/treasure_classes.v0.json                        - only if deeper effective-depth coverage needs test hooks
shared/protocol/session_snapshot.v1.schema.json              - monster entity rarity metadata if schema is not already sufficient
shared/protocol/state_delta.v1.schema.json                   - monster entity rarity metadata if schema is not already sufficient
shared/protocol/examples/session_snapshot.json               - example monster rarity
shared/protocol/examples/state_delta.json                    - example monster rarity
shared/golden/monster_rarity.json                            - pinned rarity, scaling, and loot-depth fixture
shared/golden/monster_rarity.v0.schema.json                  - fixture schema
tools/validate_shared.py                                     - rarity weights/colors/modifiers/loot offsets validation
server/internal/game/rules.go                                - parse and validate monster rarity rules
server/internal/game/dungeon_gen.go                          - roll rarity during generated monster placement
server/internal/game/sim.go                                  - apply rarity HP, damage, XP, and effective loot depth
server/internal/game/types.go                                - expose monster rarity in entity view if needed
server/internal/game/game_test.go                            - rarity generation, scaling, loot offset, determinism tests
server/internal/replay/replay.go                             - replay parity if entity snapshot shape changes
server/internal/http                                         - `/state` entity rarity parity if needed
client/scripts/main.gd                                       - apply player green tint and enemy rarity tint
client/tests/test_golden.gd                                  - optional shared golden/color checks
client/tests                                                 - presentation unit/smoke coverage if tint helpers are extracted
tools/bot/run.py                                             - assertions for monster rarity and effective loot depth if needed
tools/bot/scenarios/21_monster_rarity_loot_scaling.json      - end-to-end generated dungeon proof
PROGRESS.md                                             - lifecycle update when v30 ships
```

Protocol schema change note: v1 schemas already include an optional `rarity` field on entity views.
The plan must confirm whether existing protocol contracts are sufficient. If the current field is
only used for item rarity in practice, v30 should explicitly document and test that monster entities
also use it. No new protocol version is expected unless implementation finds the existing field
ambiguous or insufficient.

## 4. Rarity data model

Add a shared-data rarity catalog for generated dungeon monsters. The exact home may be
`dungeon_generation.v0.json` or a dedicated rules file if the plan finds a cleaner boundary, but it
must be consumed by the Go sim and validated by shared tooling.

Suggested shape:

```json
{
  "monster_rarities": [
    {
      "id": "common",
      "weight": 100,
      "color": "#f2f2ec",
      "hp_multiplier": 1.0,
      "damage_multiplier": 1.0,
      "xp_multiplier": 1.0,
      "loot_depth_offset": 0
    },
    {
      "id": "champion",
      "weight": 15,
      "color": "#9fc7ff",
      "hp_multiplier": 1.5,
      "damage_multiplier": 1.25,
      "xp_multiplier": 1.5,
      "loot_depth_offset": 1
    },
    {
      "id": "rare",
      "weight": 6,
      "color": "#ff9b9b",
      "hp_multiplier": 2.0,
      "damage_multiplier": 1.5,
      "xp_multiplier": 2.0,
      "loot_depth_offset": 2
    },
    {
      "id": "unique",
      "weight": 3,
      "color": "#ffd978",
      "hp_multiplier": 3.0,
      "damage_multiplier": 2.0,
      "xp_multiplier": 3.0,
      "loot_depth_offset": 3
    }
  ]
}
```

The weights are intentionally first-pass tuning:

```text
common: 100
champion: 15
rare: 6
unique: 3
```

This makes `unique` rare enough to stand out while still findable by pinned seeds and tests. Future
economy tuning can revise weights without changing the architecture.

Validation must enforce:

- exactly one `common`, `champion`, `rare`, and `unique` tier,
- all weights are positive integers,
- all colors are valid hex colors,
- colors are in the intended pastel family for the four tiers as a documented data decision,
- HP, damage, and XP multipliers are positive,
- loot depth offsets are non-negative integers,
- `common` has offset `0`,
- configured rarity IDs are stable and lowercase.

## 5. Procedural generation and determinism

Generated dungeon monster rarity is rolled during procedural level generation, after the monster
spawn position is selected and before the generated monster is appended to the level output. The
exact placement in the RNG stream must be pinned by tests and golden fixtures.

Deterministic order:

```text
session seed + level
  -> generated floor layout
  -> chest presence and generated source selection
  -> generated monster count and positions
  -> monster rarity roll per generated monster in stable append order
  -> monster entity spawn with rarity, scaled HP, scaled damage, scaled XP, and effective loot table
```

The implementation must not use wall-clock time, unseeded randomness, map iteration order, or any
client state in rarity selection. Same seed plus same ordered inputs must reproduce:

- monster rarity tiers,
- monster HP/max HP,
- monster damage,
- XP reward,
- selected effective loot depth and loot table,
- loot roll payloads,
- entity IDs and spawn order,
- `/state`, reconnect, and replay output.

Static world-preset monsters are not part of this v30 rarity roll. If an entity comes from
`worlds.v0.json`, it remains unchanged unless a future spec adds explicit preset rarity metadata.

## 6. Challenge scaling

Rarity modifies generated monster challenge from the `dungeon_mob` base definition:

```text
scaled_hp = round_base(max_hp * hp_multiplier)
scaled_attack_min = round_base(attack_damage.min * damage_multiplier)
scaled_attack_max = round_base(attack_damage.max * damage_multiplier)
scaled_xp = round_base(xp_reward * xp_multiplier)
```

The plan must choose and document the exact rounding function. Default: deterministic nearest
integer with a minimum of `1` for HP, attack damage, and positive XP values.

Scaling applies to generated monster entity state at spawn time. The base monster definition remains
unchanged so future content can still reason about the canonical `dungeon_mob`.

Generated monster snapshots/deltas must expose the scaled `hp` / `max_hp` as normal. The client does
not recompute scaling.

## 7. Loot-depth scaling

Monster rarity changes the effective dungeon loot depth for that monster's drops:

```text
physical_depth = abs(level_number)
effective_loot_depth = physical_depth + rarity.loot_depth_offset
```

The effective depth is used to resolve the monster drop source. For example:

| Level | Rarity | Offset | Effective loot depth |
|-------|--------|--------|----------------------|
| `-1` | `common` | `0` | `1` |
| `-1` | `champion` | `1` | `2` |
| `-2` | `rare` | `2` | `4` |
| `-5` | `unique` | `3` | `8` |

Because v29 only has coarse bands `1`, `2`, and `3+`, effective depths `3` and above will currently
route to the `3+` band. This is acceptable for v30 and should be documented as a first-pass hook for
future item-level/depth progression.

Important boundary:

- Monster rarity may improve a monster's effective loot depth.
- Monster rarity does not create unique item drops.
- Chest loot remains governed by chest source rules from v25/v29. v30 does not add chest rarity.

## 8. Protocol and presentation

Monster entity views must include rarity metadata in snapshots and spawn/update deltas wherever a
client or bot needs to observe it. If the existing optional `rarity` field in protocol v1 is used,
monster entities should emit:

```json
{
  "id": "123",
  "type": "monster",
  "monster_def_id": "dungeon_mob",
  "rarity": "rare",
  "position": { "x": 10, "y": 8 },
  "hp": 8,
  "max_hp": 8
}
```

The Godot client uses server-provided rarity for presentation only:

| Entity | Tint |
|--------|------|
| player | green |
| common monster | pastel white |
| champion monster | pastel blue |
| rare monster | pastel red |
| unique monster | pastel golden |

The client must not derive combat power, XP, or loot from tint. If rarity is missing on a legacy or
static monster, default presentation is `common`.

hard UI/camera/art problem.

## 9. Acceptance criteria

1. `make validate-shared` validates monster rarity rules, weights, colors, multipliers, and loot
   depth offsets.
2. Generated dungeon monsters roll one of `common`, `champion`, `rare`, or `unique` from shared
   data using seeded deterministic generation.
3. Rarity applies only to generated dungeon monsters in v30; static/lab monsters remain unchanged.
4. Rarity scales both HP and attack damage in server-owned entity state.
5. Rarity scales XP reward on kill.
6. Rarity changes monster loot source resolution by effective loot depth:
   `abs(level) + loot_depth_offset`.
7. A unique monster on level `-5` resolves monster loot against effective depth `8`.
8. Monster rarity appears in `/state`, reconnect snapshots, replay timelines, and any spawn delta
   needed by the Godot client.
9. The player is tinted green in the Godot client.
10. Generated dungeon monsters are tinted by rarity: pastel white, blue, red, and golden.
11. Client tinting remains presentation-only and never influences combat, loot, or XP.
12. Shared/golden tests pin rarity roll outcomes, challenge scaling, XP scaling, and loot-depth
    offsets.
13. Go tests prove deterministic rarity generation and replay-stable monster rarity/loot outcomes.
14. Protocol bot scenario `21_monster_rarity_loot_scaling.json` proves generated dungeon play with
    a non-common monster: descend, observe rarity, kill, pick up loot, `/state`, reconnect, replay,
    and fresh-session persistence.
15. The bot proof uses a pinned seed that includes a `unique` monster if practical; otherwise Go
    golden tests must explicitly cover the `unique` level `-5` -> effective depth `8` case.
16. `make ci` green.

## 10. Testing plan

1. `make validate-shared`
2. `cd server && go test ./internal/game/... -run 'Rarity|Dungeon|Loot|XP|Replay'`
3. `cd server && go test ./internal/http/... ./internal/replay/... -run 'Rarity|Replay|State'`
4. `make client-unit`
5. `make client-smoke`
6. `make bot` - includes `21_monster_rarity_loot_scaling.json`
7. `make bot-client` only if Godot client bot coverage is extended for tint/rarity visibility
8. `make ci`
9. Manual: `make play`, descend into generated dungeon floors, confirm common/champion/rare/unique
   monsters are visually distinct when encountered, and confirm the player is green.

## 11. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Rarity applies to generated dungeon monsters only. | Keeps v30 focused on PCG population and avoids surprising lab/preset fixture changes. |
| 2 | Rarity scales HP, attack damage, and XP. | Harder monsters should be both visibly and mechanically meaningful. |
| 3 | Initial weights are `common: 100`, `champion: 15`, `rare: 6`, `unique: 3`. | First-pass tuning makes higher tiers uncommon but testable with pinned seeds. |
| 4 | Loot depth offsets are `+0`, `+1`, `+2`, `+3`. | Simple, verifiable bridge from monster danger to better v29 dungeon loot bands. |
| 5 | Unique monster rarity has no relationship to unique item drops. | Avoids prematurely adding unique/set itemization while preserving a useful rarity tier name. |
| 6 | Visuals reuse the current model with pastel tints; player is green. | Gives immediate readability without production art or a new asset pipeline slice. |

## 12. As-built notes

- Rarity rules live in `shared/rules/dungeon_generation.v0.json` under `monster_rarities`.
- Generated monster rarity uses a separate deterministic RNG label derived from
  `seed + "|monster_rarity|" + abs(level)` so existing dungeon geometry, chest presence, and chest
  placement streams do not drift.
- Scaling uses deterministic nearest-integer rounding via `floor(value + 0.5)` with minimum `1`.
- Protocol v1 already had optional entity `rarity`; v30 uses it for monster entity views without a
  schema version bump.
- The bot scenario uses a pinned champion monster for a short stable end-to-end proof. The unique
  `level -5 -> effective depth 8 -> 3+ loot band` behavior is pinned by shared/golden and Go tests.

## 13. Open follow-ups

- Final item-level/depth economy beyond the v29 `1`, `2`, `3+` bands.
- Unique/set item catalogs and unique drop rules.
- Named rare/unique monsters, rare packs, minions, aura modifiers, and boss floors.
- Production enemy models, textures, VFX, and colorblind/accessibility-safe rarity presentation.
- Magic Find and player-driven loot rarity modifiers.
