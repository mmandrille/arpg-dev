# v286 As Built: Archer Retreat AI

Date: 2026-06-19
Spec: [`docs/specs/v286_spec-archer-retreat-ai.md`](../specs/v286_spec-archer-retreat-ai.md)
Plan: [`docs/plans/v286_2026-06-19-archer-retreat-ai.md`](../plans/v286_2026-06-19-archer-retreat-ai.md)

## What shipped

- Added ranged-only `preferred_min_range` to monster rules and schema.
- Set `dungeon_archer.preferred_min_range` to `3.5`, keeping existing attack range, cooldown,
  projectile speed, damage, hit chance, spawn rules, and bow presentation unchanged.
- Added deterministic ranged retreat-goal selection in `monster_ranged_positioning.go`.
- Ranged monsters now reuse cached retreat goals while those goals remain valid, so backpedal movement
  continues between path-replan ticks.
- Added `ranged_monster_retreat_lab`, a compact close-start archer proof world.
- Added Go coverage proving the archer retreats out toward preferred range while the prior blocked-shot
  reposition test remains green.
- Added protocol bot scenario `archer_retreat_ai`.

## Proof

Focused verification:

```bash
(cd server && go test ./internal/game -run 'TestRangedMonster' -count=1)
make validate-shared
make bot scenario=archer_retreat_ai
make maintainability
```

Result: green on 2026-06-19.

Full verification:

```bash
make ci
```

Result: deferred until the end of the selected autoloop queue.

## Manual visual command

```bash
make bot-visual scenario=25_ranged_monster_ai
```

## Deferred

- Cover seeking, strafing, predictive leading, elite archer packs, and ranged damage/range/cooldown
  rebalance remain deferred.
- Client animation/VFX polish for backpedaling remains deferred.
