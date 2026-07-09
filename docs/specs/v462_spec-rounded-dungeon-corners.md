# v462 Spec - Rounded Dungeon Corners

Status: Implemented
Date: 2026-07-09
Codename: rounded-dungeon-corners
Baseline: v461 `entity-locomotion-polish`

## Purpose

Reduce the "too blocky" feel of normal generated dungeon floors by adding a presentation-only
rounded-corner treatment to generated dungeon walls and room joins. The authoritative server keeps
the existing rectangular room/corridor layout and collision, while the Godot client makes eligible
wall turns and room-edge transitions read softer and more organic.

This slice is intentionally a thin visual pass, not a true non-rectangular dungeon-generation
rewrite. It should create enough visible change for user review while preserving the deterministic
server layout already shipped in v445.

Because the owner wants to give feedback during development, the slice must be implemented with
focused visual checkpoints that present multiple corner-style options before the final style is
locked in.

## Non-goals

- No server pathfinding, collision, reachability, obstacle placement, or protocol changes.
- No non-rectangular walkable space, curved collision, or polygonal room generation.
- No boss-floor layout overhaul.
- No imported art packs, plugins, trim kits, or new asset pipeline.
- No broad dungeon surface/palette refresh beyond what is needed to support readable corner
  presentation.
- No rock/column/water/hole silhouette redesign in this slice.

## Acceptance Criteria

- Normal generated dungeon floors render a visibly softer corner treatment at eligible wall joins
  instead of only hard 90-degree box turns.
- The rounded treatment is client-only and does not change authoritative wall positions, obstacle
  blocking, stairs reachability, or movement behavior.
- The first implementation pass presents at least two visually distinct corner-treatment options
  using focused captures before one style is chosen for rollout.
- The chosen style applies consistently to the selected scope of generated dungeon `room_wall`
  runs, including the approved softened top edge treatment, without affecting town presentation.
- Existing dungeon wall/floor bot proofs continue to pass without gameplay-behavior changes.
- Focused client tests prove rounded presentation activates for eligible dungeon joins and stays off
  for excluded surfaces.
- A focused client-bot or bot-visual proof shows the chosen corner treatment on a normal generated
  dungeon floor.

## Scope and Likely Files

- Client presentation:
  - `client/scripts/wall_renderer.gd`
  - `client/scripts/ground_wall_factory.gd`
  - possible focused helper extracted from `wall_renderer.gd` if the ratchet requires it
- Client tests:
  - `client/tests/test_factories.gd`
  - possibly a new focused wall-presentation unit test if current coverage becomes too indirect
- Bot / visual proof:
  - `tools/bot/scenarios/client/79_wall_floor_dungeon_rollout.json` or a new focused client
    scenario for rounded-corner proof
  - visual review support via `make bot-visual` and/or focused `/showme` captures
- Docs:
  - `docs/plans/v462_2026-07-09-rounded-dungeon-corners.md`
  - `docs/as-built/v462_rounded-dungeon-corners.md`
  - `docs/progress/slice-lifecycle.md`
  - `PROGRESS.md`

### Asset / plugin decision

- Adopt: existing procedural dungeon presentation pipeline in `GroundWallFactory` and
  `WallRenderer`, plus the current generated dungeon wall metadata.
- Borrow: existing wall/floor shader tests, dungeon rollout client scenario patterns, and focused
  visual capture workflows already used for presentation slices.
- Reject: external dungeon kits, imported curved wall assets, Godot addons, and any server-authored
  geometry format change for this slice.

## Test and Bot Proof

```bash
godot --headless --path client --script res://tests/test_factories.gd
make client-unit
HEADLESS=1 make bot-client scenario=wall_floor_dungeon_rollout
```

Expected visual-review proof during implementation:

```bash
make bot-visual scenario=wall_floor_dungeon_rollout
```

If a dedicated focused scenario is added during planning, use that scenario for both headless and
visual proof instead of overloading the rollout route.

## Open Questions and Risks

- Scope question: should the rounded treatment apply only to `room_wall` joins first, or also to
  generated perimeter joins on the same slice? Default for planning: include both if they can share
  one coherent presentation rule; otherwise start with `room_wall`.
- Presentation risk: a stronger rounded silhouette may visually imply walkable curved collision the
  server does not actually have. Planning should keep the effect readable without lying about
  blocking edges.
- Maintainability risk: `client/scripts/wall_renderer.gd` is already a likely hotspot for growth.
  If the corner logic becomes branch-heavy, extract it into a focused presentation helper rather
  than extending one large coordinator path.
- Planning note: there is a stale `docs/specs/v461_spec-dungeon-surface-kit.md` draft even though
  `PROGRESS.md` records v461 as complete for `entity-locomotion-polish`. Planning should treat this
  slice as v462 and avoid reusing the v461 number.
