# v209 As-Built: Sorcerer Teleport

Date: 2026-06-16

## What Shipped

- Added `teleport` as a Sorcerer `mobility` skill in shared rules with data-driven range, mana cost, cooldown, and placeholder visual metadata.
- Extended the shared skill schema and Go rules loader with a typed `mobility` payload.
- Added a focused server mobility execution path that reuses server-authoritative endpoint collision checks, mana spend, cooldowns, and `skill_cast` events.
- Taught skill visual/demo tooling to categorize and cast mobility skills without requiring damage events.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestSorcererTeleport|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_skill_visual.py -q`
- `make maintainability`
- `make ci`

## Scope Limits

- Teleport intentionally has no damage, stun, root, invulnerability, production VFX, or audio.
- Barbarian Leap, Paladin Charge, and Ranger Disengage are deferred to the remaining autoloop slices.
