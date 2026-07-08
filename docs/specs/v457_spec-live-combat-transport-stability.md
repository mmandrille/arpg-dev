# v457 Spec — Live Combat Transport Stability

Status: Complete

Date: 2026-07-08

Codename: `live-combat-transport-stability`

## Purpose

Close the Ranger Volley reconnect gap that v453–v456 could not reproduce by exercising the real
Godot WebSocket and frame-processing path during crowded combat.

## Acceptance criteria

1. Godot WebSocket capacity and heartbeat settings are data-driven and sized for burst envelopes.
2. Client and server logs preserve transport close/error diagnostics without logging credentials.
3. A live Godot crowded-Volley scenario casts five times and asserts that no reconnect occurred,
   including reconnects that completed before the final assertion.
4. Client-bot class setup preserves the requested class and debug progression seeding does not
   erase it.
5. Shared validation, focused Go/Godot tests, the live scenario, and final `make ci` are green.

## Non-goals

- Replacing JSON or WebSockets.
- Async persistence architecture.
- Treating protocol replay or Python-only sockets as proof of live Godot transport stability.
