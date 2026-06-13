# v125 As-Built - Tuning Friendly Bot Scenarios

Date: 2026-06-13
Spec: [`docs/specs/v125_spec-tuning-friendly-bot-scenarios.md`](../specs/v125_spec-tuning-friendly-bot-scenarios.md)
Plan: [`docs/plans/v125_2026-06-13-tuning-friendly-bot-scenarios.md`](../plans/v125_2026-06-13-tuning-friendly-bot-scenarios.md)

## What Shipped

- Added `max_rank: "from_rules"` support to bot skill progression assertions.
- The resolver loads `shared/rules/skills.v0.json` once and fails clearly when a scenario references
  an unknown skill rule.
- Migrated the Magic Bolt, Rage/Heal, and Ranger Volley scenario progression checks away from
  hardcoded `max_rank: 5` assertions.

## Proof

- `.venv/bin/pytest tools/bot/test_protocol.py -q`
- `make bot scenario=32_skill_points_and_magic_bolt`
