# v52 Plan - Ranged Monster AI

Status: Implemented
Goal: Add a generated dungeon archer that keeps range, fires server-owned projectiles at players, and has a minimal bow marker in the Godot client.
Architecture: Extend shared monster and dungeon-generation rules with ranged attack fields and a deterministic melee/ranged spawn mix. Reuse the Go sim's existing chase/leash, projectile entity, swept collision, and combat resolution paths, adding only the missing monster-owned projectile branch that can hit players. The Godot client remains presentation-only: it infers `dungeon_archer` from the authoritative `monster_def_id`, attaches a simple procedural bow marker, and exposes that marker through bot debug state.
Tech stack: shared JSON rules/golden, Go deterministic sim and replay tests, Python protocol bot, Godot client presentation/client bot, lifecycle docs.

## Baseline and shortcut decision

Baseline is v51 `mystery-seller-core` on `main`. Reuse:

- v12 player-owned projectile entity lifecycle, `owner_id` / `target_id` / `projectile_def_id`, swept wall/barrier collision, and projectile rendering.
- v17 chase/leash AI, pathfinding, and semantic monster movement tests.
- v21 generated dungeon monster attack loop and `player_damaged` event handling.
- v30 generated monster rarity scaling for HP, attack damage, XP, tint, scale, and loot depth.
- v37 combat-control fixes around projectile aggro and stable combat event helpers.
- Current `main.gd` monster presentation and bot debug state.

Godot plugin shortcut decision: **reject external plugin or asset adoption for v52 implementation**. The adoption checklist in `docs/researchs/godot-plugins-and-shortcuts.md` was reviewed. LimboAI is unnecessary because the authoritative AI is in Go, and Kenney/other asset packs are unnecessary for the requested minimal visual distinction. Build a small procedural bow marker in `main.gd` and defer production archer art/VFX/audio.

Key implementation decisions:

- Add `dungeon_archer` as a separate monster rule with `behavior: "chase"`, `attack_mode: "ranged"`, `attack_range`, `projectile_speed`, and `projectile_def_id`.
- Keep `dungeon_mob` melee. Existing monsters default to melee when `attack_mode` is omitted.
- Extend `monster_placement` with a data-driven spawn mix:
  - `monster_pool`: weighted entries, initially melee-heavy with `dungeon_mob` and `dungeon_archer`.
  - `minimum_monsters`: initially one `dungeon_archer` on normal generated dungeon floors when count is positive.
  - Keep `monster_def_id` as the default melee id for legacy expectations and champion common minions.
- Boss floors keep the existing boss-floor population in v52.
- Monster-owned projectiles should target the selected player only; they should not damage monsters or players on other levels.
- Avoid a protocol bump. Existing schemas already represent monster `monster_def_id`, projectile `owner_id`, projectile `target_id`, and combat events.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `docs/specs/v52_spec-ranged-monster-ai.md` | Slice spec |
| Add | `docs/plans/v52_2026-06-10-ranged-monster-ai.md` | This implementation plan |
| Modify | `docs/PROGRESS.md` | Lifecycle update when v52 ships |
| Modify | `shared/rules/monsters.v0.json` | Add `dungeon_archer` and ranged attack fields |
| Modify | `shared/rules/monsters.v0.schema.json` | Validate monster ranged fields |
| Modify | `shared/rules/dungeon_generation.v0.json` | Configure melee/ranged monster mix |
| Modify | `shared/rules/dungeon_generation.v0.schema.json` | Validate spawn mix fields |
| Add if useful | `shared/golden/ranged_monster_ai.json` | Pinned semantic archer fixture |
| Add if useful | `shared/golden/ranged_monster_ai.v0.schema.json` | Fixture schema |
| Modify | `tools/validate_shared.py` | Cross-rule validation for ranged monster/spawn mix/golden |
| Modify | `server/internal/game/rules.go` | Parse and validate ranged monster fields and spawn mix |
| Modify | `server/internal/game/dungeon_gen.go` | Guarantee and roll generated archer spawns |
| Modify | `server/internal/game/sim.go` | Ranged standoff, line of sight, monster projectile spawn, player hit branch |
| Modify | `server/internal/game/game_test.go` | Rule, spawn, attack, blocked shot, regression tests |
| Modify if needed | `server/internal/replay/replay_test.go` | Replay proof if game tests do not cover it |
| Modify | `client/scripts/main.gd` | Procedural archer bow marker and debug state |
| Modify | `client/scripts/bot_scenario_runner.gd` | Bow marker/presentation assertion if needed |
| Modify | `client/tests/test_client_bot.gd` | Validate new client scenario/assertion shape |
| Modify if golden added | `client/tests/test_golden.gd` | Client-side fixture/rule drift check |
| Modify | `tools/bot/run.py` | Protocol helper/assertions for archer projectile proof if needed |
| Modify | `tools/bot/test_protocol.py` | Unit coverage for helper/assertion behavior |
| Add | `tools/bot/scenarios/38_ranged_monster_ai.json` | Protocol proof |
| Add | `tools/bot/scenarios/client/25_ranged_monster_ai.json` | Godot client proof |

## Task 1 - Shared rules and validation

Files:
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/rules/monsters.v0.schema.json`
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.schema.json`
- Modify: `server/internal/game/rules.go`
- Modify: `tools/validate_shared.py`
- Add if useful: `shared/golden/ranged_monster_ai.json`
- Add if useful: `shared/golden/ranged_monster_ai.v0.schema.json`

- [x] Step 1.1: Extend monster rule schema and Go `MonsterDef` with optional `attack_mode`, `attack_range`, `projectile_speed`, and `projectile_def_id`.
```bash
make validate-shared
```

- [x] Step 1.2: Validate default melee semantics: omitted `attack_mode` behaves as melee; melee monsters may not declare projectile fields; ranged monsters require chase behavior, attack damage, cooldown, range above melee reach, positive projectile speed, and non-empty projectile id.
```bash
cd server && go test ./internal/game/... -run 'Test.*Rules|Test.*Monster' -count=1
```

- [x] Step 1.3: Add `dungeon_archer` to `monsters.v0.json` with conservative HP/XP/loot parity to `dungeon_mob`, ranged damage tuned near current melee damage, `attack_range` around bow reach, and `projectile_def_id: "training_arrow"`.
```bash
make validate-shared
```

- [x] Step 1.4: Extend dungeon-generation rules/schema with `monster_pool` weighted entries and `minimum_monsters` entries under `monster_placement`; validate referenced monster ids, positive weights, non-negative minimum counts, and minimum totals not exceeding base count.
```bash
make validate-shared
```

- [x] Step 1.5: Configure normal dungeon generation so `dungeon_mob` remains dominant and at least one `dungeon_archer` appears on normal generated dungeon floors with positive monster count.
```bash
make validate-shared
```

- [x] Step 1.6: Add cross-consistency validation in `tools/validate_shared.py` so the pool references known chase monsters, includes `dungeon_archer`, and does not break existing `dungeon_monster_attack` and `monster_rarity` fixtures.
```bash
make validate-shared
```

## Task 2 - Dungeon generation and generated monster state

Files:
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 2.1: Extend generated monster output to preserve the rolled monster def id from `monster_pool` while keeping existing rarity and loot-band selection.
```bash
cd server && go test ./internal/game/... -run 'TestDungeon.*Monster|TestMonsterRarity' -count=1
```

- [x] Step 2.2: Place configured minimum monsters before weighted pool fill on normal generated floors, using the existing deterministic dungeon RNG and placement constraints.
```bash
cd server && go test ./internal/game/... -run 'TestDungeon.*Monster|TestGenerated.*Monster' -count=1
```

- [x] Step 2.3: Keep champion common minions as the default melee `monster_def_id` for v52, and keep boss-floor generation unchanged.
```bash
cd server && go test ./internal/game/... -run 'TestBossFloor|TestMonsterRarity' -count=1
```

- [x] Step 2.4: Update existing generation tests that assumed every generated monster was `dungeon_mob` so they assert rule-derived pool membership and guaranteed archer presence instead.
```bash
cd server && go test ./internal/game/... -run 'TestDungeon.*Monster|TestGenerated.*Monster|TestMonsterRarity' -count=1
```

## Task 3 - Server ranged monster combat

Files:
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`

- [x] Step 3.1: Add helpers for `monsterAttackMode`, `monsterAttackReach`, and ranged monster line of sight. Ranged reach should use `attack_range`; melee reach should remain `combat.unarmed_reach`.
```bash
cd server && go test ./internal/game/... -run 'TestMonster.*Attack|TestDungeonMonster' -count=1
```

- [x] Step 3.2: Update `monsterInAttackRange` and `findMonsterChaseGoal` so ranged monsters only consider themselves ready to fire when in range and line of sight is clear; otherwise they path toward a clear standoff candidate.
```bash
cd server && go test ./internal/game/... -run 'TestMonster.*Standoff|TestMonster.*Blocked|TestMonsterChase' -count=1
```

- [x] Step 3.3: Add monster projectile spawn from `advanceMonsterAttack`: snapshot scaled attack damage, set `owner_id` to the monster, `target_id` to the player, use the monster projectile speed/range/id, apply cooldown at fire time, and emit the existing projectile spawn change.
```bash
cd server && go test ./internal/game/... -run 'TestMonster.*Projectile|TestDungeonMonster' -count=1
```

- [x] Step 3.4: Extend projectile collision so player-owned projectiles keep hitting monsters exactly as before, while monster-owned projectiles can hit only their living target player on the same level.
```bash
cd server && go test ./internal/game/... -run 'TestRangedProjectile|TestMonster.*Projectile' -count=1
```

- [x] Step 3.5: Resolve monster projectile hits through existing monster combat stats and player defense stats, emitting `player_damaged`, `player_killed`, `attack_missed`, or `attack_blocked` and updating player HP through the same server-owned path as melee monster attacks.
```bash
cd server && go test ./internal/game/... -run 'TestMonster.*Projectile|TestCombatStatEffects' -count=1
```

- [x] Step 3.6: Add a focused blocked-shot test with a wall or closed barrier between archer and player; assert no player damage occurs through cover and the projectile blocks or the archer waits/moves instead.
```bash
cd server && go test ./internal/game/... -run 'TestMonster.*Blocked|TestRangedProjectile' -count=1
```

- [x] Step 3.7: Add regression tests proving existing `training_bow` player projectile behavior, `projectile_busy`, and ranged-lab golden cases still pass.
```bash
cd server && go test ./internal/game/... -run 'TestRangedProjectile|TestCombatControl' -count=1
```

## Task 4 - Protocol bot proof

Files:
- Modify: `tools/bot/run.py`
- Modify: `tools/bot/test_protocol.py`
- Add: `tools/bot/scenarios/38_ranged_monster_ai.json`

- [x] Step 4.1: Add protocol-bot assertion helpers only if current `entity_count`, `event_seen`, `combat_event_seen`, and player HP checks cannot distinguish archer-owned projectiles and ranged damage.
```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

- [x] Step 4.2: Add `38_ranged_monster_ai.json`: descend to a generated dungeon, assert live `dungeon_archer`, move within archer aggro/range without attacking, wait for archer projectile/combat proof, and assert player HP decreased.
```bash
make bot scenario=38_ranged_monster_ai.json
```

- [x] Step 4.3: Ensure existing dungeon monster, monster rarity, ranged projectile, combat stat, boss floor, and mystery seller scenarios remain green.
```bash
make bot
```

## Task 5 - Godot client bow marker and client bot proof

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/tests/test_client_bot.gd`
- Add: `tools/bot/scenarios/client/25_ranged_monster_ai.json`
- Modify if golden added: `client/tests/test_golden.gd`

- [x] Step 5.1: Add a small procedural bow marker to `dungeon_archer` monster nodes in `main.gd`, derived from `monster_def_id` or loaded monster rule attack mode. Do not add client-side AI, collision, damage, or projectile prediction.
```bash
make client-unit
```

- [x] Step 5.2: Expose `has_bow_marker` or equivalent in `entities_presentation_debug`, and teach `bot_scenario_runner.gd`/tests to assert it if current presentation matching cannot.
```bash
make client-unit
```

- [x] Step 5.3: Add client scenario `25_ranged_monster_ai.json` that descends, waits for `dungeon_archer`, asserts the bow marker through debug state, and observes the ranged player-damage path.
```bash
make bot-client scenario=25_ranged_monster_ai.json HEADLESS=1
```

- [x] Step 5.4: Run the full client bot scenario set to catch regressions in existing dungeon/combat UI flows.
```bash
make bot-client HEADLESS=1
```

## Task 6 - Lifecycle docs and final verification

Files:
- Modify: `docs/specs/v52_spec-ranged-monster-ai.md`
- Modify: `docs/plans/v52_2026-06-10-ranged-monster-ai.md`
- Modify: `docs/PROGRESS.md`

- [x] Step 6.1: Update the spec status to `Implemented` after all acceptance criteria pass.
```bash
rg -n 'Status:|ranged monster|ranged-monster-ai' docs/specs/v52_spec-ranged-monster-ai.md docs/PROGRESS.md
```

- [x] Step 6.2: Mark this plan's completed checkboxes and summarize key as-built decisions.
```bash
rg -n '\\[ \\]' docs/plans/v52_2026-06-10-ranged-monster-ai.md
```

- [x] Step 6.3: Update `docs/PROGRESS.md`: latest slice v52, lifecycle row, "What each slice proved", recently closed ranged monster AI gap, and deferred follow-ups for richer ranged AI/art/balance.
```bash
rg -n 'Latest completed slice|v52|ranged monster|Open gaps' docs/PROGRESS.md
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... -count=1`
- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make client-unit`
- [x] `make bot scenario=38_ranged_monster_ai.json`
- [x] `make bot-client scenario=25_ranged_monster_ai.json HEADLESS=1`
- [x] `make bot`
- [x] `make bot-client HEADLESS=1`
- [x] `make ci`

## Deferred scope

- Production archer/bow model, attack animation, VFX, SFX, and colorblind-safe monster silhouette.
- Ranged boss patterns, elite archer packs, retreat/cover seeking, predictive leading, AoE/homing/piercing, and monster skills.
- Final ranged monster damage/range/cooldown balance and deeper dungeon composition curves.
