# v85 Spec - Skill Demo Catalog

Status: Draft
Date: 2026-06-11
Codename: skill-demo-catalog

## Purpose

Create a small tooling foundation for skill visual demos. Given a `skill_id`, repo tooling can read
the shared skill rules and presentation metadata, classify how the skill should be demonstrated,
and expose the rank targets needed by later bot-visual flows.

This slice does not launch Godot or change gameplay. It establishes the deterministic catalog that
future visual runners will consume.

## Non-goals

- No `make skill-visual` target yet.
- No generated visual replay scenario yet.
- No changes to server skill behavior, balance, protocol schemas, or Godot presentation.
- No full skill coverage report beyond the minimal catalog/test output.

## Acceptance Criteria

- A Python helper can load all manifest-listed skills and presentations from `shared/`.
- The helper can resolve a single `skill_id` and report:
  - skill id, name, class id, kind, max rank,
  - icon label/color metadata,
  - demo category: attack, heal, self buff, stat buff, or unknown,
  - rank targets `[1, max_rank]` capped to available ranks.
- The helper fails clearly for an unknown `skill_id`.
- Unit tests cover Magic Bolt, Rage, Heal, Holy Shield, and an unknown skill.
- `make test-py` passes.

## Scope And Likely Files

- `tools/bot/skill_demo.py` - new catalog helper.
- `tools/bot/test_skill_demo.py` - focused unit tests.
- `docs/specs/v85_spec-skill-demo-catalog.md`
- `docs/plans/v85_2026-06-11-skill-demo-catalog.md`
- `PROGRESS.md` and `docs/as-built/v85_skill-demo-catalog.md` at finish.

## Test And Bot Proof

- `.venv/bin/pytest tools/bot/test_skill_demo.py -q`
- `make test-py`
- `make ci` during finish.

## Open Questions And Risks

- Risk: Skill kind classification may drift when new skill kinds land. Mitigation: keep the helper
  data-driven and unit-test every current kind.
  client UI/art integration.
