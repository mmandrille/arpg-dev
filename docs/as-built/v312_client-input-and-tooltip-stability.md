# v312 As-Built - Client Input And Tooltip Stability

Status: Complete
Date: 2026-06-20

## Summary

- Added `TextInputFocusGuard` and used it from the main client input loop so focused `LineEdit` or `TextEdit` controls consume text-entry keystrokes without also moving the player.
- Added `TooltipMouseGuard` and applied it to item tooltip panel trees so tooltip controls cannot steal mouse hover from their owning item slot.
- Expanded `InventoryRenderGuard` to include stash, container, filter, sort, and unique chest tab state.
- Updated stash, corpse, account stash, and unique chest rendering to skip unchanged renders and mark post-render state, matching the inventory tooltip-stability pattern.
- Added focused client coverage for text input focus detection, account stash stable slots, unique chest stable slots, and recursive tooltip mouse-ignore behavior.
- Added the text input focus guard test to the client unit smoke gate.

## Verification

```bash
godot --headless --path client --script res://tests/test_text_input_focus_guard.gd
godot --headless --path client --script res://tests/test_stash_panel.gd
godot --headless --path client --script res://tests/test_inventory_panel.gd
make client-unit
make maintainability
```

All checks passed on 2026-06-20.
