# Spec: `use-consumable`

Status: Draft
Branch: `feature/use-consumable`
Slice: v16 — authoritative consumable use + bottom-center hotbar (keys 1–0)
Baseline: v15 `item-visuals-and-loot-presentation` (complete on `main`, `make ci` green)
Related:

- [`v15_spec-item-visuals-and-loot-presentation.md`](v15_spec-item-visuals-and-loot-presentation.md)
- [`v13_spec-inventory-ui.md`](v13_spec-inventory-ui.md) — inventory intents pattern
- [`v4_spec-take-a-hit.md`](v4_spec-take-a-hit.md) — player HP / retaliation
- [`../../PROGRESS.md`](../../PROGRESS.md)

## 1. Purpose

Let players **use consumable items** from inventory through a **bottom-center hotbar**
with keyboard shortcuts **1–9** and **0** (slot 10). The server owns heal outcomes;
the hotbar is client-only assignment UI.

After this slice:

- `use_intent { item_instance_id }` consumes one bag row and applies heal from item rules.
- `red_potion` declares `heal: { min: 5, max: 5 }` in `shared/rules/items.v0.json`.
- Server emits `item_used` + `player_healed`; player HP updates via `entity_update`.
- Rejects: `not_in_inventory`, `not_consumable`, `not_usable`, `already_full_hp`, `player_dead`.
- `heal_lab` world + bot scenario `08_heal_lab.json` proves pickup → damage → use → full HP.
- Godot `ConsumableBar` (10 slots, bottom center): drag consumables from bag to assign; press 1–0 to use.
- Client bot scenario exercises hotbar assign + key press path.

## 2. Non-goals

- No stash, vendors, crafting, stack splitting, or consumable use on equipped items.
- No mouse-click use in the 3D world (hotbar keys + drag-assign only for v16).
- No heal-over-time, buffs, or multi-effect consumables beyond flat heal.
- No production art, new animation clips, or server-side hotbar persistence.

## 3. Acceptance criteria

1. `make validate-shared` and `make ci` green.
2. Bot `08_heal_lab.json` passes with reconnect + replay.
3. Client bot hotbar scenario passes headless.
4. Hotbar assignment is client-only; use always sends `use_intent`.
