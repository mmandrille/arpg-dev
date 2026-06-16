# v205 As-Built: Boss Enrage Phase

Date: 2026-06-15
Status: Complete

## What shipped

- Added Cave Warden enrage data to `shared/rules/boss_templates.v0.json` with a 50% health threshold and 0.5 future-cooldown multiplier.
- Added schema and Go validation for enrage thresholds `(0,1]` and positive cooldown multipliers.
- Added server-owned boss enrage state. Boss entity views now expose `enraged` and `enrage_health_ratio_threshold`.
- Added a one-shot `boss_enraged` event carrying boss id, target id, boss template id, and health-ratio threshold.
- Applied the enrage cooldown multiplier only when scheduling future boss pattern cooldowns, with a minimum 1-tick cooldown for positive base cooldowns.
- Added `tools/bot/scenarios/87_boss_enrage_phase.json` as the focused protocol proof.

## Proof

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestBossEnrage|TestBossSummonedAdds'`
- `make bot scenario=boss_enrage_phase`
- `make bot scenario=boss_floor_gate`

- `make ci`

## Notes

- No client UI or runner changes were needed. Existing protocol bot event matching can assert `boss_enraged` with `boss_template_id`.
- Enrage does not increase damage or add an instant attack. Boss telegraph timing remains unchanged; only future pattern cooldown scheduling changes.
