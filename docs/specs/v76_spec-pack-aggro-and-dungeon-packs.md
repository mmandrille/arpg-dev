# v76 Spec — Pack Aggro and Dungeon Packs

Status: Complete
Date: 2026-06-11
Codename: `pack-aggro-and-dungeon-packs`

## Purpose

Dungeon floors should create fewer, more coherent fights. Instead of isolated random monster singles, normal generated dungeon floors spawn 5 to 7 monster packs, each with 2 to 5 monsters. When the hero damages any member, nearby pack members join through a server-owned `assist_radius` so the fight reads as one encounter.

## Non-goals

- Elite/champion pack leaders, aura modifiers, or named pack identities beyond the existing rarity system.
- New monster abilities, new monster definitions, or new client VFX.
- Protocol/schema changes for exposing pack IDs to the client.
- Boss-floor trash rewrites.
- Final combat balance for pack damage, density, or loot economy.

## Acceptance Criteria

- Shared dungeon generation rules define pack count, pack size, spawn spread, and assist radius as data.
- Regular generated dungeon floors spawn between 5 and 7 packs.
- Every generated pack has 2 to 5 monsters.
- Pack members are positioned close enough that damaging one member can activate the rest through `assist_radius`.
- `assist_radius` is separate from passive `aggro_radius`; passive detection does not need to grow just to make pack fights coherent.
- Server aggro-on-hit uses assist radius for nearby combat joins, while passive chase still uses aggro radius.
- Generated floors remain reachable with stairs, teleporters, chests, monsters, and obstacles valid.
- Golden/unit coverage proves deterministic pack counts, member counts, spacing, and assist activation.
- Protocol bot coverage proves that pulling one monster from a generated pack emits multiple `monster_aggro` events after damage starts.

## Scope and Likely Files

- `shared/rules/dungeon_generation.v0.json`
- `shared/rules/dungeon_generation.v0.schema.json`
- `shared/rules/monsters.v0.json`
- `shared/rules/monsters.v0.schema.json`
- `server/internal/game/rules.go`
- `server/internal/game/dungeon_gen.go`
- `server/internal/game/sim.go`
- `server/internal/game/game_test.go`
- `tools/bot/scenarios/42_pack_aggro_and_dungeon_packs.json`
- `tools/bot/run.py` if a small assertion helper is needed
- `docs/plans/v76_2026-06-11-pack-aggro-and-dungeon-packs.md`
- `docs/as-built/v76_pack-aggro-and-dungeon-packs.md`
- `PROGRESS.md`

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game/...`
- `godot --headless --path client --script res://tests/test_golden.gd` if shared golden fixtures change.
- `make bot` with a new pack aggro scenario.
- `make ci` before finish.

## Open Questions and Risks

- No blocking questions. Default tuning for v76: `pack_count 5..7`, `pack_size 2..5`, `assist_radius 8.0`, and spawn spread small enough to guarantee pack-wide assist while preserving reachability.
- Risk: generated obstacles may make tight packs harder to place. The implementation should retry pack centers/members deterministically and fail generation only after configured attempts.
- Risk: increasing floor monster count changes difficulty. This slice proves encounter shape; final balance stays deferred.
