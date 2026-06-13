# v123 As-Built - Ranger piercing and pinning shots

Date: 2026-06-13
Status: Complete

## What Shipped

- Added schema-backed Ranger skill payloads for `pierce` and `root`.
- Added `piercing_shot` as a physical bow projectile that hits multiple line targets with
  data-owned max hits and falloff.
- Added `pinning_shot` as a physical bow projectile that roots the first damaged target using
  server-owned effect state.
- Prevented rooted monsters from moving while keeping damage, death, and effect expiry normal.
- Added green Ranger skill icons, projectile variants, and a pinning-root monster marker in Godot.
- Added protocol scenario `59_ranger_piercing_and_pinning_shots`, proving root start/expiry and a
  multi-target Piercing Shot.
- Tightened the protocol bot `cast_skill` event wait so skill-specific casts cannot pass by seeing
  an earlier event with the same event type.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestRanger|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=59_ranger_piercing_and_pinning_shots`
- `make client-unit`

## Scope Limits

- `Volley` and the full Ranger visual showcase scenario are deferred to v124.
- The new projectiles and root marker are code-native placeholder VFX, consistent with the current
  client presentation style.
- Balance values are intentionally conservative and data-owned in `shared/rules/skills.v0.json`.

