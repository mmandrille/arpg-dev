# v7 — Gear before combat scenario

**Proves:** The server can own multiple deterministic initial world presets, and replay/resume/debug timelines reconstruct the selected preset instead of drifting to the default world.

- Shared `worlds.v0.json` defines `vertical_slice` and `gear_before_combat` initial layouts.
- Sessions persist `world_id`; create defaults to `vertical_slice`, rejects unknown worlds, and resume returns the persisted world.
- `game.NewSimWithWorld` spawns the player, initial loot, and monsters from rules data; `NewSim` remains a default wrapper.
- Replay reconstruction, `/state`, replay timeline, and WebSocket fresh/resume paths use the persisted world.
- Bot scenario catalog now runs `01_vertical_slice.json` then `02_gear_before_combat.json`.
- Gear scenario walks to initial `rusty_sword`, picks it up, equips it, kills `training_dummy_reward`, picks up `training_badge`, and asserts two inventory items.

**Explicit non-goals (still true):** no pickup range gate, no `world_id` in WebSocket snapshots, no Godot inventory UI for non-visual items.
