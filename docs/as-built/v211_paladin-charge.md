# v211 As-Built: Paladin Charge

Date: 2026-06-16

## What Shipped

- Added `charge` as a Paladin mobility skill.
- Charge moves to a server-authoritative collision-safe endpoint, damages monsters near the shield-smash endpoint, and applies the shared `stun` effect.
- Added shared presentation/i18n metadata and Paladin class-foundation bot coverage.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestPaladinCharge|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- `make bot scenario=paladin_class_foundation`
- `make maintainability`
- `make ci`

## Scope Limits

No shield-equipment requirement, wall-breaking collision, production VFX/audio, or Ranger Disengage shipped here.
