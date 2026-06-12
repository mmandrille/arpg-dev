# v87 As-built - Skill Visual Matrix

## What shipped

`make skill-visual-list` now prints a skill visual coverage matrix derived from the shared skill
catalog and the v86 command mappings.

The matrix reports:

- skill id, class, category, icon label, rank targets, and mapped scenario;
- whether rank-1 visual command coverage exists;
- whether rank-5 visual coverage exists;
- whether buff stat-delta visual coverage exists.

## What it proves

- Every current skill has rank-1 visual command coverage through `make skill-visual skill=<id>`.
- Rank-5 visual coverage is explicitly not yet implemented.
- Buff stat-delta visual coverage is explicitly not yet implemented.
- The remaining skill-demo work is visible from tooling instead of hidden in docs.

## Verification

```bash
.venv/bin/pytest tools/bot/test_skill_visual.py -q
make skill-visual-list
make test-py
make maintainability
make ci
```

## Deferred

- Generated rank-5 skill demos.
- Godot character-page stat-delta visual proof before/after buff casts.
