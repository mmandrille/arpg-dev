# v86 Plan - Skill Visual Command

Status: Ready for implementation
Goal: Add `make skill-visual skill=<id>` as a validated wrapper around the existing visual replay path.
Architecture: The command consumes v85 skill demo metadata, maps each current skill to the best existing
protocol scenario, and delegates runtime work to `scripts/bot_visual.sh`. The wrapper is testable in
dry-run mode so CI does not launch Godot outside existing gates.
Tech stack: Makefile, Python tooling, existing bot-visual shell script.

## Baseline and shortcut decision

around existing replay infrastructure.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `Makefile` | Add `skill-visual` target. |
| Create | `tools/bot/skill_visual.py` | Validate skill and delegate to bot-visual. |
| Create | `tools/bot/test_skill_visual.py` | Unit coverage for mapping and failures. |
| Modify | `PROGRESS.md` | Lifecycle close-out. |
| Create | `docs/as-built/v86_skill-visual-command.md` | As-built proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none

Decision:
- [x] Extract focused helper/module/test file as part of this slice.
- [ ] Defer extraction with rationale: N/A

Verification:
```bash
make maintainability
```

## Task 1 - Skill visual wrapper

Files:
- Create: `tools/bot/skill_visual.py`
- Modify: `Makefile`

- [x] Step 1.1: Add a mapping from current skill ids to existing scenario ids.
```bash
.venv/bin/pytest tools/bot/test_skill_visual.py -q
```

- [x] Step 1.2: Add dry-run output that reports skill id, category, scenario id, and delegated command.
```bash
.venv/bin/python -m tools.bot.skill_visual holy_shield --dry-run
```

- [x] Step 1.3: Add `make skill-visual skill=<id>` target.
```bash
make skill-visual skill=holy_shield DRY_RUN=1
```

## Task 2 - Tests

Files:
- Create: `tools/bot/test_skill_visual.py`

- [x] Step 2.1: Cover every current skill mapping.
```bash
.venv/bin/pytest tools/bot/test_skill_visual.py -q
```

- [x] Step 2.2: Cover missing and unknown skill failures.
```bash
make test-py
```

## Task 3 - Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v86_skill-visual-command.md`
- Modify: `docs/plans/v86_2026-06-11-skill-visual-command.md`

- [x] Step 3.1: Mark plan tasks complete and add as-built notes.
```bash
rg -n "v86|skill-visual-command" PROGRESS.md docs/as-built docs/plans/v86_2026-06-11-skill-visual-command.md
```

- [x] Step 3.2: Run final verification.
```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `.venv/bin/pytest tools/bot/test_skill_visual.py -q`
- [x] `make skill-visual skill=holy_shield DRY_RUN=1`
- [x] `make test-py`
- [x] `make ci`

## Deferred scope

- Generated rank-5 visual scenarios.
- Character-page stat delta proof for buffs.
- Skill presentation matrix.
