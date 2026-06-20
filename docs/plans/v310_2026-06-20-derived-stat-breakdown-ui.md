# v310 Plan - Derived Stat Breakdown UI

Status: Complete
Goal: Make the character panel derived-stat list scrollable and make hover breakdowns readable for every source row.
Architecture: Keep the server-owned `character_progression.stat_breakdowns` contract unchanged. The Godot panel only renders those rows with a permanent scrollable list and an opaque formula-only custom tooltip.
Tech stack: Godot 4 GDScript panel/test plus the existing client smoke gate.

## Baseline and asset/plugin decision

Builds on v309. Asset/plugin decision: reject external UI plugins; borrow Godot's built-in `ScrollContainer` and custom tooltip hook.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/character_stats_panel.gd` | Scrollable derived stat block and breakdown tooltip formatting |
| Create | `client/tests/test_character_stats_panel.gd` | Focused panel coverage |
| Modify | `scripts/client_smoke.sh` | Add the focused test to `make client-unit` |
| Create | `docs/specs/v310_spec-derived-stat-breakdown-ui.md` | Slice spec |
| Create | `docs/as-built/v310_derived-stat-breakdown-ui.md` | Completion proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/tests/test_coop_client.gd` expectation update only; final line count stayed below baseline.

## Task 1 - Stats panel UI

Files:
- Modify: `client/scripts/character_stats_panel.gd`

- [x] Put derived stat labels in a `ScrollContainer` with the vertical scrollbar reserved on the right.
- [x] Expose scroll/debug metadata from `get_debug_state`.
- [x] Replace compact breakdown tooltip text with an opaque formula-only tooltip.

## Task 2 - Focused client test

Files:
- Create: `client/tests/test_character_stats_panel.gd`
- Modify: `scripts/client_smoke.sh`

- [x] Prove permanent derived-stat visibility, percent formatting, scroll state, and source-rich tooltip content.
- [x] Wire the new test into the client unit suite.

## Final verification

- [x] `godot --headless --path client --script res://tests/test_character_stats_panel.gd`
- [x] `make client-unit`
- [x] `make maintainability`
