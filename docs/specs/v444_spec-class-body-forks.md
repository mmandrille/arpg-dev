# v444 Spec — Class body forks from `base_human`

Status: Complete  
Date: 2026-07-06  
Codename: class-body-forks

## Purpose

Fork the shared **`base_human`** player mesh into **five class-specific static sources**, each
re-skinned through the existing **canonical 17-bone rig** so silhouettes differ (barbarian: wide
shoulders; sorcerer: tall and thin; rogue: short and compact; etc.) while **skeleton joints remain
aligned with visible anatomy** and shared Godot animation clips keep working.

Today every class binds to `character_base_human_v0` in
`shared/assets/class_presentations.v0.json`. After this slice, each class binds to its own
`character_{class}_v0` runtime GLB generated from a forked `{class}_mesh.glb` source.

**User-visible outcome:** distinct class bodies in gameplay, character select, gear preview, and
bot-visual proofs — without floating gear, joints inside the torso, or broken attack/walk clips.

## Non-goals

- New bones, new animation clips, or protocol/server/combat changes.
- Class-specific shaders, VFX, or icon/UI polish (presentation-only mesh swap).
- Replacing the canonical bind pose in `tools/assets/canonical_skeleton.py`.
- Procedural bodies (`base_human_mesh.py` / `base_human_glb()`) as merge output — dev reference only.
- Deleting `character_base_human_v0` until all five forks pass verification (keep as fallback/reference).
- CI pack promotion unless a single representative scenario is added budget-neutrally (default: extended-only visual proofs).

## Background

| Piece | Location | Role |
|-------|----------|------|
| Shared source | `assets/characters/base_human/base_human_mesh.glb` | Tier-3 static A-pose mesh (fork starting point) |
| Shared runtime | `client/assets/characters/base_human/base_human.glb` | Canonical re-skin output |
| Frozen skeleton | `tools/assets/canonical_skeleton.py` | 17-bone hierarchy + `joint_globals()` |
| Landmark fit | `joint_globals_from_mesh()` | Shoulder/elbow/hand/knee/foot from mesh vertex clusters after height normalize |
| Weights | `vertex_weights()` + `rig_canonical_hero.py` | Segment inverse-distance weights (no hand-painted weights) |
| Orchestration | `tools/assets/rig_hero_glbs.py` | `HEROES`, `HERO_TARGET_HEIGHTS`, `CANONICAL_RIG_IDS` |

v443 proved landmark placement fixes A-pose art vs arms-down bind mismatch. Class forks must preserve
that contract while allowing **mesh morphology** to change.

## Skeleton contract (non-negotiable)

These rules apply to **every** class fork. Violating them is a slice failure.

| Rule | Rationale |
|------|-----------|
| Runtime GLB exposes exactly **17 joints** named `root`, `spine`, `chest`, `neck`, `head`, `arm_l`, `elbow_l`, `hand_l`, `arm_r`, `elbow_r`, `hand_r`, `leg_l`, `knee_l`, `foot_l`, `leg_r`, `knee_r`, `foot_r` | Shared `character_anims.tres`, sockets, and `REQUIRED_BONES` |
| Rig via `rig_glb_canonical_bytes()` only — **no pre-skinned** source GLBs | Tool rejects embedded `skins` |
| Source mesh stays **A-pose** (arms roughly horizontal) | `joint_globals_from_mesh()` band fractions assume A-pose |
| Height normalize via `HERO_TARGET_HEIGHTS[class]` before landmarks | Landmark Y fractions are relative to mesh bounds |
| **Do not** fake body shape with `class_presentations` `scale` / `idle_stance` or client `ModelRoot.scale` | Those are presentation lean only; gear/skeleton will desync |
| Landmark pass before merge: `hand_*` joints within ~12 cm of mesh hand clusters (same bar as `test_base_human_landmark_hands_near_mesh_extents`) | Catches wrist-in-torso regressions |

**Forbidden shortcuts:** importing Blender-rigged GLBs; moving torso canonical globals to “fix” art;
client-only scaling; skipping skeleton `/showme` per class.

## Per-class morphology targets

Fork by copying `base_human_mesh.glb` → `assets/characters/{class}/{class}_mesh.glb`, then edit
**vertices only** (Blender proportional edit/sculpt, or AI static re-export). Keep facing **+Y**,
feet on ground, **no rig, no animations**.

| Class | Silhouette | Initial `HERO_TARGET_HEIGHTS` (m) | Morphology notes |
|-------|------------|-------------------------------------|------------------|
| `barbarian` | Heavy, wide | **1.97** | Broader shoulders, thicker chest, wider arms, stocky legs (tallest) |
| `paladin` | Armored frame | **1.88** | Broad shoulders, upright, medium bulk (less wide than barbarian) |
| `ranger` | Lean hunter | **1.82** | Narrow shoulders, longer limbs, lighter torso |
| `sorcerer` | Tall mage | **1.86** | Taller, thinner chest/hips, slimmer arms, longer neck read |
| `rogue` | Compact assassin | **1.70** | Shortest; narrower frame; slight forward mass (stooped read, not bind-pose rotation) |

Heights are **data** in `rig_hero_glbs.py` — tune after first art pass, not only in DCC tools.

## Asset decision (adopt / borrow / reject)

- **Adopt:** `base_human` canonical rig pipeline (`rig_canonical_hero.py`, `canonical_skeleton.py`);
  manifest + `class_presentations` binding pattern from v274/v430 class mesh swaps.
- **Borrow:** Blender (or equivalent) for proportional mesh edits on forked static GLB; optional
  `skills/3dmodel/scripts/create_model_probe.py` for orientation checks.
- **Reject:** Per-class animation libraries; new bone names; runtime-fetch meshes; pre-skinned hero
  exports; morphology via presentation scale only.

## Deliverables (per class)

Repeat for `barbarian`, `paladin`, `ranger`, `rogue`, `sorcerer`:

| Artifact | Path |
|----------|------|
| Static source | `assets/characters/{class}/{class}_mesh.glb` |
| Runtime (generated, committed) | `client/assets/characters/{class}/{class}.glb` |
| Manifest entry | `character_{class}_v0` in `assets/manifests/assets.v0.json` |
| Class binding | `shared/assets/class_presentations.v0.json` → `{class}.model.asset_id` |
| Rig registration | `HEROES["{class}"]`, `HERO_TARGET_HEIGHTS["{class}"]`, `CANONICAL_RIG_IDS` |
| README (optional) | `assets/characters/{class}/README.md` — morphology + provenance |

Regenerate `shared/assets/model_preview_catalog.v0.json` via `tools/assets/model_catalog.py generate`.

## Implementation order (recommended)

One class at a time with full verification before the next:

1. **barbarian** — widest shoulder delta; validates landmark bands under heavy morph.
2. **rogue** — shortest; validates lower-body landmark bands.
3. **sorcerer** — thin/tall; validates narrow cluster detection.
4. **paladin**, **ranger** — middle silhouettes.

Plan may schedule sub-commits per class; spec is not done until all five pass acceptance.

## Acceptance criteria

### Pipeline (all classes)

1. Each `assets/characters/{class}/{class}_mesh.glb` is static (no `skins`, no animations).
2. `make gen-assets` (or `rig_hero_glbs.py`) writes each `client/assets/characters/{class}/{class}.glb` with all `REQUIRED_BONES` and segment weights.
3. Each `character_{class}_v0` manifest entry has correct `runtime_path`, `required_nodes`, `provenance.sha256`, and license/origin.
4. `shared/assets/class_presentations.v0.json` maps each class to its own `character_{class}_v0` (not `character_base_human_v0`).
5. `make validate-shared` and character sections of `make validate-assets` pass.
6. `.venv/bin/pytest tools/assets/test_rig_canonical_hero.py tools/assets/test_rig_hero_glbs.py -q` pass (extend with per-class landmark tests if bands are tuned).

### Skeleton ↔ mesh match (all classes)

7. `make model model=character_{class}_v0 CHECK=1` loads with full clip set (`idle`, `walk`, `attack`, `attack_2h`, `attack_off_hand`, `attack_ranged`, `attack_staff`, `hit`, `death`).
8. Skeleton `/showme` for each class: `hand_l` / `hand_r` at visible wrists; elbows/knees/feet inside limbs — **not** floating in torso (compare to v443 barbarian baseline quality).
9. Gear `/showme` per class with representative starter gear: weapons on palms, helm on head, no large float offset.
10. `client/tests/test_animation.gd` `_test_class_character_models` asserts each class resolves its own `character_{class}_v0` and socket bones bind correctly.

### Regression

11. `character_base_human_v0` remains valid fallback for unknown classes (`necromancer` path in `test_animation.gd`).
12. No growth of unrelated procedural/equipment SHA drift in the same slice (regen only touched class GLBs).

## Scope and files likely touched

| Area | Files |
|------|-------|
| Sources | `assets/characters/{class}/{class}_mesh.glb`, READMEs |
| Runtime | `client/assets/characters/{class}/*` (GLB, textures, `.import`) |
| Manifest / binding | `assets/manifests/assets.v0.json`, `shared/assets/class_presentations.v0.json`, `shared/assets/model_preview_catalog.v0.json` |
| Rig pipeline | `tools/assets/rig_hero_glbs.py`, optionally `tools/assets/canonical_skeleton.py` (landmark bands only, with tests) |
| Client | `client/tools/build_animations.gd` only if clip source path changes (unlikely — still one skeleton) |
| Tests | `tools/assets/test_rig_canonical_hero.py`, `tools/assets/test_rig_hero_glbs.py`, `tools/assets/test_model_catalog.py`, `client/tests/test_animation.gd`, `client/tests/test_model_viewer.gd` |
| Docs | `docs/as-built/v444_class-body-forks.md`; optional `assets/characters/base_human/README.md` cross-link |

**No server or protocol changes.**

## Test and bot proof

| Layer | Command / artifact |
|-------|-------------------|
| Asset unit | `pytest tools/assets/test_rig_canonical_hero.py tools/assets/test_rig_hero_glbs.py tools/assets/test_model_catalog.py` |
| Validation | `make validate-shared`, `make validate-assets` |
| Client unit | `make client-unit` (animation + model viewer gates) |
| Model viewer | `make model model=character_{class}_v0 CHECK=1` × 5 |
| Visual | `python3 skills/showme/scripts/render_focus.py --focus skeleton --class-id {class}` × 5 |
| Visual | `python3 skills/showme/scripts/render_focus.py --focus gear --class-id {class} --items …` × 5 |
| Batch regression | `make regen-screenshots SUITE="gear skeleton"` after all five land (inspect `.artifacts/screenshots/latest/`) |
| Bot (extended) | Reuse or extend `client/93_barbarian_tier3_visual.json` pattern per class — **extended tier by default**; promote to CI pack only if merge-blocking and budget-neutral |

## Recovery if skeleton drifts

1. Confirm source is still A-pose and arm vertices occupy expected X/Y bands (`canonical_skeleton.py` ~L165–197).
2. Adjust `HERO_TARGET_HEIGHTS[class]` and re-rig.
3. If a silhouette systematically breaks clusters, add **narrow, tested** landmark band overrides (per-class or per-hero_id parameter) — document in as-built; do not edit frozen torso globals ad hoc.

## Open questions

1. **DCC workflow:** Blender-in-repo edits vs external AI re-export — plan should pick one primary path per class for reproducibility.
2. **Landmark overrides:** Are class-specific `joint_globals_from_mesh` bands required for rogue (short) or sorcerer (thin), or is height normalize + art discipline enough? Decide after barbarian fork proof.
3. **Fallback policy:** Keep `character_base_human_v0` forever as `FALLBACK_ASSET_ID`, or repoint fallback to `character_barbarian_v0` after forks ship?
4. **Slice splitting:** Single `/execute` for all five vs five micro-slices — planner chooses based on review time; this spec stands as the full contract either way.

## Risks

| Risk | Mitigation |
|------|------------|
| Morph moves arm geometry out of landmark bands | Per-class skeleton showme gate; landmark unit test |
| AI exports T-pose or embedded rig | Probe script + `validate-assets` skin rejection |
| Five forks in one PR obscure which morph broke rig | Implement in order; commit/checkpoint per class |
| Texture/material drift from re-export | Reuse base_human materials where possible; regen-screenshots gear suite |

## References

- [`assets/characters/base_human/README.md`](../../assets/characters/base_human/README.md) — canonical pipeline
- [`docs/as-built/v443_barbarian-canonical-reskin.md`](../as-built/v443_barbarian-canonical-reskin.md) — landmark skeleton fix
- [ADR-0006](../adr/0006-asset-pipeline.md) — manifests, sockets, validation
- [ADR-0007](../adr/0007-animation-state-model.md) — client-only animation; shared clips
