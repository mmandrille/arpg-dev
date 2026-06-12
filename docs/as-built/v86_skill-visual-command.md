# v86 As-built - Skill Visual Command

## What shipped

`make skill-visual skill=<skill_id>` now validates a shared skill id through the v85 demo catalog,
passes it to the reusable `skill_visual` bot scenario, and delegates to `scripts/bot_visual.sh`.

The scenario is data-driven from the skill catalog:

- damage skills cast on a far enemy so auto-navigation is visible before impact;
- ally-targeted help skills create a visible ally three tiles away and cast on that ally;
- self-only skills cast on the hero because the rules do not allow ally targeting.

## What it proves

- Humans can launch a visual replay for a known skill with one command.
- Humans can launch all catalog skill visuals with `make skill-visual skill=all`.
- Future catalog skills can use the same scenario without adding a per-skill scenario file.
- Unknown or missing skills fail before server/Godot startup.
- Dry-run mode prints the selected scenario and delegated command for test coverage.

## Verification

```bash
.venv/bin/pytest tools/bot/test_skill_visual.py -q
make skill-visual skill=holy_shield DRY_RUN=1
make skill-visual skill=all DRY_RUN=1
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
