# v96 Plan — Town presentation polish

Status: Complete
Goal: Make the level-0 town read as a compact hub with richer ground, distributed services, cabins, and a central campfire.
Architecture: World service placement stays data-driven in `shared/rules/worlds.v0.json`. The cabins and campfire are client-only presentation nodes built from existing primitive mesh helpers, so server authority and protocol contracts do not change. `$showme` gains a focused `town` capture for step-by-step visual proof.
Tech stack: Shared JSON world rules, Godot client presentation/tests, showme visual tooling, SDD lifecycle docs.

## Baseline and shortcut decision

Builds on v95 `unique-item-catalog-seed` with current branch `main`.

Godot plugin / asset shortcut decision: reject adopting a plugin or imported asset pack for this slice. Borrow the existing in-repo procedural primitive style used by merchant, stash, stairs, and market-board presentation. Imported Kenney-style buildings remain future art-direction work because this slice only needs a compact proof and no asset-pipeline expansion.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/worlds.v0.json` | Distribute `dungeon_levels` town interactables around the central hub. |
| Modify | `client/scripts/main.gd` | Improve town ground texture and add procedural cabin/fire/town prop helpers. |
| Modify | `client/tests/test_item_visuals.gd` | Assert richer town ground variation and procedural town prop structure. |
| Modify | `skills/showme/scripts/render_focus.py` | Add `town` focus and default capture dimensions. |
| Modify | `skills/showme/scripts/visual_capture.gd` | Render a focused town composition for visual feedback. |
| Create | `docs/as-built/v96_town-presentation-polish.md` | Record shipped behavior and proof. |
| Modify | `PROGRESS.md` | Mark v96 complete and add deferred scope. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `skills/showme/scripts/visual_capture.gd`

Decision:
- [x] Defer extraction with rationale: this slice touches tightly-coupled visual factory helpers and focused capture setup; extracting a presentation module now would be larger than the requested polish. Keep growth narrow and verify with `make maintainability`.
- [x] Update the grandfathered baseline for `client/scripts/main.gd` and `skills/showme/scripts/visual_capture.gd` as the explicit v96 maintenance exception.

Verification:
```bash
make maintainability
```

## Task 1 — Shared town layout

Files:
- Modify: `shared/rules/worlds.v0.json`

- [x] Step 1.1: Move `dungeon_levels` town interactables into a distributed but nearby hub at least 5 tiles from the central fire: stash/market/vendor/mystery/bishop around center instead of a straight line.
- [x] Step 1.2: Preserve all existing interactable IDs.

```bash
make validate-shared
```

## Task 2 — Client town presentation

Files:
- Modify: `client/scripts/main.gd`

- [x] Step 2.1: Strengthen `GROUND_TEXTURE_TOWN` texel generation with path/dirt/grass variation while keeping dungeon rock unchanged.
- [x] Step 2.2: Add procedural cabin helper nodes using wood, roof, door, window, and shadow primitives.
- [x] Step 2.3: Add procedural campfire helper nodes with stones, logs, flame, ember glow, and light.
- [x] Step 2.4: Add a `make_town_preview_scene`-style helper that composes ground, current town services, cabins, and fire for showme without mutating gameplay state.

```bash
godot --headless --path client --script res://tests/test_item_visuals.gd
```

## Task 3 — Focused visual proof

Files:
- Modify: `skills/showme/scripts/render_focus.py`
- Modify: `skills/showme/scripts/visual_capture.gd`

- [x] Step 3.1: Add `--focus town` to the Python launcher and use a wide default viewport.
- [x] Step 3.2: Add `_setup_town` to the capture script using the client town preview helper.
- [x] Step 3.3: Capture the first visual proof, inspect it, then adjust composition if needed.

```bash
python3 skills/showme/scripts/render_focus.py --focus town
```

## Task 4 — Focused tests

Files:
- Modify: `client/tests/test_item_visuals.gd`

- [x] Step 4.1: Add tests proving town ground has at least three distinguishable material colors and remains distinct from dungeon ground.
- [x] Step 4.2: Add tests proving town preview contains at least two cabins, a central campfire, and the expected existing service nodes.

```bash
godot --headless --path client --script res://tests/test_item_visuals.gd
make client-unit
```

## Task 5 — Lifecycle docs and CI

Files:
- Create: `docs/as-built/v96_town-presentation-polish.md`
- Modify: `docs/specs/v96_spec-town-presentation-polish.md`
- Modify: `docs/plans/v96_2026-06-12-town-presentation-polish.md`
- Modify: `PROGRESS.md`

- [x] Step 5.1: Mark spec/plan complete and add the as-built summary.
- [x] Step 5.2: Update `PROGRESS.md` latest completed slice, CI gate, lifecycle row, and deferred art scope.
- [x] Step 5.3: Run final verification.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make validate-shared`
- [x] `godot --headless --path client --script res://tests/test_item_visuals.gd`
- [x] `python3 skills/showme/scripts/render_focus.py --focus town`
- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make ci`

## Deferred scope

- Imported town building asset packs, collision-aware decorative buildings, ambient NPC wander, fire animation polish, audio, weather, and full art-direction pass.
