# v188 As-built: Live Rare Combat Affixes

Date: 2026-06-15

## What shipped

- Added shared item-template support for rare `hit_chance`, `crit_chance`, and `evade_chance`
  percent-point rolls, including schema, validator, golden schema, and shop pricing coverage.
- Added rare roll candidates to `cave_blade` and `cave_gloves` while preserving existing
  `attack_speed_percent` behavior.
- Aggregated equipped hit, crit, and evade rolls into effective player combat stats and exposed
  them through derived stats and stat breakdowns.
- Added defender-side evade resolution after attacker hit and before block, using the existing miss
  outcome/event path.
- Moved derived-stat protocol view code out of `sim.go` into `derived_stats.go` to keep the
  maintainability ratchet green while adding `evade_chance`.
- Added client stat labels for `evade_chance` so the existing character stats panel can display it.

## Proof

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game/... -run 'RareCombatAffix|CombatStat|AttackSpeed'`
- `make bot scenario=79_live_rare_combat_affixes.json`
- `make ci`

## Follow-up

- Affix naming, crafting, and wider combat-affix distribution remain deferred.
- Skill cooldown and mana-cost affixes are intentionally left for v189.
