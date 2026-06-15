# v144 Plan: Client Bot Runner Split

Status: Complete
Goal: Move bot scenario runner dispatch and validation chains into focused helper modules without
changing scenario semantics.
Architecture: `BotScenarioRunner` remains the only public runner class bound by `BotController`.
New helper classes hold step catalogs, static validation, wait dispatch, assertion dispatch, and
action dispatch. Helpers call back into the runner for existing state, matchers, failure formatting,
and logging-sensitive context so behavior stays unchanged.
Tech stack: Godot GDScript, client unit tests, headless client bot scenarios, maintainability ratchet.

## Baseline and shortcut decision

is internal test-runner infrastructure, not UI/art/camera/asset work.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/bot_scenario_runner.gd` | Keep public runner state/API; delegate dispatch and validation. |
| Create | `client/scripts/bot_step_catalog.gd` | Step type categories and scenario/step validation. |
| Create | `client/scripts/bot_wait_handlers.gd` | Wait-step dispatch. |
| Create | `client/scripts/bot_assertion_handlers.gd` | Assertion-step dispatch. |
| Create | `client/scripts/bot_action_handlers.gd` | Action-step dispatch. |
| Modify | `client/tests/test_client_bot.gd` | Preserve/extend compatibility coverage for runner constants and validation delegates. |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `bot_scenario_runner.gd` baseline. |
| Modify | `docs/CODEMAP.md` | Add runner helper modules to Bot / scenarios. |
| Create | `docs/as-built/v144_client-bot-runner-split.md` | Close-out proof and deferred scope. |
| Modify | `PROGRESS.md` | Mark v144 complete and update current status. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] `client/tests/test_client_bot.gd` if modified
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 - Extract step catalog and validation

Files:
- Create: `client/scripts/bot_step_catalog.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `client/tests/test_client_bot.gd` if compatibility coverage needs adjustment

- [x] Step 1.1: Move step type arrays, `ALL_STEP_TYPES`, `validate_scenario`, and
  `validate_step` into `BotStepCatalog`.
- [x] Step 1.2: Keep `BotScenarioRunner` constants and static validation methods as compatibility
  delegates.
```bash
make client-unit
```

## Task 2 - Extract wait/assert/action dispatch

Files:
- Create: `client/scripts/bot_wait_handlers.gd`
- Create: `client/scripts/bot_assertion_handlers.gd`
- Create: `client/scripts/bot_action_handlers.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`

- [x] Step 2.1: Move wait-step `match` dispatch into `BotWaitHandlers.evaluate`.
- [x] Step 2.2: Move assertion-step `match` dispatch into `BotAssertionHandlers.evaluate`.
- [x] Step 2.3: Move action-step dispatch into `BotActionHandlers.queue`.
- [x] Step 2.4: Keep runner-owned helper methods, memory, failure formatting, and pending-action
  storage unchanged behind the helper calls.
```bash
make client-unit
```

## Task 3 - Ratchet, CODEMAP, lifecycle, and CI

Files:
- Modify: `.maintainability/file-size-baseline.tsv`
- Modify: `docs/CODEMAP.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v144_client-bot-runner-split.md`
- Modify: `docs/specs/v144_spec-client-bot-runner-split.md`
- Modify: `docs/plans/v144_2026-06-13-client-bot-runner-split.md`

- [x] Step 3.1: Lower `bot_scenario_runner.gd` baseline to the post-extraction line count.
- [x] Step 3.2: Update CODEMAP and lifecycle docs.
- [x] Step 3.3: Run final verification.
```bash
make bot-client HEADLESS=1
make maintainability
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make bot-client HEADLESS=1`
- [x] `make maintainability`
- [x] `make ci`

## Deferred scope

- Splitting domain-specific matcher helpers such as shop/stash/market row matching into separate
  assertion-domain modules remains future paydown.
- Python `tools/bot/run.py` assertion dispatch remains v145.
