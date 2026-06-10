# Spec: `equipment-requirements-and-preview`

Status: Draft
Date: 2026-06-09
Branch: `main`
Slice: v43 - equipment requirements and preview
Baseline: v42 `vendor-appraisal-and-item-comparison`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, deterministic replay
- [`../adr/0008-world-structure-and-dungeon-progression.md`](../adr/0008-world-structure-and-dungeon-progression.md) - character-scoped progression
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - client UI shortcut checklist
- [`v23_spec-item-templates-and-rolled-drops.md`](v23_spec-item-templates-and-rolled-drops.md) - template requirements and rolled payloads
- [`v26_spec-character-stats-and-leveling.md`](v26_spec-character-stats-and-leveling.md) - durable base stats and derived stat views
- [`v42_spec-vendor-appraisal-and-item-comparison.md`](v42_spec-vendor-appraisal-and-item-comparison.md) - server-authored shop appraisals and direct comparisons

## 1. Purpose

v42 makes vendor offers and sell rows understandable, but equipment requirements are still a
placeholder: item templates only allow `level`, the Go rules loader rejects levels above `1`, and
equip checks do not consider character stats. This slice turns requirements into a real
server-authoritative progression gate and extends item previews so the player can tell why an item
can or cannot be equipped before spending stat points.

Equipment templates can require `level`, `str`, `dex`, `vit`, or `magic`. The server validates and
enforces those requirements at equip time. Unmet gear may still be bought, picked up, sold, and
stored in the bag, but equip rejects with `requirements_not_met` without mutating equipment,
inventory, hotbar, or derived stats. Once the character satisfies the same item requirements through
leveling or stat allocation, equip succeeds through the normal authoritative path.

The server also emits presentation-ready requirement status and direct derived-stat previews for
inventory rows, loot item metadata, shop offers, and sell appraisals. The Godot client only renders
this data; it does not decide whether an item is usable.

## 2. Non-goals

- No affix grammar, procedural item names, unique/set catalog, crafting, repair, buyback, stash, or
  trade.
- No active skills, mana spenders, mana regeneration, attack-speed gameplay, or skill bar behavior.
- No broad item/economy rebalance. Add only the minimum non-trivial requirement fixtures needed for
  the proof.
- No main-menu character summaries or old-session resume UI.
- No production item art, inventory art, vendor art, animation, VFX, or audio pass.
- No external inventory/shop plugin adoption. The plan must record the adoption checklist result;
  the expected default is to extend the in-repo inventory and shop panels.

## 3. Acceptance Criteria

1. Shared item template requirements accept `level`, `str`, `dex`, `vit`, and `magic`, and shared
   validation rejects unsupported keys, negative values, and invalid requirement fixtures.
2. At least one existing generated equipment template has a non-trivial requirement that a fresh
   character does not satisfy, while existing bot-critical templates remain equippable by default.
3. `equip_intent` rejects unmet requirements with `requirements_not_met` before any equipment,
   inventory capacity, hotbar, HP, mana, or derived-stat mutation.
4. After the same character satisfies the requirement through existing stat allocation/leveling
   paths, the same item equips successfully.
5. Item instance views expose server-authored requirement status: each requirement's required value,
   current character value, and whether it is met.
6. Server-authored inventory/shop previews include direct derived-stat deltas against the current
   character state when the item would be equipped, including support for hand-occupancy slots.
7. `shop_opened` still lets players buy unmet gear; purchase authority, gold checks, capacity checks,
   sell appraisals, and v42 direct item comparisons continue to work.
8. Godot inventory and shop rows render requirement status and preview/comparison lines without text
   overlap on the existing panels.
9. Protocol bot proof covers pickup/buy of unmet gear, equip rejection, stat allocation, equip
   success, `/state`, reconnect, replay, and fresh-session persistence.
10. Shared validation, Go tests, client unit tests, protocol bot, client bot, and `make ci` pass.

## 4. Scope And Likely Files

```text
docs/specs/v43_spec-equipment-requirements-and-preview.md - this spec
docs/plans/v43_2026-06-09-equipment-requirements-and-preview.md - implementation plan
PROGRESS.md - lifecycle update when v43 ships

shared/rules/item_templates.v0.schema.json - broaden requirements
shared/rules/item_templates.v0.json - add minimal non-trivial requirement
shared/protocol/state_delta.v5.schema.json - requirement status and preview payloads if v4 cannot be extended cleanly
shared/protocol/session_snapshot.v4.schema.json - same item view payloads if needed
shared/protocol/examples/state_delta.json - update detailed item/shop example
shared/protocol/examples/session_snapshot.json - update inventory item example
shared/golden/equipment_requirements.json - deterministic requirement/preview fixture
shared/golden/equipment_requirements.v0.schema.json - fixture schema
tools/validate_shared.py - validate requirements and golden fixture

server/internal/game/types.go - requirement status and stat preview protocol views
server/internal/game/rules.go - parse/validate requirement keys and fixture drift
server/internal/game/sim.go - enforce requirements and emit item preview views
server/internal/game/shop.go - include requirement status and equip preview in shop rows
server/internal/game/game_test.go - equip reject/success, no mutation, preview coverage
server/internal/game/shop_test.go - shop unmet gear and preview coverage

tools/bot/run.py - assertions for requirement status and preview deltas
tools/bot/scenarios/31_equipment_requirements_and_preview.json - protocol proof

client/scripts/inventory_panel.gd - render requirement and preview lines
client/scripts/shop_panel.gd - render requirement and preview lines
client/tests/test_inventory_panel.gd - row detail tests if available
client/tests/test_shop_panel.gd - shop requirement/preview row tests
client/scripts/bot_controller.gd, client/scripts/bot_scenario_runner.gd - client bot assertions if needed
tools/bot/scenarios/client/17_equipment_requirements_and_preview.json - client UI proof
```

Protocol note: the default is to use v5 schema files if requirement-status or preview payloads are
not a clean additive extension of v4. No new client intent is expected; equip continues through the
existing `equip_intent`.

## 5. Data Shape Draft

Item requirements remain part of the rolled item payload:

```json
{
  "requirements": {
    "level": 2,
    "str": 6
  }
}
```

Item views and shop rows gain optional server-authored requirement status:

```json
{
  "requirements": { "level": 2, "str": 6 },
  "requirement_status": [
    { "stat": "level", "required": 2, "current": 1, "met": false },
    { "stat": "str", "required": 6, "current": 5, "met": false }
  ],
  "requirements_met": false
}
```

Equipment previews use direct derived-stat deltas from the current character state:

```json
{
  "equip_preview": {
    "slot": "main_hand",
    "requirements_met": false,
    "deltas": [
      { "stat": "damage_min", "current": 4.0, "preview": 7.0, "delta": 3.0 },
      { "stat": "damage_max", "current": 7.0, "preview": 11.0, "delta": 4.0 }
    ]
  }
}
```

Direct v42 item comparison can stay in shop rows. Equip preview answers a different question:
"what would my character-derived stats look like if this item were equipped?" It must be computed on
the server from the same stat path used by combat.

## 6. Test And Bot Proof

- `make validate-shared` validates broadened requirement schemas and the new equipment requirement
  golden fixture.
- Go tests prove rule validation, unmet equip rejection, no mutation on rejection, success after
  stat allocation, preview deltas, and shop rows for unmet gear.
- `make bot scenario=equipment_requirements_and_preview` proves protocol-visible requirement status,
  reject, allocation, equip success, `/state`, reconnect, replay, and fresh-session persistence.
- `make client-unit` covers inventory/shop rendering helpers for requirement and preview lines.
- `HEADLESS=1 make bot-client scenario=17_equipment_requirements_and_preview.json` proves the real
  Godot client exposes requirement/preview rows.
- `make ci` is the final gate.

## 7. Resolved Questions And Risks

| # | Question / risk | Decision |
|---|-----------------|----------|
| Q-1 | Should unmet gear still be purchasable? | Yes. Only equip is gated. |
| Q-2 | Which requirements ship first? | Support `level`, `str`, `dex`, `vit`, and `magic`; use one or two non-trivial template examples. |
| Q-3 | Protocol version bump? | Use v5 if the schema change is not a clean v4 extension. |
| Q-4 | Preview scope? | Inventory and shop only; no main-menu summaries. |
| R-1 | Requirement changes can break older scenarios that equip generated gear. | Keep bot-critical templates equippable by default and isolate the new unmet proof to one pinned template/item. |
| R-2 | Client could accidentally duplicate authority. | Server emits `requirements_met`, per-stat status, and preview deltas; client renders only. |
