# v369 Spec: Wayfarer's Accord Set

Status: Complete
Date: 2026-06-29
Codename: wayfarers-accord-set

## Purpose

Add a third enabled five-piece set item package whose pieces occupy only armor and jewelry slots so
every class can wear the full set while keeping its class weapon. Set bonuses use primary stats,
resource pools, and skill-wide modifiers that benefit barbarian, sorcerer, paladin, and ranger builds
equally.

## Non-goals

- New set mechanics, class-specific set rules, or weapon-slot set pieces.
- Production set art, new client UI surfaces, or collection-panel redesign.
- Mystery-seller rarity rebalance, boss drop rotation changes beyond adding elite drop entries, or
  economy tuning outside the new catalog rows.

## Acceptance Criteria

- `shared/rules/set_items.v0.json` defines an enabled `ready` five-piece `wayfarers_accord` set with
  unique piece ids, distinct non-weapon slots (head, chest, gloves, boots, amulet), and bonuses that
  include all four primary stats plus cross-class skill/resource value at full set.
- The debug unique chest and elite special drop pool include the new pieces using existing
  rule-derived catalog wiring.
- Focused Go tests cover payload identity, partial bonuses, full-set bonuses, and `all_skills`
  rank lift on at least one equipped skill.
- A protocol bot scenario opens the town unique chest, takes one Wayfarer's Accord piece by display
  name, and asserts it reaches inventory.
- `make validate-shared`, focused Go tests, the new bot scenario, and focused client unit tests pass.

## Scope And Likely Files

- Shared rules: `shared/rules/set_items.v0.json`, `shared/rules/treasure_classes.v0.json`.
- Server tests: `server/internal/game/unique_chest_test.go`.
- Bot proof: new extended scenario under `tools/bot/scenarios/`.
- Docs: spec, plan, as-built, lifecycle row, `PROGRESS.md`.

## Test And Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestWayfarersAccord|TestUniqueTestChest|TestSetItem' -count=1`
- `make bot scenario=84_wayfarers_accord_set.json`
- `make client-unit`

## Open Questions And Risks

- No blocking questions. Reuse existing set presentation and server bonus aggregation; reject new
  client plugins/assets for this data-only extension.
- Risk: tests that assumed exactly two enabled sets must remain rule-derived while asserting the new
  package explicitly.
