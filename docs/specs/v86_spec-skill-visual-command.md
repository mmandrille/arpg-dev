# v86 Spec - Skill Visual Command

Status: Draft
Date: 2026-06-11
Codename: skill-visual-command

## Purpose

Add a single command entrypoint for inspecting a skill through the existing visual replay path:

```bash
make skill-visual skill=holy_shield
```

The command validates `skill` through the v85 skill demo catalog, selects the best existing bot
scenario for that skill, and runs the same infrastructure as `make bot-visual`.

## Non-goals

- No generated rank-5 scenario yet.
- No new Godot UI bot flow for opening the skill tree or character page.
- No new server behavior, skill balance, protocol schema, or presentation asset.
- No coverage matrix beyond command mapping tests.

## Acceptance Criteria

- `make skill-visual skill=magic_bolt|rage|heal|holy_shield` delegates to `scripts/bot_visual.sh`
  with the mapped scenario id.
- Unknown skills fail before launching the server or Godot.
- Missing `skill=` fails with a clear message.
- A dry-run mode can be unit-tested without launching Godot.
- Unit tests cover every current skill mapping and unknown/missing input.

## Scope And Likely Files

- `Makefile`
- `tools/bot/skill_visual.py`
- `tools/bot/test_skill_visual.py`
- `docs/specs/v86_spec-skill-visual-command.md`
- `docs/plans/v86_2026-06-11-skill-visual-command.md`
- `PROGRESS.md` and `docs/as-built/v86_skill-visual-command.md` at finish.

## Test And Bot Proof

- `.venv/bin/pytest tools/bot/test_skill_visual.py -q`
- `.venv/bin/python -m tools.bot.skill_visual holy_shield --dry-run`
- `make test-py`
- `make maintainability`
- `make ci` during finish.

Manual visual command for this slice:

```bash
make skill-visual skill=holy_shield
```

## Open Questions And Risks

- Risk: Existing scenarios show rank-1 behavior only. Mitigation: v87 will add generated rank target
  demos and buff stat proof.
