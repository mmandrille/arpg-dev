# v151 Plan — Extraction Independence Gate

Status: Complete
Goal: Add a CI-backed gate that prevents new helper-global namespace laundering from counting as structural extraction.
Architecture: Keep the existing file-size ratchet intact and add a second focused ratchet for extraction coupling. Existing `tools/bot/run.py` helper-global wrappers are grandfathered by count; new occurrences or stale baselines fail maintainability. The policy lives in `CLAUDE.md` so future specs/plans must require importable, unit-testable extracted modules.
Tech stack: Python tooling, pytest, Make maintainability target, SDD docs.

## Baseline and Shortcut Decision

Builds on v150 engineering review and its critique of v145-v149 line-count-only review. No Godot
applicable.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `scripts/check-extraction-coupling-ratchet.py` | Count and ratchet helper-global extraction coupling. |
| Create | `.maintainability/extraction-coupling-baseline.tsv` | Baseline current legacy coupling sites. |
| Modify | `make/ci.mk` | Run the new gate from `make maintainability`. |
| Create | `tools/test_extraction_coupling_ratchet.py` | Prove pass/fail behavior for the new gate. |
| Modify | `CLAUDE.md` | Add extraction independence policy. |
| Modify | `PROGRESS.md` | Record the freeze and lifecycle closeout. |
| Create | `docs/as-built/v151_extraction-independence-gate.md` | As-built proof. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `server/internal/game/game_test.go`
- [x] `tools/bot/run.py`
- [x] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice: add a small ratchet script and
  its focused pytest file without touching grandfathered source files.

Verification:
```bash
make maintainability
```

## Task 1 — Gate Script and Baseline

Files:
- Create: `scripts/check-extraction-coupling-ratchet.py`
- Create: `.maintainability/extraction-coupling-baseline.tsv`

- [x] Step 1.1: Count tracked source/tool occurrences of `helpers=globals()` with a TSV baseline.
- [x] Step 1.2: Fail when a baseline path grows, shrinks without a lowered baseline, disappears, or when an unbaselined source file contains the pattern.
```bash
python3 scripts/check-extraction-coupling-ratchet.py
```

## Task 2 — Maintainability Wiring and Tests

Files:
- Modify: `make/ci.mk`
- Create: `tools/test_extraction_coupling_ratchet.py`

- [x] Step 2.1: Run the new gate from `make maintainability`.
- [x] Step 2.2: Add pytest coverage for pass, growth failure, stale-baseline failure, and unbaselined coupling failure.
```bash
.venv/bin/pytest tools/test_extraction_coupling_ratchet.py -q
make maintainability
```

## Task 3 — Process Policy

Files:
- Modify: `CLAUDE.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Document that an extracted module only counts when it is importable and unit-testable without importing its source file or receiving the full source namespace.
- [x] Step 3.2: Record that the dedicated `run.py` split campaign is frozen unless a future typed-context slice replaces helper-global wrappers directly.
```bash
python3 scripts/check-extraction-coupling-ratchet.py
```

## Task 4 — Lifecycle Docs and CI

Files:
- Modify: `docs/specs/v151_spec-extraction-independence-gate.md`
- Modify: `docs/plans/v151_2026-06-14-extraction-independence-gate.md`
- Create: `docs/as-built/v151_extraction-independence-gate.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark the spec and plan complete after verification.
- [x] Step 4.2: Add the v151 lifecycle row/as-built and update current status.
```bash
make ci
```

## Final Verification

- [x] `python3 scripts/check-extraction-coupling-ratchet.py`
- [x] `.venv/bin/pytest tools/test_extraction_coupling_ratchet.py -q`
- [x] `make maintainability`
- [x] `make ci`
