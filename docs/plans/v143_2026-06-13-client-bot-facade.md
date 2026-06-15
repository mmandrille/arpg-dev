# v143 Plan: Client Bot Facade

Status: Complete
Goal: Move client bot adapter implementations out of `main.gd` while preserving the public bot API.
Architecture: `BotController` continues calling `main.gd` `bot_*` methods. `main.gd` delegates the
panel/action implementation to a focused `BotFacade` helper so future bot adapter growth has a
small home outside the main scene coordinator.
Tech stack: Godot GDScript, client smoke/unit tests, client bot scenarios, maintainability ratchet.

## Baseline and shortcut decision

add maintenance cost and would not replace the existing `BotController` API.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/bot_facade.gd` | Focused bot adapter implementation for panels, hotbar, stats, skills, and skill direction. |
| Modify | `client/scripts/main.gd` | Keep public `bot_*` wrappers and delegate to `BotFacade`. |
| Create | `client/tests/test_bot_facade.gd` | Headless fake-panel unit coverage. |
| Modify | `scripts/client_smoke.sh` | Run the new facade test in `make client-unit`/client smoke. |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `main.gd` grandfathered baseline. |
| Modify | `docs/CODEMAP.md` | Add `bot_facade.gd` to Bot / scenarios client files. |
| Create | `docs/as-built/v143_client-bot-facade.md` | Close-out proof and deferred scope. |
| Modify | `PROGRESS.md` | Mark v143 complete and update current status. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 - Extract BotFacade

Files:
- Create: `client/scripts/bot_facade.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 1.1: Move selected shop/stash/bishop/blacksmith/market/hotbar/stat/skill adapter bodies into `BotFacade`.
- [x] Step 1.2: Replace each moved `main.gd` method body with a same-signature delegate to `BotFacade`.
- [x] Step 1.3: Keep direct `main.gd` logic for unrelated state assembly, menu flow, pending events, and shadow replay.
```bash
make client-unit
```

## Task 2 - Add facade unit coverage

Files:
- Create: `client/tests/test_bot_facade.gd`
- Modify: `scripts/client_smoke.sh`

- [x] Step 2.1: Add fake panel/main classes that record facade calls.
- [x] Step 2.2: Cover representative panel adapters, hotbar adapters, skill direct cast, and skill-bar fallback.
- [x] Step 2.3: Add the test to the client-unit gate.
```bash
make client-unit
```

## Task 3 - Ratchet, CODEMAP, lifecycle, and CI

Files:
- Modify: `.maintainability/file-size-baseline.tsv`
- Modify: `docs/CODEMAP.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v143_client-bot-facade.md`
- Modify: `docs/specs/v143_spec-client-bot-facade.md`
- Modify: `docs/plans/v143_2026-06-13-client-bot-facade.md`

- [x] Step 3.1: Lower the `main.gd` baseline to the post-extraction line count.
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

Note: the watched, non-headless `make bot-client` run ended `38 passed, 2 failed` in
`main_menu_flow` and `client_boss_phase_readability`, both outside the extracted facade surface.
The CI-equivalent headless command passed all 40 scenarios.

## Deferred scope

- Splitting `client/scripts/bot_scenario_runner.gd` remains v144.
- Removing compatibility wrappers from `main.gd` remains deferred until `BotController` has a new stable facade binding.
