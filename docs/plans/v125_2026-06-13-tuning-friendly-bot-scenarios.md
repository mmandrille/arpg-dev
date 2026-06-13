# v125 Plan - Tuning Friendly Bot Scenarios

Status: Complete
Goal: Let bot skill progression assertions derive rule-owned skill max ranks from shared data.
Architecture: Keep scenario JSON declarative; teach the Python bot assertion layer to resolve a
small sentinel value against `shared/rules/skills.v0.json`.
Tech stack: Python protocol bot, JSON scenarios, SDD docs.

## Baseline And Shortcut Decision

Baseline is v124 `ranger-volley-and-visual-scenario` on `main`, committed before this loop.

Godot plugin adoption checklist: not applicable. This slice touches protocol bot tooling and JSON
scenario assertions only.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `tools/bot/run.py` | Resolve rule-derived skill max-rank expectations. |
| Modify | `tools/bot/test_protocol.py` | Cover runtime/state assertion paths and failure mode. |
| Modify | `tools/bot/scenarios/32_skill_points_and_magic_bolt.json` | Remove hardcoded Magic Bolt max-rank tuning locks. |
| Modify | `tools/bot/scenarios/39_rage_and_heal_skills.json` | Remove hardcoded Rage/Heal max-rank tuning locks. |
| Modify | `tools/bot/scenarios/60_ranger_volley_and_visual_showcase.json` | Remove hardcoded Volley max-rank tuning lock. |
| Add | `docs/as-built/v125_tuning-friendly-bot-scenarios.md` | Record implementation and proof. |
| Modify | `PROGRESS.md` | Lifecycle closeout. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files likely touched:
- [x] `tools/bot/run.py`
- [x] `tools/bot/test_protocol.py`

Decision:
- [x] Keep the resolver tiny and local to bot assertions; defer broader scenario/assertion module
  extraction to a later tools split.

## Task 1 - Rule Resolver

- [x] Step 1.1: Add cached loading for `shared/rules/skills.v0.json`.
- [x] Step 1.2: Add `max_rank: "from_rules"` support for skill progression assertions.
- [x] Step 1.3: Make missing skill rows fail clearly.

## Task 2 - Scenario Migration

- [x] Step 2.1: Migrate Magic Bolt scenario skill progression checks.
- [x] Step 2.2: Migrate Rage/Heal scenario skill progression checks.
- [x] Step 2.3: Migrate the Ranger Volley showcase progression check.

## Task 3 - Verification And Closeout

- [x] Step 3.1: Add focused protocol tests for runtime/state assertion paths.
- [x] Step 3.2: Run focused pytest and the affected bot scenario where local services allow it.
- [x] Step 3.3: Mark docs complete, update `PROGRESS.md`, and commit.

## Final Verification

- [x] `.venv/bin/pytest tools/bot/test_protocol.py -q`
- [x] `make bot scenario=32_skill_points_and_magic_bolt`
