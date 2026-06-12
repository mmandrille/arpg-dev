# v86 As-built - Skill Visual Command

## What shipped

`make skill-visual skill=<skill_id>` now validates a shared skill id through the v85 demo catalog,
maps it to the best current bot scenario, and delegates to `scripts/bot_visual.sh`.

Current mappings:

- `magic_bolt` -> `skill_points_and_magic_bolt`
- `rage` -> `rage_and_heal_skills`
- `heal` -> `paladin_heal_skill`
- `holy_shield` -> `paladin_holy_shield`

## What it proves

- Humans can launch a visual replay for a known skill with one command.
- Unknown or missing skills fail before server/Godot startup.
- Dry-run mode prints the selected scenario and delegated command for test coverage.

## Verification

```bash
.venv/bin/pytest tools/bot/test_skill_visual.py -q
make skill-visual skill=holy_shield DRY_RUN=1
make test-py
make maintainability
make ci
```

Manual visual command:

```bash
make skill-visual skill=holy_shield
```

## Deferred

- Generated rank-5 visual scenarios.
- Character-page stat delta proof for buff skills.
- Skill presentation matrix.
