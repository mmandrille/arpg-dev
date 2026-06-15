# v168 Plan — Bot Step Validation Split

Status: Complete
Goal: Extract action-step validation from `BotStepCatalog.validate_step`.
Architecture: `BotStepCatalog.validate_step` remains the public validation API. A new helper returns
an empty string for valid handled action steps, an error string for invalid handled action steps, and
an unhandled sentinel for non-action steps.
Tech stack: Godot GDScript client bot and headless client unit tests.

## Baseline and shortcut decision

with no UI/art/camera presentation change.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/bot_action_step_validator.gd` | Action-step validation helper |
| Modify | `client/scripts/bot_step_catalog.gd` | Delegate action validation |
| Verify | `client/tests/test_client_bot.gd` | Existing invalid-step coverage |
| Modify | `PROGRESS.md` | Slice lifecycle and summary |
| Add | `docs/as-built/v168_bot-step-validation-split.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/tests/test_client_bot.gd` if touched
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused helper module as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 — Action validation helper

Files:
- Add: `client/scripts/bot_action_step_validator.gd`
- Modify: `client/scripts/bot_step_catalog.gd`

- [x] Step 1.1: Move action validation checks for key presses, clicks, drags, hotbar use, menu
  actions, stash/search/sort, market actions, and blacksmith upgrade into the helper.
- [x] Step 1.2: Keep wait/assert validation in `BotStepCatalog.validate_step`.
```bash
make client-unit
```

## Task 2 — Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v168_bot-step-validation-split.md`

- [x] Step 2.1: Update lifecycle docs and write the as-built note.
- [x] Step 2.2: Run final verification.
```bash
make maintainability
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make ci`
