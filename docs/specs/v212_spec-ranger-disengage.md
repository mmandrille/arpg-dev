# v212 Spec: Ranger Disengage

Status: Complete
Date: 2026-06-16
Codename: ranger-disengage

## Purpose

Give Rangers a class escape skill, `disengage`, that vaults away from danger and leaves a short snare at the starting position. This gives the ranged class a way to escape monster surrounds while reinforcing kiting and spacing instead of adding another direct damage shot.

## Baseline

Builds on v209 mobility support plus v210/v211 class mobility examples. Asset/plugin decision: borrow existing code-native skill presentation and visual-demo paths; reject external assets/plugins and production VFX.

## Non-goals

- No Sorcerer, Barbarian, Paladin, or Rogue changes.
- No trap inventory, deployable persistence, stealth, or invulnerability.
- No production VFX/audio.

## Acceptance Criteria

- `disengage` is a Ranger mobility skill with data-driven range, start-area snare radius/duration, mana cost, cooldown, and presentation metadata.
- Casting Disengage moves the Ranger to a collision-safe endpoint.
- Monsters near the starting position receive `disengage_snare` without taking damage.
- Monsters only near the endpoint remain unaffected.
- Ranger class-foundation bot coverage includes Disengage.
- Focused Go tests prove movement, start-area snare, no damage, and endpoint unaffected behavior.

## Scope and Likely Files

- Shared rules/assets/i18n: Disengage catalog entries.
- Server mobility: start-origin impact for `mode: disengage`.
- Server tests: focused mobility test.
- Bot scenario: Ranger class foundation.
- Lifecycle docs.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestRangerDisengage|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- `make bot scenario=ranger_class_foundation`
- `make ci`

Visual manual check: `make bot-visual scenario=skill_visual ARPG_SKILL_VISUAL_SKILL_ID=disengage`.

## Open Questions and Risks

- No blocking questions.
