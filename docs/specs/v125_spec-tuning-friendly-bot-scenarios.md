# v125 Spec - Tuning Friendly Bot Scenarios

Status: Complete
Date: 2026-06-13
Codename: `tuning-friendly-bot-scenarios`

## Purpose

Reduce bot scenario brittleness by letting skill progression assertions reference shared rule data
for rule-owned tuning values. Scenario proofs should continue to verify behavior while avoiding
hardcoded skill caps that must be edited whenever balance data changes.

## Non-goals

- No gameplay balance changes.
- No protocol schema changes.
- No broad scenario rewrite.
- No removal of exact assertions that intentionally own API contracts or deterministic event shape.

## Acceptance Criteria

1. Bot skill progression assertions support a rule-derived `max_rank` expectation.
2. At least the main skill progression scenarios use rule-derived `max_rank` assertions instead of
   hardcoded skill cap numbers.
3. Missing or malformed skill rule references fail with clear bot assertion errors.
4. Existing runtime and state assertion paths both support the new expectation.
5. Focused protocol bot tests pass.

## Likely Files

- `tools/bot/run.py`
- `tools/bot/test_protocol.py`
- `tools/bot/scenarios/32_skill_points_and_magic_bolt.json`
- `tools/bot/scenarios/39_rage_and_heal_skills.json`
- `tools/bot/scenarios/60_ranger_volley_and_visual_showcase.json`
- `docs/as-built/v125_tuning-friendly-bot-scenarios.md`
- `PROGRESS.md`

## Test And Bot Proof

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
make bot scenario=32_skill_points_and_magic_bolt
```
