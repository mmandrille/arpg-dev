# v195 Spec: Boss And Elite Special Drops

Status: Complete - make ci green on 2026-06-15
Date: 2026-06-15
Codename: boss-and-elite-special-drops

## Purpose

Give boss and elite reward sources explicit authored item drops instead of only reusing ordinary depth treasure classes. Treasure classes should be able to reference ready named uniques and set pieces directly, allowing special rewards to be deterministic, data-driven, and testable.

## Non-goals

- Final boss/elite economy tuning, drop rates, pity systems, reservations, or personal loot.
- Random set rarity drops from normal monster templates.
- New client presentation beyond existing unique/set item payload rendering.
- New bosses, elite behaviors, or chest UI.

## Acceptance Criteria

- Treasure class schemas and server rule validation accept exactly one of `item_def_id`, `item_template_id`, `unique_item_id`, or `set_item_id`.
- `boss_drop_tier_1` resolves through a boss-specific treasure class that always drops one named unique and one set piece in addition to an existing equipment roll.
- The elite objective reward chest resolves through an elite-specific treasure class that always drops one set piece in addition to an existing chest-style equipment reward.
- Spawned authored drops carry the same fixed unique/set payloads used by the unique chest, including display name, rarity, stats, requirements, and effect ids.
- Focused Go tests prove treasure-class resolution, spawned payload identity, and unchanged normal monster drop-rate math.
- A protocol bot scenario kills the Cave Warden and asserts the named unique/set drops spawn.

## Scope And Likely Files

- Shared rules/schemas: `shared/rules/treasure_classes.v0.schema.json`, `shared/rules/loot_tables.v0.json`, `shared/rules/treasure_classes.v0.json`, `shared/rules/dungeon_generation.v0.json` if the elite objective table id changes.
- Server: `server/internal/game/rules.go`, `server/internal/game/sim.go`, focused loot tests.
- Bot: new protocol scenario under `tools/bot/scenarios/`.
- Docs: spec, plan, as-built, `PROGRESS.md`.

## Test And Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TreasureClass|SpecialDrop|Boss|EliteObjective' -count=1`
- `make bot scenario=84_boss_special_drops.json`
- `make ci`

## Open Questions And Risks

- No blocking questions. Use existing named unique `Conduit Staff` and set pieces for the first special rewards.
- Risk: bot assertions may need an existing loot entity display-name assertion; if absent, keep bot proof to kill boss and assert dropped entity payload fields via existing assertion support or add a small generic assertion helper.
