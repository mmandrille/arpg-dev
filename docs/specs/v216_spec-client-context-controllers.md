# v216 — Client Context-Object Extraction

**Status:** Draft  
**Date:** 2026-06-16  
**Codename:** client-context-controllers

---

## Purpose

Establish the **typed context object** pattern and use it to extract two stateful domains from
`main.gd`:

1. **`TownNodeFactory`** — pure mesh construction for all town/dungeon interactable nodes
   (door, chest, merchant, bishop, blacksmith, market board, stair, teleporter, town cabin,
   campfire, market badge).
2. **`BossVisualsController`** — boss health bar management, phase display, telegraph marker
   synchronisation, and boss entity queries.

**Pattern definition:** Every extracted controller receives a narrow, typed context object at
construction time instead of a reference to `main.gd`. The context holds only the state the
controller needs. This makes each controller independently `preload`-able and unit-testable.

```
main.gd  →  creates context  →  controller.new(ctx)  →  operates on ctx fields
```

No `helpers=globals()`. No passing `self` as a generic `Node`. The context IS the interface
contract between main.gd and the controller.

**Maintainability obligation:** After v215, `main.gd` should be ≈ 6188 lines. This slice
targets a further reduction of ≈ 450 lines → `main.gd` ≈ 5738. Update
`.maintainability/file-size-baseline.tsv` to the new actual count.

---

## Non-goals

- Input handling, skill system, menu/character flow, combat effects, and visual replay
  extraction are explicitly deferred. Each requires a larger context surface; they are
  the next candidates once this pattern is proven.
- No change to game logic, rendering output, or observable server behaviour.
- No changes to Go server, shared rules, or Python tools.
- No new Godot autoloads.

---

## Typed context pattern (normative)

Each context is a `class_name Foo extends RefCounted` holding only the fields that domain
needs. `main.gd` constructs the context from its own vars and passes it to the controller:

```gdscript
# boss_visuals_context.gd
class_name BossVisualsContext extends RefCounted
var entities: Dictionary          # read-only access to entity records
var asset_manifest: Dictionary    # for model/socket lookups
var health_bars_layer: CanvasLayer
```

```gdscript
# in main.gd _ready():
var _ctx_boss := BossVisualsContext.new()
_ctx_boss.entities = entities
_ctx_boss.asset_manifest = asset_manifest
_ctx_boss.health_bars_layer = health_bars_layer
_boss_visuals = BossVisualsController.new(_ctx_boss, boss_health_bar)
```

**Rules for context objects:**

1. A context object is a plain data carrier — no methods, no logic.
2. It holds references (Dictionary, Node), not copies. The controller reads live state.
3. It must not hold a reference to `main.gd` itself or any parent node.
4. Each context type is narrow: include only what the controller provably needs.
5. Context objects have their own files (`*_context.gd`) so they can be `preload`ed in tests
   independently of the controller.

---

## Acceptance criteria

1. `make ci` green after extraction.
2. `main.gd` line count ≤ the post-v215 baseline; `.maintainability/file-size-baseline.tsv`
   updated to new actual count.
3. `TownNodeFactory` and `BossVisualsController` files are each ≤ 400 lines.
4. Context files (`BossVisualsContext`) are ≤ 30 lines each.
5. No `main.gd` symbol is accessible from any new file except through its typed context
   parameter, `ItemRulesLoader` singleton, or `ClientConstants`.
6. A headless unit test can `preload` each new file directly without `main.gd` in scope and
   invoke at least one method. `BossVisualsContext` must be constructable with a mock dict.
7. `make bot` and `make client-smoke` pass.
8. `make bot-visual` completes without visual regression (boss health bar, telegraph marker,
   town node appearances unchanged).

---

## Scope and files touched

### New files

| File | `class_name` | Contents |
|------|-------------|----------|
| `client/scripts/town_node_factory.gd` | `TownNodeFactory` | `make_door_node`, `make_chest_node`, `make_merchant_node`, `make_bishop_node`, `make_blacksmith_node`, `make_market_board_node`, `make_town_preview_scene`, `make_town_cabin_node`, `make_town_campfire_node`, `make_market_badge`, `add_merchant_box`, `add_merchant_cylinder`, `merchant_material`, `make_stair_node`, `add_stair_box`, `stair_base_color`, `stair_material`, `make_teleporter_node`. Constructor: no params (all functions are pure mesh builders taking explicit input params). |
| `client/scripts/boss_visuals_context.gd` | `BossVisualsContext` | Data carrier: `entities: Dictionary`, `asset_manifest: Dictionary`, `health_bars_layer: CanvasLayer`. |
| `client/scripts/boss_visuals_controller.gd` | `BossVisualsController` | `hide_boss_health_bar`, `sync_boss_health_bar`, `advance_boss_phase_display`, `boss_phase_for_display`, `active_boss_entity_id`, `boss_health_bar_title`, `apply_boss_phase_started`, `apply_boss_phase_ended`, `normalize_boss_phase_metadata`, `sync_boss_telegraph_marker_from_record`, `sync_boss_telegraph_marker`, `remove_boss_telegraph_marker`. Constructor: `BossVisualsController.new(ctx: BossVisualsContext, health_bar: BossHealthBar)`. |

### Modified files

| File | Change |
|------|--------|
| `client/scripts/main.gd` | Add `var _town_factory: TownNodeFactory` (initialized in `_ready`); replace inline `_make_*` calls with `_town_factory.make_*`; add `var _boss_visuals: BossVisualsController` with context wired in `_ready`; replace inline boss-visual calls with `_boss_visuals.*`; remove extracted functions; update `.maintainability/file-size-baseline.tsv` |
| `.maintainability/file-size-baseline.tsv` | Lower `client/scripts/main.gd` entry |

### Unchanged

- All Go server, shared, and Python files.
- `_attach_pick_collider` stays in `main.gd` — called from `_upsert_entity`, not from any
  factory function. Confirmed: the `_make_*` functions do NOT call `_attach_pick_collider`.
- `_set_interactable_state` / `_apply_interactable_state_tint` stay in `main.gd` — they use
  `create_tween()` (Node method on self) and operate on live entity records, not on node
  construction.

---

## TownNodeFactory design notes

All `_make_*` functions in this domain are pure node constructors: they take explicit params
(def_id, flags), build mesh primitives, and return `Node3D`. They have **no implicit dependencies
on main.gd's state**. Confirmed by code inspection:

- `ChestPresentationScript.add_part` is already a class-level call — remains as-is.
- `_res_path` is used by loot factory code (v215) — if TownNodeFactory also needs it,
  promote it to `ClientConstants` or a shared `PathUtils` helper rather than duplicating.
- `_make_market_badge` creates Label3D nodes with text — no `self` dependency.
- `make_town_preview_scene` returns a Node3D subtree — stays pure.

Because `TownNodeFactory` has no state, it MAY be implemented as static methods (Godot 4
supports `static func`) rather than an instantiated class, allowing calls like
`TownNodeFactory.make_door_node()` without a `var _town_factory` instance. Plan decides.

---

## BossVisualsController design notes

`BossVisualsController` is stateful — it holds the `boss_health_bar` node reference and reads
`entities` live from the context dictionary.

**`active_boss_entity_id()`** iterates `ctx.entities` to find the live boss. Because this is
a Dictionary read (not a range+mutate), it does not violate Go-sim determinism rules (client-only).

**`advance_boss_phase_display(delta)`** is called from `main.gd`'s `_process`. Main.gd calls
`_boss_visuals.advance_boss_phase_display(delta)` instead of the inline call.

**`apply_boss_phase_started` / `apply_boss_phase_ended`** are called from `_apply_delta` event
dispatch. Main.gd calls `_boss_visuals.apply_boss_phase_started(entity_id, ev)`.

**`boss_health_bar_title`** uses `_item_definition` (currently in main.gd, moving to
`LootNodeFactory` in v215). Options: (a) pass item_rules dict via context so
`BossVisualsController` can look up definitions itself, (b) keep a `get_item_definition`
callable on the context. Plan decides — option (a) (add `item_rules: Dictionary` to
`BossVisualsContext`) is simpler and keeps the context as a plain data carrier.

---

## Future extraction roadmap (non-normative)

This pattern, once proven in v216, unlocks:

| Domain | Context needs | Estimated savings |
|--------|--------------|-------------------|
| Input handler | `entities`, `predicted_pos`, `player_id`, `player_anchor`, `_camera`, skill cast callable | ~380 lines |
| Skill system | `skill_progression`, `skill_cooldowns`, `equipped`, `skill_function_keys`, UI panel refs | ~280 lines |
| Combat effects layer | `entities`, `damage_numbers_layer`, `health_bars_layer`, `player_anchor` | ~250 lines |
| Menu / character flow | `gameplay_active`, `character_flow`, session-start callables | ~230 lines |
| Visual replay | `_apply_snapshot` / `_apply_delta` callables, `client`, `visual_replay_*` vars | ~130 lines |

These are explicitly **not in scope for v216** but are the natural sequence once the pattern
is established. The roadmap table should be referenced in the v217+ next-slice brief.

---

## Test and bot proof

- **Headless unit:** Extend `client/tests/test_factories.gd` (from v215) with:
  - `TownNodeFactory.make_door_node()` returns a Node3D with expected child names.
  - `TownNodeFactory.make_chest_node("town_stash")` returns a node with "ChestLidPivot".
  - `BossVisualsContext.new()` constructs without error with mock dicts.
  - `BossVisualsController.new(mock_ctx, null)` constructs without error (null health bar is
    valid for unit test — real bar is wired in gameplay).
- **Smoke:** `make client-smoke` passes.
- **Bot:** `make bot` passes. Boss-fight scenarios (if any in the bot catalog) exercise the
  controller indirectly.
- **Visual replay:** `make bot-visual` passes. Boss health bar and telegraph markers appear
  correctly if a boss-floor replay scenario is available.

---

## Open questions and risks

1. **Static vs instance for TownNodeFactory.** Static methods avoid the `var _town_factory`
   boilerplate but cannot hold state. Since town node construction is stateless, static is
   correct — but Godot 4's static func support and any limitations with `preload` in headless
   tests should be verified before the plan commits to this.

2. **`item_rules` on BossVisualsContext.** Boss health bar title lookup needs item definitions.
   Adding `item_rules: Dictionary` to the context (backed by `ItemRulesLoader.item_rules`) is
   the lowest-friction solution. Confirm during planning that `ItemRulesLoader` is already loaded
   before `_boss_visuals` is first used.

3. **Context object mutation safety.** `entities` is a Dictionary held by reference — the
   controller reads live state, which is correct, but the context must not write to `entities`
   (entity addition/removal is main.gd's responsibility). Document this as a comment in
   `BossVisualsContext`.

4. **Telegraph marker node ownership.** `_sync_boss_telegraph_marker` adds/removes a Node3D
   child on the entity's node inside `entities`. This is a write to an entity's scene subtree,
   not to the `entities` dict itself — acceptable for a display-only controller, but note it
   in the as-built.
