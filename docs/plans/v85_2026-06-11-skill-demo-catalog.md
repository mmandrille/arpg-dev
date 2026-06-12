# v85 Plan - Skill Demo Catalog

Status: Ready for implementation
Goal: Add a deterministic Python catalog helper that describes how each shared skill should be demonstrated.
Architecture: The helper reads the existing content manifest, skill rules, and skill presentation JSON. It
does not alter authoritative gameplay or client presentation. Later visual-runner slices consume this helper
instead of duplicating skill-kind and icon assumptions.
Tech stack: Python tooling, shared JSON, pytest.

## Baseline and shortcut decision

Builds on v84 `client-bot-step-registry`. Godot plugin decision: reject, because this slice is engine-free
tooling over existing shared contracts.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `tools/bot/skill_demo.py` | Load and classify skill demo metadata. |
| Create | `tools/bot/test_skill_demo.py` | Unit coverage for current skill kinds and errors. |
| Create | `server/internal/game/character_stats_test.go` | Extract focused character-stat tests from oversized `game_test.go`. |
| Modify | `server/internal/game/game_test.go` | Shrink grandfathered test file below ratchet allowance. |
| Modify | `docs/plans/v85_2026-06-11-skill-demo-catalog.md` | Track execution. |
| Modify | `PROGRESS.md` | Lifecycle close-out. |
| Create | `docs/as-built/v85_skill-demo-catalog.md` | As-built proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `server/internal/game/game_test.go`

Decision:
- [x] Extract focused helper/module/test file as part of this slice.
- [ ] Defer extraction with rationale: N/A

Verification:
```bash
make maintainability
```

## Task 1 - Skill demo catalog helper

Files:
- Create: `tools/bot/skill_demo.py`

- [x] Step 1.1: Load manifest-listed skill rule and presentation files through `tools.content_manifest`.
```bash
.venv/bin/pytest tools/bot/test_skill_demo.py -q
```

- [x] Step 1.2: Expose `skill_demo_entry(skill_id)` with id, display, icon, kind, category, class, max rank, and rank targets.
```bash
.venv/bin/pytest tools/bot/test_skill_demo.py -q
```

- [x] Step 1.3: Add a small CLI mode for human inspection: `python -m tools.bot.skill_demo magic_bolt`.
```bash
.venv/bin/python -m tools.bot.skill_demo magic_bolt
```

## Task 2 - Tests

Files:
- Create: `tools/bot/test_skill_demo.py`

- [x] Step 2.1: Cover Magic Bolt, Rage, Heal, and Holy Shield categories and icon metadata.
```bash
.venv/bin/pytest tools/bot/test_skill_demo.py -q
```

- [x] Step 2.2: Cover unknown skill errors.
```bash
make test-py
```

## Task 3 - Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v85_skill-demo-catalog.md`
- Modify: `docs/plans/v85_2026-06-11-skill-demo-catalog.md`

- [x] Step 3.1: Mark plan tasks complete and add as-built notes.
```bash
rg -n "v85|skill-demo-catalog" PROGRESS.md docs/as-built docs/plans/v85_2026-06-11-skill-demo-catalog.md
```

- [x] Step 3.2: Run final verification.
```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `.venv/bin/pytest tools/bot/test_skill_demo.py -q`
- [x] `make test-py`
- [x] `make ci`

## Deferred scope

- `make skill-visual skill=...` visual command.
- Rank 1/rank 5 Godot replay flows.
- Character-page stat delta proof for buffs.
- Skill presentation coverage matrix.
