# v310 As-Built - Derived Stat Breakdown UI

Status: Complete
Date: 2026-06-20

## Summary

- Wrapped the character panel's derived-stat rows in a Godot `ScrollContainer`, with at least six visible rows and the vertical scrollbar reserved on the right.
- Made derived stats permanently visible under a plain title and preserved existing fraction/whole percent formatting.
- Expanded derived stat hover text into an opaque custom tooltip with formula-only content: one contribution per line, source context in parentheses, and final value/cap at the bottom.
- Added readable source-category labels for item bases, item rolls, skill effects, passives, set bonuses, buffs, debuffs, caps, clamps, and unknown future source kinds.
- Made derived-stat labels explicit mouse-hover targets so Godot custom tooltips appear over the rows without the default transparent tooltip style.
- Added a focused Godot stats-panel test and wired it into `make client-unit`.

## Verification

```bash
godot --headless --path client --script res://tests/test_character_stats_panel.gd
make client-unit
make maintainability
```

All three checks passed on 2026-06-20.
