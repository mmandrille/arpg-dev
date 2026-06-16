# v210 As-Built: Barbarian Leap

Date: 2026-06-16

## What Shipped

- Added `leap` as a Barbarian mobility skill.
- Leap moves to a server-authoritative collision-safe endpoint, damages monsters near landing, and applies `leap_stun`.
- Added shared presentation/i18n metadata and Barbarian class-foundation bot coverage.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestBarbarianLeap|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- `make bot scenario=barbarian_class_foundation`
- `make maintainability`
- `make ci`

## Scope Limits

No airborne physics, invulnerability, wall jumping, production VFX/audio, Paladin Charge, or Ranger Disengage shipped here.
