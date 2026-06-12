# v85 As-built - Skill Demo Catalog

## What shipped

`tools/bot/skill_demo.py` now derives skill demo metadata from the shared content manifest, skill
rules, and skill presentation catalog. It can print one skill or all skills as JSON and reports
display metadata, class, kind, demo category, max rank, and rank targets.

The maintainability ratchet was also restored by moving character-stat Go tests out of the
oversized `server/internal/game/game_test.go` into `server/internal/game/character_stats_test.go`.

## What it proves

- Skill demo tooling can resolve current shared skills without hardcoded presentation metadata.
- Magic Bolt, Rage, Heal, and Holy Shield classify into attack, self-buff, heal, and stat-buff demo
  categories.
- Unknown skill ids fail clearly before a later visual runner tries to launch.
- `server/internal/game/game_test.go` is back within its grandfathered file-size allowance.

## Verification

```bash
.venv/bin/pytest tools/bot/test_skill_demo.py -q
make test-py
cd server && go test ./internal/game/... -run 'TestEffectiveAttackSpeedUsesWeaponAndItemPercent|TestHealthAndManaRegenUseStatsAndItemRolls'
make maintainability
make ci
```

## Deferred

- `make skill-visual skill=...`.
- Generated visual replay flows for rank 1 and rank 5.
- Character-page stat delta proof for buff skills.
- Full skill presentation coverage matrix.
