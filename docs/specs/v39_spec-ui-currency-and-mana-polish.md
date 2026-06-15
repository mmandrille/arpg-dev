# Spec: `ui-currency-and-mana-polish`

Slice: v39 - currency, mana, and UI polish follow-up

## Goals

- Treat coins as character gold, not inventory items.
- Add a single durable currency field, `gold`, with no other money types.
- Add player `mana` / `max_mana` to snapshots and deltas, plus a blue mana potion consumable.
- Change armor's character-derived formula source from VIT to DEX.
- Polish Godot UI: bigger interface text, mana bar, gold counter, mutual-exclusive panels, character rename affordance, sword hand pose, and a settings toggle for the top-right debug text.

## Contract

- `session_snapshot` gains top-level `gold`.
- Player entity views gain `mana` and `max_mana`.
- `state_delta` gains `gold_update`.
- Currency loot uses `item_def_id: "gold"` and a rolled `amount`; pickup removes the loot and emits `gold_picked_up`.
- Gold min/max rolls scale upward for generated dungeon depth and monster rarity. Static/town rewards use the base gold range.
- Consumable item definitions may declare `mana_restore`.