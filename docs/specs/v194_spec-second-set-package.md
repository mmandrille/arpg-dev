# v194 Spec: Second Set Package

Status: Complete - make ci green on 2026-06-15
Date: 2026-06-15
Codename: second-set-package

## Purpose

Add a second enabled five-piece set item package so the set system is no longer a one-off Verdant Vanguard proof. The new package should use existing item templates, appear in the debug unique chest, and grant deterministic partial/full set bonuses through the existing server-authoritative set bonus path.

## Non-goals

- Random set drops, boss-specific set rewards, set economy pricing, or production collection UI.
- New set mechanics beyond existing fixed pieces plus 2/3/4/full-set stat bonuses.
- New Godot plugins, UI surfaces, art assets, or visual rules; existing set rarity presentation remains the client path.

## Acceptance Criteria

- `shared/rules/set_items.v0.json` defines a second enabled `ready` five-piece set with unique piece ids, unique display names, five distinct equipment slots, and existing base templates.
- The debug unique chest offers both enabled set packages, with deterministic ordering and item counts derived from enabled rules rather than hard-coded to one set.
- Focused Go tests cover the new set payload, at least one partial bonus, the full-set bonus, and the existing Verdant Vanguard behavior.
- A protocol bot scenario opens the town unique chest, takes one new set item by display name, and asserts it reaches inventory.
- `make validate-shared`, focused Go tests, the new bot scenario, and `make ci` pass.

## Scope And Likely Files

- Shared rules: `shared/rules/set_items.v0.json`.
- Server tests: `server/internal/game/unique_chest_test.go`.
- Bot proof: new `tools/bot/scenarios/83_second_set_package.json`.
- Docs: spec, plan, as-built, and `PROGRESS.md`.

## Test And Bot Proof

- `make validate-shared` proves the second set follows the shared schema and server rule validation.
- `cd server && go test ./internal/game -run 'TestUniqueTestChest|TestSetItem' -count=1` proves chest inclusion/order and set bonus behavior.
- `make bot scenario=83_second_set_package.json` proves the new package is reachable through the player-facing debug chest workflow.
- `make ci` remains the final gate.

## Open Questions And Risks

- No blocking questions. Use existing item templates and existing set presentation; reject new client plugin/assets for this data-only extension.
- Risk: tests that assumed exactly five enabled set items must become rule-derived while still asserting the original set and the new set explicitly.
