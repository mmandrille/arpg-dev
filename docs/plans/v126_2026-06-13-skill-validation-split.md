# v126 Plan - Skill Validation Split

Status: Complete
Goal: Extract skill-specific shared validation into a dedicated module without changing validation
semantics.
Architecture: `validate_shared.py` remains the orchestration entrypoint. `validate_skills.py`
contains skill-domain helpers and reports through the existing `Report` interface.
Tech stack: Python shared tooling, SDD docs.

## Baseline And Shortcut Decision

Baseline is v125 `tuning-friendly-bot-scenarios` on `main`, committed as `19a75f85`.

Godot plugin adoption checklist: not applicable. This slice is Python validation tooling only.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `tools/validate_skills.py` | Skill-domain validation helpers. |
| Modify | `tools/validate_shared.py` | Delegate skill validation and drop extracted helper code. |
| Modify | `tools/test_validate_shared.py` | Add focused regression for extracted helper failure behavior. |
| Add | `docs/as-built/v126_skill-validation-split.md` | Record implementation and proof. |
| Modify | `PROGRESS.md` | Lifecycle closeout. |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files likely touched:
- [x] `tools/validate_shared.py`
- [x] `tools/test_validate_shared.py`

Decision:
- [x] Move one coherent skill block only; leave unrelated cross-check domains in place for future
  slices.

## Task 1 - Extract Skill Validator

- [x] Step 1.1: Add `tools/validate_skills.py` with `validate_skill_catalogs`.
- [x] Step 1.2: Move the skill requirement helper into the new module.
- [x] Step 1.3: Preserve existing report labels and failure order inside the extracted domain.

## Task 2 - Wire And Test

- [x] Step 2.1: Import and call the helper from `validate_shared.py`.
- [x] Step 2.2: Add a focused unit regression for a skill validation failure through the extracted
  helper.
- [x] Step 2.3: Run focused pytest and `make validate-shared`.

## Task 3 - Lifecycle

- [x] Step 3.1: Mark docs complete and add as-built notes.
- [x] Step 3.2: Run `make ci`.
- [x] Step 3.3: Update `PROGRESS.md` and commit.

## Final Verification

- [x] `.venv/bin/pytest tools/test_validate_shared.py -q`
- [x] `make validate-shared`
- [x] `make ci`
