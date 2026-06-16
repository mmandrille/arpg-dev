# v212 As-Built: Ranger Disengage

Date: 2026-06-16

## What Shipped

- Added `disengage` as a Ranger mobility skill.
- Disengage moves to a server-authoritative collision-safe endpoint and applies `disengage_snare` to monsters near the starting position without damage.
- Endpoint-only monsters remain unaffected.
- Added shared presentation/i18n metadata and Ranger class-foundation bot coverage.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestRangerDisengage|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- `make bot scenario=ranger_class_foundation`
- `make maintainability`
- `make ci`

## Scope Limits

No trap inventory, deployable persistence, stealth, invulnerability, production VFX/audio, or additional class mobility changes shipped here.
