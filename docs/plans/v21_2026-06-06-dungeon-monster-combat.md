# v21 Plan — Dungeon monster combat

Status: Ready for implementation
Goal: Make generated dungeon floors dangerous by spawning chase monsters that proactively damage the player in melee range.

Architecture: Dungeon monsters are generated from shared dungeon-generation rules and spawned only on negative dungeon levels. The Go sim remains authoritative for monster AI, cooldowns, damage rolls, HP changes, and emitted `player_damaged` / `player_killed` events. No protocol schema change is needed because existing player damage events already drive client reactions. Use a `no_drop` loot table for `dungeon_mob` so the existing monster contract stays intact while v21 defers monster rewards.
Tech stack: Shared JSON contracts and golden fixtures, Go authoritative sim, Python protocol bot, Godot golden tests.

## Baseline and shortcut decision

v21 builds on v20 `play-session-loop`: sessions start in town level `0`, dungeon floors are lazy generated negative levels, and `make play` enters the town-to-dungeon loop. It reuses v17 chase AI/pathfinding, v18 generated dungeon floors, v19/v20 level travel and scoped deltas, and v4 player damage/death events.

Godot shortcut adoption checklist:

- **Reason:** this slice has no new client UI, camera, art, or presentation system. Existing `player_damaged` / `player_killed` event mappings already trigger hit/death reactions.
- **Borrow:** existing Godot golden fixture pattern in `client/tests/test_golden.gd`; no addon or asset pack needed.

Spec review notes resolved during planning:

- Add `no_drop` to `shared/rules/loot_tables.v0.json` and set `dungeon_mob.loot_table = "no_drop"` because the current monster schema and validator require a loot table.
- Use sim-native cooldown state: prefer `lastAttackTick uint64` plus `hasAttacked bool` on internal `entity` so the first attack can occur immediately after the monster reaches melee range, then repeat every `attack_cooldown_ticks`.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/monsters.v0.schema.json` | Add optional proactive attack fields |
| Modify | `shared/rules/monsters.v0.json` | Add `dungeon_mob` chase attacker |
| Modify | `shared/rules/loot_tables.v0.json` | Add empty `no_drop` table |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Add `monster_placement` contract |
| Modify | `shared/rules/dungeon_generation.v0.json` | Configure dungeon monster placement |
| Add | `shared/golden/dungeon_monster_attack.json` | Pin proactive attack seed, tick, damage, HP |
| Modify | `tools/validate_shared.py` | Validate attack fields, no-drop table, dungeon monster golden drift |
| Modify | `server/internal/game/rules.go` | Parse and validate monster attack and placement rules |
| Modify | `server/internal/game/dungeon_gen.go` | Generate deterministic dungeon monster positions |
| Modify | `server/internal/game/sim.go` | Spawn generated monsters and advance proactive attacks |
| Modify | `server/internal/game/game_test.go` | Generated placement, proactive attack, golden tests |
| Modify | `client/tests/test_golden.gd` | Cross-check dungeon monster attack fixture |
| Modify | `tools/bot/run.py` | Add/extend scenario helpers for passive damage and generated mob targeting |
| Add | `tools/bot/scenarios/14_dungeon_monsters.json` | End-to-end dungeon threat proof |
| Modify | `PROGRESS.md` | v21 lifecycle update when complete |

## Task 1 — Shared rules and validation

Files:

- Modify: `shared/rules/monsters.v0.schema.json`
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/rules/loot_tables.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `shared/rules/dungeon_generation.v0.json`
- Add: `shared/golden/dungeon_monster_attack.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Extend monster schema with optional `attack_damage` damage range and `attack_cooldown_ticks` integer.
- [x] Step 1.2: Add `no_drop` loot table with no `drops` and no weighted `entries`; verify `LootDrops("no_drop")` returns an empty slice.
- [x] Step 1.3: Add `dungeon_mob` with `loot_table: "no_drop"`, `behavior: "chase"`, aggro/leash/move speed, `attack_damage`, and `attack_cooldown_ticks`.
- [x] Step 1.4: Extend dungeon-generation schema/rules with `monster_placement` fields: `count`, `monster_def_id`, `margin_from_wall`, `min_spawn_distance`, `max_attempts`.
- [x] Step 1.5: Add validator checks:
      `attack_damage` is valid when present; `attack_cooldown_ticks > 0`; proactive attack fields only apply to chase monsters; `monster_placement.monster_def_id` exists and references a chase monster; empty `no_drop` is valid.
- [x] Step 1.6: Add initial golden fixture shape for `dungeon_monster_attack.json`; pin final values after the Go sim test harness computes the deterministic first attack.

```bash
make validate-shared
```

## Task 2 — Server rule loading and dungeon generation

Files:

- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Add `AttackDamage *DamageRange` and `AttackCooldownTicks int` to `MonsterDef`.
- [x] Step 2.2: Add typed `MonsterPlacementRules` to `DungeonGenerationRules` and mirror shared validation in `LoadRules`.
- [x] Step 2.3: Add `generatedMonster{defID, pos}` and `generatedDungeonLevel.monsters`.
- [x] Step 2.4: Place exactly `monster_placement.count` monsters on each generated dungeon level using the local per-level RNG stream, after stairs/teleporter placement.
- [x] Step 2.5: Reject monster placement candidates inside wall margins, too close to player spawn, too close to stairs/teleporter, or overlapping already placed monsters.
- [x] Step 2.6: Add tests for deterministic generation on levels `-1` and `-2`, exact monster count, no monsters on town level `0`, and stable positions for the pinned seed.

```bash
cd server && go test ./internal/game/... -run 'Dungeon.*Monster|GenerateDungeon'
```

## Task 3 — Server proactive monster attacks

Files:

- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 3.1: Extend internal `entity` with cooldown state, preferably `lastAttackTick uint64` and `hasAttacked bool`.
- [x] Step 3.2: Spawn generated dungeon monsters in `populateDungeonLevel` with `monsterDefID`, `lootTable`, `spawnPos`, HP, and idle AI mode.
- [x] Step 3.3: Add `advanceMonsterAttack(res)` after `advanceMonsterMovement` and before `advanceProjectiles`.
- [x] Step 3.4: Iterate active-level entities with `sortedEntityIDs`; skip dead monsters, missing attack rules, dead player, leashed/returning monsters, and monsters outside `playerMeleeReach + monster interaction radius`.
- [x] Step 3.5: Roll damage with the sim seeded RNG, floor player HP at `0`, emit an `entity_update` for the player, then emit `player_damaged` or `player_killed`.
- [x] Step 3.6: Update cooldown state only when an attack is emitted; enforce repeats every `attack_cooldown_ticks`.
- [x] Step 3.7: Add behavioral tests proving passive player damage without `action_intent`, cooldown gating, deterministic replay from the same seed/inputs, and no proactive attacks for existing static/retaliation-only monsters.
- [x] Step 3.8: Use the behavioral test harness to pin `shared/golden/dungeon_monster_attack.json`.

```bash
cd server && go test ./internal/game/... -run 'ProactiveAttack|DungeonMonster|MonsterAttack|Replay'
```

## Task 4 — Golden cross-checks

Files:

- Modify: `server/internal/game/game_test.go`
- Modify: `client/tests/test_golden.gd`
- Modify: `tools/validate_shared.py`

- [ ] Step 4.1: Add Go golden test that loads `dungeon_monster_attack.json`, runs the pinned seed through level `-1`, records the first `player_damaged`, and asserts tick, damage, and HP.
- [ ] Step 4.2: Add `tools/validate_shared.py` checks that the golden monster id, seed, level, cooldown, and damage range reference valid shared rules.
- [ ] Step 4.3: Add a Godot golden check that validates fixture constants against shared rules and confirms the referenced monster has proactive attack fields.
- [ ] Step 4.4: Keep `client/tests/test_golden.gd` data-only; do not simulate server combat in GDScript.

```bash
make validate-shared
make client-unit
```

## Task 5 — Bot scenario

Files:

- Modify: `tools/bot/run.py`
- Add: `tools/bot/scenarios/14_dungeon_monsters.json`

- [ ] Step 5.1: Add or reuse bot helpers to descend from town to level `-1`, wait without sending player attacks, and assert a `player_damaged` event was observed.
- [ ] Step 5.2: Add helper support to target a generated `dungeon_mob` by `monster_def_id` after it has approached the player.
- [ ] Step 5.3: Add scenario `14_dungeon_monsters.json`: descend, wait for passive damage, kill one mob with existing `action_entity`/auto-approach flow, and assert current level plus seen `player_damaged` and `monster_killed` events.
- [ ] Step 5.4: Confirm `/state`, reconnect resume, and replay verification still run through the normal `make bot` path for scenario `14`.
- [ ] Step 5.5: Migrate older scenarios only if new dungeon monsters make `12_dungeon_levels.json` or `13_teleporter_lab.json` timing flaky; otherwise leave `01`-`13` unchanged.

```bash
make db-up
make bot
```

## Task 6 — Client verification

Files:

- Modify: `client/tests/test_golden.gd`

- [ ] Step 6.1: Run client unit tests after adding the data-only dungeon monster golden check.
- [ ] Step 6.2: Run client smoke if a server is available, confirming existing `player_damaged` / `player_killed` mappings require no code changes.

```bash
make client-unit
make client-smoke
```

## Task 7 — Lifecycle docs and CI

Files:

- Modify: `PROGRESS.md`

- [ ] Step 7.1: When implementation ships, add v21 to the lifecycle table and mark latest completed slice as `dungeon-monster-combat`.
- [ ] Step 7.2: Document the as-built proof: generated dungeon mobs, proactive attack cooldown, deterministic golden, bot scenario `14`, and no town monsters.
- [ ] Step 7.3: Record deferred follow-ups: monster loot drops, attack animation, depth scaling, ranged/AoE monsters, protocol-level town safe-zone, respawn.

```bash
make ci
```

## Final verification

- [ ] `make validate-shared`
- [ ] `cd server && go test ./internal/game/... -run 'Dungeon.*Monster|ProactiveAttack|MonsterAttack'`
- [ ] `make client-unit`
- [ ] `make bot`
- [ ] `make ci`

Manual check:

```bash
make play
# Descend to level -1, stand still near a dungeon_mob, observe HP dropping and player hit reaction.
```

## Deferred scope

- No monster loot drops beyond the empty `no_drop` table.
- No monster attack animation or production monster art.
- No ranged monsters, AoE, coordination, depth scaling, safe-zone protocol guard, respawn, or character-scoped persistence.
