# v210 Spec: Barbarian Leap

Status: Complete
Date: 2026-06-16
Codename: barbarian-leap

## Purpose

Give Barbarians a class escape skill, `leap`, that relocates in an aimed direction and smashes the landing area for damage plus a short stun/root. This complements existing stationary Earthbreaker with true mobility for escaping monster surrounds.

## Baseline

Builds on v209 `mobility` skill support. Asset/plugin decision: borrow existing code-native skill presentation and visual-demo paths; reject external assets/plugins and production VFX.

## Non-goals

- No Sorcerer, Paladin, Ranger, or Rogue changes.
- No airborne physics, invulnerability, or pathfinding over walls.
- No production VFX/audio.

## Acceptance Criteria

- `leap` is a Barbarian mobility skill with data-driven range, impact radius, damage percent, stun duration, mana cost, cooldown, and presentation metadata.
- Casting Leap moves the player to a collision-safe endpoint.
- Monsters near the landing point take server-authoritative damage and receive the shared `stun` effect.
- Barbarian class-foundation bot coverage includes Leap.
- Focused Go tests prove movement, damage, stun, and unaffected distant monsters.

## Scope and Likely Files

- Shared rules/assets/i18n: Leap catalog entries.
- Server tests: focused mobility test.
- Bot scenario: Barbarian class foundation.
- Lifecycle docs.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestBarbarianLeap|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- `make bot scenario=barbarian_class_foundation`
- `make ci`

Visual manual check: `make bot-visual scenario=skill_visual ARPG_SKILL_VISUAL_SKILL_ID=leap`.

## Open Questions and Risks

- No blocking questions.
