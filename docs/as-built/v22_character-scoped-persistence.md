# v22 — Character-scoped persistence

**Proves:** Default-character item instances, equipped weapon state, and waypoint unlocks can
survive fresh sessions while replay remains pinned to a session-start progression snapshot.

- Postgres now has character-owned item instances with `location`, `equipped`, `slot`, and
  future-ready `rolled_stats`, plus character waypoint rows keyed by level.
- Fresh session creation freezes the character's current items and waypoints into immutable
  session-start snapshot tables; WebSocket fresh attach, `/state`, replay, and timeline all load
  that snapshot before applying session inputs.
- Live inventory add/update/remove changes persist against the session character; dropped and
  consumed items are removed durably for v22.
- Teleporter discovery changes persist as character waypoints; town level `0` remains always
  available even when not explicitly stored.
- Same-session reconnect continues to reconstruct from recorded inputs, not mutable live
  character rows, so historical replay does not drift after later fresh-session progression.
- Bot scenario `15_character_persistence.json` proves gear/equipment persistence, persisted
  level `-1` waypoint access, fresh-session level generation, `/state`, reconnect, and replay.

**Explicit non-goals:** no character picker, player-facing old-session resume, stash UI,
vendors/gold/crafting/quests, character stats/skills/XP, respawn/checkpoints, durable dungeon
maps/monsters/floor drops/HP, or random item stat generation.
