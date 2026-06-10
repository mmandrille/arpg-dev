# v10 — Click action and melee range

**Proves:** A single left-click action can cover combat, loot pickup, and interactable activation
while the server enforces melee reach and mutable world object state deterministically.

- `action_intent { target_id }` replaces active `attack_intent` / `pick_up_intent` protocol use.
- Shared combat/item rules define `combat.unarmed_reach` and weapon `reach`; Go and GDScript
  consume `shared/golden/melee_reach.json`.
- Server rejects in-world actionable targets beyond reach with `out_of_range`.
- `wooden_door` interactables spawn from shared rules, block movement while closed, open through
  an authoritative action, emit `interactable_activated`, and unblock passage.
- Godot left-click ray-picks monsters, loot, and doors through per-entity pick colliders; doors are
  rendered as simple in-repo panels that tween open from authoritative state.
- Bot scenarios `01`-`03` now use action steps; `04_door_lab` proves far reject, door open,
  passage, loot pickup, reconnect resume, and replay.
- `make ci` green on 2026-06-05.

**Explicit non-goals (still true):** no click-to-move, pathfinding, ranged weapons, key/lock
puzzles, door closing, inventory UI, or production door art.
