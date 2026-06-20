# v312 Spec - Client Input And Tooltip Stability

Status: Complete
Date: 2026-06-20
Codename: client-input-and-tooltip-stability

## Purpose

Stop gameplay movement keys from leaking through focused text inputs, and make inventory-style item tooltips stable across stash, account stash, unique chest, and related panels.

## Non-goals

- No gameplay input remapping or new control settings.
- No tooltip visual redesign.
- No new external UI plugins or assets; adopt/reject decision: reject external UI plugins and use small shared Godot helper classes for focus and tooltip mouse filtering.

## Scope And Likely Files

- `client/scripts/text_input_focus_guard.gd` - shared helper to detect focused `LineEdit` or `TextEdit` controls.
- `client/scripts/main.gd` - suppress hotkeys and continuous movement while text input owns focus.
- `client/scripts/tooltip_mouse_guard.gd` - shared helper to make tooltip trees ignore mouse input.
- `client/scripts/item_tooltip_panel.gd` - apply recursive mouse-ignore behavior to item tooltips.
- `client/scripts/inventory_render_guard.gd` - include stash/container/filter state in stable render keys.
- `client/scripts/stash_panel.gd` - reuse the render guard for account stash, corpses, and unique chest views.
- `client/scripts/inventory_panel.gd` - mark post-render inventory state so stable renders do not rebuild hovered slots.
- `client/tests/test_text_input_focus_guard.gd`, `client/tests/test_stash_panel.gd`, and `client/tests/test_inventory_panel.gd` - focused client proof.

## Acceptance Criteria

- When a `LineEdit` or `TextEdit` has focus, `WASD` and other gameplay key handling do not move the player or trigger gameplay hotkeys.
- Item tooltips recursively ignore mouse input so the tooltip cannot steal hover from the item slot that owns it.
- Account stash and unique chest panels do not rebuild unchanged item slot controls while hovering.
- Existing inventory tooltip stability remains covered.

## Testing Plan

- Run focused Godot tests:
  ```bash
  godot --headless --path client --script res://tests/test_text_input_focus_guard.gd
  godot --headless --path client --script res://tests/test_stash_panel.gd
  godot --headless --path client --script res://tests/test_inventory_panel.gd
  ```
- Run the client test gate:
  ```bash
  make client-unit
  ```
- Run maintainability:
  ```bash
  make maintainability
  ```
