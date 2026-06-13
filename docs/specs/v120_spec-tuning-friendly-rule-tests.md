# v120 Spec: Tuning-Friendly Rule Tests

Status: Complete
Date: 2026-06-13
Codename: `tuning-friendly-rule-tests`

## Purpose

Remove a focused test tuning lock so gameplay balance can change in shared data without unrelated
UI tests failing. The first target is the Godot skills panel test, which duplicated Magic Bolt stat
requirements, mana cost, and max-rank values that are already owned by `shared/rules/skills.v0.json`.

## Non-goals

- No gameplay tuning changes.
- No production UI changes.
- No broad rewrite of all historical exact-value tests.
- No changes to protocol/schema/golden tests where exact values are the subject under test.

## Acceptance criteria

- `client/tests/test_skills_panel.gd` derives Magic Bolt requirement, mana-cost, and max-rank
  expectations from `SkillRulesLoader` instead of hardcoding current tuning values.
- The test still proves visible UI behavior: requirements met/blocked, tooltip requirement text,
  missing-stat diff text, max rank propagation, and spend-button state.
- Focused Godot unit coverage passes.
- `make maintainability` and `make ci` pass before commit.

## Verification

- `make client-unit`
- `make maintainability`
- `make ci`

## Review gate

v120 is the next engineering-review milestone. After the slice passes CI, write a fresh review set
under `docs/reviews/` and update `PROGRESS.md` review pointers before finishing.
