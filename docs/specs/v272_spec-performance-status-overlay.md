# v272 Spec - Performance Status Overlay

## Status

Approved for implementation on 2026-06-18.

## Problem

The crowded-lightning slices now produce useful backend performance data in logs, but during live
play the top-right status toggle still shows old client debug text: connection mode, inventory,
weapon visual state, and control hints. That makes it hard to validate crowded combat feel while
watching the game.

## Goals

- Replace the existing status text contents with a top-right **Performance Status** panel.
- Keep the status toggle as the user-facing control, but remove the old debug/control/weapon data
  from the panel.
- Show the backend crowded-combat fields during gameplay:
  `ai_ms`, `pathfind_ms`, `combat_ms`, `broadcast_ms`, `persist_ms`, `path_requests`,
  `path_cache_hits`, `path_nodes_visited`, `monsters_moved`, and tick-budget overrun state.
- Also show client FPS and a best-effort WebSocket intent ping.
- Preserve the multiplayer model: server owns AI, navigation, combat, loot, and backend perf
  truth; clients only render and estimate local presentation/network values.

## Protocol Surface

`shared/protocol/state_delta.v8.schema.json` gains an optional `performance` object. The object is
sent by the authoritative session loop at a throttled debug/status cadence and can arrive in a
`state_delta` with empty `changes` and `events`.

The payload is server-owned and includes:

- tick timing: `tick`, `total_ms`, `sim_ms`, `ai_ms`, `pathfind_ms`, `combat_ms`, `broadcast_ms`,
  `persist_ms`, `tick_budget_ms`, `tick_over_budget`, `tick_overrun_ms`, `degradation_applied`
- path/crowd counters: `path_requests`, `path_cache_hits`, `path_nodes_visited`, `monsters_moved`
- room shape: `game_level`, `entities`, `players`, `monsters`, `live_monsters`, `companions`,
  `projectiles`, `loot`, `interactables`, `walls`
- loop shape: `inputs`, `results`, `changes`, `events`, `acks`, `rejects`, `clients`

This is a backward-compatible optional v8 extension for the existing coordinated server/client
repo. No legacy client compatibility is required during active development.

## Client Behavior

- The label is anchored at the top right, below the level label.
- Header text is exactly `Performance Status`.
- FPS is read locally from Godot.
- Ping is estimated by recording send time for client envelopes and consuming the round trip when
  the matching `intent_accepted` or `intent_rejected` arrives.
- When no performance payload has arrived yet, the panel still shows FPS, ping, connection state,
  tick, and a waiting backend line.

## Non-Goals

- Historical graphs, dashboards, percentile metrics, or production observability export.
- Client-side authority over monster movement, AI, combat, loot, or backend timing values.
- Replacing the crowded-lightning log/bot proof added by v268-v271.

## Acceptance

- Shared protocol validation accepts `state_delta.performance`.
- Backend tests prove the payload includes timing, counters, room shape, tick-budget, and fanout
  behavior without changes/events.
- Client tests prove the status panel title/content replaces old debug text and that ping timing is
  tracked from accepted/rejected intents.
- Focused verification passes for shared validation, realtime backend tests, and Godot client unit
  tests touched by this slice.
