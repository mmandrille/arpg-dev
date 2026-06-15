# v96 Spec — Town presentation polish

Status: Complete
Date: 2026-06-12
Codename: `town-presentation-polish`

## Purpose

Make the level-0 town read as a small settled hub instead of a row of service nodes. The slice improves the town ground texture, distributes the existing town interactables around the center, and adds client-side cabin and campfire presentation props that do not affect authoritative gameplay.

## Non-goals

- No new town services, interactions, quests, combat rules, or safe-zone rules.
- No protocol schema version bump.
- No collision or pathing changes for decorative cabin/fire props.
- No production art pass beyond in-repo procedural Godot primitives.

## Acceptance criteria

- The level-0 town ground uses a richer generated texture with grass, dirt-worn path variation, and flecks instead of the current mostly uniform tiled grass.
- `dungeon_levels` town services are distributed around the central town space, with each visible service at least 5 tiles from the central campfire while still staying in the town hub.
- Existing services still use their current interactable IDs and remain reachable by current interaction flows.
- The client can render non-authoritative town cabins and a central campfire as visual props.
- A focused `$showme` capture can render the town composition with ground, services, cabins, and campfire in one frame.
- Focused client visual tests assert the new procedural props and stronger town ground variation.

## Scope and likely files

- `shared/rules/worlds.v0.json` — adjust level-0 town interactable positions for `dungeon_levels`; keep IDs unchanged.
- `client/scripts/main.gd` — improve town ground texel generation and add reusable procedural town prop helpers.
- `client/tests/test_item_visuals.gd` — add focused assertions for town ground variation and town props.
- `skills/showme/scripts/render_focus.py` and `skills/showme/scripts/visual_capture.gd` — add `town` focus for step-by-step visual proof.
- `docs/plans/`, `docs/as-built/`, `PROGRESS.md` — SDD lifecycle docs.

## Test and visual proof

- `make validate-shared`
- `godot --headless --path client --script res://tests/test_item_visuals.gd`
- `python3 skills/showme/scripts/render_focus.py --focus town`
- `make client-unit`
- `make maintainability`
- `make ci`

No bot scenario is required because this slice changes presentation and static world layout only; service IDs and protocol payloads remain unchanged.

## Open questions and risks

- `client/scripts/main.gd` and `skills/showme/scripts/visual_capture.gd` are grandfathered over the maintainability line. The plan must either keep growth within the baseline allowance or document the focused-deferral rationale.
- Decorative props must be visually clear without becoming gameplay-obstacle-looking blockers; keep them off the exact central service interaction points.
