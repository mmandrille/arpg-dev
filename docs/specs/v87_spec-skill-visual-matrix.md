# v87 Spec - Skill Visual Matrix

Status: Draft
Date: 2026-06-11
Codename: skill-visual-matrix

## Purpose

Extend the skill visual tooling with a matrix that shows, for every shared skill, the visual command,
current scenario coverage, rank targets, and whether buff stat-delta proof is still missing.

This closes the compact autoloop batch by making the remaining rank-5 and character-page requirements
visible and data-driven instead of implicit.

## Non-goals

- No generated rank-5 farming scenario yet.
- No Godot character-page automation for buff stat deltas yet.
- No server/debug shortcut to grant skill ranks.
- No new protocol or gameplay behavior.

## Acceptance Criteria

- A CLI command prints a stable skill visual matrix for all current skills.
- The matrix includes skill id, class, category, icon label, rank targets, scenario id, rank-1 coverage,
  rank-5 coverage, and buff stat-delta coverage.
- `make skill-visual-list` exposes the matrix.
- Tests verify current coverage: all current skills have rank-1 visual command coverage; rank-5 and buff
  stat-delta coverage are reported as not yet covered.
- `make test-py`, `make maintainability`, and `make ci` pass.

## Scope And Likely Files

- `tools/bot/skill_visual.py`
- `tools/bot/test_skill_visual.py`
- `Makefile`
- `docs/specs/v87_spec-skill-visual-matrix.md`
- `docs/plans/v87_2026-06-11-skill-visual-matrix.md`
- `PROGRESS.md` and `docs/as-built/v87_skill-visual-matrix.md` at finish.

## Test And Bot Proof

- `.venv/bin/pytest tools/bot/test_skill_visual.py -q`
- `make skill-visual-list`
- `make test-py`
- `make ci` during finish.

## Open Questions And Risks

- Risk: This reports rank-5/buff-stat gaps rather than implementing them. Mitigation: the matrix is explicit
  deferred scope and the next slice can target generated rank demos without rediscovery.
