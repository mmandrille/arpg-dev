# Spec: `combat-stat-effects-and-feedback`

Status: Implemented
Branch: `main`
Slice: v31 - real combat stat effects, mitigation, crits, blocks, monster stats, and floating feedback
Baseline: v30 `monster-rarity-and-loot-scaling`
Related:

- [`v21_spec-dungeon-monster-combat.md`](v21_spec-dungeon-monster-combat.md) - generated hostile dungeon mobs and proactive monster attacks
- [`v23_spec-item-templates-and-rolled-drops.md`](v23_spec-item-templates-and-rolled-drops.md) - rolled item payloads and weapon damage rolls
- [`v26_spec-character-stats-and-leveling.md`](v26_spec-character-stats-and-leveling.md) - durable stats, derived substats, XP, and character sheet
- [`v28_spec-full-equipment-and-belt-hotbar.md`](v28_spec-full-equipment-and-belt-hotbar.md) - paper-doll equipment and display-only armor/block rolls
- [`v30_spec-monster-rarity-and-loot-scaling.md`](v30_spec-monster-rarity-and-loot-scaling.md) - generated monster rarity and scaled HP/damage/XP
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - dungeon progression and character-scoped state
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - plugin adoption checklist if client UI work expands
- [`../../PROGRESS.md`](../../PROGRESS.md)

## 1. Purpose

The game now has durable character stats, rolled equipment, full paper-doll slots, generated dungeon
loot, and monster rarity scaling. Several combat-relevant numbers are still mostly cosmetic:
`armor`, `block_percent`, `hit_chance`, `crit_chance`, `crit_damage`, and most monster-side combat
stats do not yet drive authoritative outcomes.

This slice makes those stats real.

After v31:

- Player outgoing attacks use authoritative hit and critical strike rolls.
- Player equipped gear contributes to effective combat stats, including `armor`, `block_percent`,
  and `max_hp`.
- Incoming monster attacks can miss, crit, be mitigated by armor, or be fully blocked by shield
  block.
- Monsters have their own effective combat stats, so enemies can also use hit chance, crit chance,
  crit damage, armor, and block chance.
- Block chance is globally capped at **75%** after all sources are summed.
- Successful, non-blocked hits never deal less than **1 damage**. Full block is the only combat
  outcome that reduces a landed attack to `0`.
- Combat events expose enough outcome metadata for bots, replay, and client floating combat text.
- The Godot client shows floating combat text for damage, misses, blocks, and crits. Crit text is
  slightly larger and tilted.
- Floating combat text can be disabled from settings.
- Character stat hover UI can explain how an effective stat was assembled from base stats,
  allocated points, equipment, rolled stats, caps, and clamps.

The proof is: shared stat aggregation -> deterministic hit/crit/block/mitigation rolls ->
authoritative player and monster combat outcomes -> floating combat feedback -> bot/replay/golden
coverage.

## 2. Non-goals

- No attack-speed gameplay, attack cooldown overhaul, or animation-speed scaling.
- No movement-speed gameplay or pathfinding/prediction speed changes.
- No dodge, elemental resistances, status effects, life steal, thorns, pierce, AoE, or homing.
- No spell system, mana consumers, mana regeneration, or magic-damage skill loop.
- No affix grammar, procedural names, unique/set items, or special-effect execution.
- No enemy equipment inventory or player-visible monster inspection panel.
- No polished item comparison UI. Tooltip/stat breakdown data is required; full comparison UX is
  deferred.
- No final balance pass. Numbers are first-pass deterministic tuning only.
- No production VFX/audio for crits, misses, or blocks.
- No Protobuf migration.

## 3. Files to create or modify

```text
docs/specs/v31_spec-combat-stat-effects-and-feedback.md       - this slice contract
docs/plans/v31_<YYYY-MM-DD>-combat-stat-effects-and-feedback.md - implementation plan
shared/rules/combat.v0.schema.json                            - hit/crit/block/mitigation caps and damage-floor rules
shared/rules/combat.v0.json                                   - first-pass combat stat constants
shared/rules/character_progression.v0.schema.json             - effective-stat breakdown/cap metadata if needed
shared/rules/character_progression.v0.json                    - first-pass stat formula tuning
shared/rules/item_templates.v0.schema.json                    - combat stat roll validation if new keys are added
shared/rules/item_templates.v0.json                           - ensure existing armor/block/max_hp stats aggregate
shared/rules/monsters.v0.schema.json                          - monster combat stats
shared/rules/monsters.v0.json                                 - default monster hit/crit/armor/block values
shared/rules/dungeon_generation.v0.json                       - if rarity should scale monster combat stats beyond HP/damage/XP
shared/protocol/session_snapshot.v1.schema.json               - effective stat breakdowns if current progression view is insufficient
shared/protocol/state_delta.v1.schema.json                    - combat outcome metadata and stat breakdown updates
shared/protocol/examples/session_snapshot.json                - effective stat breakdown example
shared/protocol/examples/state_delta.json                     - hit/crit/block/miss event examples
shared/golden/combat_stat_effects.json                        - pinned hit/crit/block/mitigation/min-damage fixture
shared/golden/combat_stat_effects.v0.schema.json              - fixture schema
tools/validate_shared.py                                      - combat stat caps, block cap, damage-floor, golden drift validation
server/internal/game/rules.go                                 - parse combat stat constants and monster stats
server/internal/game/sim.go                                   - effective stat aggregation and combat resolution
server/internal/game/types.go                                 - stat breakdown and combat event view types if needed
server/internal/game/game_test.go                             - deterministic stat combat tests
server/internal/replay/replay.go                              - replay parity if event/snapshot shapes change
server/internal/http                                          - `/state` parity for effective stats and events if needed
client/scripts/main.gd                                        - floating combat text from authoritative events
client/scripts/character_stats_panel.gd                       - hover breakdowns for effective stats
client/scripts/settings_menu.gd                               - floating combat text toggle, or existing menu settings file
client/tests/test_golden.gd                                   - combat stat fixture checks
client/tests                                                  - focused floating-text/settings tests if helpers are extracted
tools/bot/run.py                                              - assertions for combat outcome metadata and settings if needed
tools/bot/scenarios/22_combat_stat_effects.json               - protocol end-to-end proof
tools/bot/scenarios/client/11_combat_feedback.json            - optional Godot client UI proof
PROGRESS.md                                              - lifecycle update when v31 ships
```

Protocol note: v31 may extend existing v1 events with optional fields rather than bumping protocol
version. If the current schema cannot represent combat outcomes clearly, the plan must call out the
coordinated schema change and update every producer/consumer in the same slice.

## 4. Combat model

### 4.1 Effective combat stats

Both players and monsters resolve combat from effective stats.

Minimum effective stat catalog for v31:

| Stat | Player sources | Monster sources | Used for |
|------|----------------|-----------------|----------|
| `damage_min` / `damage_max` | weapon base/rolls, STR-derived bonus | monster rules, rarity scaling if configured | raw outgoing damage range |
| `hit_chance` | derived DEX formula, equipment if present | monster rules | hit/miss roll |
| `crit_chance` | derived DEX formula, equipment if present | monster rules | crit roll after hit |
| `crit_damage` | derived STR formula, equipment if present | monster rules | crit multiplier |
| `armor` | derived VIT formula, equipped armor/shield/jewelry rolls | monster rules, rarity if configured | flat damage mitigation |
| `block_percent` | equipped shield and future skills only | monster rules | full block chance |
| `max_hp` | derived VIT formula and equipped item stats | monster rules and rarity scaling | HP cap / max HP |

`attack_speed`, `movement_speed`, `max_mana`, and future spell stats may remain displayed but must
not affect gameplay in v31 unless the plan explicitly narrows and proves the change.

### 4.2 Player stat aggregation

The authoritative player effective stat view is assembled from:

1. Base character stats from progression rules.
2. Allocated character stat points.
3. Derived formulas from `character_progression.v0.json`.
4. Equipped static item `base_stats`.
5. Equipped instance `rolled_stats`.
6. Global caps and clamps from shared combat/stat rules.

The current v26 `derived_stats` snapshot should become the final effective values used by combat.
If the implementation needs to preserve "character-only derived stats" separately, add explicit
names rather than overloading one field.

### 4.3 Monster stat aggregation

Monsters get effective combat stats from rules data. Generated dungeon monster rarity may continue
to scale existing HP/damage, and may also scale or override hit/crit/armor/block if the plan finds a
small, deterministic data shape.

Static lab monsters should remain simple but can opt into explicit combat stats so fixtures can pin
miss, crit, block, and mitigation outcomes without relying on generated dungeon placement.

Monster stats are server-authoritative. The client does not predict monster combat outcomes.

## 5. Damage resolution

### 5.1 Outgoing attack order

All damaging attacks, including player melee, player projectile impact, monster proactive attacks,
and monster retaliation, use the same conceptual order:

```text
attacker effective stats + defender effective stats
  -> hit roll
  -> block roll if hit
  -> raw damage roll
  -> crit roll if hit and not blocked
  -> armor mitigation
  -> minimum damage floor
  -> HP mutation and authoritative event
```

The implementation may consume random draws in a slightly different internal order only if the
order is documented in the plan and pinned by golden fixtures. The key requirement is replay-stable
determinism: the same seed and ordered inputs must reproduce the same hit, miss, block, crit,
damage, HP, death, loot, XP, and event metadata.

### 5.2 Hit and miss

`hit_chance` is a probability in `[0.0, 1.0]`.

- If the hit roll fails, no damage roll is applied and the event outcome is `miss`.
- Misses do not trigger block, crit, armor, retaliation, death, loot, or XP.
- Existing `attack_missed` may remain, but the event must include enough metadata for client text
  and bots to identify attacker, defender, and correlation.

### 5.3 Block

`block_percent` is a percentage stat represented in item data as `0..100`.

- Effective block chance is converted to `0.0..1.0`.
- Effective block chance is capped at **75%** after all gear, stat, and future skill sources are
  summed.
- A successful block fully negates damage.
- A block produces a `0` damage combat outcome and clear event metadata.
- Block does not trigger armor mitigation or crit damage.
- Only a successful block may reduce a landed attack to `0`.

### 5.4 Critical hits

`crit_chance` is a probability in `[0.0, 1.0]`; `crit_damage` is a multiplier.

- Crit is rolled only after a hit that was not blocked.
- Raw damage is multiplied by `crit_damage` before armor mitigation unless the implementation plan
  deliberately chooses the opposite and pins the result.
- Crit event metadata must be visible to bots, replay, and floating combat text.
- Crit floating text is slightly larger and tilted compared with normal damage text.

### 5.5 Armor mitigation and minimum damage

Armor is flat mitigation in v31.

Suggested first-pass formula:

```text
mitigated_damage = raw_or_crit_damage - effective_armor
final_damage = max(1, mitigated_damage)
```

This keeps armor easy to reason about while avoiding zero-damage stalemates. The only exception is
block:

```text
final_damage = 0 when blocked
```

If future balance needs smoother scaling, percentage or curve-based mitigation can replace this
formula in shared data. v31 should not introduce that complexity.

### 5.6 Death, retaliation, loot, and XP

Existing death, loot, and XP semantics remain:

- A killed monster emits `monster_killed`, drops loot through the current loot path, and awards XP
  once.
- A killed player emits `player_killed` and rejects further gameplay intents as today.
- Retaliation should not occur when the player misses, when the monster blocks, or when the monster
  dies.
- Retaliation may occur after a successful non-lethal hit, including a crit.

## 6. Protocol and event metadata

Combat events must let clients and bots distinguish:

- `miss`
- `block`
- normal damage
- critical damage
- killed target
- attacker entity id
- defender entity id
- final damage
- raw damage and mitigated amount if useful for debugging
- correlation id

Preferred shape is to extend existing event payloads with optional fields and keep event names
stable where possible:

```json
{
  "event_type": "monster_damaged",
  "entity_id": "2",
  "source_entity_id": "1",
  "damage": 5,
  "outcome": "crit",
  "raw_damage": 8,
  "mitigated_damage": 3,
  "blocked": false,
  "critical": true,
  "correlation_id": "attack-1"
}
```

For misses and blocks, either reuse existing event names with richer metadata or add explicit event
types. The plan must choose one consistent approach and update schemas, examples, bots, replay, and
Godot parsing together.

## 7. Effective stat breakdown UI

The character stats panel should support hover breakdowns for effective stats.

Minimum v31 behavior:

- Hovering a stat shows a compact breakdown of active sources.
- The breakdown includes base/derived character contribution, equipment contribution, rolled stat
  contribution, and final caps/clamps.
- Block chance explicitly shows the **75% cap** when the uncapped value exceeds the cap.
- The values shown in UI come from server snapshots/deltas or shared-rule evaluation that is golden
  checked against the server. The client must not invent hidden gameplay math.

Suggested data shape:

```json
{
  "key": "block_percent",
  "value": 75,
  "uncapped_value": 82,
  "cap": 75,
  "sources": [
    { "label": "Cave Shield", "value": 8, "kind": "equipment_base" },
    { "label": "Rolled block", "value": 4, "kind": "equipment_roll" },
    { "label": "Future skill", "value": 70, "kind": "skill" }
  ]
}
```

The plan may choose a smaller initial source catalog if no skills exist yet, but the data model
should not require a rewrite when skills arrive.

## 8. Floating combat text and settings

The Godot client displays floating combat text from authoritative combat events only.

Minimum visual outcomes:

| Outcome | Presentation |
|---------|--------------|
| normal damage | readable number near target |
| crit | larger number, slight tilt, distinct color |
| miss | short `Miss` text |
| block | short `Block` text, `0`, or both if the plan chooses |
| player damage | distinguishable from monster damage |

Floating combat text must be possible to disable in settings.

Settings requirements:

- Add a `floating_combat_text` boolean, default `true`.
- Expose it in the existing menu/settings UI.
- Persist it through the same local settings path used by current menu settings.
- When disabled, authoritative combat still occurs and events are still processed; only the
  presentation text is suppressed.

## 9. Determinism and replay

This slice changes combat RNG consumption and therefore has high determinism risk.

The implementation must:

- Use only the seeded sim RNG in authoritative combat.
- Avoid wall-clock time, unseeded randomness, map iteration order, or client state in combat
  resolution.
- Pin draw order in golden fixtures for hit, block, raw damage, crit, and any future combat roll.
- Keep melee and projectile impact behavior replay-stable.
- Ensure `/state`, reconnect resume, replay timeline, and `make replay` agree after stat-driven
  combat outcomes.

If changing draw order breaks older golden fixtures, update all fixtures and document why in the
plan. Backward compatibility is not required for its own sake during active development, but drift
must be intentional and covered.

## 10. Bot and golden proof

### 10.1 Golden fixture

Create `shared/golden/combat_stat_effects.json` with pinned cases for:

1. Player hit chance miss.
2. Player critical hit against a monster.
3. Monster armor mitigation reducing player damage but respecting minimum `1`.
4. Player armor mitigation reducing monster damage but respecting minimum `1`.
5. Shield block fully negating incoming damage.
6. Block cap clamping an uncapped value above `75%`.
7. Monster crit against player.
8. Monster block against player.
9. Projectile attack using the same hit/crit/mitigation model at impact.

Go and Godot golden tests must agree on effective stat aggregation and formula output. Go owns
authoritative RNG resolution; Godot only needs to prove shared formula and display-relevant
interpretation equivalence.

### 10.2 Protocol bot

Add `tools/bot/scenarios/22_combat_stat_effects.json`.

High-level flow:

1. Start in a deterministic combat-stat lab world or pinned dungeon seed.
2. Pick up and equip known armor, shield, and weapon items.
3. Allocate stat points if needed to create visible hit/crit/armor differences.
4. Fight pinned monsters that demonstrate player miss, player crit, monster mitigation, monster
   block, monster crit, player mitigation, player block, and minimum damage.
5. Assert final HP, monster HP/death, event metadata, `/state`, reconnect resume, replay, and
   character persistence.

### 10.3 Client bot

Add `tools/bot/scenarios/client/11_combat_feedback.json` if the plan can keep it focused.

Minimum client proof:

- Floating damage text appears for a normal hit.
- Crit text uses a distinct larger/tilted style.
- Settings toggle disables new floating combat text.
- Character stat hover shows at least one effective stat breakdown with equipment contribution.

If client-bot automation makes the slice too large, the plan may defer the client bot but must keep
unit/headless coverage for formatting helpers and settings persistence.

## 11. Acceptance criteria

1. Player attacks use authoritative hit and crit rolls for melee and projectile damage.
2. Monster attacks use authoritative hit and crit rolls.
3. Player and monster armor reduce incoming damage with the v31 flat formula.
4. Successful non-blocked hits always deal at least `1` damage after mitigation.
5. Block fully negates damage and cannot exceed `75%` effective chance.
6. Player equipment `armor`, `block_percent`, and `max_hp` contribute to final effective stats.
7. Monsters have rule-defined effective combat stats for hit, crit, crit damage, armor, and block.
8. Combat events expose outcome metadata sufficient for bots, replay, and floating combat text.
9. Character stat hover UI can show source breakdowns for at least damage, armor, block, and max HP.
10. Floating combat text shows normal damage, crit, miss, and block outcomes from authoritative
    events.
11. Floating combat text can be disabled and the setting persists locally.
12. `/state`, reconnect resume, replay timeline, and replay verification remain deterministic.
13. `make validate-shared`, Go tests, Godot golden/unit tests, protocol bot, optional client bot,
    and `make ci` pass.

## 12. Validation commands

Expected final gate:

```bash
make validate-shared
go test ./internal/game/...
make client-unit
make bot
make bot-client
make ci
```

During planning, narrow commands may be used for focused iteration, but the slice is not complete
until `make ci` is green.

## 13. Open questions resolved

| Question | Decision |
|----------|----------|
| Armor behavior | Flat mitigation, with successful non-blocked hits floored to `1` damage. |
| Block behavior | Full damage negation, globally capped at `75%` effective chance. |
| Crit feedback | Show larger, slightly tilted floating combat text. |
| Floating combat text setting | Add a settings toggle; default enabled. |
| Monster stats | Monsters also get hit, crit, crit damage, armor, and block. |
| Stat breakdowns | Effective stats should expose enough source data for hover explanations. |

## 14. ADR alignment

- ADR-0001 D2: the server remains authoritative for every combat outcome.
- ADR-0001 D6: combat math stays in shared rules/data with bounded evaluators and golden fixtures.
- ADR-0001 D8: deterministic combat roll order is pinned for replay and agent debugging.
- ADR-0008: character progression remains character-scoped; session combat state remains replayable.

The plan must run the Godot plugin adoption checklist only if it expands client UI beyond the
existing in-repo Control/menu/stat-panel patterns. Expected decision: reject gameplay/stat plugins;
borrow only small UI presentation ideas if useful.
