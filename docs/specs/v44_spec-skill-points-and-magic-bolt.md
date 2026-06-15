# Spec: `skill-points-and-magic-bolt`

Status: Draft
Date: 2026-06-09
Branch: `main`
Slice: v44 - skill points, attack-speed foundation, and first active skill
Baseline: v43 `equipment-requirements-and-preview`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - character-scoped progression
- [`v26_spec-character-stats-and-leveling.md`](v26_spec-character-stats-and-leveling.md) - durable level, stat points, stats, derived stat formulas
- [`v28_spec-full-equipment-and-belt-hotbar.md`](v28_spec-full-equipment-and-belt-hotbar.md) - hotbar, paper-doll equipment, handedness
- [`v31_spec-combat-stat-effects-and-feedback.md`](v31_spec-combat-stat-effects-and-feedback.md) - effective stat aggregation and combat event metadata
- [`v39_spec-ui-currency-and-mana-polish.md`](v39_spec-ui-currency-and-mana-polish.md) - player mana and mana potion baseline
- [`v43_spec-equipment-requirements-and-preview.md`](v43_spec-equipment-requirements-and-preview.md) - requirements, stat display, and equip previews

## 1. Purpose

The game now has persistent character stats, mana, full equipment, rolled item stats, combat stat
breakdowns, and requirement-gated equipment. It still has no active skills, no spendable skill
points, and `attack_speed` is displayed but not yet a real timing foundation.

This slice adds the first server-authoritative active skill loop and the minimum progression model
needed to support it:

- Characters gain **3 stat points per level** instead of 5.
- Characters gain **1 skill point every 3 levels**.
- The only spendable skill in v44 is `magic_bolt`.
- Spending a skill point on `magic_bolt` increases its rank and improves the skill's stats.
- `magic_bolt` is available to all characters for now; class selection is deferred.
- `magic_bolt` consumes mana, deals server-authoritative damage, and has a cooldown.
- `magic_bolt` cooldown is **2x the character's current attack interval**.
- The Godot skill slot disables immediately after cast and gradually re-enables as cooldown
  recovers.
- Attack speed becomes a server-authored effective stat derived from DEX, equipped weapon speed, and
  item stats.
- Every weapon declares an attack-speed property. Two-handed swords are at least 30% slower than
  comparable one-handed swords, and long bows are slower than short bows.

The proof is: level progression grants stat and skill points -> skill rank persists -> attack speed
is derived authoritatively from stats and gear -> `magic_bolt` uses that attack interval for
cooldown -> mana/cooldown/damage are replay-safe -> Godot renders the single-skill UI state from
server-owned data.

## 2. Non-Goals

- No class selection, class restrictions, class starting stats, or class-specific skill access.
- No skill tree layout. The skills interface shows only `magic_bolt`.
- No respec, skill-point refund, alternate skill branches, or passive skills.
- No multiple active skills, skill tabs, socketed skills, buffs, debuffs, damage-over-time, AoE,
  homing, summons, traps, auras, or status effects.
- No mana regeneration, potion rebalance, or broader resource system beyond spending mana on
  `magic_bolt`.
- No global basic-attack cooldown rebalance unless the plan finds a small, deterministic path. v44
  must make attack speed authoritative for effective stats and `magic_bolt` cooldown; applying it
  to normal attacks can be deferred.
- No animation-speed scaling or production VFX/audio. Placeholder projectile/impact presentation is
  acceptable.
- No final combat balance pass. Numbers are first-pass deterministic tuning only.
  adoption checklist result; expected default is to extend in-repo UI and use shared rules for
  gameplay.

## 3. Acceptance Criteria

1. Shared progression rules grant 3 stat points per level, and all affected golden fixtures/tests
   are updated from the old 5-point contract.
2. Shared progression rules grant 1 skill point every 3 levels, starting when the character reaches
   level 3.
3. Fresh and persisted characters expose `unspent_skill_points` and skill ranks in snapshots,
   deltas, `/state`, reconnect, and replay.
4. `allocate_skill_point_intent` or equivalent server intent spends exactly one unspent skill point
   on `magic_bolt`, increments its rank, persists the rank, and rejects invalid spends without
   mutation.
5. `magic_bolt` rank improves skill stats through data-driven shared rules. v44 must at minimum
   increase `magic_bolt` damage per rank.
6. `cast_skill_intent` or equivalent server intent validates known skill id, known rank, mana cost,
   cooldown, live player state, usable target/direction, and range before mutation.
7. Successful `magic_bolt` casts subtract mana, start cooldown, resolve server-authoritative damage,
   emit combat/skill events, and can kill monsters through the normal death/loot/XP path.
8. Recasting `magic_bolt` before cooldown ends rejects with a clear reason and does not subtract
   mana, damage targets, or reset cooldown.
9. `magic_bolt` cooldown equals 2x the current server-authored attack interval at cast time.
10. Effective `attack_speed` uses DEX, equipped weapon speed, and item stats; the resulting attack
   interval is exposed in stat breakdowns or skill/cooldown views for bot and UI inspection.
11. Every weapon template/static weapon has an attack-speed property. Shared validation rejects
   missing weapon attack speed and invalid speed relationships.
12. Two-handed swords are at least 30% slower than comparable one-handed swords; long bows are
   slower than short bows.
13. Item template `base_stats` and `rollable_stats` support an attack-speed modifier that can
   increase or decrease effective attack speed within validated bounds.
14. Godot shows a skills interface with one skill choice, `magic_bolt`, current rank, available
   skill points, and a spend action.
15. Godot shows one usable `magic_bolt` skill slot near the hotbar. After cast, the slot is disabled
   and gradually re-enables as server cooldown time passes.
16. Protocol bot proof covers leveling to a skill-point threshold, spending a skill point, casting
   `magic_bolt`, mana spend, cooldown rejection, cooldown recovery, damage/kill, `/state`,
   reconnect, replay, and fresh-session persistence.
17. Client bot proof covers the real Godot skills interface, skill-point spend, skill-slot cooldown
   disabled/recharge presentation, and cast feedback.
18. `make validate-shared`, Go tests, client unit tests, protocol bot, client bot, and `make ci`
   pass.

## 4. Scope And Likely Files

```text
docs/specs/v44_spec-skill-points-and-magic-bolt.md - this spec
docs/plans/v44_2026-06-09-skill-points-and-magic-bolt.md - implementation plan
PROGRESS.md - lifecycle update when v44 ships

shared/rules/character_progression.v0.schema.json - stat points per level and skill point cadence
shared/rules/character_progression.v0.json - change stat points to 3 and add skill point cadence
shared/rules/skills.v0.schema.json - active skill catalog, rank scaling, mana cost, cooldown rule
shared/rules/skills.v0.json - `magic_bolt` definition and rank table
shared/rules/item_templates.v0.schema.json - weapon attack speed and attack-speed roll stat
shared/rules/item_templates.v0.json - weapon speed values and attack-speed roll support
shared/rules/items.v0.schema.json - static weapon attack speed if static items still need it
shared/rules/items.v0.json - static weapon speed values
shared/protocol/messages.v5.schema.json - skill-point spend and skill cast intents if v4 cannot be extended cleanly
shared/protocol/session_snapshot.v5.schema.json - skill ranks, skill points, cooldowns, effective attack interval if needed
shared/protocol/state_delta.v5.schema.json - skill events, cooldown updates, skill-point/rank updates if needed
shared/protocol/examples/session_snapshot.json - skill/rank/cooldown example
shared/protocol/examples/state_delta.json - skill cast/cooldown/reject examples
shared/golden/skill_points_and_magic_bolt.json - progression, attack speed, cooldown, rank, damage fixture
shared/golden/skill_points_and_magic_bolt.v0.schema.json - fixture schema
tools/validate_shared.py - skill rules, weapon speed relationships, golden drift validation

server/migrations/00XX_character_skills.sql - durable skill points/ranks and session-start snapshots
server/internal/store/models.go - skill progression persistence models
server/internal/store/interfaces.go - skill progression repo methods
server/internal/store/repos.go - Postgres skill progression implementation
server/internal/store/store_test.go - skill progression persistence tests
server/internal/game/rules.go - parse skill rules, attack speed, skill-point cadence
server/internal/game/types.go - skill state, cooldown, attack-speed view types
server/internal/game/sim.go - skill points, skill spend, magic bolt cast, cooldown, damage
server/internal/game/game_test.go - skill points, cooldown, attack speed, damage, replay-safe behavior
server/internal/realtime/runner.go - persist skill-point/rank mutations by character
server/internal/replay/replay.go - session-start skill progression snapshots
server/internal/http/session.go - load skill progression on fresh session create/attach

client/scripts/main.gd - parse skill snapshots/deltas, send skill intents, route skill events
client/scripts/skills_panel.gd - one-skill spend interface
client/scripts/skill_bar.gd - magic_bolt slot, disabled/recharge presentation
client/scripts/character_stats_panel.gd - display attack speed / attack interval if needed
client/scripts/stat_labels.gd - display labels for attack speed / attack interval / skill points
client/scripts/bot_scenario_runner.gd - client bot steps/assertions for skills and cooldown
client/tests/test_golden.gd - skill/attack-speed golden fixture checks
client/tests/test_skills_panel.gd - skill spend UI model tests if helpers are extracted
client/tests/test_skill_bar.gd - cooldown presentation helper tests if helpers are extracted

tools/bot/run.py - skill-point, skill-cast, cooldown, and attack-speed assertions
tools/bot/scenarios/32_skill_points_and_magic_bolt.json - protocol end-to-end proof
tools/bot/scenarios/client/19_skill_points_and_magic_bolt.json - Godot client UI proof
```

Protocol note: v44 probably needs a coordinated schema bump for new intents, skill state, and
cooldown payloads. Use `v5` only if it is still the next protocol version when planning starts; if
v43 landed these fields as additive v4 changes and no v5 files exist, the plan should decide the
cleanest coordinated version.

## 5. Data Shape Draft

### 5.1 Progression rules

`shared/rules/character_progression.v0.json` updates the existing point grant:

```json
{
  "points_per_level": 3,
  "skill_points": {
    "points_per_grant": 1,
    "grant_every_levels": 3,
    "first_grant_level": 3
  }
}
```

Skill point grants are derived from level thresholds and must be deterministic. If one XP gain
crosses multiple qualifying levels, the character receives every earned stat-point and skill-point
grant in stable event order.

### 5.2 Skill rules

New file: `shared/rules/skills.v0.json`.

Example:

```json
{
  "version": 0,
  "skills": {
    "magic_bolt": {
      "name": "Magic Bolt",
      "kind": "projectile",
      "max_rank": 5,
      "targeting": "direction_or_target",
      "range": 9.0,
      "projectile_speed": 10.0,
      "mana_cost": {
        "base": 3,
        "per_rank": 1
      },
      "damage": {
        "min_base": 2,
        "max_base": 4,
        "min_per_rank": 1,
        "max_per_rank": 2,
        "scales_with": { "magic": 0.25 }
      },
      "cooldown": {
        "type": "attack_interval_multiplier",
        "multiplier": 2.0
      }
    }
  }
}
```

The concrete numbers are first-pass defaults, not final balance. The contract is that the server
loads the skill from shared data, rank modifies the skill through data, and the client renders from
server state rather than duplicating authority.

### 5.3 Skill progression view

Snapshots and `/state` expose skill progression:

```json
{
  "skill_progression": {
    "unspent_skill_points": 1,
    "skills": [
      {
        "skill_id": "magic_bolt",
        "rank": 1,
        "max_rank": 5,
        "can_spend": true
      }
    ]
  }
}
```

`state_delta` emits updates when skill points are earned or spent:

```json
{
  "type": "skill_rank_updated",
  "skill_id": "magic_bolt",
  "rank": 2,
  "unspent_skill_points": 0
}
```

### 5.4 Skill cooldown view

Cooldown state is server-owned and visible enough for the client to draw the disabled/recharge
state:

```json
{
  "skill_cooldowns": [
    {
      "skill_id": "magic_bolt",
      "remaining_ticks": 24,
      "total_ticks": 40
    }
  ]
}
```

The client may animate the radial/progress fill between server updates, but the authoritative
availability check is the server cooldown.

### 5.5 Attack speed and attack interval

Effective attack speed combines character stats, weapon speed, and item stats. The exact formula is
for the plan to pin in golden fixtures, but the expected shape is:

```text
character_attack_speed = DEX-derived attack_speed from character_progression
weapon_attack_speed_multiplier = equipped weapon attack_speed multiplier
item_attack_speed_modifier = sum equipped base/rolled attack_speed modifiers
effective_attack_speed = clamp(character_attack_speed * weapon_attack_speed_multiplier + item_attack_speed_modifier)
attack_interval_ticks = round(base_attack_interval_ticks / effective_attack_speed)
magic_bolt_cooldown_ticks = attack_interval_ticks * 2
```

The plan may choose additive or multiplicative item modifiers, but it must document the choice and
pin examples in the golden fixture. The user-facing requirement is that DEX and items can increase
or decrease effective attack speed, while slower weapon families visibly produce longer
`magic_bolt` cooldowns.

Weapon data example:

```json
{
  "item_type": "greatsword",
  "attack_speed": 0.70
}
```

Validation requirements:

- Every weapon template/static weapon declares `attack_speed`.
- Weapon `attack_speed` must be positive and inside a bounded range.
- Comparable two-handed sword templates are at most `0.70x` the speed of comparable one-handed
  sword templates.
- Long bows are slower than short bows.
- Attack-speed rolled stats have bounded min/max values and may increase or decrease speed.

## 6. Server Behavior

### 6.1 Level-up rewards

The existing XP and level-up path changes from 5 stat points per level to 3 stat points per level.
Skill points are granted on qualifying levels:

```text
level 2 -> +3 stat points, +0 skill points
level 3 -> +3 stat points, +1 skill point
level 4 -> +3 stat points, +0 skill points
level 6 -> +3 stat points, +1 skill point
```

The server emits progression updates in stable order so replay and bots can assert exact outcomes.

### 6.2 Skill point spend

Skill point spend validates:

- actor owns the character
- character has at least one unspent skill point
- requested skill id exists
- requested skill is spendable in v44
- current rank is below max rank

On success, the server decrements `unspent_skill_points`, increments `magic_bolt` rank, persists the
mutation, and emits a skill rank update. Rejections must not mutate skill points or ranks.

### 6.3 Magic bolt cast

`magic_bolt` cast validates:

- actor is alive and allowed to act
- skill id is `magic_bolt`
- rank is at least 1 or v44 treats rank 0 as usable rank 1 only if explicitly chosen in the plan
- player has enough mana
- skill is not on cooldown
- target or direction is valid
- target/range/path collision rules are valid for the chosen projectile model

On success:

- subtract mana
- start cooldown using current attack interval * 2
- spawn/resolve the projectile through deterministic server logic
- apply authoritative combat damage on hit
- emit skill cast, cooldown, mana, projectile/impact, and combat events as needed

The plan should prefer reusing the existing v12 projectile path if that keeps the slice smaller and
deterministic.

## 7. Client Behavior

The skills interface is intentionally small:

- One openable panel lists `Magic Bolt`.
- The row shows rank, max rank, unspent skill points, and a spend button when available.
- The spend button sends the server skill-point intent and waits for authoritative update.
- The player cannot spend points into any other skill in v44.

The skill slot is similarly small:

- One `magic_bolt` slot appears near the existing hotbar.
- Activating it sends a skill-cast intent using the current target or facing/direction.
- After the server accepts the cast, the slot becomes disabled and displays cooldown recovery.
- The slot gradually re-enables based on `remaining_ticks / total_ticks`.
- If the server rejects the cast, the slot returns to the last authoritative available state and
  shows the existing kind of status/error feedback.
## 8. Test And Bot Proof

- `make validate-shared` validates progression changes, skill rules, weapon attack speed, and the
  new golden fixture.
- Go tests prove stat point grant changes, skill point cadence, skill spend success/rejection,
  `magic_bolt` rank scaling, mana spend, cooldown rejection/recovery, attack speed calculations,
  weapon speed validation, and deterministic damage.
- Store tests prove skill points/ranks persist by character and session-start snapshots preserve
  replay determinism.
- Replay tests prove `magic_bolt` cast, cooldown, mana spend, damage, death/loot/XP, skill spend,
  and reconnect reconstruct correctly.
- `make bot scenario=skill_points_and_magic_bolt` proves leveling, skill point earn/spend,
  attack-speed-derived cooldown, cast, cooldown reject/recover, kill, `/state`, reconnect, replay,
  and fresh-session persistence.
- `make client-unit` covers skill panel and skill slot model helpers.
- `HEADLESS=1 make bot-client scenario=18_skill_points_and_magic_bolt.json` proves the real Godot
  client shows the one-skill panel, spends a point, casts, disables the slot, and re-enables it.
- `make ci` is the final gate.

## 9. Resolved Questions And Risks

| # | Question / risk | Decision |
|---|-----------------|----------|
| Q-1 | First skill type? | `magic_bolt`, available to all characters for v44. |
| Q-2 | Do classes ship before skills? | No. Build the active skill foundation first; classes can gate/modify skills later. |
| Q-3 | Skill UI shape? | One small skills panel with only `magic_bolt`, plus one skill slot near the hotbar. |
| Q-4 | Cooldown model? | `magic_bolt` cooldown is 2x current attack interval, derived from authoritative attack speed. |
| Q-5 | Stat points per level? | Change from 5 to 3. Update rules, fixtures, tests, and docs together. |
| Q-6 | Skill point cadence? | Grant 1 skill point every 3 character levels. |
| Q-7 | What does rank improve first? | At minimum, `magic_bolt` damage. Mana/cooldown scaling can be added only if still small and data-driven. |
| R-1 | v26/v31 goldens and bot scenarios may assume 5 stat points per level. | Update every affected fixture/scenario in the same slice; do not preserve stale compatibility. |
| R-2 | Attack speed can expand into a full combat timing rewrite. | v44 requires attack speed for stat views and skill cooldown; normal basic attack timing is deferred unless the plan keeps it small. |
| R-3 | Cooldown UI could drift from server state. | Server owns cooldown availability; client interpolation is presentation only and reconciles to server updates. |
| R-4 | Skill rules could become executable logic. | Keep skill definitions declarative and schema-validated; Go and GDScript consume shared data for display/goldens, with Go owning outcomes. |
