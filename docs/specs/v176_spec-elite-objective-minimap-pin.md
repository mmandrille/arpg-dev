# v176 Spec — Elite Objective Minimap Pin

Status: Approved for planning
Date: 2026-06-15
Codename: elite-objective-minimap-pin

## Purpose

Make the generated elite side-objective reward chest easier to locate while playing. When the current floor has a closed `elite_objective` chest, the Godot HUD should show a compact minimap-style panel with a player dot and an objective pin projected from the chest position relative to the local player.

## Non-goals

- No server protocol, quest, objective, or world-generation changes.
- No full minimap, fog-of-war, route line, compass, or persistent floor map.
- No changes to elite objective locking, leader count rules, loot, or reward chest generation.
- No external Godot minimap plugin adoption for this compact overlay.

## Acceptance Criteria

- A compact top-right minimap-style widget appears when a closed `elite_objective` chest exists on the active floor.
- The widget draws a stable player marker and an objective pin whose debug position is derived from the chest position relative to the local player.
- The minimap hides when no elite objective chest exists or when the objective chest is open/complete.
- Bot/debug state exposes minimap visibility, pin presence, status, and normalized pin coordinates.
- A pinned client bot scenario descends to the existing elite objective floor and asserts the minimap pin is visible and active.

## Scope and Files Likely Touched

- Client UI: new `client/scripts/elite_objective_minimap.gd`.
- Client state helper: new `client/scripts/elite_objective_minimap_state.gd`.
- Client wiring: `client/scripts/main.gd` creates and updates the minimap and exposes debug state.
- Client tests: new `client/tests/test_elite_objective_minimap.gd` and `scripts/client_smoke.sh` gate.
- Bot tooling: minimap assertion helper, wait/assert registration, and client scenario.
- Docs: this spec, matching plan, as-built notes, and `PROGRESS.md`.

## Test and Bot Proof

- `make client-unit` covers minimap hidden, active pin, clamped pin coordinates, and complete states.
- `make bot-client scenario=45_elite_objective_minimap_pin.json` asserts the minimap on the pinned elite objective floor.
- `make maintainability` proves grandfathered client files stay within the ratchet.
- Final `make ci` passes before commit.

## Open Questions and Risks

- No blocking questions.
- Risk: this is a presentation-only position projection, not a navigational route. Future map or compass slices should promote objective routing into an explicit shared/client map model rather than expanding this compact widget indefinitely.
