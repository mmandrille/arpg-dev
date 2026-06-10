# v53 Spec: Boss Health Bar UI

Status: Accepted
Date: 2026-06-10
Codename: boss-health-bar-ui

## Purpose

Boss floors already expose authoritative boss metadata (`is_boss`, `boss_template_id`, `hp`, `max_hp`) through the existing session snapshot and state delta contracts. The client currently renders bosses as monsters with a small world-space health bar, which makes the boss encounter feel like a normal mob and gives poor visibility during movement.

Add a dedicated screen-anchored boss health bar so the active dungeon boss is always readable while present, without changing combat authority or protocol shape.

## Scope

- Add an in-repo Godot `Control` for a top-center boss health bar.
- Show the bar when a live boss monster exists in the current entity set.
- Populate the bar from client entity records already fed by server state:
  - boss entity id
  - `boss_template_id`
  - display title derived from `boss_template_id`
  - `hp`
  - `max_hp`
- Update the bar whenever boss hp changes.
- Hide the bar when the boss dies, is removed, the level changes, or gameplay state is torn down.
- Expose boss bar debug state through `get_bot_state()`.
- Add bot scenario support for asserting/waiting on boss health bar state.
- Add a focused client bot scenario that reaches the first boss floor and asserts the bar is visible for `cave_warden`.

## Non-goals

- No protocol, server combat, boss generation, or shared schema changes.
- No boss portraits, phase timers, special VFX, or bespoke boss art.
- No replacement of existing world-space monster health bars.
- No multi-boss layout beyond deterministic selection if more than one live boss ever appears.
- No health prediction; the bar displays only server-owned hp/max hp already received by the client.

## Acceptance Criteria

1. In `dungeon_levels`, reaching the first boss floor shows a screen-anchored boss health bar while the boss is alive.
2. The boss bar displays a readable boss title derived from `boss_template_id`, plus numeric hp/max hp.
3. The boss bar fill ratio updates from authoritative hp/max hp and clamps safely between 0 and 1.
4. The boss bar hides when there is no live boss entity in the current level, including boss death, entity removal, level change, or gameplay teardown.
5. Normal monster world-space health bars continue to render as before.
6. `get_bot_state()` includes a `boss_health_bar` dictionary with visibility, boss id, template id, title, hp, max hp, and ratio.
7. A `godot_client` bot scenario can wait/assert the boss bar state on the first boss floor.

## Plugin And Asset Adoption

Decision: reject external Godot plugin or asset adoption for this slice.

Reason: the client already has in-repo `Control`/`CanvasLayer` UI patterns and a monster health bar implementation. The slice is a narrow presentation layer over existing server state, so adopting a plugin or asset pack would add surface area without reducing implementation risk.

## Test Proof

- `CLIENT_UNIT_ONLY=1 make client-smoke` covers server-independent client and bot-runner assertions.
- `make bot-client SCENARIO=tools/bot/scenarios/client/26_boss_health_bar_ui.json HEADLESS=1` proves the client reaches the first boss floor and exposes the boss bar for `cave_warden`.
- `make ci` is the final finish gate.

## Risks

- Boss floor traversal can be slow in headless bot runs. The scenario should use the existing `click_entity` stair automation and generous per-floor timeouts.
- If a future ruleset produces multiple live bosses, this slice should pick a deterministic live boss rather than adding multi-boss UI.
