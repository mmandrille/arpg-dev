# v124 As-Built - Ranger volley and visual scenario

Date: 2026-06-13
Status: Complete

## What Shipped

- Added schema-backed `volley` payload data with arrow count and spread angle.
- Added Ranger `Volley` as a physical fan projectile skill with data-owned cost, cooldown,
  requirements, damage, projectile visual, and presentation text.
- Implemented server-authoritative Volley fan resolution in the Ranger skill helper: each arrow ray
  can hit one monster, duplicate target damage is prevented within a cast, and damage events carry
  `skill_id: volley`.
- Added a green Volley projectile cue and multi-arrow skill icon.
- Added `ranger_showcase_lab` and protocol scenario `60_ranger_volley_and_visual_showcase`, proving
  Ranger starter bow, Pinning Shot root start, Piercing Shot multi-hit, and Volley multi-target
  damage in a visual-friendly replay.
- Updated Ranger class foundation coverage to reference all three Ranger skills.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestRanger|TestLoadRules'`
- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=60_ranger_volley_and_visual_showcase`
- `make client-unit`
- `make maintainability`

## Visual Verification

Run the focused replay with:

```bash
make bot-visual scenario=60_ranger_volley_and_visual_showcase
```

## Scope Limits

- Root-expiry proof remains in v123 scenario `59_ranger_piercing_and_pinning_shots`; the v124
  showcase is shortened to stay under the protocol bot scenario budget.
- Volley uses code-native placeholder VFX consistent with existing skill presentation.

