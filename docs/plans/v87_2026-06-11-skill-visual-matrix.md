# v87 Plan - Skill Visual Matrix

Status: Ready for implementation
Goal: Add a data-driven matrix that reports current skill visual command coverage and remaining rank/stat gaps.
Architecture: Reuse v85/v86 metadata and scenario mappings. Coverage booleans are conservative and explicit:
current scenarios prove rank 1, while rank 5 and buff stat-delta visual proofs remain false until dedicated
scenario support exists.
Tech stack: Makefile, Python tooling, pytest.

## Baseline and shortcut decision

existing coverage.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `tools/bot/skill_visual.py` | Add matrix rows and CLI list mode. |
| Modify | `tools/bot/test_skill_visual.py` | Cover matrix content and conservative gaps. |
| Modify | `Makefile` | Add `skill-visual-list` target. |
| Modify | `PROGRESS.md` | Lifecycle close-out. |
| Create | `docs/as-built/v87_skill-visual-matrix.md` | As-built proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none

Decision:
- [x] Keep changes in focused files below 600 lines.
- [ ] Defer extraction with rationale: N/A

Verification:
```bash
make maintainability
```

## Task 1 - Matrix reporting

Files:
- Modify: `tools/bot/skill_visual.py`
- Modify: `Makefile`

- [x] Step 1.1: Add matrix row generation for every current skill.
```bash
.venv/bin/pytest tools/bot/test_skill_visual.py -q
```

- [x] Step 1.2: Add `--list` CLI and `make skill-visual-list`.
```bash
make skill-visual-list
```

## Task 2 - Tests

Files:
- Modify: `tools/bot/test_skill_visual.py`

- [x] Step 2.1: Assert all current skills have scenario/rank-1 coverage.
```bash
.venv/bin/pytest tools/bot/test_skill_visual.py -q
```

- [x] Step 2.2: Assert rank-5 and buff stat-delta coverage are false for current mappings.
```bash
make test-py
```

## Task 3 - Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v87_skill-visual-matrix.md`
- Modify: `docs/plans/v87_2026-06-11-skill-visual-matrix.md`

- [x] Step 3.1: Mark plan tasks complete and add as-built notes.
```bash
rg -n "v87|skill-visual-matrix" PROGRESS.md docs/as-built docs/plans/v87_2026-06-11-skill-visual-matrix.md
```

- [x] Step 3.2: Run final verification.
```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `.venv/bin/pytest tools/bot/test_skill_visual.py -q`
- [x] `make skill-visual-list`
- [x] `make test-py`
- [x] `make ci`

## Deferred scope

- Generated rank-5 skill demos.
- Godot character-page stat-delta visual proof before/after buff casts.
