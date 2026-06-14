# v147 Plan: Bot Wait Runtime Split

Status: Complete
Goal: Move protocol bot wait/pump helpers out of `run.py` with no scenario behavior change.
Architecture: `tools.bot.run` remains the executable bot and keeps wait/pump helper wrappers. A new
`tools/bot/wait_runtime.py` owns the moved implementations and receives existing helper globals from
wrappers to avoid reverse-importing `tools.bot.run` during `python -m tools.bot.run`.
Tech stack: Python protocol bot, pytest, maintainability ratchet, full bot/CI gates.

## Baseline and shortcut decision

Builds on v146 `bot-movement-runtime-split`. No Godot/plugin shortcut decision is needed; this slice
touches Python bot tooling only.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `tools/bot/run.py` | Keep executable bot and compatibility wrappers. |
| Create | `tools/bot/wait_runtime.py` | Own wait/pump helper implementations. |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `tools/bot/run.py` baseline. |
| Modify | `docs/CODEMAP.md` | Add wait runtime module to Bot / scenarios tooling files. |
| Create | `docs/as-built/v147_bot-wait-runtime-split.md` | Close-out proof and deferred scope. |
| Modify | `PROGRESS.md` | Mark v147 complete and update current status. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `tools/bot/run.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 - Extract wait runtime helpers

Files:
- Modify: `tools/bot/run.py`
- Create: `tools/bot/wait_runtime.py`

- [x] Step 1.1: Move wait/pump helper bodies into `wait_runtime.py`.
- [x] Step 1.2: Keep `run.py` wrapper functions with the same names/signatures.
- [x] Step 1.3: Avoid helper-to-`run.py` imports; pass existing helper globals from wrappers.
```bash
python -m py_compile tools/bot/run.py tools/bot/wait_runtime.py
```

## Task 2 - Bot and Python verification

Files:
- Modify: `tools/bot/run.py`
- Create: `tools/bot/wait_runtime.py`

- [x] Step 2.1: Run Python unit checks.
- [x] Step 2.2: Run full protocol bot proof.
```bash
make test-py
make bot
```

## Task 3 - Ratchet, CODEMAP, lifecycle, and CI

Files:
- Modify: `.maintainability/file-size-baseline.tsv`
- Modify: `docs/CODEMAP.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v147_bot-wait-runtime-split.md`
- Modify: `docs/specs/v147_spec-bot-wait-runtime-split.md`
- Modify: `docs/plans/v147_2026-06-14-bot-wait-runtime-split.md`

- [x] Step 3.1: Lower `tools/bot/run.py` baseline to the post-extraction line count.
- [x] Step 3.2: Update CODEMAP and lifecycle docs.
- [x] Step 3.3: Run final verification.
```bash
.venv/bin/python tools/validate_codemap.py
make maintainability
make ci
```

## Final verification

- [x] `python -m py_compile tools/bot/run.py tools/bot/wait_runtime.py`
- [x] `make test-py`
- [x] `make bot`
- [x] `.venv/bin/python tools/validate_codemap.py`
- [x] `make maintainability`
- [x] `make ci`

## Deferred scope

- Splitting action dispatch, state ingestion, co-op orchestration, and replay helpers out of
  `tools/bot/run.py` remains future paydown.
