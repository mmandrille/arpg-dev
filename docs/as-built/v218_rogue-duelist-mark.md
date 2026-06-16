# v218 As-Built - Rogue Duelist Mark

Date: 2026-06-16

## Shipped

- Extended `poison_stab` with a data-driven Rogue mark that increases all player-owned damage against the marked monster while active.
- Applied the mark bonus to direct weapon hits, active skill damage, and poison tick damage.
- Added `executioner`, a Rogue passive skill that rolls a configurable execute on every damaging hit when the target is below a rank-scaled health threshold.
- Added data-driven Dash stun fields and applied root/stun control to live Dash targets.
- Added a procedural red skull marker above living monsters carrying the `rogue_mark` effect id.
- Added DEX as a standard crit-damage derived-stat contributor for every class through `character_progression.v0.json`.
- Updated Rogue catalog presentation, i18n, skill panel fixtures, bot skill demo metadata, and the Rogue class-foundation scenario.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestRogue|TestLoadRules|TestCritDamageUsesDexterityAsStandardDerivedStat|TestDerivedStats|TestEffectiveAttackSpeedUsesWeaponAndItemPercent' -count=1`
- `.venv/bin/pytest tools/bot/test_skill_demo.py tools/bot/test_protocol.py::test_load_scenarios_discovers_rogue_class_foundation tools/bot/test_protocol.py::test_class_foundation_scenarios_cover_every_class_skill -q`
- `make bot scenario=rogue_class_foundation`
- `godot --headless --path client --script res://tests/test_status_effect_presentation.gd`
- `make client-unit`

## Visual Verification

- `make bot-visual scenario=rogue_class_foundation`

## Notes

- `executioner` is passive-only and rejects direct cast attempts with `passive_skill_not_castable`.
- Execute tuning is data-owned: threshold starts at 10%, gains 5% per rank, and currently rolls 35% chance on qualifying hits.
- Poison mark tuning is data-owned: 25% bonus, 50 ticks, effect id `rogue_mark`.
