# v312 Plan - Client Input And Tooltip Stability

Status: Complete
Goal: Prevent focused text inputs from leaking keyboard input to gameplay controls and make item tooltip hover stability reusable across inventory-style panels.
Architecture: Keep the behavior client-side. Use shared helper classes for focus detection and tooltip mouse filtering, and extend the existing inventory render guard so stash-like panels can skip unchanged renders while the pointer is over an item slot.
Tech stack: Godot 4 GDScript client scripts and headless panel tests.

## Baseline and asset/plugin decision

Builds on v311. Asset/plugin decision: reject external UI plugins and reuse the existing Godot control tree plus small shared helper scripts.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/text_input_focus_guard.gd` | Detect focused text-entry controls |
| Modify | `client/scripts/main.gd` | Suppress gameplay hotkeys and movement polling while text input owns focus |
| Create | `client/scripts/tooltip_mouse_guard.gd` | Recursively make tooltip controls ignore mouse input |
| Modify | `client/scripts/item_tooltip_panel.gd` | Apply tooltip mouse guard after tooltip content is built |
| Modify | `client/scripts/inventory_render_guard.gd` | Track inventory, stash, container, filter, and tab state in stable render keys |
| Modify | `client/scripts/inventory_panel.gd` | Mark post-render state after guarded inventory renders |
| Modify | `client/scripts/stash_panel.gd` | Use guarded stable renders for stash, corpse, and unique chest views |
| Create | `client/tests/test_text_input_focus_guard.gd` | Focused keyboard-focus proof |
| Modify | `client/tests/test_stash_panel.gd` | Stash and unique chest slot stability proof |
| Modify | `scripts/client_smoke.sh` | Include the new focus guard test in client unit smoke |
| Create | `docs/as-built/v312_client-input-and-tooltip-stability.md` | Completion proof |

## Maintenance ratchet

Target: source/test/tool files stay within the existing ratchet. Keep helper behavior in small class scripts rather than growing `main.gd`, `stash_panel.gd`, or tooltip scripts with duplicated logic.

## Task 1 - Text input focus guard

- [x] Add a reusable focus guard for `LineEdit` and `TextEdit`.
- [x] Skip gameplay key handling in `_unhandled_input` while text entry owns focus.
- [x] Skip continuous movement polling while text entry owns focus.
- [x] Add a focused headless Godot test.

## Task 2 - Tooltip hover stability

- [x] Add a recursive tooltip mouse guard.
- [x] Apply it to item tooltip panels after tooltip content is built.
- [x] Extend the render guard to account for stash/container/filter/tab state.
- [x] Use guarded renders in stash, corpse, and unique chest modes.
- [x] Add tests proving stable slot controls for account stash and unique chest rerenders.

## Final verification

- [x] `godot --headless --path client --script res://tests/test_text_input_focus_guard.gd`
- [x] `godot --headless --path client --script res://tests/test_stash_panel.gd`
- [x] `godot --headless --path client --script res://tests/test_inventory_panel.gd`
- [x] `make client-unit`
- [x] `make maintainability`
