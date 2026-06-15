# v191 As-built: Affix Name Grammar

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Added deterministic affix-style display names for magic-or-higher non-unique, non-set rolled
  item payloads.
- The first grammar chooses one readable prefix from the highest-priority rolled stat family, such
  as `Focused Rare Sorcerer Staff` for skill cooldown or mana-cost rolls.
- Kept common item names, authored unique names, and authored set-piece names unchanged.
- Kept the protocol unchanged by continuing to persist and render names through
  `ItemRollPayload.display_name`.
- Added the naming helper in `server/internal/game/affix_names.go` so `shop.go` remains within its
  maintainability ratchet allowance.
- Updated rolled item and generated shop goldens plus the shared validator's deterministic shop
  roll mirror for the intended display-name drift.
- Extended protocol bot scenario `80_skill_affix_rolls.json` to assert the generated display name
  alongside its stat-key proof.

## Verification

- `cd server && go test ./internal/game -run 'TestAffixName|TestSkillAffix' -count=1`
- `make bot scenario=80_skill_affix_rolls.json`
- `make maintainability`
- `make validate-shared`
- `make ci`

## Deferred

- Multi-affix prefix/suffix grammar.
- Localized generated affix text.
- Procedural names for authored unique and set items.
