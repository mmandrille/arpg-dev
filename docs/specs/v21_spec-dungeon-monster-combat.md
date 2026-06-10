# Spec: `dungeon-monster-combat`

Status: Draft — pending review
Branch: `feature/dungeon-monster-combat`
Slice: v21 — dungeon floors spawn hostile monsters that chase and attack the player
Baseline: v20 `play-session-loop`
Related:

- [`v17_spec-monster-chase-movement.md`](v17_spec-monster-chase-movement.md) — chase infrastructure reused here
- [`v18_spec-dungeon-levels-and-stairs.md`](v18_spec-dungeon-levels-and-stairs.md) — dungeon gen extended here
- [`v20_spec-play-session-loop.md`](v20_spec-play-session-loop.md) — playable loop this slice threatens
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) — D3 PCG density, D5 town safe zone
- [`../../PROGRESS.md`](../../PROGRESS.md)

## 1. Purpose

Post-v20 the dungeon is entirely passive: floors contain loot and stairs, but zero monsters.
Chase infrastructure (v17) allows monsters to follow the player; they never initiate damage.
The play loop (town → descend → explore) has no threat and no meaningful risk.

This slice closes that gap. Every generated dungeon floor spawns a small number of `dungeon_mob`
entities. `dungeon_mob` uses the existing `chase` behavior and adds a new proactive attack: when
the monster is within melee reach of the player and its per-entity tick-based attack cooldown has
elapsed, it deals damage and emits the same `player_damaged` / `player_killed` events that
retaliation already uses. The player's hit and death animations (ADR-0007) fire without any client
change.

Town (level `0`) is unaffected: no monsters spawn there. The dungeon is now the danger zone the
play loop requires.

## 2. Non-goals

- No monster loot drops. `dungeon_mob` is a pure obstacle for v21; reward drops deferred.
- No monster attack animation. Client monster walk/idle presentation is unchanged; player hit
  reaction already fires from `player_damaged` events.
- No ranged monster attacks, AoE, or multi-monster coordination.
- No monster stat scaling by dungeon depth. All floors use the same flat `dungeon_mob` definition.
- No town safe-zone protocol guard. Town is safe because monsters simply are not placed there.
- No character-scoped persistence or respawn.
- No production monster art.

## 3. Files to create or modify

```text
docs/specs/v21_spec-dungeon-monster-combat.md              — this slice contract
docs/plans/v21_2026-06-06-dungeon-monster-combat.md        — implementation plan
shared/rules/monsters.v0.json                              — add dungeon_mob definition
shared/rules/dungeon_generation.v0.json                    — add monster_placement block
shared/golden/dungeon_monster_attack.json                  — pin proactive attack tick and damage
server/internal/game/dungeon_gen.go                        — generatedDungeonLevel gains monsters field
server/internal/game/sim.go                                — advanceMonsterAttack; lastAttackTick on entity
server/internal/game/game_test.go                          — golden test for dungeon_monster_attack.json
client/tests/test_golden.gd                                — GDScript cross-check of dungeon_monster_attack.json
tools/bot/scenarios/14_dungeon_monsters.json               — prove proactive hit end-to-end
PROGRESS.md                                           — lifecycle update when v21 ships
```

No protocol schema changes. `player_damaged` / `player_killed` event types already exist in
`shared/protocol/state_delta.v1.schema.json`.

## 4. Data shapes

### Monster rule addition

Add to `shared/rules/monsters.v0.json`:

```json
"dungeon_mob": {
  "name": "Cave Wraith",
  "max_hp": 4,
  "behavior": "chase",
  "aggro_radius": 6.0,
  "leash_radius": 10.0,
  "move_speed": 1.0,
  "attack_damage": { "min": 1, "max": 2 },
  "attack_cooldown_ticks": 40
}
```

`attack_damage` and `attack_cooldown_ticks` are new optional fields in the monster schema.
Monsters without these fields never initiate attacks (all existing monsters retain their current
behavior; `retaliation_damage` is unchanged).

### Dungeon generation rule addition

Add to `shared/rules/dungeon_generation.v0.json`:

```json
"monster_placement": {
  "count": 2,
  "monster_def_id": "dungeon_mob",
  "margin_from_wall": 2.0,
  "min_spawn_distance": 6.0,
  "max_attempts": 64
}
```

`min_spawn_distance` is checked against the player spawn point to prevent instant-aggro spawns.

### Golden fixture

`shared/golden/dungeon_monster_attack.json`:

```json
{
  "description": "dungeon_mob proactive attack: player takes damage without sending any action_intent",
  "session_seed": "<pinned at implementation time>",
  "level": -1,
  "tick_of_first_player_damaged": <pinned>,
  "damage": <pinned>,
  "player_hp_after": <pinned>
}
```

Values are pinned deterministically during implementation by running the sim headless with the
chosen seed and recording the first `player_damaged` emission on the default play loop.

### Entity field addition

The internal `entity` struct in the Go sim gains a `lastAttackTick` field (type `int`, zero value
= never attacked). This is purely sim-internal state; it does not appear on the wire.

## 5. Architecture and flow

### Monster placement (dungeon gen)

```text
GenerateDungeonLevel(seed, levelNum, rules)
  — existing: place walls, down stair, [up stair -1 only], teleporter, training_badge loot
  — new:      place rules.MonsterPlacement.Count monsters at random positions
              (margin from wall, min_spawn_distance from player spawn,
               collision-free from stairs/teleporter; same retry logic as stair placement)
  — returns:  generatedDungeonLevel now includes []generatedMonster{defID, pos}
```

The sim's existing `ensureDungeonLevel` / `spawnDungeonLevel` path spawns monsters from
`generatedDungeonLevel.monsters` the same way it already spawns loot entities.

### Proactive attack (sim tick loop)

```text
Sim.Tick (per 20 Hz tick, per active level)
  advanceMonsterMovement   ← v17, unchanged
  advanceMonsterAttack     ← new
    for each live monster m with attack_damage defined:
      player := activeLevel.entities[playerID]
      if player dead or on different level: skip
      if distance(m.pos, player.pos) > playerMeleeReach + m.interactionRadius: skip
      if currentTick - m.lastAttackTick < m.attackCooldownTicks: skip
      dmg := rollRange(m.attackDamage)          ← uses sim seeded RNG
      player.hp -= dmg  (floor 0)
      m.lastAttackTick = currentTick
      emit player_damaged / player_killed event
  advanceProjectiles       ← v12, unchanged
```

Attack range uses `playerMeleeReach + monster.interactionRadius` (the same geometry as player
melee) so the monster attacks the moment it can be attacked back. The seeded RNG call in
`rollRange` is deterministic and replay-safe.

### Client (no change)

`player_damaged` / `player_killed` events already drive player hit/death animations via the
existing `AnimationController` priority machine. No new client code is needed.

## 6. Acceptance criteria

1. Generated dungeon levels `-1`, `-2`, and below each contain exactly
   `dungeon_generation.monster_placement.count` (2) `dungeon_mob` entities, placed by deterministic
   PCG from the level seed.
2. A `dungeon_mob` in aggro range moves toward the player each tick (v17 chase, unchanged).
3. When a `dungeon_mob` is within melee reach and `currentTick - lastAttackTick >=
   attack_cooldown_ticks`, it emits a `player_damaged` or `player_killed` event without the player
   sending any `action_intent`.
4. The damage roll uses the sim seeded RNG: same seed + same inputs → identical damage value and
   tick on every replay.
5. `shared/golden/dungeon_monster_attack.json` is verified by Go test and `test_golden.gd`.
6. Bot scenario `14_dungeon_monsters.json` completes: descend, receive a `player_damaged` event
   with no player attack sent, kill the mob, and pass `/state` + reconnect resume + replay.
7. Town level `0` has no monsters; existing bot scenarios `01`–`13` are unaffected.
8. Monsters already in combat lab worlds (`chase_lab`, `chase_maze`, `leash_lab`) retain their
   existing behavior; `retaliation_damage` on all other monster defs is unchanged.
9. `make ci` green.

## 7. Testing plan

1. `make validate-shared` — schema validates updated `monsters.v0.json` and
   `dungeon_generation.v0.json`.
2. `cd server && go test ./internal/game/... -run 'DungeonMonster|ProactiveAttack|MonsterAttack'`
   — golden fixture test.
3. `make bot` — scenario `14_dungeon_monsters.json` + no regression in `01`–`13`.
4. `make client-unit` — `test_golden.gd` dungeon monster attack cross-check.
5. `make ci` — final gate.
6. Manual: `make play` — descend to level `-1`, stand still, observe player HP dropping and hit
   animation firing without clicking.

## 8. Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | `attack_cooldown_ticks` is a per-monster rules field, not a global combat constant. | Different monster types will need different speeds. Flat default of 40 ticks (2 s) for `dungeon_mob`. |
| 2 | Attack range = `playerMeleeReach + monster.interactionRadius` (same geometry as player melee). | Symmetric reach: monster attacks when the player can also reach it, avoiding asymmetric dead zones. |
| 3 | `lastAttackTick` is sim-internal; not on the wire. | Replay reconstructs it from the input stream + tick counter; no protocol change needed. |
| 4 | Existing `retaliation_damage` on static monsters is unchanged. | Non-chase monsters never reach the player, so retaliation remains their only damage vector. |
| 5 | `dungeon_mob` has no loot drop. | Deferred; killing mobs is its own reward for v21 (survival). |
| 6 | Monster spawn positions use the same margin/retry pattern as stair placement. | Re-uses proven PCG code; no new algorithmic complexity. |

## 9. Open questions

| # | Question | Resolution |
|---|----------|------------|
| D-1 | Monster stat scaling by dungeon depth. | Deferred — all floors use the same `dungeon_mob` def for v21. |
| D-2 | Monster loot drops on dungeon mobs. | Deferred to a future slice. |
| D-3 | Monster attack animation. | Deferred — client already handles player reaction via existing events. |
| D-4 | Town safe-zone as a protocol-level combat rejection (ADR-0008 D5). | Deferred — town is safe because monsters simply don't spawn there. |
