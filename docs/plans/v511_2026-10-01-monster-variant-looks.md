# v511 Plan — Monster variant looks

- **Spec:** [v511 spec](../specs/v511_spec-monster-variant-looks.md)
- **Baseline commit:** `425b9ae4` (detached worktree; no branch)
- **Prerequisites:** none unintegrated (v501 death rim flash and v507 silhouettes are on the base). All tasks may start now.
- **Plan date:** 2026-10-01

## Spec review gate (result)

| Check | Result |
|---|---|
| Scope / non-goals | Presentation-only, five existing families, no new assets. Boss path and elite-command aura lights explicitly excluded. |
| Acceptance to test mapping | See the table below; every criterion has a unit/validator/capture owner. |
| Protocol / schema / golden | None. Only `kit_monster_presentation.v0.schema.json` (asset catalog) changes. `rarity`, `visual_scale`, `visual_tint` already on the wire. |
| Go determinism / replay / world presets | Untouched (no `game/` edits). |
| Shared data ownership | Variant tuning in `kit_monster_presentation.v0.json`; any new hardcoded tuning in GDScript is a defect per the Data-Driven Configuration Policy. Validators cross-check rarity ids and palette ids against `dungeon_generation.v0.json`. |
| Server authority | Unchanged. |
| Client asset adopt/borrow/reject | Recorded in the spec. |
| Maintainability ratchet | `main.gd`, `validate_shared.py` are grandfathered: call-through edits only, new logic in new files <= 600 lines. |
| Overlap with siblings | See conflict table. |

Correction recorded: the brief says "scale"; the server already sends `visual_scale` (champion 1.25, unique 1.5), so the catalog scale is a *multiplier* on it, avoiding double scaling. No product decision blocks planning; the three spec open questions have recommended defaults and are tracked in task 1/2.

### Acceptance-to-test map

| Criterion | Owner |
|---|---|
| Catalog + schema, cross-checks | `make validate-shared validate-assets`; python tests in `tools/assets/test_validate_assets.py` and `tools/test_validate_shared.py` (negative fixtures: unknown rarity, palette, family) |
| Resolver semantics | `client/tests/test_monster_variant_looks.gd` (derived from loaded catalog) |
| Detail tint composes, texture kept, status/hit-flash still work | same test + existing `test_kit_monsters.gd`, `test_look_and_feel_polish.gd` |
| Eye glow / aura / death cleanup | same test (structure) + showme captures + client scenario |
| Scale / health bar alignment | test + capture review |
| Back-compat (bosses, companions, dummies) | same test (node equivalence vs base fixture) |
| Visual proof | showme matrix + player-camera capture |
| Performance | v495/v497 A/B harness |

## File map and ownership

New (all ours, no sibling contention):
- `client/scripts/monster_variant_resolver.gd` — pure `resolve(visual_key, rarity, palette_id) -> Dictionary`.
- `client/scripts/monster_variant_look.gd` — `apply(node, look)` / `clear(node)`: detail tint, eye markers, aura ring; owns marker names for idempotency.
- `client/tests/test_monster_variant_looks.gd`.
- Showme additions in `client/scripts/showme/` (matrix capture) if `visual_capture.gd` would otherwise grow past its baseline.
- Optional client scenario `tools/bot/scenarios/client/v511_monster_variant_pack.json`.

Edited (shared; conflict risk in the next section):
- `shared/assets/kit_monster_presentation.v0.json` (+ `.schema.json`) — add top-level `variants` only; do not touch `clip_profiles`/`monsters`.
- `client/scripts/kit_monster_presentation_loader.gd` — add `variant_config()` accessors only.
- `client/scripts/main.gd` — in `_make_entity_node` (monster branch) and the record update path where rarity/palette is known, one call-through each; `_entity_base_tint` neutralizes the albedo base for catalogued families (decision D2).
- `tools/validate_shared.py`, `tools/assets/validate_assets.py` — cross-checks; prefer a new `tools/validate_monster_variants.py` called from both, as `validate_rarity_cues.py` was.
- `scripts/client_smoke.sh` (register the new test), `docs/CODEMAP.md`, `skills/showme/scripts/render_focus.py` only if a new focus/suite name is added.
- `.maintainability/*` only if a grandfathered file moves (expected none).
- Not edited unless forced: `model_reaction_controller.gd`, `kit_monster_visual.gd`, `model_tint.gd`. If forced, record the lines in the handoff for v513.

## Tasks

- [x] **1. Inventory spike (no code).** With `make model model=<asset_id>` or `skills/showme` find, for the three skeleton scenes plus wolf and bat: head/eye bone names, mesh/material layout (is the atlas shared? how many surfaces?), and a plausible eye offset. Record in the as-built; decide per family whether eye glow is included (spec Q3). Check: bone names resolve via `Skeleton3D.find_bone` in a throwaway headless script (scratchpad, not committed).
- [x] **2. Catalog + schema.** Add `variants`: `rarities` (tint color/strength, scale_multiplier, eye color/energy, aura ring), `depth` keyed by palette id (tint hue shift/strength multiplier), `families` overrides (eye bone/offset/size, aura radius/height, `albedo_neutral` flag). Update the schema (`additionalProperties: false`, hex patterns, numeric ranges). Check: `make validate-shared`.
- [x] **3. Validators.** New `tools/validate_monster_variants.py`: rarity ids are exactly the `monster_rarities` ids; depth keys are the `biome_palettes` ids; family keys exist in `monster_visuals`; ranges sane; every family has a resolvable neutral look. Python tests with negative fixtures. Check: `python3 -m pytest tools/test_validate_monster_variants.py -q` and `make validate-assets`.
- [x] **4. Loader accessors + pure resolver.** Test-first: write `test_monster_variant_looks.gd` resolver cases (distinct per rarity, deterministic, unknown -> neutral, depth differs for skeletons), then implement. Check: `godot --headless --path client --script res://tests/test_monster_variant_looks.gd` (see `scripts/client_smoke.sh` for the exact invocation).
- [x] **5. Look application.** `monster_variant_look.gd`: detail tint via `ModelDetailTint.set_detail` on each mesh; eye markers (shared unshaded emissive primitive on a BoneAttachment3D named per family); aura ring (shared torus mesh + cached material, modeled on `MonsterFamilyAccent`); idempotent re-apply on rarity change; `clear()` removes everything. Material cache capped and keyed by `(source, look)`; resolve the cap-vs-v495 interaction explicitly (raise `ModelTint.MAX_CACHED_TINTS` or use a separate cache; record why). Tests: texture preserved, idempotent, clear restores base, markers are not lights, no extra materials beyond cache bound for a 5x4x3 sweep.
- [x] **6. Wire into `main.gd`.** Apply after model build in `_make_entity_node` and when a record's rarity/palette is known; compute effective scale as server `visual_scale` × catalog multiplier; neutral albedo base for catalogued families, unchanged path for bosses/companions/dummies/uncatalogued. Hide eye/aura nodes on death alongside the corpse treatment (check `enter_death` / dissolve ordering so v501 rim flash is not fought). Tests: back-compat node equivalence, death cleanup, health bar offset still above the scaled head. Check: `make client-unit`.
- [x] **7. Showme matrix.** Extend the `monsters` focus (or add a suite) to render 5 families × 4 rarities at one palette, plus skeleton at 3 palettes, color and grayscale. Check: `make regen-screenshots SUITE="monsters"` then inspect every image (grayscale rarity ordering, floor contrast, eye/aura readability).
- [x] **8. Real-camera proof.** One client scenario with a mixed-rarity pack through the actual player camera; headless pass plus one visible-renderer run (`make bot-client SCENARIO=<id> HEADLESS=1`, `make bot-visual scenario=<id>` with `BOT_STEP_DELAY=0.0` if the default delay times out as in v507). Movement-contract: use a lab world / `start_level` placement, no incidental stair walking; register in the scenario catalog and movement audit only if a new scenario is added.
- [x] **9. Performance A/B.** Set budgets first (see spec). Reuse the v495 first-spawn repeat runner and the v497 frame-pacing probe on `sorcerer_multigroup_perf_probe`: base vs this slice, 10 fresh-process trials per side, same host, quality Balanced and Performance. Record first-spawn process p95, frame p50/p95, draw calls, resources, memory. If over budget, reduce to shared meshes/materials and re-measure before claiming any result; record limits honestly (single host).
- [x] **10. Docs and handoff.** `docs/CODEMAP.md` rows (Assets, Hero light & visuals); `docs/as-built/v511_monster-variant-looks.md` with captures paths (ignored evidence under `.artifacts/`), inventory findings, measurements, and limits; tick these checkboxes. Do not edit `PROGRESS.md` or the lifecycle index (coordinator).

> Execution notes: deviations (reaction-controller skip meta, in-game scenario on the benchmark arena, showme focus instead of a regen suite, A/B toggle by emptying `variants.families`) are recorded in [the as-built](../as-built/v511_monster-variant-looks.md).

## Final gate: focused slice verification

```bash
make validate-shared validate-assets
python3 -m pytest tools/test_validate_monster_variants.py tools/assets/test_validate_assets.py tools/test_validate_shared.py -q
make client-unit
make maintainability
git diff --check
make regen-screenshots SUITE="monsters"
make bot-client SCENARIO=<v511 scenario> HEADLESS=1   # if a scenario was added
```

The coordinator runs the combined `make ci` only after every accepted slice is integrated. This worker does not run `make ci`, `make ci-full`, commit, push, or `/finish`.

## Integration and handoff notes

Handoff must list every changed, deleted, and untracked path and any ignored evidence (`.artifacts/v511-*`, benchmark run dirs), exact commands with outcomes, measured limits, unmet acceptance items, and the shared-file conflict list below.

### Shared-file conflicts with siblings

| File | v511 touch | Likely sibling touch | Mitigation |
|---|---|---|---|
| `shared/assets/kit_monster_presentation.v0.json` + schema | adds top-level `variants` | v512 (boss entries under `monsters`, possibly boss clip/attachment), v513 (`clip_profiles`, new clips) | Different top-level keys; merge as JSON and re-run the schema validators; diff the final file against each handoff. If v512 adds boss `monsters` rows, v511 validator must either accept bosses without variants or v512 adds a neutral entry. |
| `client/scripts/kit_monster_presentation_loader.gd` | new accessors | v512/v513 accessors for their sections | Append-only functions; trivial merge. |
| `client/scripts/kit_monster_visual.gd` | none planned (hook lives in new file) | v513 (clip aliasing), v512 | Avoid edits here; if needed keep to one call. |
| `client/scripts/model_reaction_controller.gd` | none planned | v513 (heavy: reaction/clip polish), v512 | Variant tint uses the detail layer precisely so this file need not change; re-verify hit-flash/status composition after v513 lands. |
| `client/scripts/main.gd` (`_make_entity_node`, `_entity_base_tint`, record update path) | two call-throughs + neutral-base rule | v512 (boss presence in the same node-build path, `BOSS_VISUAL_MODEL`), v513 | Keep the v511 hunk isolated and guarded by "has catalog entry and not a boss"; resolve against the integrated state by hand. |
| `monster_family_accent.gd` | read as pattern only | v512 may reuse for bosses | No edit. |
| `tools/validate_shared.py`, `tools/assets/validate_assets.py` | one call line each | v512/v513 add their own cross-checks | Separate validator module; trivial merge. |
| `scripts/client_smoke.sh`, `docs/CODEMAP.md`, showme `visual_capture.gd`/`render_focus.py` | one registration line / rows / matrix | all three plus UI slices v514-v517 likely add lines | Append-only; re-run `validate_codemap` after merge. |
| UI slices v514-v517 | none | `main.gd`, `scripts/client_smoke.sh`, `docs/CODEMAP.md`, showme registrations | Registry-only overlap; no semantic conflict expected. |

**Recommended integration order:** v513 → v511 → v512. v513 changes `model_reaction_controller`/clip profiles that v511's tint composition must be re-verified against; v512 should land last so boss presence reuses the `unique` variant look and rebases its small `main.gd` hunk on top of both. v511 and v513 can run in parallel safely (disjoint JSON sections); only v512's boss entries need coordination with the `variants` schema. UI slices v514-v517 can integrate in any position after the registry files are reconciled.
