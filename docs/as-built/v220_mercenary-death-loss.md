# v220 As-Built - Mercenary Death Loss

Date: 2026-06-16

## What shipped

- Monsters can now select an engaged owned companion as their melee attack target when that
  companion is already attacking them and is in range.
- Monster damage against companions uses the existing monster combat stat path and emits
  `companion_damaged`, `companion_killed`, or `attack_missed` events.
- A hired `mercenary_guard` killed by monster damage is removed from active entity state and emits
  `mercenary_lost` with the lost companion id, killer source id, `service=mercenary`,
  `offer_id=fixed:mercenary_guard`, and `monster_def_id=mercenary_guard`.
- Players can rehire through the existing mercenary board after a loss; gold spend remains the
  existing v206 `mercenary_hired` path.

## Proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'TestMercenary|TestMonsterDamagesCompanion'
make bot scenario=mercenary_death_loss
make maintainability
```

All focused checks passed on 2026-06-16 during `$autoloop`. The enclosing batch-level `make ci`
also passed after the selected v219-v223 feature queue completed.

Manual visual proof, if desired:

```bash
make bot-visual scenario=mercenary_death_loss
```

## Scope limits

- Loss is active-session entity loss only; no durable mercenary roster, recovery timer, refund,
  revive, or insurance system shipped.
- Ranged monster attacks still target players only. This slice adds melee companion interception for
  an engaged companion, not a broad monster threat/aggro rewrite.
- Client UI changes were unnecessary because v219 already renders companion roster state from
  authoritative entities.
