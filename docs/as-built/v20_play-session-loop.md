# v20 — Play session loop

**Proves:** The generated dungeon can be entered from a static town and used as the default fresh
interactive play loop without changing the authoritative client/server boundary.

- `dungeon_levels` now starts at town level `0`, built from `worlds.v0.json` with a down stair and
  a teleporter; level `-1` is generated lazily on first descent.
- Town teleporter discovery is initialized server-side and appears in snapshots as level `0`
  discovered; protocol v1 now allows `target_level: 0`.
- Generated level `-1` now has a `stairs_up` landing at the dungeon player spawn, so
  `0 -> -1 -> -2 -> -1 -> 0` is replayable and golden-tested.
- `ascend_intent` from level `-1` returns to town at the town down stair; teleporting to town lands
  at the town teleporter when the current floor has an active discovered teleporter.
- `scripts/play.sh` launches a fresh `dungeon_levels` run by default, and the interactive Godot
  client requests `dungeon_levels` when no world is specified.
- Godot renders dungeon perimeter walls only below level `0`; town remains an open placeholder hub
  with the existing waypoint panel and level-HUD behavior.
- Bot scenarios `12_dungeon_levels` and `13_teleporter_lab`, replay goldens, and client golden
  checks were updated for the town preamble and town waypoint row.

**Explicit non-goals:** no character-scoped persistence, player-facing resume, safe-zone combat
rules, NPCs/vendors/stash, production town art, or plugin adoption.
