# v228 Plan - Resource Wallet Panel

Status: Complete
Goal: Add a compact HUD wallet readout for account resources.
Architecture: Server-owned wallet balances remain unchanged. The client displays existing
`resource_wallet` state through the current HUD bar and proves it with a client-bot assertion.
Tech stack: Godot GDScript client, client bot scenario proof, SDD docs.

## Baseline and shortcut decision

Reuse the existing `CharacterBar` HUD rather than adding a new window or service panel. Asset/plugin
decision: adopt code-native Godot controls; reject external UI assets/plugins.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/character_bar.gd` | Render wallet rows in the HUD bar |
| Modify | `client/scripts/main.gd` | Pass existing `resource_wallet` state to the HUD |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Add wallet HUD assertion |
| Modify | `client/scripts/bot_step_catalog.gd` | Register assertion type |
| Add | `client/tests/test_character_bar.gd` | Unit proof for wallet render/debug state |
| Modify | `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json` | Live proof for pickup and spend |
| Modify | `PROGRESS.md` | Current status after completion |
| Modify | `docs/progress/slice-lifecycle.md` | Lifecycle row |
| Add | `docs/as-built/v228_resource-wallet-panel.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: main should only pass existing state into an existing HUD
  component; new wallet rendering lives in `character_bar.gd`.

Verification:
```bash
make maintainability
```

## Task 1 - HUD wallet rendering

Files:
- Modify: `client/scripts/character_bar.gd`
- Add: `client/tests/test_character_bar.gd`

- [x] Step 1.1: Render nonzero wallet resources as compact rows and expose debug state.
```bash
make client-unit
```

## Task 2 - State wiring and bot proof

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/bot_assertion_handlers.gd`
- Modify: `client/scripts/bot_step_catalog.gd`
- Modify: `tools/bot/scenarios/client/39_blacksmith_upgrade_ui.json`

- [x] Step 2.1: Sync `resource_wallet` into `CharacterBar` after snapshots/deltas and assert pickup
  plus spend in the blacksmith client scenario.
```bash
make bot-client scenario=39_blacksmith_upgrade_ui.json HEADLESS=1
```

## Task 3 - Lifecycle docs

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Add: `docs/as-built/v228_resource-wallet-panel.md`

- [x] Step 3.1: Record v228 as complete with focused proof and note the final batch CI is pending.
```bash
make maintainability
```

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make bot-client scenario=39_blacksmith_upgrade_ui.json HEADLESS=1`
- [x] Batch-level `make ci` passed after the selected v226-v232 `$autoloop` queue.
