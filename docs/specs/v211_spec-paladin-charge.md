# v211 Spec: Paladin Charge

Status: Complete
Date: 2026-06-16
Codename: paladin-charge

## Purpose

Give Paladins a class escape skill, `charge`, that surges in an aimed direction and shield-smashes enemies near the endpoint for damage plus a short stun. This gives the shield class a decisive reposition tool when surrounded without changing its support identity.

## Baseline

Builds on v209 `mobility` skill support and v210 Leap proof. Asset/plugin decision: borrow existing code-native skill presentation and visual-demo paths; reject external assets/plugins and production VFX.

## Non-goals

- No Sorcerer, Barbarian, Ranger, or Rogue changes.
- No shield-equipment requirement or wall-breaking collision behavior.
- No production VFX/audio.

## Acceptance Criteria

- `charge` is a Paladin mobility skill with data-driven range, impact radius, damage percent, stun duration, mana cost, cooldown, and presentation metadata.
- Casting Charge moves the player to a collision-safe endpoint.
- Monsters near the endpoint take server-authoritative damage and receive `charge_stun`.
- Paladin class-foundation bot coverage includes Charge.
- Focused Go tests prove movement, damage, stun, and unaffected distant monsters.

## Scope and Likely Files

- Shared rules/assets/i18n: Charge catalog entries.
- Server tests: focused mobility test.
- Bot scenario: Paladin class foundation.
- Lifecycle docs.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestPaladinCharge|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- `make bot scenario=paladin_class_foundation`
- `make ci`

Visual manual check: `make bot-visual scenario=skill_visual ARPG_SKILL_VISUAL_SKILL_ID=charge`.

## Open Questions and Risks

- No blocking questions.
