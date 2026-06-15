# v72 Plan — Monster Visual Catalog

Status: Complete
Goal: Add a data-driven monster visual catalog and prove it with quadruped/flyer monsters plus boss visual variety.
Architecture: Monster mechanics remain in shared rules and the Go sim; monster presentation moves to shared asset metadata consumed by the Godot client. New assets are generated deterministically through the existing GLB pipeline and validated through the manifest. Boss model selection stays deterministic and server-authored, while client rendering resolves all monster scenes through one loader/resolver.
Tech stack: Shared JSON/schema, Go deterministic sim, Python asset validators/generators, Godot client scenes/scripts/tests, Python/Godot bot/showme tooling.

## Baseline and Shortcut Decision

Builds on v71 `class-picker-and-sprites` with v72 as the next free slice. Current code has one monster scene (`monster_dummy.tscn`), a hardcoded client model branch in `main.gd`, and boss visuals pinned to `current_humanoid_player` in `boss_templates.v0.json` plus validation. The slice replaces those model-choice branches with shared visual metadata.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/assets/monster_visuals.v0.schema.json` | Monster visual metadata contract |
| Create | `shared/assets/monster_visuals.v0.json` | Monster def to scene/asset/scale/height/profile mapping |
| Modify | `tools/validate_shared.py` | Validate new shared asset file/schema |
| Modify | `assets/manifests/assets.v0.json` | Register quadruped/flyer GLB assets |
| Modify | `tools/assets/gen_glb.py` | Generate deterministic quadruped/flyer GLBs |
| Modify | `tools/assets/validate_assets.py` | Cross-check monster visuals against manifest/runtime assets |
| Modify | `tools/assets/test_validate_assets.py` | Validator coverage for monster visuals |
| Create | `assets/monsters/quadruped/README.md`, `assets/monsters/tiny_flyer/README.md` | Provenance notes |
| Create | `client/assets/monsters/quadruped/monster_quadruped.glb`, `client/assets/monsters/tiny_flyer/monster_tiny_flyer.glb` | Runtime assets |
| Modify | `shared/rules/monsters.v0.json` | Add `dungeon_wolf`, `dungeon_bat` |
| Modify | `shared/rules/dungeon_generation.v0.json` | Add dungeon spawn weights/minimums |
| Modify | `shared/rules/boss_templates.v0.json`, `.schema.json`, `server/internal/game/rules.go` | Deterministic boss visual options |
| Modify | `server/internal/game/dungeon_gen.go`, `server/internal/game/sim.go`, tests/goldens | Emit selected boss visual model key |
| Modify | `shared/rules/worlds.v0.json` | Add compact visual catalog lab world |
| Create | `client/scripts/monster_visuals_loader.gd` | Client-side shared visual resolver |
| Create/Modify | `client/scenes/monster_quadruped.tscn`, `client/scenes/monster_tiny_flyer.tscn`, existing monster scene refs | Scene loading and animation profiles |
| Modify | `client/scripts/main.gd` | Instantiate monster visuals through resolver |
| Modify | `client/tests/test_animation.gd` and/or new `client/tests/test_monster_visuals.gd` | Scene/metadata coverage |
| Modify | `skills/showme/scripts/render_focus.py`, `skills/showme/scripts/visual_capture.gd` | Monster lineup focus |
| Create | `tools/bot/scenarios/NN_monster_visual_catalog.json` and optional client scenario | Bot/client proof |
| Modify | `PROGRESS.md`, `docs/as-built/v72_monster-visual-catalog.md` | Lifecycle docs |

## Task 1 — Shared Visual Contract

Files:
- Create: `shared/assets/monster_visuals.v0.schema.json`
- Create: `shared/assets/monster_visuals.v0.json`
- Modify: `tools/validate_shared.py`
- Modify: `tools/assets/validate_assets.py`
- Modify: `tools/assets/test_validate_assets.py`

- [x] Step 1.1: Define a schema with version and `monster_visuals` entries keyed by monster def id, including `asset_id`, `scene`, `scale`, `height_offset`, and `animation_profile`.
- [x] Step 1.2: Add mappings for current monster defs to the dummy scene, and planned `dungeon_wolf` / `dungeon_bat` to new scene keys.
- [x] Step 1.3: Validate that each referenced monster visual asset exists in `assets.v0.json` and has type `monster`.
- [x] Step 1.4: Add negative validator tests for unknown monster asset ids and wrong asset types.

```bash
make validate-shared
.venv/bin/pytest tools/assets/test_validate_assets.py -q
```

## Task 2 — Deterministic Monster Assets

Files:
- Modify: `tools/assets/gen_glb.py`
- Modify: `assets/manifests/assets.v0.json`
- Create: `assets/monsters/quadruped/README.md`
- Create: `assets/monsters/tiny_flyer/README.md`
- Create: `client/assets/monsters/quadruped/monster_quadruped.glb`
- Create: `client/assets/monsters/tiny_flyer/monster_tiny_flyer.glb`

- [x] Step 2.1: Add deterministic low-poly GLB generators for a quadruped predator and tiny flyer.
- [x] Step 2.2: Register `monster_quadruped_predator_v0` and `monster_tiny_flyer_v0` manifest entries with provenance and hashes.
- [x] Step 2.3: Run the asset generator and update committed runtime GLBs.
- [x] Step 2.4: Validate asset manifest, GLB node presence, and provenance hashes.

```bash
make gen-assets
make validate-assets
```

## Task 3 — Data-Driven Monster and Boss Rules

Files:
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/rules/dungeon_generation.v0.json`
- Modify: `shared/rules/boss_templates.v0.schema.json`
- Modify: `shared/rules/boss_templates.v0.json`
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/dungeon_gen.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/game_test.go`
- Modify: `shared/golden/boss_floor_-5.json` and related goldens if required

- [x] Step 3.1: Add `dungeon_wolf` and `dungeon_bat` as chase monsters mechanically equivalent to the current melee dungeon mob.
- [x] Step 3.2: Add both to dungeon monster pool with conservative weights and minimum deterministic coverage where existing tests need stable presence.
- [x] Step 3.3: Replace boss visual validation that pins `current_humanoid_player` with data validation for a small boss visual option/pool.
- [x] Step 3.4: Make boss visual option selection deterministic in the Go generation path and emitted to entity snapshots/deltas.
- [x] Step 3.5: Update Go tests/goldens for dungeon generation and boss visual metadata.

```bash
make validate-shared
cd server && go test ./internal/game/...
```

## Task 4 — Godot Monster Visual Resolver and Scenes

Files:
- Create: `client/scripts/monster_visuals_loader.gd`
- Create: `client/scenes/monster_quadruped.tscn`
- Create: `client/scenes/monster_tiny_flyer.tscn`
- Modify: `client/scripts/main.gd`
- Modify: `client/tests/test_animation.gd`
- Create/Modify: `client/tests/test_monster_visuals.gd`

- [x] Step 4.1: Add a static loader/resolver for `shared/assets/monster_visuals.v0.json`.
- [x] Step 4.2: Add quadruped and tiny flyer scenes with required clips: `idle`, `walk`, `hit`, `death`.
- [x] Step 4.3: Route `main.gd` monster instancing through the resolver using `monster_def_id` and any server-provided boss visual key.
- [x] Step 4.4: Apply catalog scale/height offset plus server rarity/boss visual scale without moving authoritative server coordinates.
- [x] Step 4.5: Keep archer bow marker as a narrow compatibility overlay if needed; do not expand ranged overlays in this slice.
- [x] Step 4.6: Add client tests for mapping validity and scene clip coverage.

```bash
make client-unit
```

## Task 5 — Showme Monster Approval Loop

Files:
- Modify: `skills/showme/scripts/render_focus.py`
- Modify: `skills/showme/scripts/visual_capture.gd`

- [x] Step 5.1: Add `--focus monsters` to render a lineup of dummy, quadruped, tiny flyer, and boss-scale variants.
- [x] Step 5.2: Run the focused capture and inspect the screenshot for nonblank render, readable silhouettes, relative scale, and flyer height/wing pose.
- [x] Step 5.3: If the lineup is questionable, stop and ask the user for visual feedback before continuing broad integration. If acceptable, record the screenshot path in the execution notes. Approved screenshot: `.artifacts/showme/20260611-123159-monsters.png`.

```bash
python3 skills/showme/scripts/render_focus.py --focus monsters
```

## Task 6 — Bot and Client Scenario Proof

Files:
- Modify: `shared/rules/worlds.v0.json`
- Create: `tools/bot/scenarios/NN_monster_visual_catalog.json`
- Optionally create: `tools/bot/scenarios/client/NN_monster_visual_catalog.json`
- Modify: `tools/bot/run.py` or `client/scripts/bot_scenario_runner.gd` only if existing assertions cannot cover visual metadata.

- [x] Step 6.1: Add a compact `monster_visual_catalog_lab` with dummy/wolf/bat monsters and clear positions.
- [x] Step 6.2: Add a protocol bot scenario asserting the new monster definitions exist and can be killed or reached under existing mechanics.
- [x] Step 6.3: Add a client scenario assertion for distinct visual model/debug names if current client bot supports it; otherwise cover visual resolver through Godot unit tests.

```bash
make bot
make client-smoke
```

## Task 7 — Lifecycle Docs and CI

Files:
- Modify: `docs/specs/v72_spec-monster-visual-catalog.md`
- Modify: `docs/plans/v72_2026-06-11-monster-visual-catalog.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v72_monster-visual-catalog.md`

- [x] Step 7.1: Mark the spec status complete once implementation is verified.
- [x] Step 7.2: Update plan checkboxes as tasks complete.
- [x] Step 7.3: Add v72 lifecycle/as-built notes and defer true flying behavior / richer monster attacks.
- [x] Step 7.4: Run the final CI gate.

```bash
make ci
```

## Final Verification

- [x] `make validate-shared`
- [x] `make validate-assets`
- [x] `.venv/bin/pytest tools/assets/test_validate_assets.py -q`
- [x] `cd server && go test ./internal/game/...`
- [x] `make client-unit`
- [x] `python3 skills/showme/scripts/render_focus.py --focus monsters`
- [x] `make bot`
- [x] `make client-smoke`
- [x] `make ci`

## Deferred Scope

- True flying gameplay and vertical/pathing differences.
- Quadruped pounce, bat swarm/dive, or other new AI behavior.
- General ranged-weapon overlays for all monster visuals beyond keeping the existing archer marker working.
- Production-grade external art pass.
