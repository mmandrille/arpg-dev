# v145 Plan: Bot Runtime Assertion Split

Status: Complete
Goal: Move Python bot assertion dispatch out of `run.py` while preserving public helper imports and
scenario behavior.
Architecture: `tools.bot.run` remains the executable module and keeps `run_assertions` /
`run_runtime_assertions` wrappers. A new `tools/bot/runtime_assertions.py` module owns the long
snapshot/runtime dispatch chains. The wrappers pass the existing helper globals into the extracted
module to avoid importing `tools.bot.run` from the helper during `python -m tools.bot.run`.
Tech stack: Python protocol bot, pytest, maintainability ratchet, full bot/CI gates.

## Baseline and shortcut decision

Builds on v144 `client-bot-runner-split`. No Godot/plugin shortcut decision is needed; this slice
touches Python bot tooling only.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `tools/bot/run.py` | Keep executable bot and compatibility wrappers. |
| Create | `tools/bot/runtime_assertions.py` | Extract snapshot and runtime assertion dispatch. |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `tools/bot/run.py` baseline. |
| Modify | `docs/CODEMAP.md` | Add assertion dispatch module to Bot / scenarios tooling files. |
| Create | `docs/as-built/v145_bot-runtime-assertion-split.md` | Close-out proof and deferred scope. |
| Modify | `PROGRESS.md` | Mark v145 complete and update current status. |

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

## Task 1 - Extract runtime assertion dispatch

Files:
- Modify: `tools/bot/run.py`
- Create: `tools/bot/runtime_assertions.py`

- [x] Step 1.1: Move `run_assertions` dispatch into the new module.
- [x] Step 1.2: Move `run_runtime_assertions` dispatch into the new module.
- [x] Step 1.3: Keep `run.py` wrapper functions with the same names/signatures.
- [x] Step 1.4: Avoid helper-to-`run.py` imports; pass existing helper globals from wrappers.
```bash
.venv/bin/pytest tools/bot/test_item_assertions.py tools/bot/test_stash_assertions.py
```

## Task 2 - Bot and Python verification

Files:
- Modify: `tools/bot/run.py`
- Create: `tools/bot/runtime_assertions.py`

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
- Create: `docs/as-built/v145_bot-runtime-assertion-split.md`
- Modify: `docs/specs/v145_spec-bot-runtime-assertion-split.md`
- Modify: `docs/plans/v145_2026-06-13-bot-runtime-assertion-split.md`

- [x] Step 3.1: Lower `tools/bot/run.py` baseline to the post-extraction line count.
- [x] Step 3.2: Update CODEMAP and lifecycle docs.
- [x] Step 3.3: Run final verification.
```bash
make maintainability
make ci
```

## Final verification

- [x] `.venv/bin/pytest tools/bot/test_item_assertions.py tools/bot/test_stash_assertions.py`
- [x] `make test-py`
- [x] `make bot`
- [x] `make maintainability`
- [x] `make ci`

## Deferred scope

- Splitting action execution, movement helpers, co-op orchestration, and replay helpers out of
  `tools/bot/run.py` remains future paydown.
