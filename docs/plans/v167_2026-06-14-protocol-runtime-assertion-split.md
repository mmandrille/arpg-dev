# v167 Plan — Protocol Runtime Assertion Split

Status: Complete
Goal: Extract runtime shop/stash economy assertions into a focused Python helper.
Architecture: `run_runtime_assertions` remains the public entrypoint. A new helper receives the
runtime state, assertion, where-label, and existing narrow helper bindings, then returns whether it
handled the assertion.
Tech stack: Python protocol bot tests and full repo CI.

## Baseline and shortcut decision

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `tools/bot/runtime_economy_assertions.py` | Runtime shop/stash assertion domain |
| Modify | `tools/bot/runtime_assertions.py` | Delegate economy assertions |
| Verify | `tools/bot/test_protocol.py` | Existing runtime assertion coverage |
| Modify | `PROGRESS.md` | Slice lifecycle and summary |
| Add | `docs/as-built/v167_protocol-runtime-assertion-split.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `tools/bot/test_protocol.py` if touched
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused helper module as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 — Economy assertion helper

Files:
- Add: `tools/bot/runtime_economy_assertions.py`
- Modify: `tools/bot/runtime_assertions.py`

- [x] Step 1.1: Move runtime `stash_item_count`, `stash_gold`, `stash_capacity`,
  `shop_offer_count`, `shop_offer_details`, `shop_sell_appraisal_count`,
  `shop_sell_appraisal_details`, `shop_event`, and `stash_event` handling into the helper.
- [x] Step 1.2: Keep unhandled assertion types in `runtime_assertions.py`.
```bash
.venv/bin/pytest tools/bot/test_protocol.py
```

## Task 2 — Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v167_protocol-runtime-assertion-split.md`

- [x] Step 2.1: Update lifecycle docs and write the as-built note.
- [x] Step 2.2: Run final verification.
```bash
make maintainability
make ci
```

## Final verification

- [x] `.venv/bin/pytest tools/bot/test_protocol.py`
- [x] `make maintainability`
- [x] `make ci`
