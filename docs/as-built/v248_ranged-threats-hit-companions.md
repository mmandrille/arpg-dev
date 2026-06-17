# v248 As-Built - Ranged Threats Hit Companions

Date: 2026-06-17

## What shipped

- Reused the existing engaged-companion targeting rule for ranged monsters, so an archer can select
  a companion that is actively engaging it.
- Routed monster projectile creation through the selected attack target instead of always targeting
  the player.
- Added targeted companion collision handling for monster-owned projectiles, resolving hits through
  the existing `damageCompanionByMonster` path.
- Added a focused ranged-companion combat lab using a non-attacking shared companion fixture, which
  keeps the proof stable without pinning balance values.
- Added `91_ranged_threats_hit_companions.json` as the protocol bot proof for archer-sourced
  companion combat events.

## Proof

```bash
cd server && go test ./internal/game -run 'RangedMonster.*Companion|RangedMonsterProjectile' -count=1
make validate-shared
make bot scenario=91_ranged_threats_hit_companions.json
make maintainability
```

All focused checks passed on 2026-06-17 during `$autoloop`. The selected v241-v250 batch-level
`make ci` also passed on 2026-06-17 after v250.

## Scope limits

- No companion threat table, taunt system, ranged AI retarget policy beyond existing engaged
  companions, projectile avoidance, new monster types, damage tuning, protocol schema changes, or
  client VFX shipped.
