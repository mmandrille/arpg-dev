# v79 Spec — Elite Pack Roles

Status: Complete
Date: 2026-06-11
Codename: `elite-pack-roles`

## Purpose

Generated dungeon packs should have a small amount of encounter structure beyond raw clustering. This slice adds a server-owned foundation for pack leaders and monster role composition so future elite modifiers and richer pack behavior have clean data hooks.

## Non-goals

- New monster abilities, aura effects, affixes, or named elite UI.
- Client-side threat/readability presentation.
- Protocol exposure of pack IDs, leader IDs, or role labels.
- Final balance for elite frequency or pack difficulty.

## Acceptance Criteria

- Shared monster rules identify generated dungeon monster roles.
- Dungeon generation rules define elite pack chance and composition constraints.
- Generated non-boss packs choose deterministic role-aware members instead of purely random definitions.
- Each generated pack has at least one frontline monster and no more than the configured ranged count.
- Elite pack foundation marks at most one planned pack member as leader and biases that member toward champion rarity without creating new abilities.
- Existing champion follower behavior remains best-effort and separate from planned pack size.
- Server tests prove deterministic roles, leader bounds, and composition constraints.
- Protocol bot still proves generated pack aggro after the role/leader changes.

## Likely Files

- `shared/rules/monsters.v0.json`
- `shared/rules/monsters.v0.schema.json`
- `shared/rules/dungeon_generation.v0.json`
- `shared/rules/dungeon_generation.v0.schema.json`
- `server/internal/game/rules.go`
- `server/internal/game/dungeon_gen.go`
- `server/internal/game/game_test.go`
- `tools/bot/scenarios/42_pack_aggro_and_dungeon_packs.json`
- `docs/plans/v79_2026-06-11-elite-pack-roles.md`
- `docs/as-built/v79_elite-pack-roles.md`
- `PROGRESS.md`

## Verification

- `make validate-shared`
- `cd server && go test ./internal/game/...`
- `ARPG_ADDR=:8888 SCENARIO=pack_aggro_and_dungeon_packs ./scripts/bot_local.sh`
- `make ci`
