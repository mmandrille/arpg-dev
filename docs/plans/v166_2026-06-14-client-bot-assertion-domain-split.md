# v166 Plan — Client Bot Assertion Domain Split

Status: Complete
Goal: Split UI/menu client bot assertion dispatch into a focused helper without DSL changes.
Architecture: The existing `BotAssertionHandlers.evaluate` remains the public assertion dispatcher.
A new helper handles a coherent subset and returns a handled/result pair; unhandled assertions stay
in the existing match.
Tech stack: Godot GDScript client bot and headless client unit tests.

## Baseline and shortcut decision

Builds on v165. Godot plugin adoption: reject, because this is internal client bot test harness
code with no UI/art/camera presentation change.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/bot_ui_assertion_handlers.gd` | UI/menu assertion domain helper |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Delegate handled UI/menu assertion types |
| Verify | `client/tests/test_client_bot.gd` | Existing coverage for moved assertions |
| Modify | `PROGRESS.md` | Slice lifecycle and summary |
| Add | `docs/as-built/v166_client-bot-assertion-domain-split.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/bot_scenario_runner.gd` if touched
- [x] `client/tests/test_client_bot.gd` if touched
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused helper module as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 — UI assertion helper

Files:
- Add: `client/scripts/bot_ui_assertion_handlers.gd`
- Modify: `client/scripts/bot_assertion_handlers.gd`

- [x] Step 1.1: Move main-menu, character-panel/session, multiplayer-panel/filter, settings, pause,
  and basic character-info visibility dispatch into the helper.
- [x] Step 1.2: Keep unhandled assertion types in the existing dispatcher.
```bash
make client-unit
```

## Task 2 — Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v166_client-bot-assertion-domain-split.md`

- [x] Step 2.1: Update lifecycle docs and write the as-built note.
- [x] Step 2.2: Run final verification.
```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make ci`
