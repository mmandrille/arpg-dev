# v137 Plan - Bot Assertion Domain Split

Status: Complete
Goal: Split stash and unique-chest bot assertion helpers out of the broad protocol runner without changing scenario behavior.
Architecture: Keep the new helper module pure and synchronous. `run.py` remains the action dispatcher and passes its count comparator into helper assertions so the extracted module does not import the runner.
Tech stack: Python protocol bot tooling, pytest, lifecycle docs.

## Baseline and shortcut decision

Builds on v136 `unique-chest-client-proof` and the v130 review finding that bot runtime assertions
should be split by domain before more unique/chest assertions land. No client UI, camera,

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `tools/bot/stash_assertions.py` | Stash/chest item filtering, selection, and count/event assertions |
| Modify | `tools/bot/run.py` | Import helper functions and remove duplicated stash helper bodies |
| Modify | `tools/bot/test_protocol.py` | Existing assertion path coverage |
| Create | `tools/bot/test_stash_assertions.py` | Direct helper tests |
| Create | `tools/bot/test_item_assertions.py` | Rolled inventory display-name suffix regression coverage |
| Modify | `scripts/bot_local.sh`, `scripts/bot_client.sh`, `scripts/bot_client_local.sh`, `scripts/ci.sh` | Keep debug-gated bot scenarios deterministic in local and CI launchers |
| Modify | `tools/bot/scenarios/61_purple_town_unique_chest.json`, `tools/bot/scenarios/client/32_town_bishop_respec_panel.json` | Remove brittle entity assumptions from existing scenarios |
| Modify | `client/scripts/bot_controller.gd`, `client/scripts/main.gd`, `client/tests/test_golden.gd` | Keep client bot entity filtering, bot dispatch, and unique-name golden checks aligned with current contracts |
| Modify | `server/internal/replay/replay_test.go` | Derive replay guest ids from session players instead of world entity counts |
| Create | `docs/as-built/v137_bot-assertion-domain-split.md` | As-built summary |
| Modify | `PROGRESS.md` | Lifecycle closeout |
| Modify | `docs/specs/v137_spec-bot-assertion-domain-split.md` | Status closeout |
| Modify | `docs/plans/v137_2026-06-13-bot-assertion-domain-split.md` | Checkbox closeout |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `tools/bot/run.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `tools/bot/test_protocol.py`; pre-existing `server/internal/game/sim.go` drift was also baselined after user approval.

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 - Extract stash assertion helper

Files:
- Create: `tools/bot/stash_assertions.py`
- Modify: `tools/bot/run.py`

- [x] Step 1.1: Move stash-item filtering, stash selection, stash id lookup, and stash event/count assertion logic into the helper module.
- [x] Step 1.2: Keep `run.py` dispatch behavior and public imports used by tests stable.

```bash
.venv/bin/pytest tools/bot/test_protocol.py -q
```

## Task 2 - Focused helper coverage

Files:
- Create: `tools/bot/test_stash_assertions.py`
- Modify: `tools/bot/test_protocol.py`

- [x] Step 2.1: Add direct tests for filtering, selecting, missing selection errors, and stash event assertion behavior.
- [x] Step 2.2: Keep existing snapshot/delta assertion tests passing through `run_assertions` and `run_runtime_assertions`.
- [x] Step 2.3: Add focused coverage that rolled inventory display-name suffix assertions are opt-in.

```bash
.venv/bin/pytest tools/bot/test_protocol.py tools/bot/test_stash_assertions.py -q
```

## Task 2b - CI closeout repairs

Files:
- Modify: `scripts/bot_local.sh`, `scripts/bot_client.sh`, `scripts/bot_client_local.sh`, `scripts/ci.sh`
- Modify: `tools/bot/scenarios/61_purple_town_unique_chest.json`
- Modify: `tools/bot/scenarios/client/32_town_bishop_respec_panel.json`
- Modify: `client/scripts/bot_controller.gd`, `client/scripts/main.gd`, `client/tests/test_golden.gd`
- Modify: `server/internal/replay/replay_test.go`

- [x] Step 2b.1: Keep debug-gated unique chest and bishop bot scenarios deterministic under local and CI launchers.
- [x] Step 2b.2: Remove brittle entity-count/id assumptions from replay and existing bot scenarios.
- [x] Step 2b.3: Keep client bot filtered entity clicks from falling back to unrelated interactables while preserving auto-approach.
- [x] Step 2b.4: Align the Godot golden item-roll check with unique display-name rules.

## Task 3 - Lifecycle docs and CI

Files:
- Create: `docs/as-built/v137_bot-assertion-domain-split.md`
- Modify: `PROGRESS.md`
- Modify: `docs/specs/v137_spec-bot-assertion-domain-split.md`
- Modify: `docs/plans/v137_2026-06-13-bot-assertion-domain-split.md`

- [x] Step 3.1: Record v137 completion and next-slice pointer.
- [x] Step 3.2: Mark spec and plan complete.

```bash
make maintainability
make ci
```

## Final verification

- [x] `.venv/bin/pytest tools/bot/test_protocol.py tools/bot/test_stash_assertions.py tools/bot/test_item_assertions.py -q`
- [x] `make maintainability`
- [x] `make ci`

## Deferred scope

- Broader protocol bot action extraction.
- Client bot runner splits.
- New unique/chest gameplay assertions beyond the extracted helper coverage.
