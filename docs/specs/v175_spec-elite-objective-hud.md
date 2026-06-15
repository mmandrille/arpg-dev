# v175 Spec — Elite Objective HUD

Status: Approved for planning
Date: 2026-06-14
Codename: elite-objective-hud

## Purpose

Make generated elite side objectives readable without opening a panel. When the current floor has an elite-objective reward chest, the Godot HUD should show the objective state: defeat remaining elite leaders, claim the unlocked chest, or complete after the chest opens.

## Non-goals

- No minimap pins, compass arrows, or floor-map marker routing.
- No server quest/objective protocol changes; the client derives display state from existing entity metadata.
- No changes to objective rules, required leader count, chest locking, loot, or generation tuning.
- No persistent quest journal entries or multi-objective list.

## Acceptance Criteria

- A compact HUD tracker appears only when the active floor has an `elite_objective` chest.
- While any `monster_pack_leader` on the floor is alive, the HUD shows a defeat-leaders objective with a remaining count.
- After all leaders are dead and the objective chest remains closed, the HUD shows a claim-chest objective.
- After the objective chest opens, the HUD shows complete.
- Bot/debug state exposes the tracker visibility, status, and remaining leader count.
- A pinned client bot scenario descends to the existing elite objective floor and asserts the HUD active state.

## Scope and Files Likely Touched

- Client UI: new `client/scripts/elite_objective_tracker.gd`.
- Client wiring: `client/scripts/main.gd` derives state and updates the tracker.
- Client tests: new `client/tests/test_elite_objective_tracker.gd` and `scripts/client_smoke.sh`.
- Bot tooling: assertion helper and scenario for tracker debug state.
- Docs: this spec, matching plan, as-built notes, and `PROGRESS.md`.

## Test and Bot Proof

- `make client-unit` covers tracker active, claim, complete, and hidden states.
- `make bot-client scenario=44_elite_objective_hud.json` asserts the tracker on the pinned elite objective floor.
- `make maintainability` proves the grandfathered client wiring stays within the ratchet.
- Final `make ci` passes before commit.

## Open Questions and Risks

- No blocking questions.
- Risk: the HUD status is inferred client-side from entity metadata; if future server objective types need richer state, they should add explicit objective protocol rather than expanding this inference indefinitely.
