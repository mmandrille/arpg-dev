# v43 Plan — Equipment Requirements and Preview

Status: Complete
Spec: `docs/specs/v43_spec-equipment-requirements-and-preview.md`
Baseline: v42 `vendor-appraisal-and-item-comparison` on `main`
Date: 2026-06-09

## Goal

Make equipment requirements real server-authoritative gates and expose server-authored requirement
status plus equip previews in inventory and shop presentation.

## Architecture

Requirements remain declarative item-template data. The Go sim validates supported requirement keys,
enforces them before equip mutation, and computes all requirement status and derived-stat preview
payloads. The Godot client renders these payloads only; it continues to send existing `equip_intent`,
`shop_buy_intent`, and `shop_sell_intent` messages. Protocol changes are additive to the current
coordinated `state_delta.v4` and `session_snapshot.v4` schemas.

## Tech stack

Shared JSON schemas/rules/goldens, Go authoritative sim/tests, Python protocol bot, Godot GDScript
inventory/shop panels and unit tests, Godot client bot, and lifecycle docs.

## Baseline And Shortcut Decision

This builds directly on v23 item template requirement payloads, v26 durable stats, v31 effective
stat breakdowns, v42 shop appraisals, and existing inventory/shop tooltip panels. Godot plugin
checklist result: **reject external inventory/shop plugin** for this slice. GLoot/Godot-Inventory
would add adapter work while the existing panels already render server-owned item metadata and
tooltips.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/specs/v43_spec-equipment-requirements-and-preview.md` | Slice contract |
| Create | `docs/plans/v43_2026-06-09-equipment-requirements-and-preview.md` | Implementation plan |
| Modify | `shared/rules/item_templates.v0.schema.json` | Broaden requirements to level/base stats |
| Modify | `shared/rules/item_templates.v0.json` | Add a non-trivial requirement fixture |
| Modify | `shared/protocol/state_delta.v4.schema.json` | Requirement status and equip preview fields |
| Modify | `shared/protocol/session_snapshot.v4.schema.json` | Requirement status and equip preview fields |
| Modify | `shared/protocol/examples/state_delta.json` | Detailed requirement/preview example |
| Modify | `shared/protocol/examples/session_snapshot.json` | Inventory requirement/preview example |
| Create | `shared/golden/equipment_requirements.json` | Deterministic requirement fixture |
| Create | `shared/golden/equipment_requirements.v0.schema.json` | Fixture schema |
| Modify | `tools/validate_shared.py` | Validate requirement keys and golden fixture |
| Modify | `server/internal/game/types.go` | Requirement status and equip preview protocol views |
| Modify | `server/internal/game/rules.go` | Requirement validation |
| Modify | `server/internal/game/sim.go` | Requirement enforcement and item view previews |
| Modify | `server/internal/game/shop.go` | Shop requirement status and equip preview payloads |
| Modify | `server/internal/game/game_test.go` | Equip reject/success/no-mutation/preview tests |
| Modify | `server/internal/game/shop_test.go` | Shop unmet gear and preview tests |
| Modify | `tools/bot/run.py` | Requirement status and preview assertions |
| Create | `tools/bot/scenarios/31_equipment_requirements_and_preview.json` | Protocol proof |
| Modify | `client/scripts/inventory_panel.gd` | Render requirement status and equip previews |
| Modify | `client/scripts/shop_panel.gd` | Render requirement status and equip previews |
| Modify | `client/tests/test_shop_panel.gd` | Tooltip/detail coverage |
| Modify | `client/scripts/main.gd` | Preserve requirement/preview payloads on client loot entities |
| Modify | `client/scripts/bot_scenario_runner.gd` | Client bot assertions |
| Modify | `client/tests/test_golden.gd` | GDScript golden consumption for equipment requirements |
| Modify | `tools/bot/scenarios/client/16_vendor_item_comparison.json` | Client vendor requirement/preview assertions |
| Create | `tools/bot/scenarios/client/17_equipment_requirements_and_preview.json` | Client UI proof |
| Modify | `PROGRESS.md` | Lifecycle closeout |

## Task 1 — Shared Rules, Schemas, And Golden Fixture

Files:
- Modify: `shared/rules/item_templates.v0.schema.json`
- Modify: `shared/rules/item_templates.v0.json`
- Modify: `shared/protocol/state_delta.v4.schema.json`
- Modify: `shared/protocol/session_snapshot.v4.schema.json`
- Modify: `shared/protocol/examples/state_delta.json`
- Modify: `shared/protocol/examples/session_snapshot.json`
- Create: `shared/golden/equipment_requirements.json`
- Create: `shared/golden/equipment_requirements.v0.schema.json`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Broaden item-template requirements to support `level`, `str`, `dex`, `vit`, and
  `magic`.
- [x] Step 1.2: Add one non-trivial requirement to a controlled generated template while preserving
  existing bot-critical equip paths.
- [x] Step 1.3: Add `requirement_status`, `requirements_met`, and `equip_preview` schema fields to
  item, loot, shop offer, and sell appraisal views.
- [x] Step 1.4: Add an equipment requirements golden fixture covering unmet requirements, met-after
  allocation, and preview deltas.
- [x] Step 1.5: Extend shared validation for requirement keys and golden fixture drift.

```bash
make validate-shared
```

## Task 2 — Server Requirement Enforcement And Preview Views

Files:
- Modify: `server/internal/game/types.go`
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/shop.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `server/internal/game/shop_test.go`

- [x] Step 2.1: Add protocol structs for per-requirement status and equip-preview stat deltas.
- [x] Step 2.2: Replace the v23 `level <= 1` validation cap with supported-key validation.
- [x] Step 2.3: Enforce all item requirements before equip side effects and preserve old rejection
  semantics for unsupported/non-equippable/wrong-slot items.
- [x] Step 2.4: Compute requirement status from current character level/base stats.
- [x] Step 2.5: Compute direct derived-stat preview deltas through the same effective-stat path used
  by combat.
- [x] Step 2.6: Include requirement status and preview data on inventory item views, loot views,
  shop offers, and sell appraisals.
- [x] Step 2.7: Add Go tests for unmet rejection, no mutation on rejection, success after stat
  allocation, shop purchase of unmet gear, and preview payloads.

```bash
cd server && go test ./internal/game/... -run 'Test.*Requirement|TestShop'
```

## Task 3 — Protocol Bot Proof

Files:
- Modify: `tools/bot/run.py`
- Create: `tools/bot/scenarios/31_equipment_requirements_and_preview.json`

- [x] Step 3.1: Add protocol bot assertions for inventory/shop requirement status and equip preview
  deltas.
- [x] Step 3.2: Add a scenario that obtains unmet gear, proves equip rejection, levels/allocates
  enough stat points, equips successfully, and verifies `/state`, reconnect, replay, and fresh
  session persistence.
- [x] Step 3.3: Keep existing vendor buy/sell and full-equipment scenarios green.

```bash
make bot scenario=equipment_requirements_and_preview
make bot
```

## Task 4 — Godot Inventory And Shop Presentation

Files:
- Modify: `client/scripts/inventory_panel.gd`
- Modify: `client/scripts/shop_panel.gd`
- Modify: `client/tests/test_shop_panel.gd`
- Modify: `client/tests/test_golden.gd`
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/bot_scenario_runner.gd`
- Modify: `tools/bot/scenarios/client/16_vendor_item_comparison.json`
- Create: `tools/bot/scenarios/client/17_equipment_requirements_and_preview.json`

- [x] Step 4.1: Render requirement status lines from server payloads in shared inventory/shop
  tooltip paths.
- [x] Step 4.2: Render equip-preview deltas separately from v42 direct item comparison lines.
- [x] Step 4.3: Expose client debug state/assertions for requirement and preview rows.
- [x] Step 4.4: Add or update Godot unit tests for requirements and previews.
- [x] Step 4.5: Add a client bot scenario proving the real UI exposes unmet requirement and preview
  rows.

```bash
make client-unit
HEADLESS=1 make bot-client scenario=17_equipment_requirements_and_preview.json
```

## Task 5 — Lifecycle Docs And Final Verification

Files:
- Modify: `PROGRESS.md`
- Modify: `docs/plans/v43_2026-06-09-equipment-requirements-and-preview.md`

- [x] Step 5.1: Add v43 lifecycle row, numbering note, summary, bot scenario list, and deferred
  gaps to `PROGRESS.md`.
- [x] Step 5.2: Mark plan checkboxes complete as tasks pass.
- [x] Step 5.3: Run final verification.

```bash
make validate-shared
cd server && go test ./...
make client-unit
make bot scenario=equipment_requirements_and_preview
HEADLESS=1 make bot-client scenario=17_equipment_requirements_and_preview.json
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `cd server && go test ./...`
- [x] `.venv/bin/python -m pytest tools/bot/test_protocol.py`
- [x] `make client-unit`
- [x] `make bot scenario=equipment_requirements_and_preview`
- [x] `make bot scenario=vendor_appraisal_quotes`
- [x] `HEADLESS=1 make bot-client scenario=16_vendor_item_comparison.json`
- [x] `HEADLESS=1 make bot-client scenario=17_equipment_requirements_and_preview.json`
- [x] `make ci`

## Deferred

- Affix grammar, procedural names, unique/set items, item-level economy, crafting, repair, buyback,
  stash, trade, and loot filters.
- Active skills, mana spenders/regeneration, attack-speed gameplay, skill bar behavior, and passive
  skill requirement sources.
- Main-menu character summaries and broader inventory/shop UI plugin adoption.
