# v204 Plan — Set Collection Panel

Status: Complete
Goal: Add a client set collection panel that summarizes owned/equipped set progress from existing item summary payloads.
Architecture: Keep set collection as presentation-only client state derived from item rows already sent by the server. Put parsing, aggregation, and panel rendering in a focused `set_collection_panel.gd` file so the over-limit inventory coordinator only owns a small integration button. Bot proof reads panel debug state rather than adding server/protocol contracts.
Tech stack: Godot GDScript client UI/tests, Godot client bot scenario, docs.

## Baseline and shortcut decision

Builds on v181/v194 set item payloads, v195 set drops, and v203 clean CI. No shared/server/store/protocol work is planned. Godot plugin/assets decision: reject external plugins/assets because existing Control widgets and set summary lines are enough for a text/progress panel.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `client/scripts/set_collection_panel.gd` | Parse set summary lines, aggregate owned/equipped pieces, render the panel, expose debug state. |
| Modify | `client/scripts/inventory_panel.gd` | Add a compact set collection button, feed current inventory/equipment rows into the panel, expose debug state. |
| Add | `client/tests/test_set_collection_panel.gd` | Unit coverage for parsing, owned/equipped states, bonus active flags, and refresh behavior. |
| Add | `client/tests/test_set_collection_panel.gd.uid` | Godot test metadata. |
| Modify | `scripts/client_smoke.sh` | Add the focused GDScript test to client smoke. |
| Modify | `client/scripts/bot_ui_assertion_handlers.gd` | Verify set collection state through existing inventory panel assertion. |
| Modify | `client/scripts/bot_action_step_validator.gd` | Allow the bot-only unique chest set-tab selector mode. |
| Modify | `client/scripts/bot_facade.gd` | Reuse the stash sort bot action to select the unique chest set tab. |
| Add | `tools/bot/scenarios/client/46_set_collection_panel.json` | Live client proof through the unique chest/inventory flow. |
| Modify | `PROGRESS.md` | Mark v204 complete when shipped. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v204 lifecycle row. |
| Add | `docs/as-built/v204_set-collection-panel.md` | Capture shipped behavior and proof. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/inventory_panel.gd`
- [x] `client/scripts/bot_scenario_runner.gd` not touched after planning; set collection assertion landed in smaller handler files.
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected.
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice: `client/scripts/set_collection_panel.gd` owns parsing, aggregation, rendering, and debug state. If integration grows an over-limit file, remove equivalent local helper/detail lines or keep the net delta within the ratchet allowance.

Verification:
```bash
make maintainability
```

## Task 1 — Focused set collection panel

Files:
- Create: `client/scripts/set_collection_panel.gd`
- Create: `client/tests/test_set_collection_panel.gd`
- Create: `client/tests/test_set_collection_panel.gd.uid`
- Modify: `scripts/client_smoke.sh`

- [x] Step 1.1: Implement summary-line parsing for `Set: <name> (X/Y equipped)` and `<N>-piece set bonus: ... (active|inactive)`.
- [x] Step 1.2: Aggregate rows into per-set progress with piece display names, owned/equipped/missing state, owned/equipped counts, total pieces, and bonus rows.
- [x] Step 1.3: Render a compact panel using existing Godot controls/colors and expose `get_debug_state()`.
- [x] Step 1.4: Add focused unit tests and wire them into client smoke.
```bash
make client-unit
```

## Task 2 — Inventory integration

Files:
- Modify: `client/scripts/inventory_panel.gd`

- [x] Step 2.1: Add a small set collection control near inventory metadata/actions.
- [x] Step 2.2: Feed current inventory and equipped item dictionaries into `SetCollectionPanel` after `set_inventory_state()` and render updates.
- [x] Step 2.3: Expose set collection debug state through `InventoryPanel.get_debug_state()`.
- [x] Step 2.4: Keep the integration narrow enough for the file-size ratchet.
```bash
make client-unit
make maintainability
```

## Task 3 — Client bot proof

Files:
- Modify: `client/scripts/bot_ui_assertion_handlers.gd`
- Modify: `client/scripts/bot_action_step_validator.gd`
- Modify: `client/scripts/bot_facade.gd`
- Add: `tools/bot/scenarios/client/46_set_collection_panel.json`

- [x] Step 3.1: Add a client bot assertion/action that can verify set collection state by set display name, owned/equipped counts, and bonus row state.
- [x] Step 3.2: Add a scenario that opens the town unique chest, takes one set piece, and verifies owned progress in inventory set collection debug state.
```bash
SCENARIO=set_collection_panel HEADLESS=1 ./scripts/bot_client_local.sh
```

## Task 4 — Lifecycle docs and CI

Files:
- Modify: `docs/specs/v204_spec-set-collection-panel.md`
- Modify: `docs/plans/v204_2026-06-15-set-collection-panel.md`
- Add: `docs/as-built/v204_set-collection-panel.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark the spec and plan complete after implementation.
- [x] Step 4.2: Add as-built proof and lifecycle row.
- [x] Step 4.3: Update `PROGRESS.md` latest completed slice and CI gate.
```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make client-unit`
- [x] `SCENARIO=set_collection_panel HEADLESS=1 ./scripts/bot_client_local.sh`
- [x] `make ci`

## Deferred scope

- Structured server set metadata, account-wide collection persistence, set rewards, collection achievements, new set art, and collection window layout persistence remain future itemization/UI work.
