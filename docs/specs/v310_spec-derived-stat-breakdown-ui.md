# v310 Spec - Derived Stat Breakdown UI

Status: Complete
Date: 2026-06-20
Codename: derived-stat-breakdown-ui

## Purpose

Improve the character panel derived-stat section so it remains usable as the stat list grows and each derived stat explains how the final value was produced from server-authored breakdown rows.

## Non-goals

- No server authority, protocol, shared-rule, or formula changes.
- No new external assets, plugins, or UI dependencies; adopt/reject decision: reject external UI plugins and borrow the existing Godot `ScrollContainer` plus custom tooltip behavior.
- No redesign of the full character window or base-stat allocation flow.

## Scope And Likely Files

- `client/scripts/character_stats_panel.gd` - render derived stats inside a right-side-scrollable area and format richer breakdown tooltips.
- `client/tests/test_character_stats_panel.gd` - focused panel proof for scroll/debug state and tooltip content.
- `scripts/client_smoke.sh` - include the focused test in `make client-unit`.

## Acceptance Criteria

- The derived-stat block is permanently visible under a plain title and shows at least six rows inside a scroll container with the vertical scrollbar on the right.
- The derived-stat block keeps existing percent formatting for fraction and whole-percent stats.
- Hovering a derived-stat row with `stat_breakdowns` shows an opaque tooltip with only a readable formula: one source contribution per line, source context in parentheses, and final value/cap at the bottom.
- Current and future source kinds such as items, skills, buffs, debuffs, passives, set bonuses, caps, and clamps are categorized in readable text without requiring protocol changes.
- `make client-unit` includes a focused stats-panel gate.

## Testing Plan

- Run the focused Godot panel test:
  ```bash
  godot --headless --path client --script res://tests/test_character_stats_panel.gd
  ```
- Run the client unit suite:
  ```bash
  make client-unit
  ```
