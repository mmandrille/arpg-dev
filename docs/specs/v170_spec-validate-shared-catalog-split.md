# v170 Spec — Validate Shared Catalog Split

Status: Complete
Date: 2026-06-14
Codename: `validate-shared-catalog-split`

## Purpose

Move one coherent rules-catalog validation domain out of `tools/validate_shared.py` into a focused
helper module without changing shared-data semantics or CLI behavior. This slice targets
`main_config` gameplay validation.

## Non-goals

- No shared rules, schema, protocol, golden, gameplay, or content changes.
- No broad rewrite of `validate_shared.py`.
- No changes to validation labels or expected success/failure messages beyond helper ownership.

## Acceptance Criteria

- `main_config` gameplay validation lives in a focused helper module.
- `tools/validate_shared.py` remains the shared-validation entrypoint and delegates to the helper.
- A focused Python regression covers a bad `main_config` gameplay value through the helper.
- `validate_shared.py` line count and maintainability baseline are lowered.
- Focused pytest, `make validate-shared`, `make maintainability`, and `make ci` pass.

## Scope and Files

- Tools: `tools/validate_shared.py`, `tools/validate_main_config.py`
- Tests: `tools/test_validate_shared.py`
- Maintainability: `.maintainability/file-size-baseline.tsv`
- Docs: `PROGRESS.md`, `docs/as-built/v170_validate-shared-catalog-split.md`

## Test and Bot Proof

Focused proof:

```bash
.venv/bin/pytest tools/test_validate_shared.py -q
make validate-shared
```

Full proof: `make ci`.

## Open Questions and Risks

- Risk: `validate_shared.py` local variables can be reused later in the file. Keep the shared
  orchestration scalar needed by later shop validation local to `validate_shared.py`.
